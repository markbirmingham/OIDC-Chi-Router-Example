package main

import (
	"net/http"
)

func Callback(auth *Authenticator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		session, err := store.Get(r, "auth-session")
		if err != nil {
			panic(err)
		}

		if r.URL.Query().Get("state") != session.Values["state"] {
			http.Error(w, "Invalid state parameter", http.StatusBadRequest)
			return
		}

		token, err := auth.Exchange(r.Context(), r.URL.Query().Get("code"))
		if err != nil {
			http.Error(w, "Failed to exchange an authorization code for a token.", http.StatusUnauthorized)
			return
		}

		idToken, err := auth.VerifyIDToken(r.Context(), token)
		if err != nil {
			http.Error(w, "Failed to verify ID Token.", http.StatusInternalServerError)
			return
		}

		var profile map[string]interface{}
		if err := idToken.Claims(&profile); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		session.Values["access_token"] = token.AccessToken
		session.Values["profile"] = profile

		if err := session.Save(r, w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/user", http.StatusTemporaryRedirect)
		return
	}
}
