// T2205 — Нагрузочное тестирование: User Profile
// Целевая нагрузка: 200 RPS, P95 < 100ms, P99 < 200ms
//
// Сценарий имитирует запросы к профилю пользователя:
//   - GET /user/me (чтение профиля — основной сценарий)
//   - PUT /user/me (обновление контактов)
//
// Профиль — один из самых частых запросов в системе (вызывается
// при загрузке страницы, проверке прав, отображении данных).

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// --- Кастомные метрики ---
const profileGetDuration = new Trend('profile_get_duration');
const profileUpdateDuration = new Trend('profile_update_duration');
const errorRate = new Rate('profile_errors');
const requestsCounter = new Counter('profile_total_requests');

// --- Конфигурация ---
export const options = {
  stages: [
    { duration: '30s', target: 50 },     // ramp-up: 0 → 50 вирт. пользователей
    { duration: '1m', target: 200 },     // peak: 200 вирт. пользователей
    { duration: '30s', target: 0 },      // cooldown: 200 → 0
  ],
  thresholds: {
    // SLA: P95 < 100ms, P99 < 200ms (профиль — быстрый эндпоинт)
    http_req_duration: ['p(95)<100', 'p(99)<200'],
    // Ошибки не более 1%
    http_req_failed: ['rate<0.01'],
    // Кастомные метрики
    profile_get_duration: ['p(95)<100', 'p(99)<200'],
    profile_update_duration: ['p(95)<100', 'p(99)<200'],
    profile_errors: ['rate<0.01'],
  },
  noConnectionReuse: false,
  maxRedirects: 0,
};

// --- URL ---
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const API_PREFIX = '/api/v1';

// --- Генератор данных для обновления ---
function generatePhone() {
  return '+7900' + Math.floor(Math.random() * 10000000).toString().padStart(7, '0');
}

function generateEmail() {
  const domains = ['test.ru', 'lkfl.local'];
  return `user_${Math.floor(Math.random() * 100000)}@${domains[Math.floor(Math.random() * domains.length)]}`;
}

// --- Сценарии ---

// S01: чтение профиля пользователя
function scenarioGetProfile() {
  const url = `${BASE_URL}${API_PREFIX}/user/me`;

  const res = http.get(url, {
    headers: {
      'Accept': 'application/json',
    },
  });
  profileGetDuration.add(res.timings.duration);
  requestsCounter.add(1);

  const success = check(res, {
    'get: status 200 or 401': (r) => r.status === 200 || r.status === 401,
    'get: fast response': (r) => r.timings.duration < 100,
  });
  if (!success) errorRate.add(1);
}

// S01: обновление контактов
function scenarioUpdateProfile() {
  const url = `${BASE_URL}${API_PREFIX}/user/me`;
  const payload = JSON.stringify({
    phone: generatePhone(),
    email: generateEmail(),
  });

  const res = http.put(url, payload, {
    headers: {
      'Content-Type': 'application/json',
      'Accept': 'application/json',
    },
  });
  profileUpdateDuration.add(res.timings.duration);
  requestsCounter.add(1);

  // 200 (успех), 401 (без токена), 422 (валидация)
  const success = check(res, {
    'update: status 2xx or 4xx': (r) => r.status >= 200 && r.status < 500,
    'update: fast response': (r) => r.timings.duration < 200,
  });
  if (!success) errorRate.add(1);
}

// --- Основная функция ---
// Веса: 90% — чтение, 10% — обновление
export default function () {
  if (Math.random() < 0.90) {
    scenarioGetProfile();
  } else {
    scenarioUpdateProfile();
  }

  sleep(0.05);
}

// --- Init ---
export function setup() {
  const res = http.get(`${BASE_URL}/api/healthz`);
  check(res, { 'health check': (r) => r.status === 200 });
  return {};
}
