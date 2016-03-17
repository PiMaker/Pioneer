package main

import (
    "./commands"
    
    "net/http"
    "html/template"
    "fmt"
    "github.com/twinj/uuid"
    "time"
    
    "io/ioutil"
    "encoding/json"
    
    "strings"
    "strconv"
)

type TemplateModel struct {
    Motd string
    Commands []commands.DisplayCommand
}

var (
    templateCollection *template.Template
    validTokens []*http.Cookie
    config commands.JsonObject
    templateModel TemplateModel
)

const pioneerAccessToken = "pioneer-access-token"

func main() {
    loadConfig()
    commands.ParseCommands(config)
    
    templateModel = TemplateModel{Motd: config["motd"].(string), Commands: commands.CommandsAvailable}
    templateCollection = template.Must(template.ParseFiles("html/login.html", "html/main.html"))
    
    http.HandleFunc("/login", loginHandler)
    http.HandleFunc("/main", mainHandler)
    http.HandleFunc("/api/", apiHandler)
    http.Handle("/js/", http.FileServer(http.Dir("./assets")))
    http.Handle("/css/", http.FileServer(http.Dir("./assets")))
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
    if cerr == nil && cookieIsValid(cookie) {
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
    if err != nil || !cookieIsValid(cookie) {
        http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
        return
    }
    
    terr := templateCollection.ExecuteTemplate(w, "main.html", templateModel)
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
        if error == nil && string(body) == config["password"] {
            u := uuid.NewV4()
            cookie := &http.Cookie{Name: pioneerAccessToken, Value: u.String(), Expires: time.Now().Add(30*time.Minute), Path: "/"}
            validTokens = append(validTokens, cookie)
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
            if cookie.Value == validTokens[i].Value {
                validTokens[i].Expires = time.Now().Add(-1 * time.Minute)
            }
        }
        
        break
    case "cmd":
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
        
        for _, cmd := range commands.CommandsAvailable {
            if cmd.ID == id {
                if cmd.ExecutableCommand.Execute(string(body)) {
                    return
                }
                
                http.Error(w, "Command returned an error", 500)
                return
            }
        }
        
        http.Error(w, "Command not found", 404)
        
        break
    }
}

func cookieIsValid(cookie *http.Cookie) bool {
    for i := 0; i < len(validTokens); i++ {
        if time.Now().After(validTokens[i].Expires) {
            validTokens = append(validTokens[:i], validTokens[i+1:]...)
            i--
            continue
        }
        if cookie.Value == validTokens[i].Value {
            return true
        }
    }
    
    return false
}

func loadConfig() {
    configString, err := ioutil.ReadFile("config.json")
    if err != nil {
        panic("config.json could not be loaded!")
    }
    
    var c interface{}
    json.Unmarshal(configString, &c)
    config = c.(map[string]interface{})
}