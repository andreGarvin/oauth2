package oauth2

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

var (
	// come default scopes when intating oauth
	Scopes []string = []string{"profile", "email"}

	// enums for oauth actions
	OauthActionSignin        string = "SIGNIN"
	OauthActionAuthorize     string = "AUTHORIZE"
	OauthActionCreateAccount string = "CREATE_ACCOUNT"
)

// UserOauthToken is token information returned after requesting to fetch the access token
type UserOauthToken struct {
	IDToken      string `json:"id_token"`
	ExpiresAt    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// OauthState is a struct that will represent the JWT payload for the oauth state
type OauthState struct {
	Action string `json:"action"`
}

// Oauth returns a new the methods initate oauth witha IDP
type Oauth struct {
	clientID     string
	oauthURL     string
	tokenURL     string
	callbackURL  string
	clientSecret string
	scopes       []string
}

// New returns the Oauth struct
func New(clientID, oauthURL, tokenURL, callbackURL, clientSecret string, scopes []string) *Oauth {
	return &Oauth{
		scopes:       scopes,
		clientID:     clientID,
		oauthURL:     oauthURL,
		tokenURL:     tokenURL,
		callbackURL:  callbackURL,
		clientSecret: clientSecret,
	}
}

// CreateOauthURL returns a formatted oauth url
func (oauth *Oauth) CreateOauthURL(state string) (string, error) {
	parsedOauthURL, err := url.Parse(oauth.oauthURL)
	if err != nil {
		return "", err
	}

	query := parsedOauthURL.Query()

	query.Set("scope", strings.Join(oauth.scopes, " "))
	query.Set("redirect_uri", oauth.callbackURL)
	query.Set("client_id", oauth.clientID)
	query.Set("access_type", "offline")
	query.Set("response_type", "code")
	query.Set("prompt", "consent")
	query.Set("state", state)

	parsedOauthURL.RawQuery = query.Encode()

	return parsedOauthURL.String(), nil
}

// FetchAccessToken fetches the access token when obtaining the access after the user has oauth from the oauth prodiver
func (oauth *Oauth) FetchAccessToken(accessCode string) (UserOauthToken, error) {
	var oauthAccessTokenInfo UserOauthToken

	var query url.Values = make(url.Values)
	query.Set("client_secret", oauth.clientSecret)
	query.Set("grant_type", "authorization_code")
	query.Set("redirect_uri", oauth.callbackURL)
	query.Set("client_id", oauth.clientID)
	query.Set("code_verifier", "")
	query.Set("code", accessCode)

	err := fetchOauthToken(oauth.tokenURL, query.Encode(), &oauthAccessTokenInfo)
	if err != nil {
		return oauthAccessTokenInfo, err
	}

	return oauthAccessTokenInfo, nil
}

// RefreshAccessToken requests for a new access token from the oauth provider
func (oauth *Oauth) RefreshAccessToken(refreshToken string) (UserOauthToken, error) {
	var oauthAccessTokenInfo UserOauthToken

	var query url.Values = make(url.Values)
	query.Set("client_secret", oauth.clientSecret)
	query.Set("refresh_token", refreshToken)
	query.Set("grant_type", "refresh_token")
	query.Set("client_id", oauth.clientID)

	err := fetchOauthToken(oauth.tokenURL, query.Encode(), &oauthAccessTokenInfo)
	if err != nil {
		return oauthAccessTokenInfo, err
	}

	return oauthAccessTokenInfo, nil
}

// fetchOauthToken this a obstraction from making the request to the oauth provider
func fetchOauthToken(oauthTokenURL string, urlEncodedBody string, val interface{}) error {
	payload := bytes.NewBuffer([]byte(urlEncodedBody))

	request, err := http.NewRequest("POST", oauthTokenURL, payload)

	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if err != nil {
		return err
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	if response.StatusCode != 200 {
		if response.StatusCode == 400 {
			return errors.New("oauth: error fetching token, endpoint does not exist")
		}

		var badRequest struct {
			Error       string `json:"error"`
			Description string `json:"error_description"`
		}

		err = json.Unmarshal(body, &badRequest)
		if err != nil {
			return err
		}

		return fmt.Errorf("oauth: error fetching token %v %v", badRequest.Error, badRequest.Description)
	}

	err = json.Unmarshal(body, &val)
	if err != nil {
		return err
	}

	return nil
}
