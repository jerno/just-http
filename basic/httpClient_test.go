package basic

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetRaw(t *testing.T) {
	server := createServer(0)
	defer server.Close()

	testCases := []testCaseDefinition[getRequestArguments, []byte]{
		{
			name:    "Test valid URL",
			args:    getRequestArguments{url: server.URL + "/" + "valid-url"},
			want:    []byte(`{"Cluster_name": "cl1", "Pings": 2}`),
			wantErr: false,
		},
		{
			name:    "Test invalid URL",
			args:    getRequestArguments{url: "ht%$://invalid-url"},
			want:    []byte(""),
			wantErr: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got, err := GetRaw(testCase.args.url)
			if (err != nil) != testCase.wantErr {
				t.Errorf("TestGetRaw() error = %v, wantErr %v", err, testCase.wantErr)
				return
			}
			if !bytes.Equal(got, testCase.want) {
				t.Errorf("TestGetRaw() = %v, want %v", got, testCase.want)
			}
		})
	}
}

func TestGetString(t *testing.T) {
	server := createServer(0)
	defer server.Close()

	testCases := []testCaseDefinition[getRequestArguments, string]{
		{
			name:    "Test valid URL",
			args:    getRequestArguments{url: server.URL + "/" + "valid-url"},
			want:    `{"Cluster_name": "cl1", "Pings": 2}`,
			wantErr: false,
		},
		{
			name:    "Test invalid URL",
			args:    getRequestArguments{url: "ht%$://invalid-url"},
			want:    "",
			wantErr: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got, err := GetString(testCase.args.url)
			if (err != nil) != testCase.wantErr {
				t.Errorf("TestGetString() error = %v, wantErr %v", err, testCase.wantErr)
				return
			}
			if got != testCase.want {
				t.Errorf("TestGetString() = %v, want %v", got, testCase.want)
			}
		})
	}
}

func createServer(delayInMilliseconds int) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		time.Sleep(time.Duration(delayInMilliseconds) * time.Millisecond)
		if req.URL.Path == "/valid-url" {
			rw.Write([]byte(`{"Cluster_name": "cl1", "Pings": 2}`))
		}
		if req.URL.Path == "/valid-post-url" {
			rw.Write([]byte(`{"Cluster_name": "server cluster", "Pings": 202}`))
		}
		if req.URL.Path == "/internal-server-error" {
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Header().Set("Content-Type", "application/json")
			resp := make(map[string]string)
			resp["message"] = "Some Error Occurred"
			jsonResp, err := json.Marshal(resp)
			if err != nil {
				log.Fatalf("Error happened in JSON marshal. Err: %s", err)
			}
			rw.Write(jsonResp)
		}
	}))
	return server
}

type getRequestArguments struct {
	url string
}

type testCaseDefinition[T1 any, T2 any] struct {
	name    string
	args    T1
	want    T2
	wantErr bool
}
