# M04 — API под новую архитектуру льгот

## Цель

Пересобрать API-контракты и артефакты под 5 сущностей льгот (T0301):
Benefit, BenefitPlan, ActivationFlow, BenefitCollection, UserBenefit.

## Описание

M03 пересобрал архитектуру льгот. API остался на старой модели — endpoints
активировали benefit, а не plan; DMS/matkapital захардкожены; не было CRUD для plans/flows.

Решение: переписать api.md + обновить артефакты.md.

## Задачи

| Задача | Файл-мишень | Что делаем | Статус |
|--------|-------------|------------|--------|
| T0401-пересборка-api | спецификация/api.md, спецификация/артефакты.md | Переписать endpoints под 5 сущностей. Убрать DMS/matkapital. Generic wizard. | ✅ выполнено |

## Exit criteria

- [x] T0401 report.md — «выполнено»
- [x] api.md — все endpoints обновлены
- [x] артефакты.md — S10a generic wizard, S10b обновлён
