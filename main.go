package main

import (
	"crypto/tls"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var rootPath = "./site/"

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := filepath.Join(rootPath, r.URL.Path[1:])

	if strings.Contains(path, "/secrets/") {
		c, err := r.Cookie("cookie")
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if c.Value != "yummy" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
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

func Login(w http.ResponseWriter, r *http.Request) {
	username, pwd, ok := r.BasicAuth()
	if ok {
		if username == "balk" && pwd == "123" {
			http.SetCookie(w, &http.Cookie{Name: "cookie", Value: "yummy", MaxAge: 1800, Secure: true, SameSite: http.SameSiteNoneMode})
			w.Header().Add("success", "123")
			w.WriteHeader(http.StatusOK)
		} else {
			http.Error(w, "Bad credentials", http.StatusUnauthorized)
		}
	} else {
		http.Error(w, "Unable to parse credentials", http.StatusBadRequest)
	}
}

func main() {
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
