package models

import "database/sql"

type URL struct {
	Key string `json:"key"`
	URL string `json:"url"`
}

type DB struct {
	sql.DB
	Driver string
	Login string
	Password string
	Host string
	Port string
	Database string
}
