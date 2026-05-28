// Seed — загрузка начальных данных для разработки и staging.
//
// Создаёт tenant «sdek» с brand config, пользователями, категориями,
// типами энгейджментов и офферами.
// Idempotent: повторный запуск не ломает данные.
// Только для dev/staging окружения, не для production.
//
// Запуск:
//
//	make seed
//	или
//	DB_DSN="postgresql://..." go run ./cmd/seed/
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func main() {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		fmt.Println("DB_DSN required")
		os.Exit(1)
	}

	conn, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect error: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = conn.Close(context.Background()) }()

	ctx := context.Background()

	// ─── Миграции ───
	fmt.Println("Running migrations...")
	if mErr := runMigrations(ctx, conn); mErr != nil {
		fmt.Fprintf(os.Stderr, "migrations error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("  Migrations OK")

	// ─── Tenant ───
	sdekID, err := upsertTenant(ctx, conn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "upsert tenant error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("  Tenant: sdek (id=%s)\n", sdekID)

	// ─── Brand config ───
	if err := upsertBrandConfig(ctx, conn, sdekID); err != nil {
		fmt.Fprintf(os.Stderr, "upsert brand config error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("  Brand config: OK")

	// ─── Категории ───
	if err := seedCategories(ctx, conn, sdekID); err != nil {
		fmt.Fprintf(os.Stderr, "seed categories error: %v\n", err)
		os.Exit(1)
	}

	// ─── Пользователи ───
	if err := seedUsers(ctx, conn, sdekID); err != nil {
		fmt.Fprintf(os.Stderr, "seed users error: %v\n", err)
		os.Exit(1)
	}

	// ─── Типы энгейджментов ───
	if err := seedEngagementTypes(ctx, conn, sdekID); err != nil {
		fmt.Fprintf(os.Stderr, "seed engagement types error: %v\n", err)
		os.Exit(1)
	}

	// ─── Офферы ───
	if err := seedEngagementOffers(ctx, conn, sdekID); err != nil {
		fmt.Fprintf(os.Stderr, "seed engagement offers error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nSeed completed successfully!")
}

// ─── Миграции ───

func runMigrations(ctx context.Context, conn *pgx.Conn) error {
	// Create schema
	_, err := conn.Exec(ctx, "CREATE SCHEMA IF NOT EXISTS lkfl_platform")
	if err != nil {
		return fmt.Errorf("create schema: %w", err)
	}

	// Create migration tracking table
	_, err = conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS lkfl_platform.schema_migrations (
			id SERIAL PRIMARY KEY,
			filename VARCHAR(255) NOT NULL UNIQUE,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	// Find migrations directory
	migrationDir := findMigrationsDir()
	if migrationDir == "" {
		fmt.Println("  Warning: migrations directory not found, skipping")
		return nil
	}

	entries, err := os.ReadDir(migrationDir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	var migrationFiles []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") && !strings.HasSuffix(e.Name(), ".sql.down") {
			migrationFiles = append(migrationFiles, e.Name())
		}
	}
	sort.Strings(migrationFiles)

	applied := 0
	skipped := 0
	for _, name := range migrationFiles {
		// Check if already applied
		var exists bool
		err = conn.QueryRow(ctx, `
			SELECT EXISTS(SELECT 1 FROM lkfl_platform.schema_migrations WHERE filename = $1)
		`, name).Scan(&exists)
		if err != nil {
			return fmt.Errorf("check migration %s: %w", name, err)
		}

		if exists {
			skipped++
			continue
		}

		data, err := os.ReadFile(filepath.Join(migrationDir, name))
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}

		_, err = conn.Exec(ctx, string(data))
		if err != nil {
			return fmt.Errorf("apply migration %s: %w", name, err)
		}

		// Record migration
		_, err = conn.Exec(ctx, `
			INSERT INTO lkfl_platform.schema_migrations (filename) VALUES ($1)
		`, name)
		if err != nil {
			return fmt.Errorf("record migration %s: %w", name, err)
		}

		fmt.Printf("  Applied: %s\n", name)
		applied++
	}

	if applied == 0 && skipped > 0 {
		fmt.Printf("  All %d migrations already applied\n", skipped)
	} else if applied > 0 {
		fmt.Printf("  Applied %d, skipped %d\n", applied, skipped)
	}

	return nil
}

func findMigrationsDir() string {
	candidates := []string{
		"migrations",
		"../migrations",
		"../../migrations",
		"../../../migrations",
		"../../../../migrations",
	}
	for _, c := range candidates {
		info, err := os.Stat(c)
		if err == nil && info.IsDir() {
			return c
		}
	}
	return ""
}

// ─── Tenant ───

func upsertTenant(ctx context.Context, conn *pgx.Conn) (uuid.UUID, error) {
	var existingID uuid.UUID
	err := conn.QueryRow(ctx, `
		SELECT id FROM lkfl_platform.tenants WHERE slug = $1
	`, "sdek").Scan(&existingID)

	if err == nil {
		return existingID, nil
	}

	sdekID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	_, err = conn.Exec(ctx, `
		INSERT INTO lkfl_platform.tenants (id, slug, name, status, settings)
		VALUES ($1, 'sdek', 'СДЭК', 'active', '{}')
	`, sdekID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("insert tenant sdek: %w", err)
	}

	return sdekID, nil
}

func upsertBrandConfig(ctx context.Context, conn *pgx.Conn, tenantID uuid.UUID) error {
	_, err := conn.Exec(ctx, `
		INSERT INTO lkfl_platform.tenant_brand_config (
			tenant_id, primary_color, secondary_color, brand_name, css_variables
		) VALUES ($1, '#E30613', '#FFFFFF', 'СДЭК Льготы', $2)
		ON CONFLICT (tenant_id) DO UPDATE SET
			primary_color = '#E30613',
			secondary_color = '#FFFFFF',
			brand_name = 'СДЭК Льготы',
			css_variables = $2,
			updated_at = NOW()
	`, tenantID, `{
			"--april-color-primary": "#E30613",
			"--april-color-primary-hover": "#C50510"
		}`)
	if err != nil {
		return fmt.Errorf("upsert brand config: %w", err)
	}

	return nil
}

// ─── Категории ───

func seedCategories(ctx context.Context, conn *pgx.Conn, tenantID uuid.UUID) error {
	categories := []struct {
		slug      string
		name      string
		icon      string
		sortOrder int
	}{
		{"dms", "ДМС", "Shield", 1},
		{"fitness", "Фитнес", "Dumbbell", 2},
		{"food", "Питание", "Utensils", 3},
		{"edu", "Развитие", "BookOpen", 4},
		{"merch", "Мерч", "ShoppingBag", 5},
	}

	count := 0
	for _, c := range categories {
		res, err := conn.Exec(ctx, `
			INSERT INTO lkfl_platform.engagement_categories (tenant_id, slug, name, icon, sort_order)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (tenant_id, slug) DO UPDATE SET
				name = EXCLUDED.name,
				icon = EXCLUDED.icon,
				sort_order = EXCLUDED.sort_order,
				updated_at = NOW()
		`, tenantID, c.slug, c.name, c.icon, c.sortOrder)
		if err != nil {
			return fmt.Errorf("seed category %s: %w", c.slug, err)
		}
		// Count both inserts and updates
		if res.RowsAffected() > 0 {
			count++
		}
	}

	fmt.Printf("  Categories: %d inserted/updated\n", count)
	return nil
}

// ─── Пользователи ───

type seedUser struct {
	email       string
	firstName   string
	lastName    string
	phone       string
	keycloakSub string
	roles       []string
}

func seedUsers(ctx context.Context, conn *pgx.Conn, tenantID uuid.UUID) error {
	users := []seedUser{
		// Admin
		{
			email:       "admin@sdek.local",
			firstName:   "Александр",
			lastName:    "Петров",
			phone:       "+79001000001",
			keycloakSub: "kc-admin-001",
			roles:       []string{"admin"},
		},
		// HR Manager
		{
			email:       "hr@sdek.local",
			firstName:   "Елена",
			lastName:    "Иванова",
			phone:       "+79001000002",
			keycloakSub: "kc-hr-001",
			roles:       []string{"hr"},
		},
		// Catalog Manager
		{
			email:       "catalog@sdek.local",
			firstName:   "Дмитрий",
			lastName:    "Сидоров",
			phone:       "+79001000003",
			keycloakSub: "kc-catalog-001",
			roles:       []string{"catalog_manager"},
		},
		// Employee 1
		{
			email:       "ivanov@sdek.local",
			firstName:   "Иван",
			lastName:    "Петров",
			phone:       "+79001000010",
			keycloakSub: "kc-emp-001",
			roles:       []string{"employee"},
		},
		// Employee 2
		{
			email:       "petrova@sdek.local",
			firstName:   "Мария",
			lastName:    "Петрова",
			phone:       "+79001000011",
			keycloakSub: "kc-emp-002",
			roles:       []string{"employee"},
		},
		// Employee 3
		{
			email:       "sidorov@sdek.local",
			firstName:   "Сергей",
			lastName:    "Сидоров",
			phone:       "+79001000012",
			keycloakSub: "kc-emp-003",
			roles:       []string{"employee"},
		},
		// Employee 4
		{
			email:       "kozlova@sdek.local",
			firstName:   "Анна",
			lastName:    "Козлова",
			phone:       "+79001000013",
			keycloakSub: "kc-emp-004",
			roles:       []string{"employee"},
		},
		// Employee 5
		{
			email:       "novikov@sdek.local",
			firstName:   "Артём",
			lastName:    "Новиков",
			phone:       "+79001000014",
			keycloakSub: "kc-emp-005",
			roles:       []string{"employee"},
		},
	}

	userIDs := make([]uuid.UUID, len(users))
	// Deterministic IDs for reproducibility
	baseUUID := [16]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x10}
	for i := range users {
		baseUUID[15] = byte(100 + i)
		userIDs[i] = uuid.UUID(baseUUID)
	}

	// Insert users
	for i, u := range users {
		var userID uuid.UUID
		err := conn.QueryRow(ctx, `
			INSERT INTO lkfl_platform.users (id, tenant_id, email, first_name, last_name, phone, keycloak_sub)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (keycloak_sub) DO UPDATE SET updated_at = NOW()
			RETURNING id
		`, userIDs[i], tenantID, u.email, u.firstName, u.lastName, u.phone, u.keycloakSub).Scan(&userID)
		if err != nil {
			return fmt.Errorf("seed user %s: %w", u.email, err)
		}
		userIDs[i] = userID

		// Insert account
		_, err = conn.Exec(ctx, `
			INSERT INTO lkfl_platform.accounts (user_id, total_balance)
			VALUES ($1, $2)
			ON CONFLICT (user_id) DO UPDATE SET updated_at = NOW()
		`, userID, balanceForUser(u))
		if err != nil {
			return fmt.Errorf("seed account for %s: %w", u.email, err)
		}

		// Insert roles
		for _, role := range u.roles {
			_, err = conn.Exec(ctx, `
				INSERT INTO lkfl_platform.user_roles (user_id, role)
				VALUES ($1, $2)
				ON CONFLICT (user_id, role) DO NOTHING
			`, userID, role)
			if err != nil {
				return fmt.Errorf("seed role %s for %s: %w", role, u.email, err)
			}
		}
	}

	fmt.Printf("  Users: %d inserted (with accounts + roles)\n", len(users))
	return nil
}

// balanceForUser возвращает стартовый баланс для пользователя.
func balanceForUser(u seedUser) int64 {
	switch {
	case contains(u.roles, "admin"):
		return 100000 // 1000.00
	case contains(u.roles, "hr"):
		return 100000
	case contains(u.roles, "catalog_manager"):
		return 100000
	default:
		return 50000 // 500.00 для сотрудников
	}
}

func contains(slice []string, val string) bool {
	for _, s := range slice {
		if s == val {
			return true
		}
	}
	return false
}

// ─── Типы энгейджментов ───

type seedEngagementType struct {
	slug         string
	name         string
	description  string
	categorySlug string
	typ          string
	status       string
	// price — цена из прототипа:
	//   amount = число (баллы или рубли без дробной части)
	//   unit = "points" | "rub"
	//   period = "month" | "day" | "year" | "once" | "included"
	//   display = строка для фронтенда, например "3 500 ₽ / мес"
	priceAmount  int
	priceUnit    string
	pricePeriod  string
	priceDisplay string
	providerName string
	iconName     string // Lucide icon name
	badge        string // badge text: Активна, Доступна, Ожидает, Новинка
	badgeColor   string // badge color: green, gray, yellow, blue
}

func seedEngagementTypes(ctx context.Context, conn *pgx.Conn, tenantID uuid.UUID) error {
	// Удалить старые типы и офферы (CASCADE удалит офферы)
	_, err := conn.Exec(ctx, `
		DELETE FROM lkfl_platform.engagement_types WHERE tenant_id = $1
	`, tenantID)
	if err != nil {
		return fmt.Errorf("cleanup engagement types: %w", err)
	}

	// 12 льгот из прототипа «ЛК физика»
	// price: amount (число), unit ("points"|"rub"), period ("month"|"day"|"year"|"once"|"included"), display (строка)
	// iconName: Lucide icon name для отображения на карточке
	// badge: Активна, Доступна, Ожидает, Новинка
	types := []seedEngagementType{
		// ── ДМС ──
		{
			slug: "dms-base", name: "ДМС — Базовая программа",
			description:  "Полис добровольного медицинского страхования. Включает амбулаторное лечение, стационар и скорую помощь.",
			categorySlug: "dms", typ: "benefit", status: "active",
			priceAmount: 0, priceUnit: "rub", pricePeriod: "included", priceDisplay: "Включено в пакет",
			providerName: "АльфаСтрахование", iconName: "heart-pulse", badge: "Активна", badgeColor: "green",
		},
		{
			slug: "dms-extended", name: "ДМС — Расширенная программа",
			description:  "Расширенный пул клиник, стоматология, офтальмология и выезд врача на дом.",
			categorySlug: "dms", typ: "benefit", status: "active",
			priceAmount: 3500, priceUnit: "rub", pricePeriod: "month", priceDisplay: "3 500 ₽ / мес",
			providerName: "АльфаСтрахование", iconName: "shield-plus", badge: "Доступна", badgeColor: "gray",
		},
		{
			slug: "dms-family", name: "ДМС Семья",
			description:  "Добавьте супруга, детей или родителей к программе ДМС с персональным покрытием.",
			categorySlug: "dms", typ: "benefit", status: "active",
			priceAmount: 1800, priceUnit: "rub", pricePeriod: "month", priceDisplay: "от 1 800 ₽ / мес",
			providerName: "АльфаСтрахование", iconName: "users", badge: "Доступна", badgeColor: "gray",
		},

		// ── Фитнес ──
		{
			slug: "fitness-worldclass", name: "Фитнес — World Class",
			description:  "Доступ в сеть фитнес-клубов World Class с групповыми занятиями и бассейном.",
			categorySlug: "fitness", typ: "benefit", status: "active",
			priceAmount: 500, priceUnit: "points", pricePeriod: "month", priceDisplay: "500 баллов / мес",
			providerName: "World Class", iconName: "dumbbell", badge: "Активна", badgeColor: "green",
		},
		{
			slug: "fitness-sbersport", name: "СберСпорт — мультиспорт",
			description:  "Доступ к 1 000+ спортивным объектам: фитнес, йога, бассейн, единоборства.",
			categorySlug: "fitness", typ: "benefit", status: "active",
			priceAmount: 800, priceUnit: "points", pricePeriod: "month", priceDisplay: "800 баллов / мес",
			providerName: "СберСпорт", iconName: "bike", badge: "Доступна", badgeColor: "gray",
		},

		// ── Питание ──
		{
			slug: "food-lunch", name: "Обеды в офисе",
			description:  "Компенсация обедов в офисе или доставки до рабочего места до 300 ₽ в день.",
			categorySlug: "food", typ: "benefit", status: "active",
			priceAmount: 300, priceUnit: "rub", pricePeriod: "day", priceDisplay: "300 ₽ / день",
			providerName: "Яндекс Еда for Business", iconName: "utensils", badge: "Доступна", badgeColor: "gray",
		},
		{
			slug: "food-coffee", name: "Кофе и снеки",
			description:  "Компенсация покупки кофе и перекусов в офисе до 100 ₽ в день.",
			categorySlug: "food", typ: "benefit", status: "active",
			priceAmount: 100, priceUnit: "rub", pricePeriod: "day", priceDisplay: "100 ₽ / день",
			providerName: "Яндекс Еда for Business", iconName: "coffee", badge: "Доступна", badgeColor: "gray",
		},

		// ── Развитие ──
		{
			slug: "edu-skillbox", name: "Обучение — Skillbox",
			description:  "Доступ к профессиональным курсам: разработка, дизайн, маркетинг, управление проектами.",
			categorySlug: "edu", typ: "benefit", status: "active",
			priceAmount: 1200, priceUnit: "points", pricePeriod: "once", priceDisplay: "1 200 баллов",
			providerName: "Skillbox", iconName: "graduation-cap", badge: "Доступна", badgeColor: "gray",
		},
		{
			slug: "edu-psychologist", name: "Психолог онлайн",
			description:  "4 сессии с профессиональным психологом онлайн. Анонимно и конфиденциально.",
			categorySlug: "edu", typ: "benefit", status: "active",
			priceAmount: 600, priceUnit: "points", pricePeriod: "once", priceDisplay: "600 баллов",
			providerName: "Яндекс Психотерапия", iconName: "brain", badge: "Ожидает", badgeColor: "yellow",
		},
		{
			slug: "edu-english", name: "Английский язык",
			description:  "Индивидуальные онлайн-занятия с преподавателем. Уровни от Beginner до Advanced.",
			categorySlug: "edu", typ: "benefit", status: "active",
			priceAmount: 1500, priceUnit: "points", pricePeriod: "month", priceDisplay: "1 500 баллов / мес",
			providerName: "Skyeng", iconName: "languages", badge: "Доступна", badgeColor: "gray",
		},

		// ── Мерч ──
		{
			slug: "merch-sdek", name: "Мерч СДЭК",
			description:  "Фирменная одежда и аксессуары СДЭК: худи, футболки, кружки, термосы.",
			categorySlug: "merch", typ: "benefit", status: "active",
			priceAmount: 200, priceUnit: "points", pricePeriod: "once", priceDisplay: "от 200 баллов",
			providerName: "СДЭК Store", iconName: "shopping-bag", badge: "Доступна", badgeColor: "gray",
		},

		// ── Стоматология (в категории ДМС) ──
		{
			slug: "dms-dental", name: "Стоматология",
			description:  "Профилактические осмотры, чистка и лечение в сети стоматологических клиник.",
			categorySlug: "dms", typ: "benefit", status: "promo",
			priceAmount: 950, priceUnit: "points", pricePeriod: "year", priceDisplay: "950 баллов / год",
			providerName: "Мать и дитя", iconName: "smile", badge: "Новинка", badgeColor: "blue",
		},
	}

	count := 0
	for _, t := range types {
		var categoryID uuid.UUID
		err := conn.QueryRow(ctx, `
			SELECT id FROM lkfl_platform.engagement_categories
			WHERE tenant_id = $1 AND slug = $2
		`, tenantID, t.categorySlug).Scan(&categoryID)
		if err != nil {
			return fmt.Errorf("get category %s for type %s: %w", t.categorySlug, t.slug, err)
		}

		// metadata: icon_name, badge, badge_color, price (amount, unit, period, display)
		metadata := fmt.Sprintf(`{"icon_name":"%s","badge":"%s","badge_color":"%s","price":{"amount":%d,"unit":"%s","period":"%s","display":"%s"}}`,
			t.iconName, t.badge, t.badgeColor,
			t.priceAmount, t.priceUnit, t.pricePeriod, t.priceDisplay)

		_, err = conn.Exec(ctx, `
			INSERT INTO lkfl_platform.engagement_types
				(tenant_id, category_id, slug, name, description, type, status, cost_cents, provider_name, image_url, metadata)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11::jsonb)
		`, tenantID, categoryID, t.slug, t.name, t.description, t.typ, t.status, nil, t.providerName, nil, metadata)
		if err != nil {
			return fmt.Errorf("seed engagement type %s: %w", t.slug, err)
		}
		count++
	}

	fmt.Printf("  Engagement types: %d inserted\n", count)
	return nil
}

// ─── Офферы ──

type seedOffer struct {
	name        string
	description string
	costCents   int64
	sortOrder   int
}

func seedEngagementOffers(ctx context.Context, conn *pgx.Conn, tenantID uuid.UUID) error {
	// Удалить старые офферы
	_, err := conn.Exec(ctx, `
		DELETE FROM lkfl_platform.engagement_offers WHERE tenant_id = $1
	`, tenantID)
	if err != nil {
		return fmt.Errorf("cleanup engagement offers: %w", err)
	}

	// Офферы для каждого типа (из прототипа)
	offers := map[string][]seedOffer{
		"dms-base": {
			{"Включено в пакет", "Амбулаторное лечение, стационар и скорая помощь", 0, 1},
		},
		"dms-extended": {
			{"Расширенная", "Пул клиник + стоматология + офтальмология", 3500, 1},
			{"Расширенная + дом", "Расширенная + выезд врача на дом", 5000, 2},
		},
		"dms-family": {
			{"Супруг(а)", "ДМС для супруга/супруги", 1800, 1},
			{"Супруг(а) + ребёнок", "ДМС для семьи из 3 человек", 3200, 2},
			{"Семья полная", "ДМС для семьи из 4+ человек", 4500, 3},
		},
		"fitness-worldclass": {
			{"Абонемент", "Доступ в клуб World Class", 500, 1},
		},
		"fitness-sbersport": {
			{"Абонемент", "Доступ к 1 000+ объектам", 800, 1},
		},
		"food-lunch": {
			{"До 300 ₽/день", "Компенсация обеда", 300, 1},
		},
		"food-coffee": {
			{"До 100 ₽/день", "Кофе и перекусы в офисе", 100, 1},
		},
		"edu-skillbox": {
			{"Курс", "Любой профессиональный курс", 1200, 1},
		},
		"edu-psychologist": {
			{"1 сессия", "Одна консультация психолога онлайн", 150, 1},
			{"4 сессии", "Пакет из 4 консультаций", 600, 2},
		},
		"edu-english": {
			{"1 занятие", "Индивидуальный урок 60 мин", 250, 1},
			{"Пакет 8", "8 занятий (−10%)", 1800, 2},
		},
		"merch-sdek": {
			{"Футболка", "Фирменная футболка СДЭК", 200, 1},
			{"Худи", "Корпоративное худи", 500, 2},
			{"Термос", "Термос 500 мл", 300, 3},
			{"Рюкзак", "Рюкзак с логотипом", 700, 4},
		},
		"dms-dental": {
			{"Осмотр + чистка", "Профилактический визит", 450, 1},
			{"Лечение", "Лечение + профилактика", 950, 2},
		},
	}

	totalInserted := 0
	for engTypeSlug, typeOffers := range offers {
		var engTypeID uuid.UUID
		err := conn.QueryRow(ctx, `
			SELECT id FROM lkfl_platform.engagement_types
			WHERE tenant_id = $1 AND slug = $2
		`, tenantID, engTypeSlug).Scan(&engTypeID)
		if err != nil {
			return fmt.Errorf("get engagement type %s for offers: %w", engTypeSlug, err)
		}

		for _, o := range typeOffers {
			_, err = conn.Exec(ctx, `
				INSERT INTO lkfl_platform.engagement_offers
					(tenant_id, engagement_type_id, name, description, cost_cents, sort_order)
				VALUES ($1, $2, $3, $4, $5, $6)
			`, tenantID, engTypeID, o.name, o.description, o.costCents, o.sortOrder)
			if err != nil {
				return fmt.Errorf("seed offer %s/%s: %w", engTypeSlug, o.name, err)
			}
			totalInserted++
		}
	}

	fmt.Printf("  Engagement offers: %d inserted\n", totalInserted)
	return nil
}
