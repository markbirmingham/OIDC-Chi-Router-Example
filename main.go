package main

import (
	"embed"
	"encoding/gob"
	"html/template"
	"io/fs"
	"log"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
)

var (
	//go:embed static/js
	static embed.FS
	//go:embed template
	templates embed.FS
)

var (
	key   = []byte("secret")
	store = sessions.NewCookieStore(key)
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Fatalf("Failed to load the env vars: %v", err)
	}

	auth, err := NewAuthenticator()
	if err != nil {
		log.Fatalf("Failed to initialize the authenticator: %v", err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// Use gob register to store custom types in our cookies
	gob.Register(map[string]interface{}{})

	fsys, _ := fs.Sub(static, "static")
	staticContent := http.FileServer(http.FS(fsys))
	r.Handle("/public/*", http.StripPrefix("/public", staticContent))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFS(templates, "template/home.html"))
		tmpl.Execute(w, nil)
	})

	r.Get("/login", Login(store, auth))
	r.Get("/callback", Callback(store, auth))
	r.Get("/user", isAuthenticated(User))
	r.Get("/protected", isAuthenticated(Protected))
	r.Get("/logout", Logout)
	r.Get("/bye", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFS(templates, "template/logout.html"))
		tmpl.Execute(w, nil)
	})

	http.ListenAndServe(":9000", r)
}

func isAuthenticated(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := store.Get(r, "auth-session")
		if err != nil {
			panic(err)
		}

		profile := session.Values["profile"]

		if profile == nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, r)
	}
}
