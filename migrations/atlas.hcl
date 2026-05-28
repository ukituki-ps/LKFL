// Atlas Migration Engine — https://atlasgo.io
//
// Atlas config для проекта LKFL.
// Миграции хранятся в формате plain SQL (без аннотаций Goose).
//
// Apply:  atlas migrate apply --url "$DB_DSN" --dir file://migrations
// Undo:   atlas migrate undo --url "$DB_DSN" --dir file://migrations --count 1
// Status: atlas migrate status --url "$DB_DSN" --dir file://migrations
//
// external_schema подключается в M18+ когда будет готов internal/schema/.
// Сейчас — plain SQL миграции без diff-генерации.
