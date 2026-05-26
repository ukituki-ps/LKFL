# T0704 — Generic REST adapter для providers/ — отчёт

## Статус

✅ выполнено

## Что сделано

- `архитектура/модули.md` — providers/: 11 hard-coded модулей → Generic REST adapter + YAML config
  - 9 адаптеров → YAML config (alpha, worldclass, yandex, yandex-psych, skillbox, sber, sdek-store, motherchild, mts, skyeng)
  - 2 адаптера остаются hard-coded: SAML (Ready4, PrimeZone) + проприетарный (Подарок в квадрате)
- `архитектура/интеграции.md` — новый раздел «Generic REST adapter» (стр.286):
  - YAML-config format с полным примером (alpha.yaml)
  - Как работает: Call(endpoint, data) + Transform(request, response) + Health()
  - Исключения: SAML, SOAP, проприетарные протоколы
- `контекст/настраиваемость.md` — «Новый адаптер провайдера (REST)» → ❌ requires code
- `контекст/философия.md` — «M07 T0704 (ADR-017): REST-провайдер — без кода!»
- Создан ADR-017: обоснование generic adapter (9/11 адаптеров используют 2 метода из 7)
- `архитектура/README.md` — ADR-017 добавлен в таблицу
- `задачи/README.md` — статус M07 обновлён

## Проблемы

- Потеря compile-time type safety → компенсируется runtime validation + health checks (задокументировано в ADR-017)
