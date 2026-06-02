-- Создаёт database keycloak_dev для Keycloak dev стенда
-- Выполняется один раз при первой инициализации PostgreSQL dev
SELECT 'CREATE DATABASE keycloak_dev'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'keycloak_dev')\gexec
