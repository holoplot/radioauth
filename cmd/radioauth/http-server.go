package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/dchest/uniuri"
	"github.com/holoplot/sw__radioauth/account"
	"golang.org/x/oauth2"
)

const html = `
<!doctype html>
<html lang="en">
	<head>
		<meta charset="utf-8">
		<meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
		<link rel="stylesheet" href="https://stackpath.bootstrapcdn.com/bootstrap/4.3.1/css/bootstrap.min.css" integrity="sha384-ggOyR0iXCbMQv3Xipma34MD+dH/1fQ784/j6cY/iJTQUOhcWr7x9JvoRxT2MZw1T" crossorigin="anonymous">
		<title>RadiOauth</title>
	</head>
	<style>
		body {
			margin-top: 100px;
		}
	</style>
	<body>
		<div class="container">
			<div class="card card-block" style="width: 22rem;" align-self-center>
				<div class="card-body">
					<h2 class="card-title">{{.Title}}</h2>
					<p class="card-text">{{.Text}}</p>

					{{if .Password -}}
						<p><code>{{.Password}}</code></p>
					{{- end}}

					{{if .LinkTarget -}}
						<p><a href="{{.LinkTarget}}">{{.LinkText}}</a></p>
					{{- end}}

					{{if .ButtonTarget -}}
						<p><a href="{{.ButtonTarget}}" role="button" class="btn btn-primary">{{.ButtonText}}</a></p>
					{{- end}}
				</div>
			</div>
		</div>
	</body>
</html>
`

type templateVars struct {
	Title        string
	Text         string
	Password     string
	LinkTarget   string
	LinkText     string
	ButtonTarget string
	ButtonText   string
}

func runHTTPServer() {
	ctx := context.Background()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		cookie := http.Cookie{
			Name:   "radioauth-csrf",
			Value:  uniuri.NewLen(20),
			MaxAge: 30,
		}
		http.SetCookie(w, &cookie)
		http.Redirect(w, r, oauthConfig.AuthCodeURL(cookie.Value, oauth2.AccessTypeOffline), http.StatusFound)
	})

	http.HandleFunc(relativeCallbackURL, func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("radioauth-csrf")
		if err != nil {
			http.Error(w, "can't get CSRF cookie", http.StatusBadRequest)
			return
		}

		if r.URL.Query().Get("state") != cookie.Value {
			http.Error(w, "state did not match", http.StatusBadRequest)
			return
		}

		oauth2Token, err := oauthConfig.Exchange(ctx, r.URL.Query().Get("code"))
		if err != nil {
			http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
			return
		}

		userInfo, err := provider.UserInfo(ctx, oauth2.StaticTokenSource(oauth2Token))
		if err != nil {
			http.Error(w, "Failed to get userinfo: "+err.Error(), http.StatusInternalServerError)
			return
		}

		vars := templateVars{}
		var a *account.Account

		if oauth2Token.RefreshToken != "" {
			a = account.NewWithRandomPassword(userInfo.Email)
			a.RefreshToken = oauth2Token.RefreshToken

		} else {
			a, err = accountStore.Read(userInfo.Email)
			if err != nil {
				http.Error(w, "No such account!", http.StatusInternalServerError)
				return
			}
		}

		a.AccessToken = oauth2Token.AccessToken
		err = accountStore.Write(a)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		vars.Title = "Great! You made it!"
		vars.Text = "You may now use the service using the password below. Make sure to store it in a secure place, as this is the last time it will be prompted to you."
		vars.Password = a.Password

		tmpl, err := template.New("name").Parse(html)
		if err != nil {
			http.Error(w, "tmpl error: "+err.Error(), http.StatusBadRequest)
			return
		}

		tmpl.Execute(w, vars)
	})

	log.Printf("Starting HTTP server on :%d", config.HTTPPort)
	http.ListenAndServe(fmt.Sprintf(":%d", config.HTTPPort), nil)
}
