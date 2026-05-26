# Отчёт T1509 — Mobile + Forms архитектура

## Статус

✅ Завершена

## Что сделано

- Создан `архитектура/фронтенд-mobile-forms.md` — самостоятельный архитектурный документ
- **Mobile:** `AprilMobileShellBar`, модальности (AprilModal vs AprilVaulBottomSheet), breakpoints (≥1280 / 768-1279 / <768), touch-ориентиры (44×44, safe area, iOS zoom prevention), жесты (Android back, iOS swipe-back)
- **Forms:** бизнес-формы (Zod + react-hook-form), admin-формы (AprilJsonSchemaForm / Zod), wizard-формы (Zustand store + step validation), survey-формы (динамические вопросы + optimistic submit + branching)
