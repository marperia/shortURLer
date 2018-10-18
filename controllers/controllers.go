package controllers

import (
	"database/sql"
	"github.com/marperia/shortURLer/models"
	"github.com/marperia/shortURLer/setting"
	"math"

	"crypto/sha256"
	"encoding/hex"
	"html"
	"net/http"
	"regexp"
	"strings"

	"errors"
	"log"
)

const Base64Alphabet = "123456789abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ"
const Base16Alphabet = "0123456789abcdef"

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

func TraceURL(link string) (string, error) {

	resp, err := http.Get(link)
	if err != nil {
		return "", err
	}

	finalURL := resp.Request.URL.String()

	return finalURL, nil

}

func BaseEncode(id int, alph string) (string) {

	var Encoded string
	var modInt int
	alphArray := strings.Split(alph, "")
	baseCount := len(alph)

	for (id >= baseCount) {
		modInt = id % baseCount
		id = id / baseCount
		Encoded = alphArray[modInt] + Encoded
	}

	modInt = id % baseCount
	id = id / baseCount
	Encoded = alphArray[modInt] + Encoded

	return Encoded

}

func BaseDecode(encodedStr string, alph string) (int) {

	strLen := float64(len(encodedStr) - 1)
	baseCount := float64(len(alph))
	var decodedStr int
	var charInd int

	for _, letter := range encodedStr {
		charInd = strings.Index(alph, string(letter))
		// custom power function
		//decodedStr = decodedStr + Power(baseCount, strLen)*charInd
		decodedStr = decodedStr + int(math.Pow(baseCount, strLen))*charInd
		strLen--
	}

	return decodedStr
}

func (u *models.URL) Save(w http.ResponseWriter, r *http.Request) {

	var HashingMethod string
	var gURL models.URL
	var myURL models.URL
	var response string

	r.ParseForm()

	if val := r.Form.Get("url"); val != "" {
		val, err := StandardizeURL(val)
		if err != nil {
			UserErrorHandle(err.Error(), 400, w)
			return
		}
		gURL.URL = html.EscapeString(val)
	} else if val := r.Form.Get("custom_param"); val != "" {
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
	} else if HashingMethod == "base58" {

	} else {
		UserErrorHandle("unknown hashing method", 400, w)
		return
	}

	DBConn := setting.DefaultDB.Connect()
	setting.DefaultDB.GetDataByKey("urls", "key", gURL.Key).Scan(&myURL.Key, &myURL.URL)

	if myURL.Key != gURL.Key {
		values := []string{html.EscapeString(gURL.Key), html.EscapeString(gURL.URL)}
		err := models.SetData(DBConn, "urls", values)

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
		DBConn := setting.DefaultDB.Connect()
		data = setting.DefaultDB.GetDataByKey("urls", "key", GettingHash)
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
