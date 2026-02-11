package repository

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	mysqldb "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(config MySQLConfig, migrationsPath string) error {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?multiStatements=true",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.DBName,
	)

	db, err := sql.Open("mysql", dsn)

	if err != nil {
		return err
	}

	defer db.Close()

	driver, err := mysqldb.WithInstance(db, &mysqldb.Config{})

	if err != nil {
		return err
	}

	migrationsURL := "file://" + migrationsPath

	m, err := migrate.NewWithDatabaseInstance(migrationsURL, "mysql", driver)

	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}
