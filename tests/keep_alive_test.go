package tests

import (
	"fmt"
	"io"
	"slices"
	"testing"

	"github.com/Mathious6/httpkit"
	"github.com/Mathious6/httpkit/profiles"
	http "github.com/bogdanfinn/fhttp"
	"github.com/bogdanfinn/fhttp/httptest"
	"github.com/stretchr/testify/assert"
)

func TestClient_UseSameConnection(t *testing.T) {
	testServer := getSimpleWebServer()
	testServer.Start()
	defer testServer.Close()

	client, err := httpkit.ProvideDefaultClient(httpkit.NewNoopLogger())
	if err != nil {
		t.Fatal(err)
	}

	endpoint := fmt.Sprintf("%s%s", testServer.URL, "/index")

	var ports []string
	for i := 0; i < 5; i++ {
		req, err := http.NewRequest(http.MethodGet, endpoint, nil)
		if err != nil {
			t.Fatal(err)
		}

		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		responseBody, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)

		if !slices.Contains(ports, string(responseBody)) {
			ports = append(ports, string(responseBody))
		}

		resp.Body.Close()
	}

	assert.Len(t, ports, 1)
}

func TestClient_UseDifferentConnection(t *testing.T) {
	testServer := getSimpleWebServer()
	testServer.Start()
	defer testServer.Close()

	options := []httpkit.HttpClientOption{
		httpkit.WithClientProfile(profiles.Chrome_107),
		httpkit.WithTransportOptions(&httpkit.TransportOptions{
			DisableKeepAlives: true,
		}),
	}

	client, err := httpkit.NewHttpClient(nil, options...)
	if err != nil {
		t.Fatal(err)
	}

	endpoint := fmt.Sprintf("%s%s", testServer.URL, "/index")

	var ports []string
	for i := 0; i < 5; i++ {
		req, err := http.NewRequest(http.MethodGet, endpoint, nil)
		if err != nil {
			t.Fatal(err)
		}

		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		responseBody, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)

		if !slices.Contains(ports, string(responseBody)) {
			ports = append(ports, string(responseBody))
		}

		resp.Body.Close()
	}

	assert.Len(t, ports, 5)
}

func getSimpleWebServer() *httptest.Server {
	var indexHandler = func(w http.ResponseWriter, req *http.Request) {
		fmt.Println("receive a request from:", req.RemoteAddr, req.Header)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(req.RemoteAddr))
	}

	router := http.NewServeMux()
	router.HandleFunc("/index", indexHandler)

	ts := httptest.NewUnstartedServer(router)

	return ts
}
