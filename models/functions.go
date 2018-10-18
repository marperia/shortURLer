package models

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"strings"
)

func (mydb *DB) Connect() *sql.DB {

	connectionParams := mydb.Login + ":" + mydb.Password + "@tcp" +
		"(" + mydb.Host + ":" + mydb.Port + ")/" + mydb.Database

	mydb, err := sql.Open(mydb.Driver, connectionParams)
	if err != nil {
		panic(err)
	}

	err = mydb.Ping()
	if err != nil {
		panic(err)
	}

	return mydb
}

func (db *DB) GetDataByKey(table, field, key string) (*sql.Row) {

	qRow := db.QueryRow("SELECT * FROM `"+table+"` WHERE `"+field+"`= ?", key)

	return qRow
}

func SetData(db *sql.DB, table string, values []string) error {

	qmsArr := make([]string, len(values))
	var qms string
	valuesInterface := []interface{}{}

	if len(values) == 0 {
		return errors.New("no values in setting data")
	}

	for i, v := range values {
		valuesInterface = append(valuesInterface, v)
		qmsArr[i] = "?"
	}

	qms = strings.Join(qmsArr, ",")

	stmt, err := db.Prepare("INSERT INTO `" + table + "` VALUES (" + qms + ")")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(valuesInterface...)
	if err != nil {
		return err
	}

	return nil
}