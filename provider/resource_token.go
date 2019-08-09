package provider

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type Token struct {
	Token string `json:"token"`
}

func GetToken(url, email, pwd string) (string, error) {
	body := strings.NewReader(fmt.Sprintf(`email=%s&pwd=%s`, email, pwd))
	req, err := http.NewRequest("POST", fmt.Sprintf(`%s/p/login_token`, url), body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}

		var t Token
		err = json.Unmarshal(b, &t)
		if err != nil {
			return "", err
		}
		return t.Token, nil
	} else {
		return "", errors.New(fmt.Sprintf("Login status code: %d.", resp.StatusCode))
	}
}
