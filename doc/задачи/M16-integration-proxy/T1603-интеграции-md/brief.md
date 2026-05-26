# T1603 — `архитектура/интеграции.md` rewrite

## Веха

M16-integration-proxy

## Контекст

`архитектура/интеграции.md` описывает прямые вызовы из монолита. Нужно переписать под архитектуру с proxy.

## Что сделать

Переписать документ:
1. Архитектура — diagram: Nginx → монолит + proxy → провайдеры
2. ProviderGateway → gRPC client в монолите
3. ProviderAdapter → в proxy, не в монолите
4. Generic REST adapter → в proxy
5. Webhook handling → proxy only
6. Admin endpoints → mono делегирует proxy через gRPC
7. Circuit breaker → реализован в proxy (не только в документации)
8. Error handling → proxy returns structured errors via gRPC

### Файлы-мишени

| Действие | Файл |
|----------|------|
| Переписать | `архитектура/интеграции.md` |

### Критерии приёмки

- [ ] Архитектура описана с proxy
- [ ] Diagram обновлена
- [ ] ProviderGateway → gRPC client
- [ ] ProviderAdapter → proxy
- [ ] Webhook handling → proxy
- [ ] Circuit breaker → реализован в proxy
- [ ] 0 ссылок на прямые вызовы из монолита
