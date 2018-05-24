package models

type URL struct {
	Key string `json:"key"`
	URL string `json:"url"`
}

type DB struct {
	Driver string
	Login string
	Password string
	Schema string
	Host string
	Port string
	Database string
}
