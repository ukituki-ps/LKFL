# T0401 — Пересборка API-контрактов под новую архитектуру льгот

## Веха

M04-api-под-льготы

## Контекст

M03 (T0301) пересобрал архитектуру льгот: одна монолитная benefit-сущность заменена на 5 сущностей (Benefit, BenefitPlan, ActivationFlow, BenefitCollection, UserBenefit).

API-контракты (`спецификация/api.md`) и артефакты (`спецификация/артефакты.md`) остались на старой модели.

### Выявленные проблемы

| # | Проблема | Где | Влияние |
|---|----------|-----|---------|
| 1 | `/benefits/:id/activate` активирует benefit, а нужно plan | Catalog/Benefits | ~30 endpoints устарели |
| 2 | `/admin/benefits` POST/PUT — cost, conditions в plan, не в benefit | Admin Catalog | Менеджер не может создать plan отдельно |
| 3 | `/dms/*` (11 endpoints) — захардкожены под ДМС | DMS | Теперь ДМС = benefit + plan + activation_flow |
| 4 | `/matkapital/*` (3 endpoints) — захардкожены | Matkapital | Теперь = generic benefit с activation_flow |
| 5 | S10a (ДМС Wizard) — захардкожен | Артефакты | Нужен generic Activation Wizard |
| 6 | S10c (Маткапитал Wizard) — захардкожен | Артефакты | Нужен тот же generic wizard |

## Что сделать

1. Переписать `спецификация/api.md`:
   - Catalog/Benefits → `/benefits` (продукт) + `/benefits/:id/plans` (тарифы)
   - Activate/deactivate → `/user-benefits` (создание экземпляра сотрудника)
   - DMS/matkapital → убрать как отдельные модули (теперь generic через activation flow)
   - Admin Catalog → CRUD benefit + plan + activation_flow + collection
2. Обновить `спецификация/артефакты.md`:
   - S10a (ДМС Wizard) + S10c (Маткапитал Wizard) → S10a (Activation Wizard — generic)
   - S10b (Детали льготы) — обновить описание (табы из plan metadata)

### Критерии приёмки

- [ ] `api.md` — все endpoints обновлены под 5 сущностей льгот
- [ ] `api.md` — DMS/matkapital убраны как отдельные модули (generic через activation flow)
- [ ] `api.md` — добавлены endpoints для benefit-plans, activation-flows, user-benefits
- [ ] `артефакты.md` — S10a generic wizard, S10b обновлён
- [ ] Сводная таблица API обновлена
- [ ] Endpoints согласованы с journeys (J02–J08, J13a)
