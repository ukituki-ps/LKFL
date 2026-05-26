// Atlas Migration Engine — https://atlasgo.io
//
// Atlas config для проекта LKFL.
// Миграции хранятся в формате plain SQL (без аннотаций Goose).
//
// Apply:  atlas migrate apply --url "$DB_DSN" --dir file://migrations
// Undo:   atlas migrate undo --url "$DB_DSN" --dir file://migrations --count 1

data "external_schema" "sdk" {
  program = ["go", "run", "./internal/schema/main.go"]
}
