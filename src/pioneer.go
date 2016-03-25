package main

import (
    "./commands"
    
    "net/http"
    "html/template"
    "fmt"
    "github.com/twinj/uuid"
    "time"
    
    "os"
    "io/ioutil"
    "encoding/json"
    "github.com/DisposaBoy/JsonConfigReader"
    
    "strings"
    "strconv"
)

type TemplateModel struct {
    Motd string
    Commands []commands.DisplayCommand
}

type Token struct {
    cookie *http.Cookie
    username string
}

var (
    templateCollection *template.Template
    validTokens []Token
    config commands.JsonObject
    users map[string]string
    templateModels map[string]TemplateModel
)

const pioneerAccessToken = "pioneer-access-token"

func main() {
    loadConfig()
    commands.ParseCommands(config)
    
    templateModels = make(map[string]TemplateModel)
    for user := range users {
        templateModels[user] = TemplateModel{Motd: config["motd"].(string), Commands: getCommandsForUser(user)}
    }
    
    templateCollection = template.Must(template.ParseFiles("html/login.html", "html/main.html"))
    
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
        fmt.Println("[WARNING] SSL encryption disabled! This is not recommended, as securely logging in is not possible without it!")
        err = http.ListenAndServe(":80", nil)
    }
    
    if err != nil {
        panic(fmt.Sprintf("ListenAndServe: %s", err))
    }
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
    cookie, cerr := r.Cookie(pioneerAccessToken)
    valid, _ := cookieIsValid(cookie)
    if cerr == nil && valid {
        http.Redirect(w, r, "/main", http.StatusTemporaryRedirect)
        return
    }
    err := templateCollection.ExecuteTemplate(w, "login.html", nil)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func mainHandler(w http.ResponseWriter, r *http.Request) {
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
    if r.Method != "POST" {
        http.Error(w, "You can only POST to this api!", http.StatusBadRequest)
        return
    }
    command := r.URL.Path[len("/api/"):]
    slashIndex := strings.Index(command, "/")
    if slashIndex != -1 {
        command = command[:slashIndex]
    }
    switch command {
    case "login":
        body, error := ioutil.ReadAll(r.Body)
        valid, username := isValidLogin(string(body))
        if error == nil && valid {
            u := uuid.NewV4()
            cookie := &http.Cookie{Name: pioneerAccessToken, Value: u.String(), Expires: time.Now().Add(30*time.Minute), Path: "/"}
            validTokens = append(validTokens, Token{cookie: cookie, username: username})
            http.SetCookie(w, cookie)
            fmt.Fprintln(w, u.String())
        } else {
            http.Error(w, "Wrong login!", 403)
        }
        
        break
    case "logout":
        cookie, err := r.Cookie(pioneerAccessToken)
        if err != nil {
            http.Error(w, "Cookie error", 500)
            return
        }
        
        for i := 0; i < len(validTokens); i++ {
            if cookie.Value == validTokens[i].cookie.Value {
                validTokens[i].cookie.Expires = time.Now().Add(-1 * time.Minute)
            }
        }
        
        break
    case "cmd":
        cookie, err := r.Cookie(pioneerAccessToken)
        valid, token := cookieIsValid(cookie)
        if err != nil || !valid {
            http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
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
            retval := cmd.ExecutableCommand.Execute(string(body))
            fmt.Fprint(w, retval)
            return
        }
        
        http.Error(w, "Command not found", 404)
        
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
    
    expectedPass, ok := users[username]
    if !ok || expectedPass != password {
        return false, ""
    }
    
    return true, username
}

func cookieIsValid(cookie *http.Cookie) (bool,*Token) {
    for i := 0; i < len(validTokens); i++ {
        if time.Now().After(validTokens[i].cookie.Expires) {
            validTokens = append(validTokens[:i], validTokens[i+1:]...)
            i--
            continue
        }
        if cookie.Value == validTokens[i].cookie.Value {
            return true,&validTokens[i]
        }
    }
    
    return false,nil
}

func loadConfig() {
    var v interface{}
    f, _ := os.Open("config.json")
    r := JsonConfigReader.New(f)
    json.NewDecoder(r).Decode(&v)
    config = v.(map[string]interface{})
    
    users = make(map[string]string)
    u := config["users"].([]interface{})
    for _, us := range u {
        use := us.(map[string]interface{})
        users[use["username"].(string)] = use["password"].(string)
    }
}