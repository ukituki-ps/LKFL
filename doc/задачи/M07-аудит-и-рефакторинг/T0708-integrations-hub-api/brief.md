# T0708 — Детализация Integrations Hub API spec

## Веха

M07-аудит-и-рефакторинг

## Контекст

Integrations Hub документирован концептуально:
- «CRUD всех интеграций: name, protocol, endpoints, auth_method, status»
- «Health dashboard: статус каждой интеграции»
- «Sync control: manual trigger, schedule, error log»

Однако:
- Нет полного списка endpoints
- Нет description response schemas
- Нет pagination format'а
- Нет authentication requirements (RBAC guard для `integration_admin` role)
- Нет version'ирования API

**Проблема:** без этого executor не может написать working integration.

**Решение:**
Написать детальную API spec в стиле `спецификация/api.md`:
- Таблица endpoints: Method, Path, Description, Auth, Request, Response
- Response schemas для CRUD операций
- Pagination cursor-based (как в Platform API)
- Error format (единый с другими сервисами)
- RBAC: только `integration_admin` role

### Файлы-мишени

| Действие | Файл |
|---|-|-|
| API spec | `архитектура/модули.md` — full endpoint table для Integrations Hub |
| API spec | `спецификация/api.md` — добавить раздел "Integrations Hub API" |
| Интеграции | `архитектура/интеграции.md` — заменить conceptual API table → full spec |
| Акторы | `контекст/акторы.md` — Администратор интеграций: endpoints |
| Обновить README вехи | `задачи/README.md` — статус M07 |

### Критерии приёмки

- [x] Полная таблица endpoints Integrations Hub: 12 endpoints (см. расчёт ниже)

**Расчёт 12 endpoints:**
1. GET `/admin/providers` — list
2. GET `/admin/providers/:id` — get
3. POST `/admin/providers` — create
4. PUT `/admin/providers/:id` — update
5. DELETE `/admin/providers/:id` — delete
6. POST `/admin/providers/:id/health` — health check
7. GET `/admin/health/dashboard` — all health
8. POST `/admin/sync/trigger` — manual sync
9. GET `/admin/sync/schedule` — get schedule
10. PUT `/admin/sync/schedule` — set schedule
11. GET `/admin/sync/errors` — error log
12. GET `/admin/sync/sla` — SLA dashboard
- [x] Response schemas описаны для каждого endpoint
- [x] RBAC guard описан: `integration_admin` только
- [x] Добавлено в `спецификация/api.md` как самостоятельный раздел
- [x] Файлы-мишени все перечислены выше
