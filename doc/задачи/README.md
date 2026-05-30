# Задачи

> **🗺️ Навигация:** [`doc/NAVIGATION.md`](../NAVIGATION.md) — карта «вопрос → файл:строка»

Этот раздел содержит **рабочие задачи** проекта: `brief.md`, `plan.yaml`, `report.md`.

## Назначение

Задачи — нижний уровень документации. Каждая задача — это конкретная работа для Executor-а. Она отвечает на вопросы:
- **Что делать?** — brief.md с контекстом и планом
- **Как отслеживать прогресс?** — plan.yaml с checklist-ом
- **Что сделано?** — report.md для TeamLead

## Поток информации

```
Задачи
│  brief.md   → Executor       (что делать)
│  plan.yaml  → Executor       (прогресс)
│  report.md  → TeamLead       (отчёт)

## Связи с другими разделами

- Контекст → Архитектура → Спецификация → Задачи
- Каждый journey → тест-кейсы в критериях-приёмки
- Каждый артефакт → endpoints в api.md
```

## Структура задачи

Каждая задача находится в папке `T{MM}{NN}-{name}/`:

```
T{MM}{NN}-{name}/
  brief.md      — контекст, ссылка на раздел плана, что нужно сделать
  plan.yaml     — checklist задач, прогресс (%), зависимости
  report.md     — отчёт: что сделано, проблемы, следующие шаги
```

## Правила

1. **Задача не создаётся без brief.md** — без контекста нет задачи
2. **plan.yaml — источник прогресса** — каждая задача имеет checklist
3. **report.md — для TeamLead** — отчёт о выполнении, проблемы, риски
4. **Статус задачи** — определяется по plan.yaml (0%, 50%, 100%)
5. **Зависимости** — `depends_on` в plan.yaml ссылается на другие T{MM}{NN}

## Номменклатура

- `T{MM}{NN}` — номер задачи: MM — номер вехи, NN — порядковый номер в вехе
- `M{MM}-{slug}` — номер вехи и краткое название
- Пример: `T0101` = задача 01 из вехи M01

## Текущие вехи

| Веха | Описание | Статус |
|------|----------|--------|
| M01-создание-описания | Формирование документации по всем разделам | выполнено |
| M02-улучшение-документации | Видимость white-label архитектуры, data schema, wireframe | выполнено |
| M03-архитектура-льгот | Пересборка льгот: 5 сущностей вместо одной монолитной схемы | выполнено |
| M04-api-под-льготы | API-контракты под 5 сущностей льгот, generic wizard | выполнено |
| M05-унификация-энгейджмента | Единая абстракция Engagement: льготы + активности → одна модель (T0501+T0502) | выполнено |
| M06-разбиение-platform | Platform → 7 внутренних пакетов (engagement, notification, recommendations) | выполнено |
| M07-аудит-и-рефакторинг | Аудит всех сервисов: eligibility split (T0701), compliance (T0702), admin handlers (T0703), generic adapter (T0704), payment-gateway (T0705), wizard engine (T0706), NATS registry (T0707), Hub API (T0708), cross-check (T0709) | выполнено (документация) |
| M08-cel-and-llm-engine | CEL Engine (ADR-021) + LLM Proxy (ADR-022) — единый движок бизнес-логики | выполнено |
| M09-gamification-system | Геймификация: ачивки, уровни лояльности, бейджи, CEL-условия, XLSX import | ✅ выполнено (документация) |
| M10-рефакторинг-по-результатам-аудита | Рефакторинг по аудиту: stub recommendations (T1001), merge LLM Proxy (T1002), shared CELContext (T1003), split api/ (T1004), shared auth (T1005) | ✅ выполнено (документация) |
| M11-split-integrations | Split Integrations: provider-gateway concept (T1101), HR sync → Platform (T1102), 1C → Billing payroll (T1103), NATS consumer update (T1104) | ✅ выполнено (документация) |
| M12-переход-на-модульный-монолит | Документация → mono-архитектура: ADR-024 (T1201), модули.md (T1202), пакеты-platform.md (T1203), NATS doc removal (T1204), API spec unified (T1205), стек.md (T1206), final consistency pass (T1207) | ✅ выполнено |
| M13-survey-engine | Survey Engine: полноценный модуль опросов (T1301) — survey_schema spec, бранчинг, TagMapper API design, TagResolver extension spec, analytics API spec | ✅ выполнено (только документация) |
| M14-survey-implementation | Реализация Go-кода Survey Engine: engagement/survey/*.go, TagResolver.AggregateSurveyTags, FlowEngine.ExecuteStep integration, admin/ survey-analytics endpoint | ⛔ отменена |
| M15-архитектура-фронтенда | Архитектура React SPA: фронтенд.md (10 разделов) + фронтенд-mobile-forms.md + 4 ADR + обновление ADR-029 + навигация | ✅ выполнено |
| M16-integration-proxy | Integration Proxy: ADR-035 (T1601), gRPC proto (T1602), интеграции.md (T1603), пакеты-platform (T1604), schema (T1605), api (T1606), модули (T1607), стек (T1608), безопасность (T1609), навигация (T1610), план (T1611), контекст (T1612), final consistency (T1613) | ✅ выполнено |
| M17-authorization | Authorization + CI/CD: 5 code — T1701 (инфра), T1702 (backend auth), T1703 (frontend auth), T1704 (CI/CD), T1705 (Observability, отложить) | ⏳ в работе |
