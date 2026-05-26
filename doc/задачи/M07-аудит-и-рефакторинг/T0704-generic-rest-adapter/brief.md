# T0704 — Generic REST adapter для providers/

## Веха

M07-аудит-и-рефакторинг

## Контекст

Документация описывает 11 адаптеров провайдеров в `integrations/providers/`:
alpha, worldclass, yandex, yandex-psych, skillbox, sber, sdek-store, motherchild, mts, skyeng, giftcard.

Все реализуют один и тот же `ProviderAdapter interface` с одинаковыми методами:
Activate, Deactivate, Status, SyncCatalog, Health, Configure.

**Противоречие с философией:**
- `контекст/философия.md`: «Новый провайдер — конфигурация, не код»
- `контекст/настраиваемость.md`: «Подключение провайдера (существующий адаптер) — ❌ requires code»
- Реальность: каждый провайдер — новый Go-модуль в `providers/`

**Решение:**
Заменить hard-coded adapters на generic REST adapter с конфигурируемой схемой трансформаторов:

```yaml
# providers/sdek-store/config.yaml
name: sdek-store
protocol: rest
endpoints:
  activate: "https://api.sdek-store.ru/v1/order"
  deactivate: "https://api.sdek-store.ru/v1/order/:id/cancel"
  status: "https://api.sdek-store.ru/v1/order/:id"
  catalog: "https://api.sdek-store.ru/v1/products"
auth:
  type: bearer
  token_url: "https://auth.sdek-store.ru/token"
transform:
  activate_request: json_path("./transforms/sdek-store/activate.jsonnet")
  activate_response: json_path("./transforms/sdek-store/activate-response.jsonnet")
timeout: 30s
retry: 3
```

Тогда новый провайдер = config file + transform templates, не Go-код.
Hard-coded adapters остаются только для провайдеров со специфичными протоколами (SAML, SOAP, custom binary).

### Файлы-мишени

| Действие | Файл |
|---|---|
| Новая структура providers/ | `архитектура/модули.md` — generic adapter вместо 11 модулей |
| ProviderAdapter interface | `архитектура/модули.md` — упростить до 2 методов: Call(endpoint, data), Transform |
| Интеграции документация | `архитектура/интеграции.md` — новый формат config-файлов |
| Философия | `контекст/философия.md` — проверить, что "нулевая привязка" теперь реально работает |
| Настраиваемость | `контекст/настраиваемость.md` — "Новый провайдер REST API" → ❌ requires code → ✅ |
| Акторы | `контекст/акторы.md` — Администратор интеграций: действия по подключению нового провайдера |
| Создать ADR | `архитектура/adr/ADR-017-generic-rest-adapter.md` |
| Обновить README архитектуры | `архитектура/README.md` — ADR-017 |
| Обновить README вехи | `задачи/README.md` — статус M07 |

### Критерии приёмки

- [x] `архитектура/модули.md` — providers/ теперь generic REST adapter + 11 config-файлов как примеры
- [x] ProviderAdapter interface упрощён: 2 метода вместо 7
- [x] `архитектура/интеграции.md` — формат config файла с примерами для 3 провайдеров
- [x] `контекст/настраиваемость.md` — "Новый провайдер REST API" более не требует Go-кода
- [x] Создан ADR-017 с обоснованием generic adapter
- [x] Перечислены исключения: когда hard-coded adapter НУЖЕН (например, SAML, SOAP)
- [x] Файлы-мишени все перечислены выше
