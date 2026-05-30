# ADR-003: Keycloak (OIDC) как центральный Identity Provider

**Статус:** Accepted
**Дата:** 2026-05-22
**Контекст:** М01-создание-описания

## Контекст

Платформа требует:
- Сквозную авторизацию на 5+ внешних сервисах (ready4.ru, PrimeZone, Подарок в квадрате)
- 100 000+ пользователей
- ФСТЭК-сертификация (on-premise, данные не покидают РФ)
- Мульти-тенантность (разные компании-заказчики)

Изначально планировалась голая JWT-аутентификация. Это не покрывает SSO и комплаенс.

## Решение

**Keycloak 26.x** как центральный IdP:
- OIDC для платформы (backend — OIDC client через `go-oidc`)
- SAML / OIDC Identity Broker для внешних провайдеров (сквозная авторизация)
- On-premise deploy (ФСТЭК, 152-ФЗ)
- Realm per tenant или group-based multi-tenancy
- Password policy, MFA (опционально), session management

## Альтернативы

| Вариант | Плюсы | Минусы |
|---------|-------|--------|
| Голые JWT | Простой | Нет SSO, нет broker, нет password management |
| Auth0 | Готовый SSO | SaaS — нарушает ФСТЭК, данные за пределами РФ |
| ZITADEL | Event sourcing, open source | Меньше mature, нет SAML broker из коробки |

## Следствия

- Backend валидирует JWT от Keycloak, не эмитит свои токены
- СSO на внешние сервисы — Keycloak Identity Broker
- Keycloak деплоится в docker-compose (dev) / кластер (prod)
- Требуется отдельная БД для Keycloak (PostgreSQL)
- `KEYCLOAK_JWKS_URL`, `KEYCLOAK_ISSUER`, `KEYCLOAK_AUDIENCE` в env каждого сервиса
