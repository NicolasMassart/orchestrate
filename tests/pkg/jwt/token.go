package jwt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/go-kit/kit/transport/http/jsonrpc"
)

type accessTokenResponse struct {
	AccessToken string `json:"access_token"`
}

func GenerateJWT(idpURL, clientID, clientSecret, audience string) (string, error) {
	body := new(bytes.Buffer)
	_ = json.NewEncoder(body).Encode(map[string]interface{}{
		"client_id":     clientID,
		"client_secret": clientSecret,
		"audience":      audience,
		"grant_type":    "client_credentials",
	})

	resp, err := http.DefaultClient.Post(idpURL, jsonrpc.ContentType, body)
	if err != nil {
		return "", fmt.Errorf("failed to request JWT token. %s", err.Error())
	}
	defer resp.Body.Close()

	acessToken := &accessTokenResponse{}
	if resp.StatusCode == http.StatusOK {
		if err2 := json.NewDecoder(resp.Body).Decode(acessToken); err2 != nil {
			return "", err2
		}

		return acessToken.AccessToken, nil
	}

	// Read body
	respMsg, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return "", fmt.Errorf(string(respMsg))
}
