package main

import (
	"crypto/rand"
	"encoding/base64"
	"html/template"
	"net/url"
	"os"

	"net/http"
)

func Login(auth *Authenticator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state, err := generateRandomState()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		session, err := store.Get(r, "auth-session")
		if err != nil {
			panic(err)
		}

		session.Values["state"] = state

		if err := session.Save(r, w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			panic(err)
		}

		http.Redirect(w, r, auth.AuthCodeURL(state), http.StatusTemporaryRedirect)
		return
	}
}

func generateRandomState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	state := base64.StdEncoding.EncodeToString(b)

	return state, nil
}

func Logout(w http.ResponseWriter, r *http.Request) {
	logoutUrl, err := url.Parse(os.Getenv("OIDC_PROVIDER_URL") + os.Getenv("OIDC_DOMAIN") + os.Getenv("OIDC_LOGOUT_URL"))

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var scheme string
	if r.TLS != nil {
		scheme = "https"
	} else {
		scheme = "http"
	}

	returnTo, err := url.Parse(scheme + "://" + r.Host + "/bye")

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	parameters := url.Values{}
	parameters.Add("redirect_uri", returnTo.String())
	parameters.Add("client_id", os.Getenv("OIDC_CLIENT_ID"))
	logoutUrl.RawQuery = parameters.Encode()

	http.Redirect(w, r, logoutUrl.String(), http.StatusTemporaryRedirect)
	return
}

func User(w http.ResponseWriter, r *http.Request) {
	// Save the state inside the session.
	session, err := store.Get(r, "auth-session")
	if err != nil {
		panic(err)
	}

	profile := session.Values["profile"]

	tmpl := template.Must(template.ParseFS(templates, "template/user.html"))
	tmpl.Execute(w, profile)
	return
}

func Protected(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "auth-session")
	if err != nil {
		panic(err)
	}

	profile := session.Values["profile"]

	tmpl := template.Must(template.ParseFS(templates, "template/protected.html"))
	tmpl.Execute(w, profile)
}
