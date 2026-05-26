-- Ref: T1705 — Atlas Migrations (init)
-- Description: Создание базовой схемы и расширений PostgreSQL.

-- Schema
CREATE SCHEMA IF NOT EXISTS lkfl_platform;

-- Extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
