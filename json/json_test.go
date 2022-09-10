package json

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAgainstHttpbin(t *testing.T) {
	server := createServer(0)
	defer server.Close()

	testCases := []testCase[parameters[SampleJson], HttpBinResponse]{
		{
			name: "Test GET to a valid URL",
			parameters: parameters[SampleJson]{
				method: http.MethodGet,
				url:    "http://httpbin.org/get",
			},
			want: HttpBinResponse{
				Url:  "http://httpbin.org/get",
				Args: map[string]interface{}{},
			},
			wantErr: false,
		},
		{
			name: "Test GET to a valid URL with url parameters",
			parameters: parameters[SampleJson]{
				method: http.MethodGet,
				url:    "http://httpbin.org/get",
				options: RequestArguments{
					QueryParams: map[string]string{
						"MyTestParams": "TestValue",
					},
				},
			},
			want: HttpBinResponse{
				Url: "http://httpbin.org/get?MyTestParams=TestValue",
				Args: map[string]interface{}{
					"MyTestParams": "TestValue",
				},
			},
			wantErr: false,
		},
		{
			name: "Test POST to a valid URL",
			parameters: parameters[SampleJson]{
				method: http.MethodPost,
				url:    "http://httpbin.org/post",
				data:   SampleJson{Cluster_name: "Hello server", Pings: 1},
			},
			want: HttpBinResponse{
				Url:  "http://httpbin.org/post",
				Args: map[string]interface{}{},
				Json: SampleJson{Cluster_name: "Hello server", Pings: 1},
			},
			wantErr: false,
		},
		{
			name: "Test PUT to a valid URL",
			parameters: parameters[SampleJson]{
				method: http.MethodPut,
				url:    "http://httpbin.org/put",
				data:   SampleJson{Cluster_name: "Hello server", Pings: 1},
			},
			want: HttpBinResponse{
				Url:  "http://httpbin.org/put",
				Args: map[string]interface{}{},
				Json: SampleJson{Cluster_name: "Hello server", Pings: 1},
			},
			wantErr: false,
		},
		{
			name: "Test DELETE to a valid URL",
			parameters: parameters[SampleJson]{
				method: http.MethodDelete,
				url:    "http://httpbin.org/delete",
				data:   SampleJson{Cluster_name: "Hello server", Pings: 1},
			},
			want: HttpBinResponse{
				Url:  "http://httpbin.org/delete",
				Args: map[string]interface{}{},
				Json: SampleJson{Cluster_name: "Hello server", Pings: 1},
			},
			wantErr: false,
		},
		{
			name: "Test invalid URL",
			parameters: parameters[SampleJson]{
				method: http.MethodGet,
				url:    "ht%$://invalid-url",
			},
			want:    HttpBinResponse{},
			wantErr: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var parsedJsonData HttpBinResponse
			var err error

			switch testCase.parameters.method {
			case http.MethodGet:
				err = Get(testCase.parameters.url, &parsedJsonData, testCase.parameters.options)
			case http.MethodPost:
				err = Post(testCase.parameters.url, testCase.parameters.data, &parsedJsonData, testCase.parameters.options)
			case http.MethodPut:
				err = Put(testCase.parameters.url, testCase.parameters.data, &parsedJsonData, testCase.parameters.options)
			case http.MethodDelete:
				err = Delete(testCase.parameters.url, testCase.parameters.data, &parsedJsonData, testCase.parameters.options)
			}

			if (err != nil) != testCase.wantErr {
				t.Errorf("TestGet() error = %v, wantErr %v", err, testCase.wantErr)
				return
			}
			assert.Equal(t, testCase.want, parsedJsonData)
		})
	}
}

func TestAgainstMockServer(t *testing.T) {
	server := createServer(1000)
	defer server.Close()

	testCases := []testCase[parameters[SampleJson], SampleJson]{
		{
			name: "Test request with a timeout of 1500ms on a server delay of 1000ms",
			parameters: parameters[SampleJson]{
				url:  server.URL + "/" + "valid-post-url",
				data: SampleJson{Cluster_name: "Hello server", Pings: 1},
				options: RequestArguments{
					TimeoutInMilliseconds: 1500,
				},
			},
			want:    SampleJson{Cluster_name: "server cluster", Pings: 202},
			wantErr: false,
		},
		{
			name: "Test request with a timeout of 500ms on a server delay of 1000ms",
			parameters: parameters[SampleJson]{
				url:  server.URL + "/" + "valid-post-url",
				data: SampleJson{Cluster_name: "Hello server", Pings: 1},
				options: RequestArguments{
					TimeoutInMilliseconds: 500,
				},
			},
			want:    SampleJson{},
			wantErr: true,
		},
		{
			name: "Test request WITHOUT a timeout on a server delay of 1000ms",
			parameters: parameters[SampleJson]{
				url:  server.URL + "/" + "valid-post-url",
				data: SampleJson{Cluster_name: "Hello server", Pings: 1},
				options: RequestArguments{
					TimeoutInMilliseconds: 0,
				},
			},
			want:    SampleJson{Cluster_name: "server cluster", Pings: 202},
			wantErr: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var parsedJsonData SampleJson
			err := Post(testCase.parameters.url, testCase.parameters.data, &parsedJsonData, testCase.parameters.options)
			if (err != nil) != testCase.wantErr {
				t.Errorf("TestTimeout() error = %v, wantErr %v", err, testCase.wantErr)
				return
			}
			if parsedJsonData != testCase.want {
				t.Errorf("TestTimeout() = %v, want %v", parsedJsonData, testCase.want)
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

type testCase[T1 any, T2 any] struct {
	name       string
	parameters T1
	want       T2
	wantErr    bool
}

type parameters[T any] struct {
	method  string
	url     string
	options RequestArguments
	data    T
}

type SampleJson struct {
	Cluster_name string `json:"Cluster_name"`
	Pings        int    `json:"Pings"`
}

type HttpBinResponse struct {
	Url  string                 `json:"url"`
	Args map[string]interface{} `json:"args"`
	Json SampleJson             `json:"json"`
}
