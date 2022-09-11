package json

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/imdario/mergo"
	betterLimitReader "github.com/luciddev13/limit_reader"
)

const sizeLimit1MB int64 = 1 << 20

var defaultArgs = RequestArguments{
	TimeoutInMilliseconds: 0,
	SizeLimit:             sizeLimit1MB,
}

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
	if err != nil {
		return err
	}

	if args.BasicAuthCredentials != (BasicAuthCredentials{}) {
		req.SetBasicAuth(args.BasicAuthCredentials.User, args.BasicAuthCredentials.Pass)
	}

	if len(args.QueryParams) > 0 {
		q := req.URL.Query()
		for k, v := range args.QueryParams {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		isCanceled := errors.Is(err, context.DeadlineExceeded)
		if isCanceled {
			return &JustHttpError{Message: fmt.Sprintf("Time limit (%s) exceeded", time.Duration(args.TimeoutInMilliseconds)*time.Millisecond)}
		}
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return &JustHttpError{Message: fmt.Sprintf("HTTP error %d", resp.StatusCode)}
	}

	maxSize := args.SizeLimit
	limitedReader := betterLimitReader.New(resp.Body, maxSize)
	decoder := json.NewDecoder(limitedReader)
	if err := decoder.Decode(&parsed); err != nil {
		if _, ok := err.(betterLimitReader.ReaderBoundsExceededError); ok {
			// Here we need to see if the error is a ReaderBoundsExceededError to determine the
			// Original Reader had more bytes then we are willing to process
			// Handle too much data error
			return &JustHttpError{Message: fmt.Sprintf("Body limit (%d bytes) exceeded", maxSize)}
		} else {
			// Handle other cases (it may be an EOF, in which case we were able to read
			// all the bytes from the original Reader)
			return err
		}
	}

	return nil
}

// getArgs merges the request arguments passed to the function
// It merges the structs, where latter values take precendence over the previous values (all field will be merged)
func getArgs(args ...RequestArguments) RequestArguments {
	merged := defaultArgs
	for _, v := range args {
		mergo.Merge(&v, merged)
		merged = v
	}
	return merged
}

type RequestArguments struct {
	TimeoutInMilliseconds int
	SizeLimit             int64
	BasicAuthCredentials  BasicAuthCredentials
	QueryParams           map[string]string
}

type BasicAuthCredentials struct {
	User string
	Pass string
}

type JustHttpError struct {
	Message string
}

func (e *JustHttpError) Error() string {
	return e.Message
}
