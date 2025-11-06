package storage

import (
	"database/sql"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rs/zerolog/log"
)

type dataBase struct {
	dataBaseDsn string
	isConnected bool
	db          *sql.DB
}

func (m *MStorage) IsConnectedDB() bool {
	return m.dataBase.isConnected
}

func (m *MStorage) initDataBase() {
	db, err := sql.Open("pgx", m.dataBase.dataBaseDsn)
	if err != nil {
		log.Print(err)
		return
	}

	err = db.Ping()
	if err != nil {
		log.Print("Database not connected...")
		return
	}

	m.dataBase.isConnected = true
	m.dataBase.db = db

	m.migrateDB()

	log.Print("Database connected...")
}

func (m *MStorage) migrateDB() {

	driver, err := postgres.WithInstance(m.dataBase.db, &postgres.Config{})
	if err != nil {
		log.Printf("error creating migration driver: %v\n", err)
		return
	}

	migrator, err := migrate.NewWithDatabaseInstance(
		"file://./migrations",
		"postgres",
		driver,
	)

	if err != nil {
		log.Printf("failed to create migrator instance: %v\n", err)
		return
	}

	err = migrator.Up()
	if err != nil && err != migrate.ErrNoChange {
		log.Printf("migration failed: %v\n", err)
		return
	}
}