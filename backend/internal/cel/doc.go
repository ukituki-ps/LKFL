// Package cel — CEL (Common Expression Language) Rule Engine: единый движок бизнес-логики.
//
// Заменяет 4 независимых механизма условий:
//   - billing YAML-array → CEL expression
//   - eligibility AND/OR/Groups → CEL expression
//   - flow condition_expr → CEL expression
//   - recommendations JSON segments → CEL expression
//
// LLM генерирует CEL из русского текста через internal/llm/.
// TagResolver вычисляет теги пользователя из HR-профиля + survey-ответов.
//
// ADR-021: CEL Engine (единый движок бизнес-логики)
package cel
