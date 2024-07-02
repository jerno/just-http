[![Go Report Card](https://goreportcard.com/badge/github.com/jerno/just-http)](https://goreportcard.com/report/github.com/jerno/just-http)

# just-http

A simple yet powerful HTTP client

# Usage

`justHttp` offers two separate packeges:

|Package                               |Description                                                             |
|-------                               |-----------                                                             |
|`github.com/jerno/just-http/basic`    |Receive raw and string data                                             |
|`github.com/jerno/just-http/json`     |Send and receive JSON data (automatically converted from structs)       |


## Basic package

- `GetRaw(url string) ([]byte, error)`

    Called with a simple `string url` paramteter, it returns the response body as a `[]byte` slice.

    Example:

    ```go
    got, err := justHttp.GetRaw("http://example.com/get")
    fmt.Printf(res)
    ```

- `GetString(url string) (string, error)`

    Called with a simple `string url` paramteter, it returns the response body as a `string`.

    Example:

    ```go
    got, err := justHttp.GetString("http://example.com/get")
    fmt.Printf(res)
    ```

## Json package

- `Get(url string, parsed *TResponse, argList ...RequestArguments) error`

    Parameters:

    |Parameter        |Type                  |Description                        |
    |-----------------|----------------------|-----------------------------------|
    |**url**          |`string`              |The URL to send the request to     |
    |**parsed**       |`*TResponse`          |A pointer to store the response    |
    |**argList**      |`...RequestArguments` |Any number of reqest arguments     |

    Example:

    ```go
    type SampleHttpBinData struct {
        Url    string `json:"url"`
        Origin string `json:"origin"`
        Args   any    `json:"args"`
        Json   any    `json:"json"`
    }

    url := "https://httpbin.org/get"
    queryParams := map[string]string{"MyParam": "test"}

    var sample SampleHttpBinData
    err := justHttpJson.Get(url, &sample, justHttpJson.RequestArguments{QueryParams: queryParams})
    ```

- `Post(url string, data TData, parsed *TResponse, argList ...RequestArguments) error`
- `Put(url string, data TData, parsed *TResponse, argList ...RequestArguments) error`
- `Delete(url string, data TData, parsed *TResponse, argList ...RequestArguments) error`

    Example:

    ```go
    type SampleHttpBinData struct {
        Url    string `json:"url"`
        Origin string `json:"origin"`
        Args   any    `json:"args"`
        Json   any    `json:"json"`
    }
    type SampleJson struct {
        Cluster_name string `json:"Cluster_name"`
        Pings        int    `json:"Pings"`
    }

    url := "https://httpbin.org/post"
    data := SampleJson{Cluster_name: "Hello server", Pings: 1}
    queryParams := map[string]string{"MyParam": "test"}

    var sample SampleHttpBinData
    err := justHttpJson.Post(url, data, &sample, justHttpJson.RequestArguments{QueryParams: queryParams})
    ```

    Parameters:

    |Parameter        |Type                  |Description                                |
    |-----------------|----------------------|-------------------------------------------|
    |**url**          |`string`              |The URL to send the request to             |
    |**data**         |`TData`               |Payload to send (will be JSON encoded)     |
    |**parsed**       |`*TResponse`          |A pointer to store the response            |
    |**argList**      |`...RequestArguments` |Any number of reqest arguments             |

## RequestArguments

```go
type RequestArguments struct {
    TimeoutInMilliseconds int
    QueryParams           map[string]string
}
```

Fields:

|Field name                |Type                |Description                                 |
|--------------------------|--------------------|--------------------------------------------|
|**TimeoutInMilliseconds** |`int`               |Timeout after the request will be cancelled |
|**QueryParams**           |`map[string]string` |URL parameters with string keys             |
