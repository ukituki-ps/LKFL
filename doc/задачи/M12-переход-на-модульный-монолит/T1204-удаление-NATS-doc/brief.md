# T1204 — Удаление NATS doc, замена на DI interfaces

## Контекст

`архитектура/nats-subjects.md` описывает 20+ NATS JetStream subjects для межсервисной коммуникации. АDR-020 — NATS master registry. При modular monolith NATS не нужен (или опционален для future split).

## План

1. Удалить `архитектура/nats-subjects.md`
2. ADR-020: пометить как Superseded by ADR-024
3. ADR-005 (NATS JetStream): добавить note — "M12: оставлен как optional. В mono-режиме не используется."
4. Обновить `архитектура/README.md` — убрать nats-subjects.md из содержимого раздела
5. Обновить overview M12 — добавить пояснение про DI interfaces вместо NATS

## Ожидаемый результат

NATS документация убрана. ADR обновлены. README консистентен.