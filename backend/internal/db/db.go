package db

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"atchannel-backend/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Connect opens the database, retrying until it accepts connections.
//
// Containers start in parallel: postgres needs a few seconds before it
// answers, and a single failed attempt would otherwise crash the process
// into a restart loop.
func Connect() {
	host := getEnv("DB_HOST", "localhost")
	user := getEnv("DB_USER", "atchannel_user")
	password := getEnv("DB_PASSWORD", "atchannel_password")
	dbname := getEnv("DB_NAME", "atchannel_db")
	port := getEnv("DB_PORT", "5432")

	dsn := "host=" + host +
		" user=" + user +
		" password=" + password +
		" dbname=" + dbname +
		" port=" + port +
		" sslmode=disable TimeZone=UTC"

	const maxAttempts = 30

	var err error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Warn),
		})
		if err == nil {
			sqlDB, dbErr := DB.DB()
			if dbErr != nil {
				err = dbErr
			} else if pingErr := sqlDB.Ping(); pingErr != nil {
				err = pingErr
			} else {
				// Keep a small pool: the API is read/write heavy but each
				// query is short, so a handful of connections is plenty.
				sqlDB.SetMaxOpenConns(20)
				sqlDB.SetMaxIdleConns(5)
				sqlDB.SetConnMaxLifetime(time.Hour)

				log.Printf("database connected (host=%s db=%s)", host, dbname)
				return
			}
		}

		if attempt == maxAttempts {
			log.Fatalf("could not connect to the database after %d attempts: %v", maxAttempts, err)
		}

		log.Printf("waiting for database (%d/%d): %v", attempt, maxAttempts, err)
		time.Sleep(time.Second)
	}
}

// Migrate brings the schema up to date.
func Migrate() {
	// The account table was renamed along with its model; this has to run
	// first, because everything below looks at the new name.
	renameTable("channelers", "users")

	// AutoMigrate can only create a NOT NULL column on an empty table, so
	// every column that has to be backfilled is added here first. See
	// ensureColumn for why this is a separate pass.
	ensureColumn("refresh_tokens", "family_id", "text",
		// Pre-existing sessions predate token families. Giving each its own
		// family keeps the row usable (it still logs out normally) while
		// admitting that no reuse detection can be applied retroactively.
		"md5(random()::text || id::text)")

	for _, table := range []string{"users", "channels", "posts", "comments"} {
		// COALESCE because created_at is itself nullable on tables created
		// before it was required: a NULL backfill would leave updated_at
		// violating the constraint it is being prepared for.
		//
		// refresh_tokens is deliberately absent: its model has no UpdatedAt
		// field, so a column there would be NOT NULL with nothing populating
		// it and every insert would fail.
		ensureColumn(table, "updated_at", "timestamp with time zone", "COALESCE(created_at, now())")
	}

	ensureSearchIndexes()

	err := DB.AutoMigrate(
		&models.User{},
		&models.Channel{},
		&models.Post{},
		&models.Comment{},
		&models.RefreshToken{},
	)

	if err != nil {
		log.Fatalf("migration error: %v", err)
	}

	log.Print("database migration completed")
}

// renameTable moves a table and its indexes to a new name, and only when the
// old one exists and the new one does not, so it is safe to run on every
// boot: a fresh install has neither and does nothing, an already-renamed
// database has the target and does nothing.
//
// Postgres does not rename indexes along with the table, so idx_ names are
// rewritten explicitly — otherwise the old name lingers and AutoMigrate adds
// a second index because it goes looking for the new one.
func renameTable(from, to string) {
	if !tableExists(from) || tableExists(to) {
		return
	}

	log.Printf("migrating: renaming table %s -> %s", from, to)

	if err := DB.Exec(fmt.Sprintf("ALTER TABLE %s RENAME TO %s", from, to)).Error; err != nil {
		log.Fatalf("migration error (rename %s -> %s): %v", from, to, err)
	}

	var names []string
	if err := DB.Raw(
		`SELECT indexname FROM pg_indexes WHERE schemaname = current_schema() AND tablename = ?`,
		to).Scan(&names).Error; err != nil {
		log.Fatalf("migration error (list indexes on %s): %v", to, err)
	}

	for _, name := range names {
		if !strings.Contains(name, from) {
			continue
		}
		renamed := strings.Replace(name, from, to, 1)
		stmt := fmt.Sprintf("ALTER INDEX %s RENAME TO %s", name, renamed)
		if err := DB.Exec(stmt).Error; err != nil {
			log.Fatalf("migration error (rename index %s): %v", name, err)
		}
	}
}

// ensureColumn guarantees a column exists, has no NULLs, and is NOT NULL —
// the state AutoMigrate cannot reach on its own.
//
// Postgres rejects "ADD COLUMN ... NOT NULL" outright when the table already
// has rows, which is exactly what happens the first time a model gains a
// required field. GORM gets as far as creating the column nullable and then
// stops, so this has to be able to finish a half-applied migration as well
// as start a fresh one: backfill and the NOT NULL constraint run every time,
// not only when the column was missing.
//
// A missing table is ignored, because on a fresh install AutoMigrate creates
// the table with the column already in place.
func ensureColumn(table, column, definition, backfill string) {
	if !tableExists(table) {
		return
	}

	if !columnExists(table, column) {
		log.Printf("migrating: adding %s.%s", table, column)

		add := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, column, definition)
		if err := DB.Exec(add).Error; err != nil {
			log.Fatalf("migration error (add %s.%s): %v", table, column, err)
		}
	}

	// Existing rows: give them a value before the constraint lands.
	update := fmt.Sprintf("UPDATE %s SET %s = %s WHERE %s IS NULL", table, column, backfill, column)
	if err := DB.Exec(update).Error; err != nil {
		log.Fatalf("migration error (backfill %s.%s): %v", table, column, err)
	}

	if columnNullable(table, column) {
		alter := fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s SET NOT NULL", table, column)
		if err := DB.Exec(alter).Error; err != nil {
			log.Fatalf("migration error (not null %s.%s): %v", table, column, err)
		}
		log.Printf("migrating: %s.%s set NOT NULL", table, column)
	}
}

// ensureSearchIndexes creates the trigram indexes that make substring search
// usable. Every listing/search endpoint filters with ILIKE '%term%', which a
// plain btree cannot serve — without these, each keystroke is a sequential
// scan over the whole table.
//
// pg_trgm is shipped with Postgres; CREATE EXTENSION needs a role allowed to
// create it, which the compose superuser is.
func ensureSearchIndexes() {
	if err := DB.Exec("CREATE EXTENSION IF NOT EXISTS pg_trgm").Error; err != nil {
		log.Printf("warning: could not enable pg_trgm, search will seq-scan: %v", err)
		return
	}

	indexes := []struct {
		name   string
		table  string
		column string
	}{
		{"idx_channels_name_trgm", "channels", "name"},
		{"idx_channels_title_trgm", "channels", "title"},
		{"idx_channels_description_trgm", "channels", "description"},
		{"idx_posts_title_trgm", "posts", "title"},
		{"idx_posts_content_trgm", "posts", "content"},
	}

	for _, idx := range indexes {
		if !tableExists(idx.table) {
			continue
		}

		stmt := fmt.Sprintf(
			"CREATE INDEX IF NOT EXISTS %s ON %s USING gin (%s gin_trgm_ops)",
			idx.name, idx.table, idx.column,
		)
		if err := DB.Exec(stmt).Error; err != nil {
			log.Printf("warning: could not create search index %s: %v", idx.name, err)
		}
	}
}

func tableExists(table string) bool {
	return exists(
		`SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = current_schema() AND table_name = ?)`,
		table)
}

func columnExists(table, column string) bool {
	return exists(
		`SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = current_schema() AND table_name = ? AND column_name = ?)`,
		table, column)
}

// columnNullable reports whether the column still accepts NULL.
func columnNullable(table, column string) bool {
	var nullable string
	err := DB.Raw(
		`SELECT is_nullable FROM information_schema.columns WHERE table_schema = current_schema() AND table_name = ? AND column_name = ?`,
		table, column).Scan(&nullable).Error
	if err != nil {
		log.Fatalf("migration error (schema probe): %v", err)
	}
	return nullable == "YES"
}

func exists(query string, args ...any) bool {
	var found bool
	if err := DB.Raw(query, args...).Scan(&found).Error; err != nil {
		log.Fatalf("migration error (schema probe): %v", err)
	}
	return found
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
