package common

import (
	"encoding/json"
	"errors"
	"io/ioutil"
)

func ReadToken(cfg string) (string, error) {
	b, _ := ioutil.ReadFile(cfg)
	var data struct {
		Token string `json:"token"`
	}
	err := json.Unmarshal(b, &data)

	if err != nil  {
		return "", err
	}
	if data.Token == "" {
		return "", errors.New("token is empty")
	}
	return data.Token, nil
}
