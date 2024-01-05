package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var rootPath = "./site/"

const cookieName = "balkAuth"

var users []User

var nilUser User

type User struct {
	Username string `json:"username"`
	Pwd      string `json:"pwd"`
	Cookie   string `json:"cookie"`
}

type Users struct {
	Users []User `json:"users"`
}

func notFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)

	f, err := os.Open(filepath.Join(rootPath, "404.html"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()

	io.Copy(w, f)
}

func unauthorized(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusUnauthorized)

	f, err := os.Open(filepath.Join(rootPath, "401.html"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()

	io.Copy(w, f)
}

func findUserByValue(val string) User {
	for i := 0; i < len(users); i++ {
		user := users[i]
		if val == user.Username || val == user.Pwd || val == user.Cookie {
			return user
		}
	}
	return nilUser
}

func isCallerAuthorized(w http.ResponseWriter, r *http.Request, path string) bool {
	if strings.Contains(path, "/secrets/") {
		c, err := r.Cookie(cookieName)
		if err != nil {
			unauthorized(w, r)
			return false
		}
		//check if sent cookie value belongs to one of the known accounts
		if findUserByValue(c.Value) == nilUser {
			unauthorized(w, r)
			return false
		}
	}
	return true
}

func validateCredentials(username string, pwd string) *http.Cookie {
	user := findUserByValue(username)

	if user != nilUser && pwd == user.Pwd {
		return &http.Cookie{Name: cookieName, Value: user.Cookie, MaxAge: 1800, Secure: true, SameSite: http.SameSiteNoneMode}
	}

	return nil
}

func Login(w http.ResponseWriter, r *http.Request) {
	username, pwd, ok := r.BasicAuth()
	if ok {
		cookie := validateCredentials(username, pwd)
		if cookie != nil {
			http.SetCookie(w, cookie)
			w.WriteHeader(http.StatusNoContent)
			return
		}
		http.Error(w, "Bad credentials", http.StatusUnauthorized)
		return
	} else {
		http.Error(w, "400 - Bad Request", http.StatusBadRequest)
		return
	}
}

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := filepath.Join(rootPath, r.URL.Path[1:])

	if !isCallerAuthorized(w, r, path) {
		return
	}

	f, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			notFound(w, r)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if f.IsDir() {
		path = filepath.Join(path, "index.html")
		if _, err := os.Stat(path); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				notFound(w, r)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	http.ServeFile(w, r, path)
}

func main() {
	// either read from a save source, i.e cloud storage, or use a proper database you nitwit (or just a auth provider..)
	file, _ := os.ReadFile("creds.json")
	creds := Users{}
	_ = json.Unmarshal([]byte(file), &creds)
	users = creds.Users

	mux := http.NewServeMux()
	mux.HandleFunc("/", ServeHTTP)
	mux.HandleFunc("/login", Login)

	cfg := &tls.Config{
		MinVersion:               tls.VersionTLS12,
		CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_AES_256_GCM_SHA384,
		},
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		TLSConfig:    cfg,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
	}

	log.Printf("Listening on port %s", port)
	log.Fatal(srv.ListenAndServe())
}
