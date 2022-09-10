package basic

import (
	"io"
	"net/http"
)

func GetRaw(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()

	bytes, _ := io.ReadAll(resp.Body)
	return bytes, nil
}

func GetString(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bytes, _ := io.ReadAll(resp.Body)
	body := string(bytes)
	return body, nil
}
