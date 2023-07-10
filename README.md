# oauth2

This is a simple oauth2 package i made that anyone can use in there project that can help people add a SSO in there applications

## Installation

To install the package, you need to install Go and set your Go workspace first.

1. The first need [Go](https://golang.org/) installed (**version 1.13+**), then you can use the Go command below.

```sh
$ go get -u github.com/andreGarvin/oauth2
```

2. Import it in your code:

```go
import "github.com/andreGarvin/oauth2"
```

Then you can start using it

## Quick start

```go
// assuming the following codes in main.go file

package main

import (
	"fmt"
	"log"
	"net/http"
	"oauth2"
)

var port = ":8080"

func main() {

	var oauth = oauth2.New(
		"<client id>",
		"<oauth url>",
		"<token url>",
		"http://localhost:8080/oauth/callback",
		"<client secret>",
		oauth2.Scopes,
	)

	// endpoints

	// intitating a oauth login request and redirecting the user to the oauth service provider to authenticate the user on our service
	http.HandleFunc("/oauth/signin", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("signin: redirecting the user to foobarbaz.com oauth page")

		// getting the formatted oauth url
		// or instead of using a string literal "sigin" use oauth2.OauthActionSignin, but this generally where state goes
		oauthURL, err := oauth.CreateOauthURL("sigin")
		if err != nil {
			fmt.Printf("error: %v", err)

			w.WriteHeader(http.StatusInternalServerError)

			fmt.Fprint(w, "internal server error")
			return
		}

		fmt.Println(oauthURL)
		http.Redirect(w, r, oauthURL, http.StatusTemporaryRedirect)
	})

	// intitating a oauth login request and redirecting the user to the oauth service provider.
	// but going to set the state of the oauth url to be a oauth action to create a account on our service when it comes back to our service
	http.HandleFunc("/oauth/create", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("create: redirecting the user to foobarbaz.com oauth page")

		// getting the formatted oauth url
		oauthURL, err := oauth.CreateOauthURL(oauth2.OauthActionCreateAccount)
		if err != nil {
			fmt.Printf("error: %v", err)

			w.WriteHeader(http.StatusInternalServerError)

			fmt.Fprint(w, "internal server error")
			return
		}

		http.Redirect(w, r, oauthURL, http.StatusTemporaryRedirect)
	})

	// and for OauthActionAuthorize you can have some authenticated endpoint that will performa action to your app to get access from the other app

	// handling oauth callback/redirect after a user has finished oauth against the IDP
	http.HandleFunc("/oauth/callback", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("handling the oauth callback")

		query := r.URL.Query()

		// this access code will allow to make a request from the service porvider to get the access token
		accessCode := query.Get("code")
		// This would be the state information we sent in the beginning of the oauth request
		state := query.Get("state")

		userOauthToken, err := oauth.FetchAccessToken(accessCode)
		if err != nil {
			fmt.Printf("failed to get user token information, %s\n", err)

			w.WriteHeader(http.StatusInternalServerError)

			fmt.Fprint(w, "internal server error")
			return
		}

		fmt.Println("oauth user access Token")
		fmt.Printf("\n%+v\n", userOauthToken)

		switch state {
		case oauth2.OauthActionCreateAccount:
			// create a new user account
			break
		case oauth2.OauthActionSignin:
			// authenticate a existing user
			break
		case oauth2.OauthActionAuthorize:
			// add some authorization stuff for your application
			break
		}

		w.WriteHeader(http.StatusOK)

		fmt.Fprint(w, "you have been authenticated")
	})

	fmt.Printf("Running on Port %s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
```

After writing the code above code you can then run the command below and test it. After testing that it works you can customize whatever bit you want in the `/oauth/callback` for your application

```
# run example.go
$ go run example.go
```
