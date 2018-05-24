package setting

import "../models"

var (
	// server setting section
	Port = ":9090"


	// app setting section
	AppName = "ShortURLer"


	HashingMethods = []string{"sha256", "custom_method"}
	DefaultHashingMethod = "sha256"


	// database setting section
	DatabaseDriver = "mysql"
	DatabaseLogin = "root"
	DatabasePassword = "12345trewq"
	DatabaseSchema = "tcp"
	DatabaseHost = "127.0.0.1"
	DatabasePort = "3306"

	DefaultDatabase = "url_shorter"
)

var DefaultDB = models.DB {
	Driver: DatabaseDriver,
	Login: DatabaseLogin,
	Password: DatabasePassword,
	Schema: DatabaseSchema,
	Host: DatabaseHost,
	Port: DatabasePort,
	Database: DefaultDatabase,
}