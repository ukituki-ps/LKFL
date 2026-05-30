// Package main — сервер LKFL: HTTP-эндпоинты, миграции, seed.
//
// Подкоманды:
//
//	lkfl-server       — запуск HTTP-сервера (порт из config)
//	lkfl-server migrate — применение SQL-миграций из migrations/
//	lkfl-server seed    — миграции + загрузка seed-данных (tenant, пользователи, льготы)
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"lkfl/internal/app"
	"lkfl/shared/pkg/migrate"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "migrate":
			runMigrate()
			return
		case "seed":
			runSeed()
			return
		}
	}

	// Default: run server
	runServer()
}

// runServer запускает HTTP-сервер (основной режим работы бинарника).
func runServer() {
	// 1. Load config
	cfg, err := app.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	// 2. Provide dependencies
	server, cleanup, err := app.Provide(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "init error: %v\n", err)
		os.Exit(1)
	}
	defer cleanup()

	// 3. Start server in goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Start()
	}()

	// 4. Wait for interrupt signal (graceful shutdown)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	case <-quit:
		fmt.Println("shutting down...")
	}

	// 5. Graceful shutdown (30s timeout)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "shutdown error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("server stopped")
}

// ─── Migrate subcommand ───

// runMigrate применяет SQL-миграции из migrations/.
func runMigrate() {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		fmt.Fprintf(os.Stderr, "DB_DSN required\n")
		os.Exit(1)
	}

	conn, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect error: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = conn.Close(context.Background()) }()

	ctx := context.Background()

	if mErr := migrate.Apply(ctx, conn, nil, true); mErr != nil {
		fmt.Fprintf(os.Stderr, "migrations error: %v\n", mErr)
		os.Exit(1)
	}
	fmt.Println("Migrations completed successfully!")
}

// ─── Seed subcommand ───

// runSeed загружает начальные данные (tenant, brand, категории, пользователи, типы, офферы).
func runSeed() {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		fmt.Fprintf(os.Stderr, "DB_DSN required\n")
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
	if mErr := migrate.Apply(ctx, conn, nil, true); mErr != nil {
		fmt.Fprintf(os.Stderr, "migrations error: %v\n", mErr)
		os.Exit(1)
	}
	fmt.Println("  Migrations OK")

	// ─── Tenant ───
	sdekID, err := seedUpsertTenant(ctx, conn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "upsert tenant error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("  Tenant: sdek (id=%s)\n", sdekID)

	// ─── Brand config ───
	if err := seedUpsertBrandConfig(ctx, conn, sdekID); err != nil {
		fmt.Fprintf(os.Stderr, "upsert brand config error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("  Brand config: OK")

	// ─── Категории ───
	if err := seedCategoriesDB(ctx, conn, sdekID); err != nil {
		fmt.Fprintf(os.Stderr, "seed categories error: %v\n", err)
		os.Exit(1)
	}

	// ─── Пользователи ───
	if err := seedUsersDB(ctx, conn, sdekID); err != nil {
		fmt.Fprintf(os.Stderr, "seed users error: %v\n", err)
		os.Exit(1)
	}

	// ─── Типы энгейджментов ───
	if err := seedEngagementTypesDB(ctx, conn, sdekID); err != nil {
		fmt.Fprintf(os.Stderr, "seed engagement types error: %v\n", err)
		os.Exit(1)
	}

	// ─── Офферы ───
	if err := seedEngagementOffersDB(ctx, conn, sdekID); err != nil {
		fmt.Fprintf(os.Stderr, "seed engagement offers error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nSeed completed successfully!")
}

// ─── Tenant ───

func seedUpsertTenant(ctx context.Context, conn *pgx.Conn) (uuid.UUID, error) {
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

func seedUpsertBrandConfig(ctx context.Context, conn *pgx.Conn, tenantID uuid.UUID) error {
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

func seedCategoriesDB(ctx context.Context, conn *pgx.Conn, tenantID uuid.UUID) error {
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

func seedUsersDB(ctx context.Context, conn *pgx.Conn, tenantID uuid.UUID) error {
	users := []seedUser{
		{
			email: "admin@sdek.local", firstName: "Александр", lastName: "Петров",
			phone: "+79001000001", keycloakSub: "kc-admin-001", roles: []string{"admin"},
		},
		{
			email: "hr@sdek.local", firstName: "Елена", lastName: "Иванова",
			phone: "+79001000002", keycloakSub: "kc-hr-001", roles: []string{"hr"},
		},
		{
			email: "catalog@sdek.local", firstName: "Дмитрий", lastName: "Сидоров",
			phone: "+79001000003", keycloakSub: "kc-catalog-001", roles: []string{"catalog_manager"},
		},
		{
			email: "ivanov@sdek.local", firstName: "Иван", lastName: "Петров",
			phone: "+79001000010", keycloakSub: "kc-emp-001", roles: []string{"employee"},
		},
		{
			email: "petrova@sdek.local", firstName: "Мария", lastName: "Петрова",
			phone: "+79001000011", keycloakSub: "kc-emp-002", roles: []string{"employee"},
		},
		{
			email: "sidorov@sdek.local", firstName: "Сергей", lastName: "Сидоров",
			phone: "+79001000012", keycloakSub: "kc-emp-003", roles: []string{"employee"},
		},
		{
			email: "kozlova@sdek.local", firstName: "Анна", lastName: "Козлова",
			phone: "+79001000013", keycloakSub: "kc-emp-004", roles: []string{"employee"},
		},
		{
			email: "novikov@sdek.local", firstName: "Артём", lastName: "Новиков",
			phone: "+79001000014", keycloakSub: "kc-emp-005", roles: []string{"employee"},
		},
	}

	userIDs := make([]uuid.UUID, len(users))
	baseUUID := [16]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x10}
	for i := range users {
		baseUUID[15] = byte(100 + i)
		userIDs[i] = uuid.UUID(baseUUID)
	}

	for i, u := range users {
		var userID uuid.UUID
		err := conn.QueryRow(ctx, `
			INSERT INTO lkfl_platform.users (id, tenant_id, email, first_name, last_name, phone, keycloak_user_id)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (tenant_id, keycloak_user_id) DO UPDATE SET updated_at = NOW()
			RETURNING id
		`, userIDs[i], tenantID, u.email, u.firstName, u.lastName, u.phone, u.keycloakSub).Scan(&userID)
		if err != nil {
			return fmt.Errorf("seed user %s: %w", u.email, err)
		}
		userIDs[i] = userID

		_, err = conn.Exec(ctx, `
			INSERT INTO lkfl_platform.accounts (user_id, total_balance)
			VALUES ($1, $2)
			ON CONFLICT (user_id) DO UPDATE SET updated_at = NOW()
		`, userID, seedBalanceForUser(u))
		if err != nil {
			return fmt.Errorf("seed account for %s: %w", u.email, err)
		}

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

func seedBalanceForUser(u seedUser) int64 {
	switch {
	case seedContains(u.roles, "admin"):
		return 100000
	case seedContains(u.roles, "hr"):
		return 100000
	case seedContains(u.roles, "catalog_manager"):
		return 100000
	default:
		return 50000
	}
}

func seedContains(slice []string, val string) bool {
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
	priceAmount  int
	priceUnit    string
	pricePeriod  string
	priceDisplay string
	providerName string
	iconName     string
	badge        string
	badgeColor   string
}

func seedEngagementTypesDB(ctx context.Context, conn *pgx.Conn, tenantID uuid.UUID) error {
	_, err := conn.Exec(ctx, `
		DELETE FROM lkfl_platform.engagement_types WHERE tenant_id = $1
	`, tenantID)
	if err != nil {
		return fmt.Errorf("cleanup engagement types: %w", err)
	}

	types := []seedEngagementType{
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
		{
			slug: "merch-sdek", name: "Мерч СДЭК",
			description:  "Фирменная одежда и аксессуары СДЭК: худи, футболки, кружки, термосы.",
			categorySlug: "merch", typ: "benefit", status: "active",
			priceAmount: 200, priceUnit: "points", pricePeriod: "once", priceDisplay: "от 200 баллов",
			providerName: "СДЭК Store", iconName: "shopping-bag", badge: "Доступна", badgeColor: "gray",
		},
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

// ─── Офферы ───

type seedOffer struct {
	name        string
	description string
	costCents   int64
	sortOrder   int
}

func seedEngagementOffersDB(ctx context.Context, conn *pgx.Conn, tenantID uuid.UUID) error {
	_, err := conn.Exec(ctx, `
		DELETE FROM lkfl_platform.engagement_offers WHERE tenant_id = $1
	`, tenantID)
	if err != nil {
		return fmt.Errorf("cleanup engagement offers: %w", err)
	}

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
