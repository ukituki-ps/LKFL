package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

type deployer struct {
	cfg Config
	sm  *stateManager
}

func newDeployer(cfg Config, sm *stateManager) *deployer {
	return &deployer{cfg: cfg, sm: sm}
}

// Deploy выполняет полный цикл деплоя:
// 1. GHCR login
// 2. Pull образов
// 3. Миграции (идемпотентные, warning при ошибке)
// 4. Up сервисов
// 5. Seed (идемпотентный, warning при ошибке)
// 6. Healthcheck
func (d *deployer) Deploy(req DeployRequest) {
	d.sm.setStatus("deploying")
	d.sm.setBranch(req.Branch)
	d.sm.setImageTag(req.ImageTag)
	d.sm.setStartedAt()
	d.sm.logf("starting deploy: branch=%s tag=%s", req.Branch, req.ImageTag)

	// 1. GHCR login
	if err := d.ghcrLogin(); err != nil {
		d.sm.fail(err.Error())
		return
	}

	// 2. Pull образов
	if err := d.composePull(req.ImageTag); err != nil {
		d.sm.fail(err.Error())
		return
	}

	// 3. Миграции (идемпотентные — не прерываем при ошибке)
	if err := d.composeRun("lkfl-migrate", req.ImageTag); err != nil {
		d.sm.logf("warning: migration failed (may already be applied): %v", err)
	}

	// 4. Up сервисов
	if err := d.composeUp(req.ImageTag); err != nil {
		d.sm.fail(err.Error())
		return
	}

	// 5. Seed (идемпотентный — не прерываем при ошибке)
	if err := d.composeRun("lkfl-seed", req.ImageTag); err != nil {
		d.sm.logf("warning: seed failed (data may already exist): %v", err)
	}

	// 6. Healthcheck
	if err := d.healthcheck(); err != nil {
		d.sm.fail("healthcheck failed: " + err.Error())
		return
	}

	d.sm.setStatus("success")
	d.sm.setFinishedAt()
	d.sm.logf("deploy completed successfully")
}

// Rollback откатывает к предыдущему тегу, запуская полный цикл deploy.
func (d *deployer) Rollback() {
	previousTag := d.sm.getPreviousTag()
	if previousTag == "" {
		d.sm.fail("no previous tag for rollback")
		return
	}

	d.sm.setStatus("rolling back")
	d.sm.setStartedAt()
	d.sm.logf("rolling back to %s", previousTag)

	// Rollback — тот же процесс deploy с предыдущим тегом
	req := DeployRequest{
		Branch:   d.sm.getBranch(),
		ImageTag: previousTag,
	}
	d.Deploy(req)
}

// ─── Docker Compose helpers ───

func (d *deployer) composeCmd() string {
	// Compose file references ./infra/ paths — нужно указать --project-directory
	return fmt.Sprintf("docker compose -p lkfl-staging -f %s --env-file .env.staging --project-directory %s", d.cfg.ComposeFile, d.cfg.ComposeDir)
}

func (d *deployer) ghcrLogin() error {
	if d.cfg.GHCRToken == "" {
		log.Println("warning: GHCR_TOKEN not set, skipping login")
		return nil
	}

	cmd := exec.Command("docker", "login", "ghcr.io", "-u", d.cfg.GHCRUsername, "--password-stdin")
	cmd.Stdin = strings.NewReader(d.cfg.GHCRToken)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ghcr login: %s", output)
	}
	d.sm.logf("GHCR login OK")
	return nil
}

func (d *deployer) composePull(imageTag string) error {
	d.sm.logf("pulling images with tag %s...", imageTag)

	env := os.Environ()
	env = append(env, fmt.Sprintf("IMAGE_TAG=%s", imageTag))

	cmd := exec.Command("sh", "-c", fmt.Sprintf("%s pull --ignore-buildable", d.composeCmd()))
	cmd.Env = env
	cmd.Dir = d.cfg.ComposeDir
	output, err := cmd.CombinedOutput()
	d.sm.logf("pull output: %s", output)
	if err != nil {
		return fmt.Errorf("compose pull: %s", output)
	}
	d.sm.logf("images pulled")
	return nil
}

func (d *deployer) composeRun(service string, imageTag string) error {
	d.sm.logf("running %s...", service)

	env := os.Environ()
	env = append(env, fmt.Sprintf("IMAGE_TAG=%s", imageTag))

	cmd := exec.Command("sh", "-c", fmt.Sprintf("%s run --rm %s", d.composeCmd(), service))
	cmd.Env = env
	cmd.Dir = d.cfg.ComposeDir
	output, err := cmd.CombinedOutput()
	d.sm.logf("%s output: %s", service, output)
	if err != nil {
		return fmt.Errorf("compose run %s: %s", service, output)
	}
	d.sm.logf("%s completed", service)
	return nil
}

func (d *deployer) composeUp(imageTag string) error {
	d.sm.logf("starting services...")

	env := os.Environ()
	env = append(env, fmt.Sprintf("IMAGE_TAG=%s", imageTag))

	// Поднимаем только основные сервисы (не deploy-worker — он сам себя пересоздаст)
	cmd := exec.Command("sh", "-c",
		fmt.Sprintf("%s up -d postgres redis keycloak lkfl-server lkfl-integration-proxy lkfl-frontend nginx", d.composeCmd()))
	cmd.Env = env
	cmd.Dir = d.cfg.ComposeDir
	output, err := cmd.CombinedOutput()
	d.sm.logf("up output: %s", output)
	if err != nil {
		return fmt.Errorf("compose up: %s", output)
	}
	d.sm.logf("services started")
	return nil
}

func (d *deployer) healthcheck() error {
	d.sm.logf("running healthcheck...")

	// Ожидание до 60 секунд (30 попыток по 2 секунды)
	for i := 0; i < 30; i++ {
		cmd := exec.Command("sh", "-c", "curl -sf http://lkfl-server:8080/healthz || curl -sf http://localhost:8080/healthz")
		if err := cmd.Run(); err == nil {
			d.sm.logf("healthcheck OK (attempt %d)", i+1)
			return nil
		}
		time.Sleep(2 * time.Second)
	}

	return fmt.Errorf("healthcheck timed out after 60s")
}
