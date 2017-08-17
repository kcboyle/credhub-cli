package uaa

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
)

// Client makes requests to the UAA server at AuthUrl
type Client struct {
	AuthUrl string
	Client  *http.Client
}

type token struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
}

// ClientCredentialGrant requests a token using client_credentials grant type
func (u *Client) ClientCredentialGrant(clientId, clientSecret string) (string, error) {
	values := url.Values{
		"grant_type":    {"client_credentials"},
		"response_type": {"token"},
		"client_id":     {clientId},
		"client_secret": {clientSecret},
	}

	token, err := u.tokenGrantRequest(values)

	return token.AccessToken, err
}

// PasswordGrant requests an access token and refresh token using password grant type
func (u *Client) PasswordGrant(clientId, clientSecret, username, password string) (string, string, error) {
	values := url.Values{
		"grant_type":    {"password"},
		"response_type": {"token"},
		"username":      {username},
		"password":      {password},
		"client_id":     {clientId},
		"client_secret": {clientSecret},
	}

	token, err := u.tokenGrantRequest(values)

	return token.AccessToken, token.RefreshToken, err
}

// RefreshTokenGrant requests a new access token and refresh token using refresh_token grant type
func (u *Client) RefreshTokenGrant(clientId, clientSecret, refreshToken string) (string, string, error) {
	values := url.Values{
		"grant_type":    {"refresh_token"},
		"response_type": {"token"},
		"client_id":     {clientId},
		"client_secret": {clientSecret},
		"refresh_token": {refreshToken},
	}

	token, err := u.tokenGrantRequest(values)

	return token.AccessToken, token.RefreshToken, err
}

func (u *Client) tokenGrantRequest(headers url.Values) (token, error) {
	request, _ := http.NewRequest("POST", u.AuthUrl+"/oauth/token", bytes.NewBufferString(headers.Encode()))
	request.Header.Add("Accept", "application/json")
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	response, err := u.Client.Do(request)

	var t token

	if err != nil {
		return t, err
	}

	defer response.Body.Close()

	decoder := json.NewDecoder(response.Body)
	err = decoder.Decode(&t)

	return t, err
}

// RevokeToken revokes the given access token
func (u *Client) RevokeToken(accessToken string) error {
	segments := strings.Split(accessToken, ".")

	if len(segments) < 2 {
		return errors.New("access token missing segments")
	}

	jsonPayload, err := base64.RawURLEncoding.DecodeString(segments[1])

	if err != nil {
		return errors.New("could not base64 decode token payload")
	}

	payload := make(map[string]interface{})
	json.Unmarshal(jsonPayload, &payload)
	jti, ok := payload["jti"].(string)

	if !ok {
		return errors.New("could not parse jti from payload")
	}

	request, _ := http.NewRequest(http.MethodDelete, u.AuthUrl+"/oauth/token/revoke/"+jti, nil)
	request.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := u.Client.Do(request)

	if err == nil {
		defer resp.Body.Close()
	}

	return err
}
