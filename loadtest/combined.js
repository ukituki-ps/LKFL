// T2205 — Нагрузочное тестирование: Комбинированная нагрузка
//
// Имитирует реальное распределение трафика по всем основным
// эндпоинтам одновременно. Проверяет что SLA держится при
// смешанной нагрузке на все модули.
//
// Распределение (на основе типичной нагрузки):
//   - 40% — Каталог (engagements)
//   - 20% — User Profile (me)
//   - 20% — User Engagements (список подключённых)
//   - 10% — Dashboard
//   - 5% — Auth (callback)
//   - 5% — Рекомендации

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// --- Кастомные метрики ---
const catalogDuration = new Trend('combined_catalog_duration');
const profileDuration = new Trend('combined_profile_duration');
const userEngagementsDuration = new Trend('combined_user_engagements_duration');
const dashboardDuration = new Trend('combined_dashboard_duration');
const authDuration = new Trend('combined_auth_duration');
const recommendationsDuration = new Trend('combined_recommendations_duration');
const errorRate = new Rate('combined_errors');
const requestsCounter = new Counter('combined_total_requests');

// --- Конфигурация ---
export const options = {
  stages: [
    { duration: '30s', target: 50 },     // ramp-up: лёгкая нагрузка
    { duration: '1m', target: 300 },     // moderate
    { duration: '2m', target: 600 },     // peak: общая нагрузка 600 VU
    { duration: '1m', target: 300 },     // moderate
    { duration: '30s', target: 100 },    // ramp-down
    { duration: '30s', target: 0 },      // cooldown
  ],
  thresholds: {
    // Общий SLA
    http_req_duration: ['p(95)<300', 'p(99)<800'],
    http_req_failed: ['rate<0.01'],
    // По модулям
    combined_catalog_duration: ['p(95)<200', 'p(99)<500'],
    combined_profile_duration: ['p(95)<100', 'p(99)<200'],
    combined_user_engagements_duration: ['p(95)<300', 'p(99)<600'],
    combined_dashboard_duration: ['p(95)<300', 'p(99)<600'],
    combined_auth_duration: ['p(95)<500', 'p(99)<1000'],
    combined_recommendations_duration: ['p(95)<300', 'p(99)<600'],
    combined_errors: ['rate<0.01'],
  },
  noConnectionReuse: false,
  maxRedirects: 0,
};

// --- URL ---
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const API_PREFIX = '/api/v1';

// --- Сценарии ---

function scenarioCatalog() {
  const page = Math.floor(Math.random() * 50) + 1;
  const perPage = [10, 20, 30, 50][Math.floor(Math.random() * 4)];
  const url = `${BASE_URL}${API_PREFIX}/engagements?page=${page}&per_page=${perPage}`;

  const res = http.get(url);
  catalogDuration.add(res.timings.duration);
  requestsCounter.add(1);

  const success = check(res, {
    'catalog: status 200': (r) => r.status === 200,
  });
  if (!success) errorRate.add(1);
}

function scenarioProfile() {
  const url = `${BASE_URL}${API_PREFIX}/user/me`;

  const res = http.get(url, {
    headers: { 'Accept': 'application/json' },
  });
  profileDuration.add(res.timings.duration);
  requestsCounter.add(1);

  const success = check(res, {
    'profile: status 200 or 401': (r) => r.status === 200 || r.status === 401,
  });
  if (!success) errorRate.add(1);
}

function scenarioUserEngagements() {
  const url = `${BASE_URL}${API_PREFIX}/user-engagements?page=1&per_page=20`;

  const res = http.get(url);
  userEngagementsDuration.add(res.timings.duration);
  requestsCounter.add(1);

  const success = check(res, {
    'user_engagements: status 200 or 401': (r) => r.status === 200 || r.status === 401,
  });
  if (!success) errorRate.add(1);
}

function scenarioDashboard() {
  const url = `${BASE_URL}${API_PREFIX}/dashboard`;

  const res = http.get(url);
  dashboardDuration.add(res.timings.duration);
  requestsCounter.add(1);

  const success = check(res, {
    'dashboard: status 200 or 401': (r) => r.status === 200 || r.status === 401,
  });
  if (!success) errorRate.add(1);
}

function scenarioAuth() {
  const code = `auth_${Math.floor(Math.random() * 100000)}`;
  const state = `state_${Math.floor(Math.random() * 100000)}`;
  const url = `${BASE_URL}${API_PREFIX}/auth/callback?code=${code}&state=${state}`;

  const res = http.get(url);
  authDuration.add(res.timings.duration);
  requestsCounter.add(1);

  const success = check(res, {
    'auth: status 302 or 4xx': (r) =>
      r.status === 302 || r.status === 400 || r.status === 401,
  });
  if (!success) errorRate.add(1);
}

function scenarioRecommendations() {
  const url = `${BASE_URL}${API_PREFIX}/recommendations`;

  const res = http.get(url);
  recommendationsDuration.add(res.timings.duration);
  requestsCounter.add(1);

  const success = check(res, {
    'recommendations: status 200 or 401': (r) => r.status === 200 || r.status === 401,
  });
  if (!success) errorRate.add(1);
}

// --- Основная функция ---
// Веса: 40% catalog, 20% profile, 20% user-engagements,
//        10% dashboard, 5% auth, 5% recommendations
export default function () {
  const rand = Math.random();

  if (rand < 0.40) {
    scenarioCatalog();
  } else if (rand < 0.60) {
    scenarioProfile();
  } else if (rand < 0.80) {
    scenarioUserEngagements();
  } else if (rand < 0.90) {
    scenarioDashboard();
  } else if (rand < 0.95) {
    scenarioAuth();
  } else {
    scenarioRecommendations();
  }

  sleep(0.05);
}

// --- Init ---
export function setup() {
  const res = http.get(`${BASE_URL}/api/healthz`);
  check(res, { 'health check': (r) => r.status === 200 });
  return {};
}
