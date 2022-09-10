package json

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/imdario/mergo"
)

func Get[TResponse any](url string, parsed *TResponse, argList ...RequestArguments) error {
	err := sendRequest(url, http.MethodGet, new(TResponse), parsed, argList...)
	if err != nil {
		return err
	}
	return nil
}

func Post[TData any, TResponse any](url string, data TData, parsed *TResponse, argList ...RequestArguments) error {
	err := sendRequest(url, http.MethodPost, data, parsed, argList...)
	if err != nil {
		return err
	}
	return nil
}

func Put[TData any, TResponse any](url string, data TData, parsed *TResponse, argList ...RequestArguments) error {
	err := sendRequest(url, http.MethodPut, data, parsed, argList...)
	if err != nil {
		return err
	}
	return nil
}

func Delete[TData any, TResponse any](url string, data TData, parsed *TResponse, argList ...RequestArguments) error {
	err := sendRequest(url, http.MethodDelete, data, parsed, argList...)
	if err != nil {
		return err
	}
	return nil
}

func sendRequest[TData any, TResponse any](url string, method string, data TData, parsed *TResponse, argList ...RequestArguments) error {
	var req *http.Request
	var err error

	var buffer bytes.Buffer
	enc := json.NewEncoder(&buffer)
	if err := enc.Encode(data); err != nil {
		return err
	}

	args := getArgs(argList...)
	if args.TimeoutInMilliseconds != 0 {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(args.TimeoutInMilliseconds)*time.Millisecond)
		defer cancel()
		req, err = http.NewRequestWithContext(ctx, method, url, &buffer)
	} else {
		req, err = http.NewRequest(method, url, &buffer)
	}

	if len(args.QueryParams) > 0 {
		q := req.URL.Query()
		for k, v := range args.QueryParams {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&parsed); err != nil {
		return err
	}

	return nil
}

// getArgs merges the request arguments passed to the function
// It merges the structs, where latter values take precendence over the previous values (all field will be merged)
func getArgs(args ...RequestArguments) RequestArguments {
	merged := RequestArguments{}
	for _, v := range args {
		mergo.Merge(&v, merged)
		merged = v
	}
	return merged
}

type RequestArguments struct {
	TimeoutInMilliseconds int
	QueryParams           map[string]string
}
