package controllers


import (
	"net/http"
	"encoding/hex"
	"crypto/sha256"
	"../setting"
	"../models"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"errors"
	"log"
	"regexp"
	"strings"
	"html"
)


func UserErrorHandle(errorText string, errorCode int, nw http.ResponseWriter) {
	err := errors.New(errorText)
	log.Println("user error:", err)
	nw.WriteHeader(errorCode)
	nw.Write([]byte(err.Error()))
}


func Sha256Checker(hash string) error {
	lowerHash := strings.ToLower(hash)
	re := regexp.MustCompile("[0-9a-f]{64}")
	if rhash := re.FindString(lowerHash); rhash == "" {
		return errors.New("invalid sha256 hash format")
	}
	return nil
}


func StandardizeURL(rawLink string) (string, error) {

	SchemeExp := "https?:\\/\\/(www\\.)?[-a-zA-Z0-9@:%._\\+~#=]{2,256}\\.[a-z]{2,6}\\b([-a-zA-Z0-9@:%_\\+.~#?&//=]*)"
	NoSchemeExp := "[-a-zA-Z0-9@:%._\\+~#=]{2,256}\\.[a-z]{2,6}\\b([-a-zA-Z0-9@:%_\\+.~#?&//=]*)"
	se := regexp.MustCompile(SchemeExp)
	nse := regexp.MustCompile(NoSchemeExp)

	if SchemeURL := se.FindString(rawLink); SchemeURL != "" {
		return strings.ToLower(SchemeURL), nil
	} else if NoSchemeURL := nse.FindString(rawLink); NoSchemeURL != "" {
		return "http://" + strings.ToLower(NoSchemeURL), nil // TODO: check for scheme online
	}

	return "", errors.New("not a valid url")
}

/*
 *
 * Section below should be replaced to an ORM
 *
 */


func ConnectToDatabase(database models.DB) *sql.DB {

	if database.Driver == "" {
		database.Driver = setting.DefaultDB.Driver
	}
	if database.Login == "" {
		database.Login = setting.DefaultDB.Login
	}
	if database.Password == "" {
		database.Password = setting.DefaultDB.Password
	}
	if database.Host == "" {
		database.Host = setting.DefaultDB.Host
	}
	if database.Port == "" {
		database.Port = setting.DefaultDB.Port
	}
	if database.Database == "" {
		database.Database = setting.DefaultDB.Database
	}

	connectionParams := database.Login + ":" + database.Password + "@tcp" +
		"(" + database.Host + ":" + database.Port + ")/" + database.Database

	db, err := sql.Open(database.Driver, connectionParams)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	return db
}


func GetDataByKey(db *sql.DB, table string, fieldName string, key string) (*sql.Row) {

	qRow := db.QueryRow("SELECT * FROM `" + table + "` WHERE `"+ fieldName +"`=?", key)

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

/*
 *
 * Section above should be replaced to an ORM
 *
 */

func SaveURL(w http.ResponseWriter, r *http.Request) {

	var HashingMethod string
	var gURL models.URL
	var urlKey string
	var response string

	r.ParseForm()

	if val := r.Form.Get("url"); val != "" {
		val, err := StandardizeURL(val)
		if err != nil {
			UserErrorHandle(err.Error(), 400, w)
			return
		}
		gURL.URL = html.EscapeString(val)
	} else 	if val := r.Form.Get("custom_param"); val != "" {
		log.Println("Custom method handler.")
	} else {
		UserErrorHandle("empty url value", 400, w)
		return
	}

	if val := r.Form.Get("method"); val != "" {
		HashingMethod = r.Form.Get("method")
	} else {
		HashingMethod = setting.DefaultHashingMethod
	}


	if HashingMethod == "sha256" {
		hash256 := sha256.New()
		hash256.Write([]byte(gURL.URL))
		gURL.Key = hex.EncodeToString(hash256.Sum([]byte(nil)))
	} else {
		UserErrorHandle("unknown hashing method", 400, w)
		return
	}

	DBConn := ConnectToDatabase(setting.DefaultDB)
	GetDataByKey(DBConn, "urls", "key", gURL.Key).Scan(&urlKey)

	if urlKey != gURL.Key {
		values := []string{html.EscapeString(gURL.Key), html.EscapeString(gURL.URL)}
		err := SetData(DBConn, "urls", values)

		DBConn.Close()

		if err != nil {
			UserErrorHandle("mysql error", 400, w)
			log.Println(err.Error())
			return
		}
	}

	response = "http://" + setting.Host + setting.Port + "/get?" + HashingMethod + "=" + gURL.Key
	w.Write([]byte(response))
	return
}


func GetURL(w http.ResponseWriter, r *http.Request) {

	var gURL models.URL
	var data *sql.Row
	var response string
	var GettingHash string
	var HashingMethod string
	var CheckPassed bool

	r.ParseForm()
	for _, i := range setting.HashingMethods {
		value := r.Form.Get(i)
		if value != "" {
			HashingMethod = i
			GettingHash = value
			break
		}
	}

	switch HashingMethod {
		case "sha256":
			err := Sha256Checker(GettingHash)
			if err != nil {
				UserErrorHandle(err.Error(), 400, w)
				return
			}
			CheckPassed = true
		case "custom_method":
			log.Println("Custom method handler.")
			CheckPassed = true
		default:
			UserErrorHandle("hashing method has not set", 400, w)
			return
	}


	if CheckPassed == true {
		DBConn := ConnectToDatabase(setting.DefaultDB)
		data = GetDataByKey(DBConn, "urls", "key", GettingHash)
		DBConn.Close()
		data.Scan(&gURL.Key, &gURL.URL)
	}

	if gURL.URL == "" {
		UserErrorHandle("no links has been found", 404, w)
		return
	}

	response = html.UnescapeString(gURL.URL)
	w.Write([]byte(response))
	return
}
