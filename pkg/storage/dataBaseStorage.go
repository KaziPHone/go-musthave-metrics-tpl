package storage

import (
	"database/sql"

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

	log.Print("Database connected...")
}
