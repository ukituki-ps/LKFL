# ADR-017 — Generic REST adapter для `providers/` вместо 11 hard-coded модулей

## Контекст

Сейчас `internal/providers/` содержит 11 hard-coded Go-модулей по одному на каждый провайдера:

```
providers/
├── alpha/
├── worldclass/
├── yandex/
├── skillbox/
└── ... (ещё 7)
```

Каждый адаптер реализует `ProviderAdapter` interface с 7 методами. Новый провайдер = новый Go-модуль + новый deploy.

При анализе 11 адаптеров выясняется:
- 9 из 11 используют только 2 метода: `Call(endpoint, data)` (для Activate/Deactivate) и `Transform(request, response)` (для нормализации JSON к `EngagementOffer`)
- Сложные протоколы (SAML, SOAP) — это 1-2 исключения
- Конфигурация провайдера (URL, headers, auth) дублируется в каждом Go-файле

## Решение

Заменить hard-coded Go-модули на YAML-config + generic adapter:

```yaml
# providers-config/alpha.yaml
name: "alpha"
protocol: "rest"
endpoints:
  activate:
    method: "POST"
    url: "https://alpha.api/activate"
    headers: {"Authorization": "${ALPHA_API_KEY}"}
  deactivate:
    method: "DELETE"
    url: "https://alpha.api/status/${userId}"
transform:
  request: "alpha/transform_request.tmpl"
  response: "alpha/transform_response.tmpl"
health:
  url: "https://alpha.api/health"
  interval: "300s"
```

Generic adapter:
- Reads YAML config at startup
- `Call(endpoint, data)` — executes configured HTTP method + URL
- `Transform(request, response)` — applies Go template for normalization
- `Health()` — periodic health check + error rate monitoring

### Исключения (когда hard-coded обязателен):

| Протокол | Почему |
|-----|-----|
| SAML (Ready4, PrimeZone) | XML + crypto + assertion validation — не выражается через REST+template |
| SOAP (если появится) | WSDL + WS-Security — требует client generation |
| Проприетарные бинарные протоколы | специфичные библиотеки |

## Последствия

- ✅ 9 из 11 адаптеров → YAML config + 1 generic adapter
- ✅ New provider = config file, not Go code
- ✅ Config versioned in Git, hot-reloadable via admin UI
- ⚠️ 2 сложных адаптера (SAML) остаются hard-coded — это корректно
- ⚠️ Потеря compile-time type safety — компенсируется runtime validation + health checks

## Статус

✅ Accepted (M07, T0704)
