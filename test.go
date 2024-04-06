/*
2. Security code review
    Часть 1. Security code review: GO
*/ 


package main

import (
    "database/sql"
    "fmt"
    "log"
    "net/http"
    "github.com/go-sql-driver/mysql"
)

var db *sql.DB
var err error

func initDB() {
    db, err = sql.Open("mysql", "user:password@/dbname")
    
    /*
	В строке выше учетные данные в явном виде представлены в коде. 
    Это может привести к утечке учетных данных, если исходный код станет доступен неавторизованным лицам, которые в свою очередь смогут выполнить несанкционированные операции с базой данных, такие как чтение, модификация или удаление данных.
    Хорошей практикой является использование переменных окружения, конфигурайионных файлов, с настроенными правами доступа, или специальных систем, к примеру Kubernetes Secrets.
    В своих проектах я чаще всего использую переменные среды.
    */
    
    if err != nil {
        log.Fatal(err)
    }

err = db.Ping()
if err != nil {
    log.Fatal(err)
    }
}

/*
Не совсем корректная обработка ошибок. Если подключение к базе данных не удается по временным причинам,
 приложение полностью прекратит работу. Вероятным решением является использование нескольких попыток подключения
 (часто используется экспоненциальная задержка между попытками), а также добавление логгирования между попытками
 и логгирование окончательной ошибки и только после этого завершения программы. 
*/

func searchHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != "GET" {
        http.Error(w, "Method is not supported.", http.StatusNotFound)
        return
    }

searchQuery := r.URL.Query().Get("query")
if searchQuery == "" {
    http.Error(w, "Query parameter is missing", http.StatusBadRequest)
    return
}

query := fmt.Sprintf("SELECT * FROM products WHERE name LIKE '%%%s%%'", searchQuery)

/*
Строка выше содержит уязвимость и может привести к SQL-инъекции. Потенциально злоумышленник может 
прочитать, модифицировать или удалить данные в БД, что может привести к компрометации всей системы.
Хорошим решением является использование подготовленных запросов (prepared statements) с плейсхолдерами для параметров.
*/

rows, err := db.Query(query)
if err != nil {
    http.Error(w, "Query failed", http.StatusInternalServerError)
    log.Println(err)
    return
}
defer rows.Close()

var products []string
for rows.Next() {
    var name string
    err := rows.Scan(&name)
    if err != nil {
        log.Fatal(err)
		/*
		Вместо log.Fatal(err) лучше обработать ошибку более гибко, например, 
		отправить пользователю сообщение об ошибке и продолжить выполнение программы
		*/
    }
    products = append(products, name)
}

fmt.Fprintf(w, "Found products: %v\n", products)
/*
Использование fmt.Fprintf для отправки данных клиенту в формате, который включает в себя строковое представление слайса products,
может потенциально привести к уязвимости в виде XSS-атаки. 
Рационально использовать html.EscapeString для экранирования имен продуктов перед отправкой клиенту
*/
}


func main() {
    initDB()
    defer db.Close()

http.HandleFunc("/search", searchHandler)
fmt.Println("Server is running")
log.Fatal(http.ListenAndServe(":8080", nil))
}

// С учетом указанных выше потенциальных уязвимостей и проблем в коде, код можно изменить и представить в таком виде

package main

import (
    "database/sql"
    "fmt"
    "html"
    "log"
    "net/http"
    "os"
    "time"

    _ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func initDB() {
    dbUser := os.Getenv("DB_USER")
    dbPassword := os.Getenv("DB_PASSWORD")
    dbName := os.Getenv("DB_NAME")
    dataSourceName := fmt.Sprintf("%s:%s@/%s", dbUser, dbPassword, dbName)

    var err error
    db, err = sql.Open("mysql", dataSourceName)
    if err != nil {
        log.Fatalf("Error opening database: %v", err)
    }

    for i := 0; i < 5; i++ {
        err = db.Ping()
        if err == nil {
            break
        }
        log.Printf("Error pinging database: %v", err)
        time.Sleep(time.Duration(i) * time.Second)
    }

    if err != nil {
        log.Fatalf("Error connecting to the database after retries: %v", err)
    }
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != "GET" {
        http.Error(w, "Method is not supported.", http.StatusNotFound)
        return
    }

    searchQuery := r.URL.Query().Get("query")
    if searchQuery == "" {
        http.Error(w, "Query parameter is missing", http.StatusBadRequest)
        return
    }

    stmt, err := db.Prepare("SELECT name FROM products WHERE name LIKE ?")
    if err != nil {
        http.Error(w, "Query preparation failed", http.StatusInternalServerError)
        log.Println(err)
        return
    }
    defer stmt.Close()

    rows, err := stmt.Query("%" + searchQuery + "%")
    if err != nil {
        http.Error(w, "Query execution failed", http.StatusInternalServerError)
        log.Println(err)
        return
    }
    defer rows.Close()

    var products []string
    for rows.Next() {
        var name string
        if err := rows.Scan(&name); err != nil {
            http.Error(w, "Failed to scan the row", http.StatusInternalServerError)
            log.Println(err)
            return
        }
        products = append(products, html.EscapeString(name))
    }

    for _, product := range products {
        fmt.Fprintf(w, "Found product: %s\n", product)
    }
}

func main() {
    initDB()
    defer db.Close()

    http.HandleFunc("/search", searchHandler)
    fmt.Println("Server is running on port 8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}