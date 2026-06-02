# Отчёт T2311 — E2E Integration Tests: Login → Dashboard → Logout

## Веха

M23 — E2E Testing

## Статус

✅ Завершена

## Изменения

### Созданные файлы

| Файл | Описание |
|------|----------|
| `frontend/e2e/integration/login-logout.spec.ts` | 4 теста реального OIDC flow |

### Изменённые файлы

| Файл | Изменение |
|------|-----------|
| `infra/keycloak/realm-lkfl-sdek.json` | Добавлен пользователь `test.employee` + `http://localhost:80/*` в redirectUris/webOrigins |
| `frontend/playwright.config.ts` | Добавлен проект `integration`, глобальный testIgnore исключает `integration/` |
| `frontend/e2e/local-login-test.spec.ts` | Заменён на placeholder с `test.skip` |
| `frontend/src/components/layout/UserMenu.tsx` | Добавлен `data-testid="user-avatar"` |
| `.env.staging` | Добавлена `E2E_BASE_URL=https://dev.april.ukituki.tech` |
| `.github/workflows/build.yml` | Добавлен job `e2e-integration` после `deploy-staging` |

## Критерии приёмки

- [x] Пользователь `test.employee` в `realm-lkfl-sdek.json`
- [x] Проект `integration` в `playwright.config.ts`
- [x] `e2e/integration/login-logout.spec.ts` — 4 теста
- [x] `E2E_BASE_URL` в `.env.staging`
- [x] `e2e/local-login-test.spec.ts` — placeholder
- [x] CI job `e2e-integration` в `build.yml`
- [x] `npx playwright test --project=integration --list` — 4 теста
- [x] `tsc --noEmit` — чистая компиляция

## Запуск

```bash
# Dev (через nginx docker compose)
cd frontend && E2E_BASE_URL=http://localhost:80 npx playwright test --project=integration

# Staging
cd frontend && E2E_BASE_URL=https://dev.april.ukituki.tech npx playwright test --project=integration
```
