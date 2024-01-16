package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"
	"strings"

	"cloud.google.com/go/storage"
)

const (
	cookieName = "balkAuth"
	rootPath   = "./site/"
)

var (
	users   []User
	nilUser User
)

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

func isAuthorizedToAccessFile(w http.ResponseWriter, r *http.Request, path string) bool {
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

func setupCredentialsDB() []User {
	// if call fails return non nil user, preventing logins but making the site available
	errorFunction := func(err error) []User {
		log.Println(err.Error())
		return []User{{Username: "error"}}
	}

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return errorFunction(err)
	}
	rc, err := client.Bucket("secret-balk-content").Object("creds.json").NewReader(ctx)
	if err != nil {
		return errorFunction(err)
	}
	slurp, err := io.ReadAll(rc)
	rc.Close()
	if err != nil {
		return errorFunction(err)
	}

	creds := Users{}
	_ = json.Unmarshal([]byte(slurp), &creds)
	users = creds.Users
	log.Println("Successfully read credentials")
	return users
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie(cookieName)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	user := findUserByValue(c.Value)
	if user != nilUser {
		data := struct {
			Username string `json:"username"`
		}{user.Username}
		json.NewEncoder(w).Encode(data)
		return
	}
	s, _ := httputil.DumpRequest(r, false)
	log.Printf("%s\n", s)
	http.Error(w, "You are an intersting person.\nWrite me a mail!", http.StatusNotAcceptable)
}

func Login(w http.ResponseWriter, r *http.Request) {
	username, pwd, ok := r.BasicAuth()
	if ok {
		cookie := validateCredentials(username, pwd)
		if cookie != nil {
			log.Printf("Successful login for: %s", username)
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

	if !isAuthorizedToAccessFile(w, r, path) {
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
	users = setupCredentialsDB()

	mux := http.NewServeMux()
	mux.HandleFunc("/", ServeHTTP)
	mux.HandleFunc("/login", Login)
	mux.HandleFunc("/user", GetUser)

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
