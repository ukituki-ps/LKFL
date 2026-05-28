// T2205 — Нагрузочное тестирование: Каталог (Engagements)
// Целевая нагрузка: 500 RPS, P95 < 200ms, P99 < 500ms
//
// Сценарий имитирует запросы сотрудников к каталогу льгот/активностей:
//   - список энгэйджментов с пагинацией
//   - фильтрация по типу (benefit/activity)
//   - поиск по ключевым словам
//   - получение деталей конкретного engagement
//   - получение категорий и типов для фильтров

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// --- Кастомные метрики ---
const catalogListDuration = new Trend('catalog_list_duration');
const catalogDetailDuration = new Trend('catalog_detail_duration');
const catalogFilterDuration = new Trend('catalog_filter_duration');
const catalogSearchDuration = new Trend('catalog_search_duration');
const errorRate = new Rate('catalog_errors');
const requestsCounter = new Counter('catalog_total_requests');

// --- Конфигурация ---
export const options = {
  stages: [
    { duration: '30s', target: 100 },   // ramp-up: 0 → 100 вирт. пользователей
    { duration: '1m', target: 500 },    // peak: удержание 500 вирт. пользователей
    { duration: '30s', target: 200 },   // ramp-down: 500 → 200
    { duration: '30s', target: 0 },     // cooldown: 200 → 0
  ],
  thresholds: {
    // SLA: P95 < 200ms, P99 < 500ms
    http_req_duration: ['p(95)<200', 'p(99)<500'],
    // Ошибки не более 1%
    http_req_failed: ['rate<0.01'],
    // Кастомные метрики
    catalog_list_duration: ['p(95)<200', 'p(99)<500'],
    catalog_detail_duration: ['p(95)<200', 'p(99)<500'],
    catalog_filter_duration: ['p(95)<200', 'p(99)<500'],
    catalog_search_duration: ['p(95)<200', 'p(99)<500'],
    catalog_errors: ['rate<0.01'],
  },
  // Базовые заголовки
  noConnectionReuse: false,
  maxRedirects: 0,
};

// --- URL ---
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const API_PREFIX = '/api/v1';

// --- Веса сценариев (реальное распределение запросов) ---
// 60% — список каталога, 20% — детали, 10% — фильтры, 10% — поиск
const scenarios = ['list', 'list', 'list', 'list', 'list', 'list', 'detail', 'filter', 'search'];

// --- Генератор параметров запроса ---
function getRandomPage() {
  return Math.floor(Math.random() * 50) + 1;
}

function getRandomPerPage() {
  const options = [10, 20, 30, 50];
  return options[Math.floor(Math.random() * options.length)];
}

function getRandomType() {
  return Math.random() > 0.5 ? 'benefit' : 'activity';
}

function getRandomSearchQuery() {
  const queries = [
    'ДМС', 'фитнес', 'спорт', 'кино', 'театр', 'ресторан', 'кафе',
    'йога', 'массаж', 'образование', 'книги', 'музыка', 'путешествие',
    'здоровье', 'бонус', 'скидка', 'льгота', 'активность', 'wellness'
  ];
  return queries[Math.floor(Math.random() * queries.length)];
}

function getRandomStatus() {
  const statuses = ['active', 'available', 'seasonal'];
  return statuses[Math.floor(Math.random() * statuses.length)];
}

// --- Сценарии ---

// S04, S09: каталог энгэйджментов — пагинация
function scenarioList() {
  const page = getRandomPage();
  const perPage = getRandomPerPage();
  const url = `${BASE_URL}${API_PREFIX}/engagements?page=${page}&per_page=${perPage}`;

  const res = http.get(url);
  catalogListDuration.add(res.timings.duration);
  requestsCounter.add(1);

  const success = check(res, {
    'list: status 200': (r) => r.status === 200,
    'list: has data': (r) => r.json && r.json().data,
  });
  if (!success) errorRate.add(1);
}

// S10: детали engagement type
function scenarioDetail() {
  // ID engagement в диапазоне реальных данных
  const id = Math.floor(Math.random() * 200) + 1;
  const url = `${BASE_URL}${API_PREFIX}/engagements/${id}`;

  const res = http.get(url);
  catalogDetailDuration.add(res.timings.duration);
  requestsCounter.add(1);

  const success = check(res, {
    'detail: status 200 or 404': (r) => r.status === 200 || r.status === 404,
  });
  if (!success) errorRate.add(1);
}

// Фильтрация по типу и категории
function scenarioFilter() {
  const type = getRandomType();
  const status = getRandomStatus();
  const page = getRandomPage();
  const url = `${BASE_URL}${API_PREFIX}/engagements?type=${type}&status=${status}&page=${page}&per_page=20`;

  const res = http.get(url);
  catalogFilterDuration.add(res.timings.duration);
  requestsCounter.add(1);

  const success = check(res, {
    'filter: status 200': (r) => r.status === 200,
  });
  if (!success) errorRate.add(1);
}

// Поиск по ключевым словам
function scenarioSearch() {
  const query = getRandomSearchQuery();
  const url = `${BASE_URL}${API_PREFIX}/engagements?search=${encodeURIComponent(query)}&page=1&per_page=20`;

  const res = http.get(url);
  catalogSearchDuration.add(res.timings.duration);
  requestsCounter.add(1);

  const success = check(res, {
    'search: status 200': (r) => r.status === 200,
  });
  if (!success) errorRate.add(1);
}

// --- Основная функция ---
export default function () {
  // Выбираем случайный сценарий по весам
  const scenario = scenarios[Math.floor(Math.random() * scenarios.length)];

  switch (scenario) {
    case 'list':
      scenarioList();
      break;
    case 'detail':
      scenarioDetail();
      break;
    case 'filter':
      scenarioFilter();
      break;
    case 'search':
      scenarioSearch();
      break;
  }

  // Небольшая пауза между запросами (имитация поведения пользователя)
  sleep(0.1);
}

// --- Init: проверка доступности ---
export function setup() {
  const res = http.get(`${BASE_URL}/api/healthz`);
  check(res, { 'health check': (r) => r.status === 200 });
  return {};
}
