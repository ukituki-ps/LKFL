# T2303 — internal/cel/ (CEL Rule Engine)

## Веха

M23-cel-engine

## Тип

code

## Контекст

CEL engine на базе `google/cel-go`. Sandbox evaluation, type-safe.
Исходник: `doc/архитектура/cel-engine.md`, ADR-021.

## Что сделать

```go
package cel

import (
    "github.com/google/cel-go/cel"
    "github.com/google/cel-go/common/types"
    "lkfl/shared/pkg/celcontext"
)

type Engine struct {
    env *cel.Env
}

// NewEngine создаёт CEL environment с type providers
func NewEngine() (*Engine, error) {
    env, err := cel.NewEnv(
        cel.Container("lkfl"),
        cel.Types(&celcontext.CELContext{}),
        // Custom functions
        cel.Function("matches", ...),
        cel.Function("contains", ...),
    )
}

// Evaluate — оценка CEL expression в sandbox
func (e *Engine) Evaluate(ctx context.Context, expression string, context celcontext.CELContext) (interface{}, error) {
    // Compile (cache)
    // Interpret
    // Return result (bool, int, list)
    // Timeout: 5s max
}
```

## Критерии приёмки

- [ ] CEL environment с type providers
- [ ] Evaluate() с sandbox
- [ ] Expression caching
- [ ] Timeout 5s
- [ ] Custom functions (matches, contains)
- [ ] Unit tests: valid expression, invalid syntax, timeout, type error
