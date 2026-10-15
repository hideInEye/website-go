package gorm

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"regexp"
	"strings"
	"time"

	"go-gin/internal/app/config"

	_ "github.com/go-sql-driver/mysql"
)

// initialSchema is the original blog's Prisma migration, without modification.
//
//go:embed schema.sql
var initialSchema string

var createTableName = regexp.MustCompile("^CREATE TABLE `([^`]+)`")

var schemaColumnName = regexp.MustCompile("(?m)^    `([^`]+)`")

var schemaIndexName = regexp.MustCompile("(?m)^    (?:UNIQUE |FULLTEXT )?INDEX `([^`]+)`")

var foreignKeyName = regexp.MustCompile("^ALTER TABLE `([^`]+)` ADD CONSTRAINT `([^`]+)`")

// Migrate creates missing blog tables and foreign keys. Existing records are not
// imported, rewritten, truncated or deleted. A database must already exist.
func Migrate(cfg config.Config) error {
	database, err := sql.Open("mysql", cfg.MySQL.DSN())
	if err != nil {
		return err
	}
	defer database.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	connection, err := database.Conn(ctx)
	if err != nil {
		return err
	}
	defer connection.Close()

	var acquired sql.NullInt64
	err = connection.QueryRowContext(ctx, "SELECT GET_LOCK(CONCAT('sunset-go:', DATABASE()), 30)").Scan(&acquired)
	if err != nil {
		return err
	}
	if !acquired.Valid || acquired.Int64 != 1 {
		return fmt.Errorf("another schema initialization is running")
	}
	defer connection.ExecContext(context.Background(), "SELECT RELEASE_LOCK(CONCAT('sunset-go:', DATABASE()))")

	statements := strings.Split(initialSchema, ";")
	if err := validateExistingSchema(ctx, connection, statements); err != nil {
		return err
	}
	return initializeMissingSchema(ctx, connection, statements)
}

func validateExistingSchema(ctx context.Context, connection *sql.Conn, statements []string) error {
	// Validate every existing table before executing any DDL. In particular this
	// rejects the legacy web_site_server.posts table, whose columns are different.
	for _, raw := range statements {
		statement := strings.TrimSpace(raw)
		match := createTableName.FindStringSubmatch(statement)
		if match == nil {
			continue
		}
		table := match[1]
		var exists int
		err := connection.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM information_schema.TABLES WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?",
			table,
		).Scan(&exists)
		if err != nil {
			return err
		}
		if exists == 0 {
			continue
		}

		for _, column := range schemaColumnName.FindAllStringSubmatch(statement, -1) {
			var found int
			err := connection.QueryRowContext(ctx,
				"SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = ?",
				table, column[1],
			).Scan(&found)
			if err != nil {
				return err
			}
			if found != 1 {
				return fmt.Errorf("table %s is incompatible: missing column %s; no existing data was changed", table, column[1])
			}
		}

		for _, index := range schemaIndexName.FindAllStringSubmatch(statement, -1) {
			var found int
			err := connection.QueryRowContext(ctx,
				"SELECT COUNT(*) FROM information_schema.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND INDEX_NAME = ?",
				table, index[1],
			).Scan(&found)
			if err != nil {
				return err
			}
			if found == 0 {
				return fmt.Errorf("table %s is missing expected index %s; check the original migration before proceeding", table, index[1])
			}
		}
	}
	return nil
}

func initializeMissingSchema(ctx context.Context, connection *sql.Conn, statements []string) error {
	for _, raw := range statements {
		statement := strings.TrimSpace(raw)
		if statement == "" {
			continue
		}
		if createTableName.MatchString(statement) {
			statement = strings.Replace(statement, "CREATE TABLE ", "CREATE TABLE IF NOT EXISTS ", 1)
		} else if match := foreignKeyName.FindStringSubmatch(statement); match != nil {
			var count int
			err := connection.QueryRowContext(ctx,
				"SELECT COUNT(*) FROM information_schema.TABLE_CONSTRAINTS WHERE CONSTRAINT_SCHEMA = DATABASE() AND TABLE_NAME = ? AND CONSTRAINT_NAME = ? AND CONSTRAINT_TYPE = 'FOREIGN KEY'",
				match[1], match[2],
			).Scan(&count)
			if err != nil {
				return err
			}
			if count > 0 {
				continue
			}
		} else {
			return fmt.Errorf("unsupported initialization statement")
		}
		if _, err := connection.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("schema initialization failed (existing rows preserved): %w", err)
		}
	}
	return nil
}
