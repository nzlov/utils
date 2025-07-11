package utils

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

func Get[T any](url string) (*T, error) {
	return GetHeader[T](url, nil)
}

func GetHeader[T any](url string, h map[string]string) (*T, error) {
	return req[T](url, "GET", nil, h)
}

func req[T any](url, method string, data any, h map[string]string) (*T, error) {
	var req *http.Request
	var err error
	if data != nil {
		db, err1 := json.Marshal(data)
		if err1 != nil {
			return nil, err1
		}
		req, err = http.NewRequest(method, url, bytes.NewReader(db))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}
	if err != nil {
		return nil, err
	}
	for k, v := range h {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	db, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	t := new(T)
	return t, json.Unmarshal(db, t)
}

func Post[T any](url string, data any) (*T, error) {
	return PostHeader[T](url, data, nil)
}

func PostHeader[T any](url string, data any, h map[string]string) (*T, error) {
	return req[T](url, "POST", data, h)
}
