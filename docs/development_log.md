# Пишем простую (но очень расширяемую) сокращалку ссылок на Golang

### Предисловие
Меня позвали на работу в Яндекс. Честно говоря, я этого совсем не ожидал, учитывая полное отсутствие в публичном доступе моих сколь-нибудь прилично написанных проектов. Тем не менее, это произошло, и я собираюсь показать всё (или *почти* всё), на что способен.

Возможно, моему будущему нанимателю или новичку в веб-разработке может понадобится ход моих мыслей во время реализации проекта, так что на всякий случай изложу его тут. К тому же, так проще анализировать ошибки.

### Теоретическая часть

Первым делом стоит определиться с функциональностью: веб-сервер (возможно, на golang) принимает на локалхост запрос вида <http://127.0.0.1:9000/save/?url=> или <http://127.0.0.1:9000/get/?sha2=>. А хэш-таблица (даже не знаю, какая БД тут больше подойдёт) хэширует ссылку и сохраняет её, либо читает хэш и выдаёт ссылку. 

Выбор Golang объясним тем, что он использует минимум необходимых зависимостей (в отличие от Django для Python), компилируемый, быстрый и довольно прост в разработке, и, разумеется, бесконечно мною любим за обработку ошибок, куда же без этого.

А выбор хэш-таблицы — тем, что чтение происходит за O(1). В качестве движка БД я выбрал MySQL InnoDB за относительные простоту настройки, скорость чтения, стабильность, а также поддержку кэширования результатов.
И, что не мало важно, я выбрал её за реляционность: поскольку мы работаем с нормализированными данными, нет никакой необходимости в использовании NoSQL-решений. А их использование может навредить во время смены бизнес-логики проекта.
Однако, использовать такие БД, как Redis, с записью на диск может оказаться оправданным решением. Действительно, при разработке для продакшена, об этом стоит задуматься. Я же не стал вдаваться в более подробный выбор БД за ненадобностью в данной конкретной задаче.

Параметры в URL самого сервиса были добавлены из-за расширяемости: можно легко реализовать на базе этого полноценный веб-сайт, если добавить обработчик home (<http://127.0.0.1:9000/>). Или реализовать новые методы и параметры чтения-записи, не влияя на всю прочую кодовую базу. Или ещё сильнее расширить функциональность без необходимости переписывать весь сервис.

**Пишем обработчик /save/**: через switch-case смотрим передаваемый параметр, обрабатывая ошибку в случае отсутствия параметра в базе (на данный момент у нас будет только url и method для определения метода кодирования ключа).
Для URL мы смотрим на его шаблон и подгоняем под стандарт: если используется не только латиница, то разумнее всего найти готовые решения в интернете, чем изобретать велосипед. Т.е. из ссылки вида «свой-велосипед.рф» должно получится «(http|https)://свой-велосипед.рф», причём в зависимости от:

1. Указанного пользователем протокола и, если не указано, то
2. Поддержки SSL на сайте.

А если шаблон не подходит совсем, то вызвать ошибку.
Не факт, что это поможет от возможных SQLinj (если они вообще предусмотрены условием), но, если нужно, можно реализовать и это. Например, банальным для PHP экранированием/кодированием «запрещённых» символов (/, \, " и т.д.). Или, опять же, посмотреть, как это уже реализовано, ведь в вопросах безопасности не стоит пологаться на собственный опыт, а использовать лучшие практики.

**Обработчик /get/**: точно так же выполняем switch-case, причём тут явно можно и нужно дополнить новыми способами кодирования ссылки (сгенерированный текст, перекодирование в base64 и т.д.).
После получения хэша (или иного идентификатора) он отправляет данные в хэш-таблицу, откуда получает в ответ ссылку.

До кучи это всё потом можно обложить тестами, потому что можем.

Чтож, самое время приступить к...

### Написание кода

**Создадим таблицу**:
```sql
CREATE TABLE `urls` (
`key` VARCHAR(64),
`url` VARCHAR(8192),
PRIMARY KEY (`key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE utf8_bin;
```
Надеюсь, всем хватит длины ссылки в 8,2 тысячи символов.

Здесь также предусмотрено возможная в будущем различие в чтении заглавных и строчных букв: при изменении метода генерации ключа, например, с кодированием в регистрозависимую, но очень компактную кодировку, не нужно будет учить БД отличать «Yandex-Key» от «yandex-key».

**Поднимем веб-сервер**:
```go
package main
 
var (
    port = ":9090"
)
 
import (
    "fmt"
    "net/http"
    "log"
)
 
func main() {
 
    // http.HandleFunc("/", home) // TODO: possible homepage
    http.HandleFunc("/save/", save)
    http.HandleFunc("/get/", get)
 
    log.Println("Web-server is running: http://127.0.0.1" + port)
 
    err := http.ListenAndServe(port, nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
    
}
```
**Обработчик /save/**:
```go
func save(w http.ResponseWriter, r *http.Request) {
    r.ParseForm()
    
    var response string
    var SavingString string
    var linkHash string
    
    SavingString = r.Form["url"] // TODO: add SQLinj-preventing
    
    linkHash = sha256.New()
    hash256.Write([]byte(SavingString))
    linkHash = hex.EncodeToString(hash256.Sum([]byte(nil)))
    
    query := "INSERT INTO `urls` VALUES ('"+ linkHash +"', '"+ SavingString +"')"
    
    db, err := ConnectToDatabase()
    if err != nil {
    	panic(err) // TODO: add better error handling
    }
    
    err = SetData(db, query)
    if err != nil {
    	panic(err) // TODO: add better error handling
    }
    
    response = "http://127.0.0.1" + setting.Port + "/get/?sha2=" + linkHash
    
    fmt.Fprintf(w, response)
}
```

Здесь реализован простейший парсинг GET- или POST-запроса, содержащего слово "url" в тексте, хэширование этой строки, а ещё представлено будущее взаимодействие с БД и написан простейший ответ веб-серверу. Пока довольно «топорно», но вроде работает.

Желательно стоит добавить проверку на то, было ли данное значение уже включено. Однако в таком случае MySQL выдаст ошибку о дублировании индекса, что позволяет избежать самого написания проверки и ещё одного запроса к БД: просто если отлавливать определённую ошибку MySQL, можно всего этого не делать. Но это решение может повлечь за собой проблемы в продакшене, так что лучше оставить это решение прождект-менеджеру.

 **TODO:** Провести тесты производительности, что эффективнее: один запрос и обработка ошибки, или два запроса. Хотя, что-то мне подсказывает...

**Обработчик /get/**:
```go
func get(w http.ResponseWriter, r *http.Request) {
    r.ParseForm()
    
    var response string
    var linkHash string
    var gURL models.URL
    
    linkHash = r.Form["sha2"] // TODO: add SQLinj-preventing
    
    query := "INSERT INTO `urls` VALUES ('"+ linkHash +"', '"+ SavingString +"')"
        
    db, err := ConnectToDatabase()
    if err != nil {
        panic(err) // TODO: add better error handling
    }
    
    data, err = GetDataByKey(db, "key", linkHash) // "key" means "key"-field in database
    if err != nil {
        panic(err) // add better error handling
    }
    data.Scan(&gURL.URL)
    
    response = gURL.URL
    
    fmt.Fprintf(w, response)
}
```
Тут также реализован простейший парсинг запроса, содержащего слово "sha2" и выдача ссылки по хэшу. В будущем мы можем добавить и другие методы 

Пожалуй, довольно разумно будет разнести код на MVC-паттерн и вынести файл с настройками. Теперь наш проект выглядит как-то так:
```
shortURLer/
---controllers/
------controllers.go
---models/
------models.go
---setting/
------setting.go
---tests/
------controllers_test.go — *todo*
------models_test.go — *todo*
------views_test.go — *todo*
---views/
------views.go — *todo*
main.go
readme.md — *todo*
```

Думаю, не стоит объяснять, что функции **save** и **get** вынесены в файл **controllers.go**.

Допишем обработчик **Home** в файле *views.go* и, заодно, добавим папку с шаблонами в корень проекта:
```go
package views
 
import (
    "net/http"
    "html/template"
    "../setting"
)
 
func Home(w http.ResponseWriter, r *http.Request)  {
    t, _ := template.ParseFiles("./templates/index.html")
    data := map[string]string {
        "Name": setting.AppName,
    }
    t.Execute(w, data)
}

```
Тут можно было использовать ```template.Must()```, но это показалось мне излишним. Возможно, это стоит позднее изменить?

Теперь файл **main.go** выглядит как-то так:
```go
package main
 
import (
    "net/http"
    "log"
    "./setting"
    "./controllers" 
    "./views" 
)
 
func main() {
    http.HandleFunc("/", views.Home)
    http.HandleFunc("/save/", controllers.Save)
    http.HandleFunc("/get/", controllers.Get)
 
    log.Println("Web-server is running: http://127.0.0.1" + setting.Port)
    err := http.ListenAndServe(setting.Port, nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}

```

Также допишем единственную модель:
```go
type URL struct {
	Key string `json:"key"`
	URL string `json:"url"`
}
```
и реализуем подключение к базе данных:
```go
func ConnectToDatabase(database ...string) *sql.DB {

	if len(database) == 0 {
		database = append(database, setting.DefaultDatabase)
	}

	connectionParams := setting.DatabaseLogPass + "@" + setting.MysqlProtocol +
		"(" + setting.MysqlHost + ":" + setting.MysqlPort + ")/" +
			database[0]

	db, err := sql.Open(setting.DatabaseDriver, connectionParams)

	if err != nil {
		panic(err)
	}

	return db
}
```
Параметр database сделан опциональным потому, что можно использовать значение из настроек, не указывая каждый раз необходимую базу данных при подключении.
Наверное, это стоит немного подправить...
Добавим ещё одну модель:
```go
type DB struct {
	Login string
	Password string
	Schema string
	Host string
	Port string
	Database string
}
```
Перепишем подключение к БД с учётом только что написанной модели:
```go
func ConnectToDatabase(database models.DB) *sql.DB {

	if database.Schema == "" {
		database = setting.DefaultDB
	}

	connectionParams := database.Login + ":" + database.Password + "@" + database.Schema +
		"(" + database.Host + ":" + database.Port + ")/" + database.Database

	db, err := sql.Open(database.Driver, connectionParams)

	if err != nil {
		panic(err)
	}

	return db
}
```
Как вы уже догадались, стандартный экземпляр структуры БД пока хранится в настройках, что идёт совсем не на пользу производительности: ведь настройки загружаются гораздо чаще, чем мы подключаемся к БД, и это занимает дополнительное место в памяти. Позднее, возможно, стоит реализовать отдельный ОРМ для всего этого, но пока не будем усложнять логику.

Допишем ещё два обработчика: один должен будет сохранять запись в БД, а другой — получать данные по ключу.
**Обработчик SetData():**
```go
func SetData(db *sql.DB, q string) error {
	if q == "" {
		return errors.New("no query in function")
	}
	insert, err := db.Query(q)
	if err != nil {
		return err
	}
	defer insert.Close()
	return nil
}
```
**Обработчик GetDataByKey():**
```go
func GetDataByKey(db *sql.DB, filed_name string, key string) (*sql.Row) {

	defer db.Close()

	qRow := db.QueryRow("SELECT `url` FROM `urls` WHERE `"+ filed_name +"`=?", key)

	return qRow
}
```
Почему я отдаю указатель на строку БД, а не саму строку? Это, опять же, поможет в реализации ОРМ, если он будет: нужно будет просто сканировать (.Scan()) строку всеми параметрами модели, что будет невозможно, если мы вернём строку.

Что ж, у нас уже есть функции, которые подключаются к БД, получают оттуда информацию по хэшу и делают новые записи, а также более-менее рабочий прототип домашней страницы. Самое время для...

### Рефакторинг
Большинство начинает рефакторинг с исправления типичных ошибок и написания новых функций. Однако такой подход, как правило, просто уничтожает уже работающий код и создаёт новые баги.

Я очень хотел бы сказать, что мы пойдём другим путём и сперва напишем unit-тесты, которые позволят определить, движемся ли мы в правильном направлении, но нет. Мы пойдём тем же неблагодарным путём, что и остальные разработчики. Ведь довольно большая часть функционала, не считая банальной обработки ошибок или защиты от SQL-инъекций, просто не реализована. Но, чтобы вы не волновались, попутно с реализацией мы будем писать тесты.

**Так выглядит структура нашего проекта сейчас:**
```
shortURLer/
---controllers/
------controllers.go
---docs/
------docs.md — *todo*
---models/
------models.go
---setting/
------setting.go
---tests/
------controllers_test.go — *todo*
------models_test.go — *todo*
------views_test.go — *todo*
---views/
------views.go
---development_log.md
---main.go
---readme.md — *todo*
```
У нас есть две структуры: база данных и ссылка — модель-шаблон для записей базы данных. Это статические структуры, которые отлично справляются со своей задачей и не будут рефакториться.
У нас также есть обработчик домашней страницы, который, по идее, ничем полезным не занимается и прекрасно работает, и я также не вижу ни одной причины, почему он может перестать, а значит тоже не нуждается в рефакторинге.
А ещё у нас есть файл с настройками. Там почему-то лежит структура базы данных с установленными по умолчанию данными из того же файла с настройками. И без того, чтобы не создать ещё один файл, где будут храниться все методы и структуры базы данных, тут не обойтись, а значит рефакторинг подождёт.
Но вишенка на торте — у нас есть ужасно построенная бизнес-логика. И вот тут всё только начинается...
Если такие функции, как **ConnectToDatabase()**, **SetData()** или **GetDataByKey()** были достаточно простыми, чтобы (почти) не нуждаться в рефакторинге, то с другими дело обстоит гораздо печальней.

Всё, что можно отрефакторить в этих трёх функциях, это добавление проверки при подключении к БД по всем пустым строкам в экземпляре базы данных на случай, если пользователь хочет лишь немного изменить параметры подключения, изменив только логин и пароль, или только хост базы данных:
```go
func ConnectToDatabase(database models.DB) *sql.DB {
 
	if database.Login == "" {
		database.Login = setting.DefaultDB.Login
	}
	if database.Password == "" {
		database.Password = setting.DefaultDB.Password
	}
	if database.Schema == "" {
		database.Schema = setting.DefaultDB.Schema
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
 
	connectionParams := database.Login + ":" + database.Password + "@" + database.Schema +
		"(" + database.Host + ":" + database.Port + ")/" + database.Database
 
	db, err := sql.Open(database.Driver, connectionParams)
 
	if err != nil {
		panic(err)
	}
 
	return db
}
```
Однако функции **SaveURL()** и **GetURL()** должны быть серьёзно исправлены. Добавим функцию, которая обрабатывает ошибки, которые нужно выводить пользователю API:
```go
func UserErrorHandle(errorText string, errorCode int, nw http.ResponseWriter) {
	err := errors.New(errorText)
	log.Println(err)
	nw.WriteHeader(errorCode)
	nw.Write([]byte(err.Error()))
}
```
На вход он принимает текст и код ошибки, а также текущий сеанс пользователя. Был ещё соблазн реализовать здесь ожидаемые и полученные параметры, но мне показалось это излишним: в конце-концов, всегда можно отформатировать строку так, чтобы она их уже содержала.
А теперь добавим функцию, которая проверяет, валиден ли sha256-хэш, который отправляет нам пользователь:
```go
func Sha256Checker(hash string) error {
	lowerHash := strings.ToLower(hash)
	re := regexp.MustCompile("[0-9a-f]{64}")
	if rhash := re.FindString(lowerHash); rhash == "" {
		return errors.New("invalid sha256 hash format")
	}
	return nil
}
```
Таким образом, мы уже можем легко и непринуждённо переписать функцию **GetURL()**:
```go
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
			} else {
				CheckPassed = true
			}
		case "custom_method":
			fmt.Println("Custom method handler.")
			CheckPassed = true
		default:
			UserErrorHandle("hashing method has not set", 400, w)
			return
	}


	if CheckPassed == true {
		data = GetDataByKey(ConnectToDatabase(setting.DefaultDB), "key", GettingHash)
		data.Scan(&gURL.URL)
	}

	if gURL.URL == "" {
		UserErrorHandle("no links has been found", 404, w)
		return
	}

	response = gURL.URL
	w.Write([]byte(response))
	return
}
```
Теперь у нас есть возможность в несколько действий добавить новые обработчики, если понадобится. Также у нас теперь есть гораздо более удобное разделение ошибок: на **фатальные**, которые зависят от сервера и не позволяют приложению больше работать (отсутствие подключения к БД или падение сервера), и **пользовательские**, которые зависят от параметров API и могут быть исправлены как пользователями API, так и разработчиками, если включено полное логирование (опустим этот этап, пожалуй).

Прежде, чем приступить к функции, которая сохраняет ссылки, нужно эти самые ссылки привести к одному общему виду, чтобы избежать дублей в БД. Самое очевидное, на первый взгляд — хранить их в lowercase с указанием протокола. Также, возможно, стоит подумать о параметрах ссылок: ведь не важно, в каком порядке стоят параметры, а их чтение и сортировка также могут помочь избежать дублей. И, разумеется, не-латинские символы в адресах сайтов тоже возможны: с этим тоже нужно что-то сделать, но я, с вашего позволения, не буду. Задача и так выглядит слишком раздутой. Ограничимся простым сравнением с регулярным выражением, добавлением протокола и привением к lowercase:
```go
func StandartizeURL(rawLink string) (string, error) {

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
```
Однако важно заметить, что при такой стандартизации ссылки с https:// могут дублировать ссылки с http://, что, с одной стороны, хорошо, потому что разным пользователям нужны разные протоколы, но с другой, возможно, стоит добавить этот параметр в функцию **GetURL()**, чтобы, например, по умолчанию выставлять https://, но, если пользователь указал, то выставлять http://.

К тому же, ссылки можно хранить и вовсе без протоколов, просто приписывая им его при выдаче пользователю.

Кстати, я терпеть не могу регулярные выражения, поэтому я взял готовое со stackoverflow: <https://stackoverflow.com/questions/3809401/what-is-a-good-regular-expression-to-match-a-url>
За результат не отвечаю, но это лучшее, что я могу на данный момент.


А ещё, будь это продакшн, тут было бы не обойтись без тестов... Наверное, их действительно стоит написать, но я заранее знаю, какой будет результат.


Приступим к рефакторингу функции **SaveURL()**:
```go
func SaveURL(w http.ResponseWriter, r *http.Request) {

	var HashingMethod string
	var gURL models.URL
	var mURL models.URL
	var response string

	r.ParseForm()

	if val := r.Form.Get("url"); val != "" {
 
		val, err := StandardizeURL(val)
		if err != nil {
			UserErrorHandle(err.Error(), 400, w)
			return
		}
		gURL.URL = template.HTMLEscapeString(val)
 
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
	GetDataByKey(DBConn, "key", gURL.Key).Scan(&mURL.Key, &mURL.URL) // check if we already have this link

	if mURL.Key != gURL.Key { // and if dont
		DBConn := ConnectToDatabase(setting.DefaultDB)
		values := []string{template.HTMLEscapeString(gURL.Key), template.HTMLEscapeString(gURL.URL)}
		i := SetData(DBConn, "urls", values)

		if i != nil {
			UserErrorHandle("mysql error", 400, w)
			log.Println(i.Error())
			return
		}
	}

	response = "http://127.0.0.1" + setting.Port + "/get?" + HashingMethod + "=" + gURL.Key
	w.Write([]byte(response))
	return
}
```

Мы добавили простейшую защиту от SQL-инъекций и сделали код более понятным. Думаю, всё ясно без пояснений.

### Окончательный вариант

```
shortURLer/
---controllers/
------controllers.go
---docs/
------docs.md
------development_log.md
---models/
------models.go
---setting/
------setting.go
---tests/
------controllers_test.go — *maybe next time*
---views/
------views.go
---main.go
---readme.md
```

У нас имеются две модели: базы данных и ссылки, которую мы сохраняем.
Только один вид: домашняя страница.
И целая куча бизнес-логики, в которую включено взаимодействие с БД: отображение ошибок пользователю, проверка валидности хэша (для экономии времени на запрос к БД), *корявый* стандартизатор ссылок, а также точки взаимодействия с сервисом (сохранение и получение ссылки). Кстати, они одинаково хорошо обрабатывают GET- и POST-запросы.
В настройках лежат данные приложения, параметры функций создания хэшей и данные о подключении к БД.

**Не** было сделано:
1. Из бизнес-логики и настроек не вынесена часть по взаимодействию с БД;
2. Не дописаны тесты; (чуть позже, возможно)
3. Не написана документация;

Хотелось бы услышать конструктивную критику.

Код: <https://github.com/marperia/shortURLer>

**UPD 25.05.18:**
1. Исправлен баг с выводом незакодированных HTML-символов (& и пр.)
2. Было переписано взаимодействие с БД;
3. В настройки добавлен параметр **Host**;
4. Код переписан более «чисто»;
5. Прочие мелкие исправления;

**UPD 14.10.18:**

I would like to apologise to everyone who had to read that terrible code. I used to be small and stupid, but now I'm just foolish.