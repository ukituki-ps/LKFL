-- Создаёт database keycloak для Keycloak IdP
-- Выполняется один раз при первой инициализации PostgreSQL
SELECT 'CREATE DATABASE keycloak'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'keycloak')\gexec
