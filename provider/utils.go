package provider

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

func GetResponse(url, email, token, method string, data io.Reader) (*[]byte, error) {
	req, err := http.NewRequest(method, url, data)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("X-User-Email", email)
	req.Header.Set("X-User-Token", token)

	timeout := time.Duration(60 * time.Second)
	client := &http.Client{
		Timeout: timeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, _ := ioutil.ReadAll(resp.Body)
	client = nil
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		return &b, nil
	}
	return &b, errors.New(fmt.Sprintf("Error occured. %d. %s", resp.StatusCode, string(b)))
}
