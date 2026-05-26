# ADR-001: Выбор Go + React для реализации

**Статус:** Accepted
**Дата:** 2026-05-22
**Контекст:** М01-создание-описания

## Контекст

Необходимо выбрать технологический стек для backend и frontend платформы гибких льгот. Система на 100 000+ пользователей, запуск лето 2026. Единая команда разработки, ограниченное время.

## Решение

**Backend:** Go 1.22 — type safety, concurrency (goroutines), low memory footprint, native Docker support, единый стек с April Profile/Worker.

**Frontend:** React 18 + Mantine + `@april/ui` — единая DS с April экосистемой, production-ready компоненты, a11y, dark mode.

## Альтернативы

| Вариант | Плюсы | Минусы | Вердикт |
|---------|-------|--------|---------|
| **Go + React** (выбрано) | Единый стек с April, type safety, зрелая экосистема | — | ✅ |
| Java (Spring Boot) + React | Enterprise-стандарт, много библиотек | Heavy, долгий startup, несовпадает с April | ❌ |
| Python (FastAPI) + React | Быстрая разработка, ML-интеграция | GIL, производительность, несовпадает с April | ❌ |
| Go + Angular | Go единый, Angular — enterprise-фронт | Angular тяжелее React, несовпадает с @april/ui | ❌ |
| Go + Vue | Vue легче React | Меньше экосистемы для enterprise, несовпадает с April | ❌ |
| TypeScript (NestJS) + React | Единый язык front/back | Нет единого стека с April (Go), overhead NestJS | ❌ |

## Аргументы «за»

- Команда Go-разработчиков работает на всех проектах (LKFL, Profile, Worker)
- Фронтенд-команда использует единый UI kit (`@april/ui`)
- Go: low memory, fast compile, simple deploy (single binary)
- React: крупнейшая экосистема, Mantine — production-ready, a11y-готов

## Аргументы «против»

- Go: меньше библиотек для UI, нет ORMs уровня Hibernate
- React: bundle size, learning curve для новых разработчиков
- Миграция на другие технологии будет дорогостоящей

## Вердикт

**Go + React.** Единственный вариант, совместимый с April экосистемой. Команда уже знает Go и React. Альтернативы либо несовместимы с April, либо вносят лишний overhead.

## Следствия

- Команда Go-разработчиков может работать на всех трёх проектах (LKFL, Profile, Worker)
- Фронтенд-команда использует единый UI kit (`@april/ui`)
- Миграция на другие технологии будет дорогостоящей
- Все новые проекты April — Go backend + React frontend по умолчанию
