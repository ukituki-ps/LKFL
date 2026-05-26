# T3601-T3610 — Integration Proxy (ADR-035)

## Веха

M36-integration-proxy

## T3601 — proto/integration/v1/integration.proto
```protobuf
syntax = "proto3";
package integration.v1;
service IntegrationService {
    rpc ResolveProvider(ResolveRequest) returns (ResolveResponse);
    rpc Activate(ActivateRequest) returns (ActivateResponse);
    rpc Deactivate(DeactivateRequest) returns (DeactivateResponse);
    rpc GetBalance(BalanceRequest) returns (BalanceResponse);
    rpc SyncCatalog(SyncRequest) returns (stream SyncResponse);
}
```

## T3602 — cmd/integration-proxy/main.go
- gRPC server :8090
- HTTP webhook receiver :8091
- DI wiring

## T3603 — integration-proxy/adapters/
- ProviderAdapter interface
- 2 hard-coded (референс)
- 3 YAML-based adapters

## T3604 — integration-proxy/circuitbreaker/
- Circuit breaker per provider
- States: closed → open → half-open
- Configurable thresholds

## T3605 — integration-proxy/webhook/
- Webhook receiver
- Signature verification
- Routing к провайдеру

## T3606 — internal/integrationclient/
- gRPC client к proxy (localhost)
- Retry, timeout
- Connection pooling

## T3607 — provider-configs/
- YAML конфиги провайдеров
- Schema validation

## T3608 — Migrations: lkfl_integration schema
```sql
CREATE SCHEMA lkfl_integration;
CREATE TABLE lkfl_integration.providers (...);
CREATE TABLE lkfl_integration.provider_credentials (...);
CREATE TABLE lkfl_integration.integration_logs (...);
CREATE TABLE lkfl_integration.dead_letters (...);
CREATE TABLE lkfl_integration.webhook_events (...);
CREATE TABLE lkfl_integration.sync_jobs (...);
```

## T3609 — docker-compose
- + lkfl-integration-proxy контейнер
- gRPC port :8090, webhook :8091

## T3610 — Prometheus metrics
- `integration_provider_latency_seconds`
- `integration_provider_errors_total`
- `integration_circuit_breaker_state`
- `integration_sync_duration_seconds`
- `integration_webhook_total`
- `integration_dead_letters_total`

## Критерии приёмки
- [ ] Все 10 задач
- [ ] gRPC contract работает
- [ ] 5 провайдеров (2 hard-coded + 3 YAML)
- [ ] Circuit breaker работает
- [ ] Webhooks receive + verify
- [ ] docker-compose 7 контейнеров
