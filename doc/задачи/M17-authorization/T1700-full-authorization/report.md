# T1700 — Отчёт

## Статус

⏳ Не начато

## Что сделано

_(пусто)_

## Проблемы

_(пусто)_

## Следующие шаги

1. Инициализация go.mod (Go 1.24+) и инфраструктуры
2. Seed Keycloak (demo realm + admin/employee users + clients)
3. Реализация shared/pkg/auth
4. Реализация internal пакетов (auth, tenant, api)
5. Фронтенд: Vite bootstrap + tenant resolution + keycloak-js + auth state machine
6. CI/CD: Dockerfiles (multi-stage) + GitHub Actions + testcontainers
7. Интеграционные тесты
8. Финальная проверка: login → /api/v1/me → 200
