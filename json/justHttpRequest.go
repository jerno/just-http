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

// Creates a new justHttp Request object
func NewJustHttpRequest[TData any, TResponse any](url string, method string, data TData, parsed *TResponse, argList []RequestArguments) justHttpRequest[TData, TResponse] {
	return justHttpRequest[TData, TResponse]{
		Url:    url,
		Method: method,
		Data:   data,
		Parsed: parsed,
		Args:   getArgs(argList),
	}
}

type justHttpRequest[TData any, TResponse any] struct {
	Url    string
	Method string
	Data   TData
	Args   RequestArguments

	req    *http.Request
	resp   *http.Response
	buffer bytes.Buffer

	Parsed *TResponse
}

// getArgs merges the request arguments passed to the function
// It merges the structs, where latter values take precedence over the previous values (all field will be merged)
func getArgs(args []RequestArguments) RequestArguments {
	merged := defaultArgs
	for _, v := range args {
		mergo.Merge(&v, merged)
		merged = v
	}
	return merged
}

// Creates and sends the HTTP request the result will be stored in `parsed *TResponse` field
func (r *justHttpRequest[TData, TResponse]) Process() error {
	encodingError := r.encodeRequestPayloadToBytes()
	if encodingError != nil {
		return encodingError
	}

	cancelContextTimeout, createRequestError := r.createRequestWithTimeout()
	if createRequestError != nil {
		return createRequestError
	}
	if cancelContextTimeout != nil {
		defer cancelContextTimeout()
	}

	r.addBasicAuthHeaders()
	r.addUrlQueryParams()

	sendRequestError := r.sendRequest()
	if sendRequestError != nil {
		return sendRequestError
	}
	defer r.resp.Body.Close()

	httpStatusCodeError := r.handleHttpStatusCodes()
	if httpStatusCodeError != nil {
		return httpStatusCodeError
	}

	responseDecodeError := r.decodeResponsePayload()
	if responseDecodeError != nil {
		return responseDecodeError
	}

	return nil
}

func (r *justHttpRequest[TData, TResponse]) encodeRequestPayloadToBytes() error {
	enc := json.NewEncoder(&r.buffer)
	err := enc.Encode(r.Data)
	if err != nil {
		return err
	}
	return nil
}

func (r *justHttpRequest[TData, TResponse]) addBasicAuthHeaders() {
	if r.Args.BasicAuthCredentials != (BasicAuthCredentials{}) {
		r.req.SetBasicAuth(r.Args.BasicAuthCredentials.User, r.Args.BasicAuthCredentials.Pass)
	}
}

func (r *justHttpRequest[TData, TResponse]) addUrlQueryParams() {
	if len(r.Args.QueryParams) > 0 {
		q := r.req.URL.Query()
		for k, v := range r.Args.QueryParams {
			q.Add(k, v)
		}
		r.req.URL.RawQuery = q.Encode()
	}
}

func (r *justHttpRequest[TData, TResponse]) createRequestWithTimeout() (context.CancelFunc, error) {
	var requestError error
	if r.Args.TimeoutInMilliseconds != 0 {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(r.Args.TimeoutInMilliseconds)*time.Millisecond)
		r.req, requestError = http.NewRequestWithContext(ctx, r.Method, r.Url, &r.buffer)
		return cancel, requestError
	} else {
		r.req, requestError = http.NewRequest(r.Method, r.Url, &r.buffer)
		return nil, requestError
	}
}

func (r *justHttpRequest[TData, TResponse]) sendRequest() error {
	var requestError error
	r.resp, requestError = http.DefaultClient.Do(r.req)
	if requestError != nil {
		isDeadlineExceeded := errors.Is(requestError, context.DeadlineExceeded)
		if isDeadlineExceeded {
			return &JustHttpError{Message: fmt.Sprintf("Time limit (%s) exceeded", time.Duration(r.Args.TimeoutInMilliseconds)*time.Millisecond)}
		}
		return requestError
	}
	return nil
}

func (r *justHttpRequest[TData, TResponse]) handleHttpStatusCodes() error {
	if r.resp.StatusCode == http.StatusUnauthorized {
		return &JustHttpError{Message: fmt.Sprintf("HTTP error %d", r.resp.StatusCode)}
	}
	return nil
}

func (r *justHttpRequest[TData, TResponse]) decodeResponsePayload() error {
	maxSize := r.Args.SizeLimit
	limitedReader := betterLimitReader.New(r.resp.Body, maxSize)
	decoder := json.NewDecoder(limitedReader)
	if err := decoder.Decode(&r.Parsed); err != nil {
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

type JustHttpError struct {
	Message string
}

func (e *JustHttpError) Error() string {
	return e.Message
}
