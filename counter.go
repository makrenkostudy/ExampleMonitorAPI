package main

import (
    "bufio"
    "database/sql"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    "strconv"
    "strings"

    _ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

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
            AddMonitorsFromFile("./monitors.txt")
            fmt.Println("OK. File products.db is created!")
            return
        case command == "--start":
            http.HandleFunc("/", ServeHTML)           // Add this line
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

func ServeHTML(w http.ResponseWriter, r *http.Request) {
    http.ServeFile(w, r, "index.html") 
}

// Remaining functions remain unchanged...

func printHelp() {
    fmt.Println()
    fmt.Println("Help:                   ./counter --help")
    fmt.Println("Create products database:  ./counter --createdb")
    fmt.Println("Start server:              ./counter --start")
    fmt.Println()
}

func CreateDB() {
    OpenDB()
    _, err := DB.Exec(`CREATE TABLE monitors (
        id INTEGER PRIMARY KEY, 
        name VARCHAR(255) NOT NULL, 
        count INTEGER
    )`)
    if err != nil {
        log.Fatal(err)
        os.Exit(2)
    }
    DB.Close()
}

func OpenDB() {
    var err error
    DB, err = sql.Open("sqlite3", "products.db")
    if err != nil {
        log.Fatal(err)
        os.Exit(1)
    }
}

func AddMonitorsFromFile(filename string) {
    var file *os.File
    var err error
    file, err = os.Open(filename)
    if err != nil {
        log.Fatal("Failed to open the file: ", err)
        os.Exit(2)
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    OpenDB()
    for scanner.Scan() {
        arr := strings.Split(scanner.Text(), ",")
        id := arr[0]
        monitorName := strings.TrimSpace(arr[1]) // Trim spaces and quotes
        monitorName = strings.Trim(monitorName, "\"") // Remove extra quotes if present
        _, err = DB.Exec("insert into monitors(id, name, count) values ($1, $2, 0)", id, monitorName)
    }
}

func AddClickForMonitor(w http.ResponseWriter, request *http.Request) {
    err := request.ParseForm()
    if err != nil {
        fmt.Fprintf(w, "{%s}", err)
    } else {
        monitorId := strings.TrimPrefix(request.URL.Path, "/category/monitor_click/")
        OpenDB()
        countValue := 0
        rows, _ := DB.Query("SELECT count FROM monitors WHERE id=" + monitorId)
        for rows.Next() {
            rows.Scan(&countValue)
        }
        countValue++
        _, err = DB.Exec("UPDATE monitors SET count = " + strconv.Itoa(countValue) + " WHERE id = " + monitorId)
    }
}

func GetMonitors(w http.ResponseWriter, request *http.Request) {
    OpenDB()
    monitors := GetMonitorsList() 
    err := request.ParseForm()
    if err != nil {
        fmt.Fprintf(w, "{%s}", err)
        return
    }

    monitorArray := make([][]interface{}, len(monitors))
    for i, monitor := range monitors {
        monitorArray[i] = []interface{}{monitor.ID, monitor.Name}
    }

    response := map[string]interface{}{
        "monitors": monitorArray,
    }
    
    w.Header().Set("Content-Type", "application/json")

    jsonResponse, err := json.MarshalIndent(response, "", "    ")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Write(jsonResponse)
}


type Monitor struct {
    ID   int
    Name string
}

func GetMonitorsList() []Monitor {
    var arr []Monitor
    var id int
    var name string
    rows, _ := DB.Query("SELECT id, name FROM monitors")
    for rows.Next() {
        rows.Scan(&id, &name)
        arr = append(arr, Monitor{ID: id, Name: name})
    }
    return arr
}

func GetStatForMonitor(w http.ResponseWriter, request *http.Request) {
    err := request.ParseForm()
    if err != nil {
        fmt.Fprintf(w, "{%s}", err)
    } else {
        countValue := 0
        monitorId := strings.TrimPrefix(request.URL.Path, "/category/monitor/")
        OpenDB()
        rows, _ := DB.Query("SELECT count FROM monitors WHERE id =" + monitorId)
        for rows.Next() {
            rows.Scan(&countValue)
        }
        strOut := "{\"id\": \"" + monitorId + "\", \"count\": \"" + strconv.Itoa(countValue) + "\"}"
        fmt.Fprintf(w, strOut)
    }
}

