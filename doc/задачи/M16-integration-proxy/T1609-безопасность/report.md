# T1609 — Отчёт

## Веха

M16-integration-proxy

## Что сделано

Добавлена секция **14. Integration Proxy — credential isolation** в `doc/архитектура/безопасность.md`.

### Содержание добавленной секции

1. **Credential isolation** — провайдерские API-креденшиалы хранятся только в proxy (таблица `providers` в schema `lkfl_integration`), монолит не знает о креденшиалах.

2. **Шифрование credentials** — AES-256-GCM для полей `credentials_encrypted`. Master key: HashiCorp Vault (production), `INTEGRATION_MASTER_KEY` env var (development).

3. **Webhook attack surface** — webhook endpoint'ы на proxy (:8091), не на монолите (:8080). Pipeline: verify signature → validate payload → store in DB → callback к монолиту. Rate limiting на уровне Nginx.

4. **gRPC security** — localhost без TLS (Docker network). Production с отдельными хостами — mTLS.

5. **STRIDE update** — добавлены 5 новых угроз: proxy compromise, mono compromise, webhook injection, gRPC MITM, credential exfiltration.

6. **Сценарии компрометации** — S4 (proxy compromise), S5 (mono compromise), S6 (webhook injection attack).

### Обновления

- TL;DR навигация: добавлена ссылка на строку 335
- plan.yaml: прогресс 0% → 100%, все чекбоксы отмечены

## Затраченное время

~15 мин

## Замечания

- Секция консистентна с ADR-035 (порты 8090 gRPC / 8091 HTTP, schema `lkfl_integration`, таблица `providers`)
- Существующие секции безопасности.md не изменены
- Сценарии S4–S6 продолжают нумерацию после S1–S3 из §13 Threat Model
