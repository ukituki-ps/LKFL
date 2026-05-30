// T2205 — Нагрузочное тестирование: Auth
// Целевая нагрузка: 100 RPS, P95 < 500ms, P99 < 1000ms
//
// Сценарий имитирует аутентификационные запросы:
//   - GET /user/me (проверка сессии, самый частый запрос)
//   - GET /auth/callback (OIDC callback после Keycloak)
//   - POST /auth/register (регистрация, редкий сценарий)
//
// Примечание: реальный OIDC flow (login → Keycloak → callback) не тестируется,
// так как требует браузера. Тестируется только backend часть —
// валидация токена и обработка callback.

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// --- Кастомные метрики ---
const meDuration = new Trend('auth_me_duration');
const callbackDuration = new Trend('auth_callback_duration');
const registerDuration = new Trend('auth_register_duration');
const errorRate = new Rate('auth_errors');
const requestsCounter = new Counter('auth_total_requests');

// --- Конфигурация ---
export const options = {
  stages: [
    { duration: '30s', target: 20 },     // ramp-up: 0 → 20 вирт. пользователей
    { duration: '1m', target: 100 },     // peak: 100 вирт. пользователей
    { duration: '30s', target: 0 },      // cooldown: 100 → 0
  ],
  thresholds: {
    // SLA: P95 < 500ms, P99 < 1000ms
    http_req_duration: ['p(95)<500', 'p(99)<1000'],
    // Ошибки не более 1%
    http_req_failed: ['rate<0.01'],
    // Кастомные метрики
    auth_me_duration: ['p(95)<500', 'p(99)<1000'],
    auth_callback_duration: ['p(95)<500', 'p(99)<1000'],
    auth_register_duration: ['p(95)<500', 'p(99)<1000'],
    auth_errors: ['rate<0.01'],
  },
  noConnectionReuse: false,
  maxRedirects: 0,
};

// --- URL ---
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const API_PREFIX = '/api/v1';

// --- Генератор тестовых данных ---
function generateEmail() {
  const domains = ['test.ru', 'lkfl.local', 'dev.k6'];
  const domain = domains[Math.floor(Math.random() * domains.length)];
  return `user_${Date.now()}_${Math.floor(Math.random() * 10000)}@${domain}`;
}

function generatePassword() {
  return `TestPass_${Math.floor(Math.random() * 100000)}!`;
}

function generateName() {
  const names = ['Иванов Иван', 'Петрова Анна', 'Сидоров Алексей', 'Козлова Мария'];
  return names[Math.floor(Math.random() * names.length)];
}

// --- Сценарии ---

// S01: профиль пользователя (самый частый auth-запрос)
// В реальных условиях вызывается с valid JWT token.
// Без токена ожидаем 401 — это нормальное поведение для load-теста без auth setup.
function scenarioMe() {
  const url = `${BASE_URL}${API_PREFIX}/user/me`;

  const res = http.get(url, {
    headers: {
      // В production здесь будет Authorization: Bearer <JWT>
      // Для load-теста без токена ожидаем 401
      'Accept': 'application/json',
    },
  });
  meDuration.add(res.timings.duration);
  requestsCounter.add(1);

  // 200 (если есть токен) или 401 (без токена) — оба допустимы
  const success = check(res, {
    'me: status 200 or 401': (r) => r.status === 200 || r.status === 401,
    'me: response time OK': (r) => r.timings.duration < 500,
  });
  if (!success) errorRate.add(1);
}

// OIDC callback — обработка возврата из Keycloak
// state — случайный параметр для имитации реального callback
function scenarioCallback() {
  const code = `auth_code_${Math.floor(Math.random() * 100000)}`;
  const state = `state_${Math.floor(Math.random() * 100000)}`;
  const url = `${BASE_URL}${API_PREFIX}/auth/callback?code=${code}&state=${state}`;

  const res = http.get(url);
  callbackDuration.add(res.timings.duration);
  requestsCounter.add(1);

  // Ожидаем перенаправление (302) или ошибку (400/401) при невалидном code
  const success = check(res, {
    'callback: status 302 or 4xx': (r) =>
      r.status === 302 || r.status === 400 || r.status === 401,
    'callback: response time OK': (r) => r.timings.duration < 1000,
  });
  if (!success) errorRate.add(1);
}

// S01: кастомная регистрация
// Редкий сценарий (1% от всех запросов), но тестируем для полноты
function scenarioRegister() {
  const url = `${BASE_URL}${API_PREFIX}/auth/register`;
  const payload = JSON.stringify({
    first_name: generateName().split(' ')[0],
    last_name: generateName().split(' ')[1],
    email: generateEmail(),
    password: generatePassword(),
    date_of_birth: '1990-01-01',
    phone: '+7900' + Math.floor(Math.random() * 10000000).toString().padStart(7, '0'),
    consents: {
      personal_data: true,
      marketing: false,
    },
  });

  const res = http.post(url, payload, {
    headers: {
      'Content-Type': 'application/json',
      'Accept': 'application/json',
    },
  });
  registerDuration.add(res.timings.duration);
  requestsCounter.add(1);

  // 201 (успех) или 409 (дубликат) или 422 (валидация)
  const success = check(res, {
    'register: status 2xx or 4xx': (r) =>
      r.status >= 200 && r.status < 500,
    'register: response time OK': (r) => r.timings.duration < 1000,
  });
  if (!success) errorRate.add(1);
}

// --- Основная функция ---
// Веса: 80% — me, 15% — callback, 5% — register
export default function () {
  const rand = Math.random();

  if (rand < 0.80) {
    scenarioMe();
  } else if (rand < 0.95) {
    scenarioCallback();
  } else {
    scenarioRegister();
  }

  sleep(0.05);
}

// --- Init ---
export function setup() {
  const res = http.get(`${BASE_URL}/api/healthz`);
  check(res, { 'health check': (r) => r.status === 200 });
  return {};
}
