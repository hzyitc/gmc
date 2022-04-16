package main

import (
	"errors"
	"io"
	"net/http"
)

func mnhv1_query(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	str := string(body)
	if str == "Not found" {
		return "", errors.New("Not found")
	}

	return str, nil
}
