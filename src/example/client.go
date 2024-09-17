package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"golang.org/x/oauth2"
)

const (
	authServerURL = "http://localhost:3000"
)

var (
	config = oauth2.Config{
		ClientID:     "c48730a3-7032-4b30-a1a4-49cbc603e418",
		ClientSecret: "5xnwQR4RdYv9zlk0o9buDxtw7Q8f9J2m",
		Scopes:       []string{"all"},
		RedirectURL:  "http://localhost:9094/oauth2",
		Endpoint: oauth2.Endpoint{
			AuthURL:   authServerURL + "/auth",
			TokenURL:  "http://localhost:8080/api/v1" + "/oauth2/token",
			AuthStyle: oauth2.AuthStyleInHeader,
		},
	}
	globalToken *oauth2.Token // Non-concurrent security
)

func GenerateVerifier() string {
	// "RECOMMENDED that the output of a suitable random number generator be
	// used to create a 32-octet sequence.  The octet sequence is then
	// base64url-encoded to produce a 43-octet URL-safe string to use as the
	// code verifier."
	// https://datatracker.ietf.org/doc/html/rfc7636#section-4.1
	data := make([]byte, 32)
	if _, err := rand.Read(data); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(data)
}

func main() {
	// This is the URL that the user's browser hits to start the OAuth flow
	// This is represented by the "Login with Sastlink" button on the client, in the frontend will generate the URL, in this
	// example we are generating the URL in the backend for simplicity
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		u := config.AuthCodeURL("random_string",
			oauth2.SetAuthURLParam("code_challenge", genCodeChallengeS256("random_string")),
			oauth2.SetAuthURLParam("code_challenge_method", "S256"))
		http.Redirect(w, r, u, http.StatusFound)
	})

	// This is the URL that the client requests to Exchange the code for a token,
	// it should be called by the frontend, in this example we are calling it in the backend for simplicity
	http.HandleFunc("/oauth2", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		state := r.Form.Get("state")
		if state != "random_string" {
			http.Error(w, "State invalid", http.StatusBadRequest)
			return
		}
		code := r.Form.Get("code")
		if code == "" {
			http.Error(w, "Code not found", http.StatusBadRequest)
			return
		}
		token, err := config.Exchange(r.Context(), code, oauth2.SetAuthURLParam("code_verifier", "random_string"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		globalToken = token

		e := json.NewEncoder(w)
		e.SetIndent("", "  ")
		e.Encode(token)
	})

	http.HandleFunc("/refresh", func(w http.ResponseWriter, r *http.Request) {
		if globalToken == nil {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		globalToken.Expiry = time.Now()
		token, err := config.TokenSource(context.Background(), globalToken).Token()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Println("AccessToken: " + token.AccessToken)
		fmt.Println("RefreshToken: " + token.RefreshToken)
		globalToken = token
		e := json.NewEncoder(w)
		e.SetIndent("", "  ")
		e.Encode(token)
	})

	http.HandleFunc("/userInfo", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("globalToken: ", globalToken)
		if globalToken == nil {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
	
		client := config.Client(context.Background(), globalToken)
		resp, err := client.Get("http://localhost:8080/api/v1/oauth2/userinfo")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()
		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		e := json.NewEncoder(w)
		e.SetIndent("", "  ")
		e.Encode(result)
	})

	log.Println("Client is running at 9094 port.Please open http://localhost:9094")
	log.Fatal(http.ListenAndServe(":9094", nil))
}

func genCodeChallengeS256(s string) string {
	s256 := sha256.Sum256([]byte(s))
	return base64.URLEncoding.EncodeToString(s256[:])
}
