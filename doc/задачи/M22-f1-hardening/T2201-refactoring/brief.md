# T2201 — Рефакторинг F1

## Веха

M22-f1-hardening

## Тип

code

## Контекст

Финальный рефакторинг после завершения всех feature задач F1.
Цель: устранить tech debt, привести код к production стандарту.

## Что сделать

### Code review checklist

- [ ] golangci-lint — все warnings исправлены (strict mode)
- [ ] ESLint — все warnings исправлены (strict mode)
- [ ] `go vet ./...` — 0 issues
- [ ] `go fmt ./...` — форматирование
- [ ] Unused imports removal
- [ ] Error wrapping (`fmt.Errorf("context: %w", err)`)
- [ ] Context propagation (все DB/Redis calls с ctx)
- [ ] Interface satisfaction check (`var _ Repository = &pgRepository{}`)
- [ ] Godoc для всех public функций
- [ ] Frontend: TypeScript strict mode, no `any`

### Refactoring tasks

1. **Error handling** — все errors wrapped с context
2. **Context propagation** — timeout на все external calls
3. **Logger injection** — structured logging во все пакеты
4. **Constants** — magic strings → const
5. **Response format** — единый формат JSON responses
6. **Validator** — request validation через go-playground/validator

## Критерии приёмки

- [ ] golangci-lint 0 issues
- [ ] ESLint 0 issues
- [ ] go vet 0 issues
- [ ] Error wrapping везде
- [ ] Context propagation везде
- [ ] Godoc для public функций
- [ ] TypeScript strict, 0 `any`
