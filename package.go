package main

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "html/template"
    "log"
    "net/http"
    "os"
    "strconv"
    "strings"

    _ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

type Monitor struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Count int    `json:"count"`
}

func main() {
    if len(os.Args) > 1 {
	switch command := strings.ToLower(os.Args[1]); {
	case command == "--help":
	    printHelp()
	case command == "--createdb":
	    if _, err := os.Stat("./monitors.txt"); os.IsNotExist(err) {
		fmt.Println("ERROR! File \"monitors.txt\" does not exist!")
		return
	    }

	    if _, err := os.Stat("./products.db"); err == nil {
		err = os.Remove("./products.db")
		if err != nil {
		    fmt.Println(err)
		    return
		}
	    }

	    CreateDB()
	    AdMonitorsFromFile("./monitors.txt")

	    fmt.Println("OK. File products.db is created!")
	    return
	case command == "--start":
	    http.HandleFunc("/category/monitors", GetMonitors)
	    http.HandleFunc("/category/monitor/", GetStatForMonitor)
	    http.HandleFunc("/category/monitor_click/", AddClickForMonitor)
	    fmt.Println("The server is running!")
	    fmt.Println("Looking forward to requests...")
	    if err := http.ListenAndServe(":8030", nil); err != nil {
		log.Fatal("Failed to start server!", err)
	    }
	default:
	    printHelp()
	}
    } else {
	printHelp()
    }
}

func printHelp() {
    fmt.Println()
    fmt.Println("Help: ./counter --help")
    fmt.Println("Create products database: ./counter --createdb")
    fmt.Println("Start server: ./counter --start")
    fmt.Println()
}

func CreateDB() {
    OpenDB()
    _, err := DB.Exec("create table monitors(id integer, name varchar(255) not null, count integer)")
    if err != nil {
	log.Fatal(err)
	os.Exit(2)
    }
    DB.Close()
}

func OpenDB() {
    db, err := sql.Open("sqlite3", "products.db")
    if err != nil {
	log.Fatal(err)
	os.Exit(1)
    }
    DB = db
}

func AdMonitorsFromFile(filename string) {
    var file *os.File
    var err error
    if file, err = os.Open(filename); err != nil {
	log.Fatal("Failed to open the file: ", err)
	os.Exit(2)
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)

    OpenDB()

    for scanner.Scan() {
	arr := strings.Split(scanner.Text(), ".")
	id, monitorName := arr[0], arr[1]
	_, err = DB.Exec("insert into monitors (id, name, count) values ($1, $2, 0)", id, monitorName)
    }
}

func AddClickForMonitor(w http.ResponseWriter, request *http.Request) {
    err := request.ParseForm()
    if err != nil {
	fmt.Fprintf(w, "{\"error\": \"%s\"}", err)
    } else {
	monitorID := strings.TrimPrefix(request.URL.Path, "/category/monitor_click/")
	OpenDB()
	countValue := 0
	rows, _ := DB.Query("select count from monitors where id=" + monitorID)
	for rows.Next() {
	    rows.Scan(&countValue)
	}
	countValue++
	_, err = DB.Exec("update monitors set count=" + strconv.Itoa(countValue) + " where id=" + monitorID)
	if err != nil {
	    fmt.Fprintf(w, "{\"error\": \"%s\"}", err)
	} else {
	    fmt.Fprintf(w, "{\"success\": true}")
	}
    }
}

func GetMonitors(w http.ResponseWriter, request *http.Request) {
    OpenDB()
    monitors := GetFromDB()
    jsonResponse, err := json.Marshal(map[string]interface{}{"monitors": monitors})
    if err != nil {
	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	return
    }

    w.Header().Set("Content-Type", "application/json")
    w.Write(jsonResponse)
}

func GetFromDB() []Monitor {
    var monitors []Monitor
    rows, _ := DB.Query("select id, name, count from monitors")
    for rows.Next() {
	var monitor Monitor
	rows.Scan(&monitor.ID, &monitor.Name, &monitor.Count)
	monitors = append(monitors, monitor)
    }
    return monitors
}

func GetStatForMonitor(w http.ResponseWriter, request *http.Request) {
    err := request.ParseForm()
    if err != nil {
	fmt.Fprintf(w, "{\"error\": \"%s\"}", err)
    } else {
	monitorID := strings.TrimPrefix(request.URL.Path, "/category/monitor/")
	OpenDB()
	countValue := 0
	rows, _ := DB.Query("select count from monitors where id=" + monitorID)
	for rows.Next() {
	    rows.Scan(&countValue)
	}
	response := map[string]interface{}{
	    "id":    monitorID,
	    "count": countValue,
	}
	jsonResponse, err := json.Marshal(response)
	if err != nil {
	    http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	    return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
    }
}
