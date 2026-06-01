-- Ref: T1714 — Add missing columns to user_roles
-- Description: Код repository.go ожидает granted_at, granted_by, expires_at,
--              но migration 20260528140000_engagement.sql создала таблицу без них.
--              Эта миграция добавляет недостающие колонки для синхронизации
--              схемы staging БД с кодом.

ALTER TABLE lkfl_platform.user_roles
    ADD COLUMN IF NOT EXISTS granted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ADD COLUMN IF NOT EXISTS granted_by UUID REFERENCES lkfl_platform.users(id),
    ADD COLUMN IF NOT EXISTS expires_at TIMESTAMPTZ;
