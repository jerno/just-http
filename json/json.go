package json

import (
	"net/http"
)

const sizeLimit1MB int64 = 1 << 20

var defaultArgs = RequestArguments{
	TimeoutInMilliseconds: 0,
	SizeLimit:             sizeLimit1MB,
}

func Get[TResponse any](url string, parsed *TResponse, argList ...RequestArguments) error {
	err := processRequest(url, http.MethodGet, new(TResponse), parsed, argList...)
	if err != nil {
		return err
	}
	return nil
}

func Post[TData any, TResponse any](url string, data TData, parsed *TResponse, argList ...RequestArguments) error {
	err := processRequest(url, http.MethodPost, data, parsed, argList...)
	if err != nil {
		return err
	}
	return nil
}

func Put[TData any, TResponse any](url string, data TData, parsed *TResponse, argList ...RequestArguments) error {
	err := processRequest(url, http.MethodPut, data, parsed, argList...)
	if err != nil {
		return err
	}
	return nil
}

func Delete[TData any, TResponse any](url string, data TData, parsed *TResponse, argList ...RequestArguments) error {
	err := processRequest(url, http.MethodDelete, data, parsed, argList...)
	if err != nil {
		return err
	}
	return nil
}

func processRequest[TData any, TResponse any](url string, method string, data TData, parsed *TResponse, argList ...RequestArguments) error {
	var request = NewJustHttpRequest(url, method, data, parsed, argList)
	err := request.Process()
	return err
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
