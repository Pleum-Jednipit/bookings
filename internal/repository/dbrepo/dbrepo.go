package dbrepo

import (
	"database/sql"

	"github.com/Pleum-Jednipit/bookings/internal/config"
	"github.com/Pleum-Jednipit/bookings/internal/repository"
)

type postgresDBRepo struct {
	App *config.AppConfig
	DB  *sql.DB
}

func NewPostgreaRepo(conn *sql.DB, a *config.AppConfig) repository.DatabaseRepo {
	return &postgresDBRepo{
		App: a,
		DB:  conn,
	}
}