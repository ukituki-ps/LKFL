package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

const previousTagFile = ".deploy-previous-tag"

// Статусы деплоя.
const (
	statusIdle      = "idle"
	statusDeploying = "deploying"
	statusSuccess   = "success"
	statusFailed    = "failed"
	statusPending   = "pending"
)

// DeployState — сериализуемое состояние деплоя для JSON-ответов.
type DeployState struct {
	Status      string `json:"status"`
	Branch      string `json:"branch,omitempty"`
	ImageTag    string `json:"imageTag,omitempty"`
	PreviousTag string `json:"previousTag,omitempty"`
	StartedAt   string `json:"startedAt,omitempty"`
	FinishedAt  string `json:"finishedAt,omitempty"`
	Error       string `json:"error,omitempty"`
}

// DeployHistoryEntry — запись в истории деплоев.
type DeployHistoryEntry struct {
	ID          int    `json:"id"`
	Status      string `json:"state"`
	Branch      string `json:"branch,omitempty"`
	ImageTag    string `json:"imageTag,omitempty"`
	PreviousTag string `json:"previousTag,omitempty"`
	StartedAt   string `json:"startedAt,omitempty"`
	FinishedAt  string `json:"finishedAt,omitempty"`
	Error       string `json:"error,omitempty"`
	Duration    string `json:"duration,omitempty"`
}

// DeployHistory — история деплоев. Хранит до 10 записей в памяти.
type DeployHistory []DeployHistoryEntry

// stateManager обеспечивает потокобезопасный доступ к состоянию деплоя.
// Использует mutex для защиты от конкурентного доступа при асинхронном деплое.
type stateManager struct {
	mu        sync.Mutex
	state     DeployState
	logs      strings.Builder
	history   DeployHistory
	historyID int
	historyMu sync.Mutex
}

func newStateManager() *stateManager {
	sm := &stateManager{
		state:   DeployState{Status: statusIdle},
		history: make(DeployHistory, 0),
	}

	// Восстановить PreviousTag из файла (переживает перезапуск)
	if tag, err := os.ReadFile(previousTagFile); err == nil && len(tag) > 0 {
		sm.state.PreviousTag = strings.TrimSpace(string(tag))
	}

	return sm
}

// canDeploy возвращает true, если можно запустить новый деплой.
func (sm *stateManager) canDeploy() bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	return sm.state.Status == statusIdle || sm.state.Status == statusFailed || sm.state.Status == statusSuccess
}

// tryAcquire atomically проверяет возможность деплоя и захватывает слот.
// Возвращает true, если деплой может быть запущен (статус установлен в "pending").
// Защищает от race condition между проверкой и запуском goroutine.
func (sm *stateManager) tryAcquire() bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if sm.state.Status == statusIdle || sm.state.Status == statusFailed || sm.state.Status == statusSuccess {
		sm.state.Status = statusPending
		return true
	}
	return false
}

// getState возвращает текущее состояние (копия, safe для JSON).
func (sm *stateManager) getState() DeployState {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	return sm.state
}

// getLogs возвращает накопленные логи последнего деплоя.
func (sm *stateManager) getLogs() string {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	return sm.logs.String()
}

// setStatus обновляет статус деплоя.
func (sm *stateManager) setStatus(status string) {
	sm.mu.Lock()

	// При старте нового деплоя сохраняем предыдущий тег и очищаем логи
	if status == statusDeploying && (sm.state.Status == statusIdle || sm.state.Status == statusSuccess || sm.state.Status == statusFailed || sm.state.Status == statusPending) {
		if sm.state.ImageTag != "" {
			sm.state.PreviousTag = sm.state.ImageTag
		}
		// Очистить логи предыдущего деплоя (ограничение роста памяти)
		sm.logs.Reset()
	}

	// Сохранить в файл для персистентности (переживает перезапуск)
	if status == statusSuccess && sm.state.ImageTag != "" {
		// Текущий ImageTag станет PreviousTag для следующего деплоя
		if err := os.WriteFile(previousTagFile, []byte(sm.state.ImageTag+"\n"), 0644); err != nil {
			log.Printf("warn: failed to save current tag: %v", err)
		}
	}

	sm.state.Status = status

	// При завершении деплоя добавляем запись в историю
	if status == statusSuccess || status == statusFailed {
		st := sm.state
		sm.mu.Unlock()
		sm.addHistoryEntry(status, st)
		sm.mu.Lock()
		return
	}
	sm.mu.Unlock()
}

// setBranch устанавливает ветку деплоя.
func (sm *stateManager) setBranch(branch string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.state.Branch = branch
}

// setImageTag устанавливает тег образа.
func (sm *stateManager) setImageTag(tag string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.state.ImageTag = tag
}

// getBranch возвращает текущую ветку.
func (sm *stateManager) getBranch() string {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	return sm.state.Branch
}

// getPreviousTag возвращает предыдущий тег для rollback.
func (sm *stateManager) getPreviousTag() string {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	return sm.state.PreviousTag
}

// setStartedAt устанавливает время начала.
func (sm *stateManager) setStartedAt() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.state.StartedAt = time.Now().Format(time.RFC3339)
}

// setFinishedAt устанавливает время завершения.
func (sm *stateManager) setFinishedAt() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.state.FinishedAt = time.Now().Format(time.RFC3339)
}

// fail устанавливает статус ошибки и время завершения.
func (sm *stateManager) fail(msg string) {
	sm.mu.Lock()
	sm.state.Status = statusFailed
	sm.state.Error = msg
	sm.state.FinishedAt = time.Now().Format(time.RFC3339)
	st := sm.state
	sm.mu.Unlock()

	sm.addHistoryEntry(statusFailed, st)
}

// addHistoryEntry добавляет запись в историю деплоев.
// Вызывается после завершения деплоя (success/failed).
// state передаётся как копия для избежания вложенной блокировки sm.mu.
func (sm *stateManager) addHistoryEntry(status string, state DeployState) {
	sm.historyMu.Lock()
	defer sm.historyMu.Unlock()

	sm.historyID++

	// Вычислить длительность
	var duration string
	if state.StartedAt != "" && state.FinishedAt != "" {
		start, err1 := time.Parse(time.RFC3339, state.StartedAt)
		finish, err2 := time.Parse(time.RFC3339, state.FinishedAt)
		if err1 == nil && err2 == nil && finish.After(start) {
			duration = finish.Sub(start).String()
		}
	}

	entry := DeployHistoryEntry{
		ID:          sm.historyID,
		Status:      status,
		Branch:      state.Branch,
		ImageTag:    state.ImageTag,
		PreviousTag: state.PreviousTag,
		StartedAt:   state.StartedAt,
		FinishedAt:  state.FinishedAt,
		Error:       state.Error,
		Duration:    duration,
	}

	sm.history = append(sm.history, entry)

	// Оставить последние 10 записей
	if len(sm.history) > 10 {
		sm.history = sm.history[len(sm.history)-10:]
	}
}

// getHistory возвращает копию истории деплоев.
func (sm *stateManager) getHistory() DeployHistory {
	sm.historyMu.Lock()
	defer sm.historyMu.Unlock()
	result := make(DeployHistory, len(sm.history))
	copy(result, sm.history)
	return result
}

// logf записывает сообщение в логи и stdout.
func (sm *stateManager) logf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	log.Println(msg)

	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.logs.WriteString(fmt.Sprintf("[%s] %s\n", time.Now().Format(time.RFC3339), msg))
}
