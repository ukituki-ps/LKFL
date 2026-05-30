// Atlas DDL migration configuration для LKFL.
//
// Используется для dev-миграций через `atlas migrate apply`.
// Production миграции выполняются через `lkfl-server migrate` (embedded).
//
// Документация: https://atlasgo.io/getting-started/datasource-postgres

data "external_schema" "lkfl_platform" {
	// Схемы задаются inline в migration SQL файлах.
	// Atlas используется для diff/plan, не для auto-migrate.
}

env "dev" {
	src = "./migrations"

	dev = "postgres://lkfl:changeme_dev_password@localhost:5432/lkfl_dev?sslmode=disable"

	migrate {
		dir = "file://migrations"
	}

	format {
		migrate {
			diff = "{{ sql .Diff \"  \" }}"
		}
	}
}
