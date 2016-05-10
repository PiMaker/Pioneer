package main

import (
    "./commands"
    
    "net/http"
    "html/template"
    "fmt"
    "github.com/twinj/uuid"
    "time"
    
    "os"
    "os/exec"
    "io/ioutil"
    "encoding/json"
    "github.com/DisposaBoy/JsonConfigReader"
    
    "strings"
    "strconv"
	"bytes"
)

type TemplateModel struct {
    Motd string
    LiveBackground bool
    LiveBackgroundInterval int
    Commands []commands.DisplayCommand
}

type Token struct {
    cookie *http.Cookie
    username string
}

type User struct {
    username string
    password string
    statusAccess bool
}

type LiveBackgroundSettings struct {
    enabled bool
    interval int
    command string
    commandArgs []string
    filename string
    users []string
}

var (
    templateCollection *template.Template
    validTokens []Token
    config commands.JsonObject
    users map[string]User
    templateModels map[string]TemplateModel
    liveBackground LiveBackgroundSettings
)

const pioneerAccessToken = "pioneer-access-token"

func main() {
    fmt.Println(time.Now().String() + " [INFO] Pioneer starting...")
    loadConfig()
    commands.ParseCommands(config)
    
    if liveBackground.enabled {
        exec.Command(liveBackground.command, liveBackground.commandArgs...)
        ticker := time.NewTicker(time.Duration(liveBackground.interval) * time.Second)
        go func() {
            fmt.Println(time.Now().String() + " [INFO] Starting live background ticker...")
            for {
                <-ticker.C
                invalidateCookies()
                if len(validTokens) > 0 {
                    exec.Command(liveBackground.command, liveBackground.commandArgs...)
                }
            }
        }()
    } else {
        fmt.Println(time.Now().String() + " [INFO] Live background disabled, no action taken.")
    }
    
    fmt.Println(time.Now().String() + " [INFO] Loading models...")
    templateModels = make(map[string]TemplateModel)
    for user := range users {
        templateModels[user] = TemplateModel{Motd: config["motd"].(string), Commands: getCommandsForUser(user),
            LiveBackground: liveBackground.enabled && in(liveBackground.users, user), LiveBackgroundInterval: liveBackground.interval}
    }
    
    templateCollection = template.Must(template.ParseFiles("html/login.html", "html/main.html"))
    
    fmt.Println(time.Now().String() + " [INFO] Registering handlers...")
    http.HandleFunc("/login", loginHandler)
    http.HandleFunc("/main", mainHandler)
    http.HandleFunc("/api/", apiHandler)
    http.Handle("/js/", http.FileServer(http.Dir("./assets")))
    http.Handle("/css/", http.FileServer(http.Dir("./assets")))
    http.Handle("/img/", http.FileServer(http.Dir("./assets")))
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request){
        http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
    })
    
    var err interface{}
    if strings.ToLower(config["ssl"].(string)) == "true" {
        fmt.Println("Starting https server...")
        err = http.ListenAndServeTLS(":443", config["certFile"].(string), config["keyFile"].(string), nil)
    } else {
        fmt.Println("Starting http server...")
        fmt.Println(time.Now().String() + " [WARN] SSL encryption disabled! This is not recommended, as secure connections are not possible without it!")
        err = http.ListenAndServe(":80", nil)
    }
    
    if err != nil {
        panic(fmt.Sprintf("ListenAndServe: %s", err))
    }
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Access-Control-Allow-Origin", "*")
    cookie, cerr := r.Cookie(pioneerAccessToken)
    valid, _ := cookieIsValid(cookie)
    if cerr == nil && valid {
        http.Redirect(w, r, "/main", http.StatusTemporaryRedirect)
        return
    }
    err := templateCollection.ExecuteTemplate(w, "login.html", nil)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
}

func mainHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Access-Control-Allow-Origin", "*")
    cookie, err := r.Cookie(pioneerAccessToken)
    valid, token := cookieIsValid(cookie)
    if err != nil || !valid {
        http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
        return
    }
    
    terr := templateCollection.ExecuteTemplate(w, "main.html", templateModels[token.username])
    if terr != nil {
        http.Error(w, terr.Error(), http.StatusInternalServerError)
    }
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
    //w.Header().Set("Access-Control-Allow-Origin", "*")
    command := r.URL.Path[len("/api/"):]
    slashIndex := strings.Index(command, "/")
    if slashIndex != -1 {
        command = command[:slashIndex]
    }
    switch command {
    case "login":
        if r.Method != "POST" {
            http.Error(w, "You can only POST to this api!", http.StatusBadRequest)
            return
        }
        body, error := ioutil.ReadAll(r.Body)
        valid, username := isValidLogin(string(body))
        if error == nil && valid {
            u := uuid.NewV4()
            cookie := &http.Cookie{Name: pioneerAccessToken, Value: u.String(), Expires: time.Now().Add(30*time.Minute), Path: "/"}
            validTokens = append(validTokens, Token{cookie: cookie, username: username})
            http.SetCookie(w, cookie)
            fmt.Fprintln(w, u.String())
            fmt.Println(time.Now().String() + " [INFO] User " + username + " logged in, received token: " + u.String() + " (valid until " + cookie.Expires.String() + ")")
        } else {
            http.Error(w, "Wrong login!", 403)
            var unameUsed string
            if error == nil {
                unameUsed = string(body)
            } else {
                unameUsed = "<Invalid Message Body>"
            }
            fmt.Println(time.Now().String() + " [INFO] Unsuccessful login attempt: " + unameUsed)
        }
        
        break
    case "logout":
        if r.Method != "POST" {
            http.Error(w, "You can only POST to this api!", http.StatusBadRequest)
            return
        }
        cookie, err := r.Cookie(pioneerAccessToken)
        if err != nil {
            http.Error(w, "Unauthorized", 403)
            return
        }
        
        for i := 0; i < len(validTokens); i++ {
            if cookie.Value == validTokens[i].cookie.Value {
                validTokens[i].cookie.Expires = time.Now().Add(-1 * time.Minute)
            }
        }
        
        break
    case "getcmds":
        cookie, err := r.Cookie(pioneerAccessToken)
        valid, token := cookieIsValid(cookie)
        if err != nil || !valid {
            http.Error(w, "Unauthorized", 403)
            return
        }
        
        retval := "{"
        for i, cmd := range templateModels[token.username].Commands {
            if i > 0 {
                retval += ","
            }
            retval += "\"" + cmd.Name + "\": " + strconv.Itoa(cmd.ID)
        }
        retval += "}"
        
        fmt.Fprint(w, retval)
        
        break;
    case "cmd":
        if r.Method != "POST" {
            http.Error(w, "You can only POST to this api!", http.StatusBadRequest)
            return
        }
        cookie, err := r.Cookie(pioneerAccessToken)
        valid, token := cookieIsValid(cookie)
        if err != nil || !valid {
            http.Error(w, "Unauthorized", 403)
            return
        }
        
        body, err := ioutil.ReadAll(r.Body)
        if err != nil {
            http.Error(w, "Parameter error", 500)
            return
        }
        
        origCmd := r.URL.Path[len("/api/"):]
        if len(origCmd) < 5 {
            http.Error(w, "Parameter error", 500)
            return
        }
        
        id, serr := strconv.Atoi(origCmd[slashIndex + 1:])
        if serr != nil {
            http.Error(w, "Parameter error", 500)
            return
        }
        
        cmd, ok := commands.CommandsAvailable[id]
        if ok && in(cmd.AllowedUsers, token.username) {
            fmt.Println(time.Now().String() + " [INFO] Executing command: " + cmd.Name)
            retval := cmd.ExecutableCommand.Execute(string(body))
            fmt.Fprint(w, retval)
            return
        }
        
        http.Error(w, "Command not found", 404)
        
        break
    case "getbck":
        cookie, err := r.Cookie(pioneerAccessToken)
        valid, token := cookieIsValid(cookie)
        if err != nil || !valid || !in(liveBackground.users, token.username) {
            http.Error(w, "Unauthorized", 403)
            return
        }
        
        streamBytes, err := ioutil.ReadFile(liveBackground.filename)

        if err != nil {
            http.Error(w, "Could not load background image...", 500)
            return
        }

        b := bytes.NewBuffer(streamBytes)
        
        _, writeerr := b.WriteTo(w)
        
        if writeerr != nil {
            http.Error(w, writeerr.Error(), 500)
        }
        
        break
    }
}

func in(haystack []string, needle string) bool {
    for _, h := range haystack {
        if h == needle {
            return true
        }
    }
    
    return false
}

func getCommandsForUser(username string) []commands.DisplayCommand {
    var retval []commands.DisplayCommand
    for _, cmd := range commands.CommandsAvailable {
        if in(cmd.AllowedUsers, username) {
            retval = append(retval, cmd)
        }
    }
    return retval
}

func isValidLogin(body string) (bool, string) {
    nPos := strings.Index(body, "\n")
    if nPos < 0 || len(body) <= nPos + 1 {
        return false, ""
    }
    
    username := body[:nPos]
    password := body[nPos + 1:]
    
    expectedUser, ok := users[username]
    if !ok || expectedUser.password != password {
        return false, ""
    }
    
    return true, username
}

func cookieIsValid(cookie *http.Cookie) (bool,*Token) {
    if cookie == nil {
        return false,nil
    }
    invalidateCookies()
    for i := 0; i < len(validTokens); i++ {
        if cookie.Value == validTokens[i].cookie.Value {
            return true,&validTokens[i]
        }
    }
    
    return false,nil
}

func invalidateCookies() {
    for i := 0; i < len(validTokens); i++ {
        if time.Now().After(validTokens[i].cookie.Expires) {
            validTokens = append(validTokens[:i], validTokens[i+1:]...)
            i--
            continue
        }
    }
}

func loadConfig() {
    fmt.Println(time.Now().String() + " [INFO] Parsing config.json...")
    
    var v interface{}
    f, _ := os.Open("config.json")
    r := JsonConfigReader.New(f)
    json.NewDecoder(r).Decode(&v)
    config = v.(map[string]interface{})
    
    users = make(map[string]User)
    u := config["users"].([]interface{})
    for _, us := range u {
        use := us.(map[string]interface{})
        uname := use["username"].(string)
        users[uname] = User{username: uname, password: use["password"].(string)}
    }
    
    lbc := config["live_background"].(map[string]interface{})
    cmd := lbc["command"].(string)
    split := strings.Split(cmd, " ")
    liveBackground = LiveBackgroundSettings{
        command: split[0],
        commandArgs: split[1:],
        filename: lbc["filename"].(string),
        interval: int(lbc["interval"].(float64)),
        users: toStringSlice(lbc["users"].([]interface{})),
        enabled: lbc["enabled"].(bool) }
        
    fmt.Println(time.Now().String() + " [INFO] Config parsed and loaded.")
}

func toStringSlice(input []interface{}) []string {
    toRet := make([]string, len(input))
    for i := range input {
        toRet[i] = input[i].(string)
    }
    return toRet
}