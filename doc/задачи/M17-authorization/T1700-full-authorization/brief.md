# T1700 — Полная система авторизации (ЭПИК)

> **Тип:** эпик. Реальная работа разбита на 5 подзадач: T1701–T1705.
> Этот файл оставлен для совместимости реестра задач.

## Контекст

Первая задача кода. Создаём **полностью рабочую систему авторизации** — от Keycloak до React SPA. Это фундамент: без этой задачи все остальные невозможны.

**ADR:** ADR-036 (Authorization System), ADR-003 (Keycloak), ADR-009 (Multi-tenancy)
**Референс:** April Ecosystem AUTHORIZATION_REFERENCE.md → адаптировано для realm per tenant

## Разбиение на подзадачи

| Код | Название | Зависит от |
|-----|----------|------------|
| **T1701** | Инфраструктура и bootstrap | — |
| **T1702** | Backend auth core (+ unit тесты) | T1701 |
| **T1703** | Frontend auth (+ Playwright E2E) | T1701 |
| **T1704** | CI/CD pipeline (Фаза A: Dockerfile+Actions, Фаза B: testcontainers+OpenAPI) | T1702, T1703 |
| **T1705** | Observability (можно отложить до M18) | T1701 |

```
T1701 (инфраструктура)
    ├── T1702 (backend auth) ───┐
    ├── T1703 (frontend auth) ──┼──→ T1704 (CI/CD: Фаза A+B)
    └── T1705 (Observability)   — можно отложить до M18
```

**Можно отложить без ущерба:** T1705 (Observability).

**Критический путь:** T1701(3d) → T1702(5d)/T1703(4d) → T1704(5d) = 13 дней.

## Результат (сборка из подзадач)

- `lkfl-server` компилируется и запускается
- Keycloak поднимается в docker-compose с демо realm (seed: admin + employee users)
- Frontend запускается, keycloak-js инициализируется (realm из subdomain)
- Login flow работает: guest → login → transition → authorized
- `/api/v1/me` возвращает профиль с ролями
- Tenant resolution работает: backend (subdomain→tenant_id) + frontend (hostname→realm)
- JWT валидация работает с realm-specific JWKS
- RBAC guard блокирует endpoints без нужных ролей
