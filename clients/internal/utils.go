package clients

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bobinette/papernet/errors"
)

type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

func UserToken(id int, client HTTPClient, baseURL string) (string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/auth/v2/users/%d/token", baseURL, id), nil)
	if err != nil {
		return "", err
	}

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		var callErr struct {
			Message string `json:"message"`
		}
		err := json.NewDecoder(res.Body).Decode(&callErr)
		if err != nil {
			return "", err
		}

		return "", errors.New(fmt.Sprintf("error in call: %v", callErr.Message), errors.WithCode(res.StatusCode))
	}

	var token struct {
		AccessToken string `json:"access_token"`
	}
	err = json.NewDecoder(res.Body).Decode(&token)
	if err != nil {
		return "", err
	}

	return token.AccessToken, nil
}
