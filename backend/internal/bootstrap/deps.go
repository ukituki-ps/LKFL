// Package bootstrap — stub для удержания зависимостей в go.mod.
//
// На этапе T1701 код модулей ещё не создан, но go mod tidy удалит
// неиспользуемые зависимости. Этот файл обеспечивает blank imports
// всех 15 production-зависимостей.
//
// Удаляется после T1702 (создание реальной структуры проекта).
package bootstrap

import (
	_ "github.com/coreos/go-oidc"
	_ "github.com/getsentry/sentry-go"
	_ "github.com/go-chi/chi/v5"
	_ "github.com/go-playground/validator/v10"
	_ "github.com/google/cel-go/cel"
	_ "github.com/jackc/pgx/v5"
	_ "github.com/jung-kurt/gofpdf"
	_ "github.com/prometheus/client_golang/prometheus"
	_ "github.com/redis/go-redis/v9"
	_ "github.com/rs/cors"
	_ "github.com/spf13/viper"
	_ "github.com/xuri/excelize/v2"
	_ "google.golang.org/grpc"
	_ "google.golang.org/protobuf/proto"
	_ "golang.org/x/crypto/bcrypt"
)
