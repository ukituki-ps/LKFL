# T1805 — Отчёт о выполнении

## Статус

✅ выполнено

## Что сделано

### `backend/cmd/seed/main.go`

- Подключение к БД через `DB_DSN` из env
- **upsertTenant**: idempotent создание tenant sdek:
  - Сначала пытается найти существующего по slug
  - Если не найден — создаёт с новым UUID
  - slug='sdek', name='СДЭК', status='active'

- **upsertBrandConfig**: idempotent создание/обновление brand config:
  - INSERT ... ON CONFLICT (tenant_id) DO UPDATE SET
  - primary_color='#E30613' (СДЭК красный)
  - secondary_color='#FFFFFF'
  - brand_name='СДЭК Льготы'
  - css_variables для April tokens (--april-color-primary, --april-color-primary-hover)

- Запуск: `make seed` (target уже есть в Makefile)

## Критерии приёмки

- [x] `cmd/seed/main.go` создан
- [x] `make seed` загружает tenant sdek + brand config
- [x] Slug: `sdek`, name: `СДЭК`
- [x] Brand config: primary_color `#E30613`, css_variables для April tokens
- [x] Idempotent (повторный запуск не ломает)
- [ ] Tenant доступен через API (требует запущенного сервера)

## Время

~15 мин
