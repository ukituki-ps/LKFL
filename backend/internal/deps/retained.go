// Package deps — retained dependencies.
//
// Зависимости, включённые в go.mod по спецификации M17 (brief T1701),
// но ещё не используемые в рабочем коде. `go mod tidy` удалит их,
// если убрать этот файл.
//
// Каждая зависимость привязана к фазе, где она будет использована:
//
//	Phase F2 (M23–M29):
//	  validator  — валидация request body (internal/user, internal/consent)
//	  cel-go     — CEL Rule Engine (internal/cel, internal/eligibility)
//	  bcrypt     — хеширование чувствительных данных
//
//	Phase F3 (M30–M38):
//	  gofpdf     — PDF генерация (документы льгот, полисы)
//	  excelize   — XLSX парсинг (HR sync, массовая загрузка)
//
//	General:
//	  rs/cors    — production CORS middleware (замена ручной corsMiddleware в app/server.go)
//	  protobuf   — gRPC proto definitions (lkfl-integration-proxy, M16+)
//
// Когда зависимость будет использована в реальном коде — удалить строку из этого файла.
package deps

import (
	_ "github.com/go-playground/validator/v10" // F2: request validation
	_ "github.com/google/cel-go/cel"           // F2: CEL Rule Engine
	_ "github.com/jung-kurt/gofpdf"            // F3: PDF generation
	_ "github.com/rs/cors"                     // General: production CORS
	_ "github.com/xuri/excelize/v2"            // F3: XLSX parsing
	_ "golang.org/x/crypto/bcrypt"             // F2: password hashing
	_ "google.golang.org/protobuf/proto"       // General: gRPC proto
)
