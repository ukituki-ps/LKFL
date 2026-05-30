-- Ref: Fix — Add missing JSONB columns to users and accounts tables
-- Description: Code expects metadata (JSONB) in users and settings (JSONB) in accounts,
-- but original migrations missed these columns.

-- Add metadata column to users table (nullable, defaults to empty JSON)
ALTER TABLE lkfl_platform.users
    ADD COLUMN IF NOT EXISTS metadata JSONB NOT NULL DEFAULT '{}'::jsonb;

-- Add settings column to accounts table
ALTER TABLE lkfl_platform.accounts
    ADD COLUMN IF NOT EXISTS settings JSONB NOT NULL DEFAULT '{}'::jsonb;
