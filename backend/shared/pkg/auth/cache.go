// Package auth — JWKS cache.
//
// go-oidc (github.com/coreos/go-oidc) автоматически кэширует JWKS внутри
// oidc.Provider. Кэш обновляется по TTL и при ошибках подписи.
// Отдельный JWKS кэш не требуется.
//
// Этот файл оставлен как placeholder для возможного future расширения
// (например, кэш tenant-specific verifier'ов или Redis-backed shared cache).
package auth

// JWKS кэширование:
// - go-oidc использует google/go-jose для подписи и кэширования JWKS
// - Кэш обновляется автоматически при истечении TTL и при signature errors
// - Для multi-tenant сценария можно добавить кэш verifier'ов по realm slug:
//
//	type VerifierCache struct {
//	    mu       sync.RWMutex
//	    verifiers map[string]*oidc.IDTokenVerifier // key = realm slug
//	}
//
// Пока не требуется — каждый tenant имеет свой verifier при инициализации.
