package main

import (
    "net/http"
    "html/template"
    "fmt"
    "github.com/twinj/uuid"
    "time"
)

var (
    templateCollection *template.Template
    validTokens []*http.Cookie
)

const pioneerAccessToken = "pioneer-access-token"

func main() {
    templateCollection = template.Must(template.ParseFiles("login.html", "main.html"))
    
    http.HandleFunc("/login", loginHandler)
    http.HandleFunc("/main", mainHandler)
    http.HandleFunc("/api/", apiHandler)
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request){
        fmt.Println("Root")
        http.Redirect(w, r, "/main", http.StatusTemporaryRedirect)
    })
    
    err := http.ListenAndServe(":80", nil)
    if err != nil {
        panic(fmt.Sprintf("ListenAndServe: %s", err))
    }
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
    err := templateCollection.ExecuteTemplate(w, "login.html", nil)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func mainHandler(w http.ResponseWriter, r *http.Request) {
    cookie, err := r.Cookie(pioneerAccessToken)
    if err != nil || !cookieIsValid(cookie) {
        http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
    }
    
    terr := templateCollection.ExecuteTemplate(w, "main.html", nil)
    if terr != nil {
        http.Error(w, terr.Error(), http.StatusInternalServerError)
    }
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
    command := r.URL.Path[len("/api/"):]
    switch command {
    case "login":
        u := uuid.NewV4()
        cookie := &http.Cookie{Name: pioneerAccessToken, Value: u.String(), Expires: time.Now().Add(30*60)}
        validTokens = append(validTokens, cookie)
        http.SetCookie(w, cookie)
        fmt.Fprintln(w, u.String())
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

// https://github.com/stianeikeland/go-rpio
// http://blog.golang.org/json-and-go