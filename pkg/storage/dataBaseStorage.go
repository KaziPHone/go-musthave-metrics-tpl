package storage

import (
	"context"
	"database/sql"

	models "github.com/KaziPHone/go-musthave-metrics-tpl/internal/model"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rs/zerolog/log"
)

type dataBase struct {
	dataBaseDsn string
	isConnected bool
	db          *sql.DB
	migratePath string
}

func newDataBase(dataBaseDsn, migratePath string) *dataBase {
	return &dataBase{
		dataBaseDsn: dataBaseDsn,
		migratePath: migratePath,
	}
}

func (d *dataBase) IsConnectedDB() bool {
	return d.isConnected
}

func (d *dataBase) initDataBase() {
	if d.dataBaseDsn == "" {
		return
	}

	db, err := sql.Open("pgx", d.dataBaseDsn)
	if err != nil {
		log.Print(err)
		return
	}

	err = db.Ping()
	if err != nil {
		log.Print("Database not connected...")
		return
	}

	d.isConnected = true
	d.db = db
	log.Print("Database connected")

	d.migrateDB()	
}

func (d *dataBase) migrateDB() {

	driver, err := postgres.WithInstance(d.db, &postgres.Config{})
	if err != nil {
		log.Printf("error creating migration driver: %v\n", err)
		return
	}

	d.setMigratePath()
	migrator, err := migrate.NewWithDatabaseInstance(
		d.migratePath,
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

	log.Print("migration success")
}

func (d *dataBase) setMigratePath() {
	if d.migratePath == "" {
		d.migratePath = "file://./migrations"
	}
}


func (d *dataBase) insertMetric(metricName, typeMetric string, val interface{}) {

	var value *float64
	var delta *int64

	if typeMetric == models.Gauge {
		value = val.(*float64)
	} else {
		delta = val.(*int64)
	}

	row := d.db.QueryRowContext(context.Background(),
	"SELECT count(id_metric) as count FROM metrics WHERE id_metric = $1", metricName)

    var id int64
    err := row.Scan(&id)

	if err != nil {
		log.Printf("insert failed: %v\n", err)
		return
	}

	if id == 0 {
		d.insert(metricName, typeMetric, value, delta)
	} else {
		d.update(metricName, value, delta)
	}

}

func (d *dataBase) insert(metricName, typeMetric string, value *float64, delta *int64) error {

	_, err := d.db.Exec("INSERT INTO metrics (id_metric, type_metric, value_metric, delta) VALUES ($1, $2, $3, $4)",
		metricName, typeMetric, value, delta)
	if err != nil {
		log.Printf("insert failed: %v\n", err)
	}
	return err
}


func (d *dataBase) update(metricName string, value *float64, delta *int64) {

	_, err := d.db.Exec("UPDATE metrics SET value_metric = $2, delta = $3 WHERE id_metric = $1",
	metricName, value, delta)
	if err != nil {
		log.Printf("insert failed: %v\n", err)
	}

}