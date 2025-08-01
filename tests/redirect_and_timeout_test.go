package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/Mathious6/httpkit"
	"github.com/Mathious6/httpkit/profiles"
	http "github.com/bogdanfinn/fhttp"
	"github.com/bogdanfinn/fhttp/httptest"
	"github.com/stretchr/testify/assert"
)

func TestClient_RedirectNoFollowWithSwitch(t *testing.T) {
	testServer := getWebServer()
	testServer.Start()
	defer testServer.Close()

	options := []httpkit.HttpClientOption{
		httpkit.WithClientProfile(profiles.Chrome_105),
		httpkit.WithNotFollowRedirects(),
	}

	client, err := httpkit.NewHttpClient(httpkit.NewNoopLogger(), options...)
	if err != nil {
		t.Fatal(err)
	}

	redirectEndpoint := fmt.Sprintf("%s%s", testServer.URL, "/redirect")

	req, err := http.NewRequest(http.MethodGet, redirectEndpoint, nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, http.StatusMovedPermanently, resp.StatusCode)

	client.SetFollowRedirect(true)

	resp, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestClient_RedirectFollowWithSwitch(t *testing.T) {
	testServer := getWebServer()
	testServer.Start()
	defer testServer.Close()

	options := []httpkit.HttpClientOption{
		httpkit.WithClientProfile(profiles.Chrome_105),
	}

	client, err := httpkit.NewHttpClient(httpkit.NewNoopLogger(), options...)
	if err != nil {
		t.Fatal(err)
	}

	redirectEndpoint := fmt.Sprintf("%s%s", testServer.URL, "/redirect")

	req, err := http.NewRequest(http.MethodGet, redirectEndpoint, nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	client.SetFollowRedirect(false)

	resp, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, http.StatusMovedPermanently, resp.StatusCode)
}

func TestClient_TestFailWithTimeout(t *testing.T) {
	testServer := getWebServer()
	testServer.Start()
	defer testServer.Close()

	options := []httpkit.HttpClientOption{
		httpkit.WithClientProfile(profiles.Chrome_105),
		httpkit.WithTimeoutSeconds(3),
	}

	client, err := httpkit.NewHttpClient(httpkit.NewNoopLogger(), options...)
	if err != nil {
		t.Fatal(err)
	}

	redirectEndpoint := fmt.Sprintf("%s%s", testServer.URL, "/timeout")

	req, err := http.NewRequest(http.MethodGet, redirectEndpoint, nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := client.Do(req)

	assert.Nil(t, resp)
	assert.Error(t, err)
}

func getWebServer() *httptest.Server {
	var indexHandler = func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	var timeoutHandler = func(w http.ResponseWriter, req *http.Request) {
		time.Sleep(5 * time.Second)
	}

	var redirectHandler = func(w http.ResponseWriter, req *http.Request) {
		http.Redirect(w, req, "/index", http.StatusMovedPermanently)
	}

	router := http.NewServeMux()
	router.HandleFunc("/timeout", timeoutHandler)
	router.HandleFunc("/redirect", redirectHandler)
	router.HandleFunc("/index", indexHandler)

	ts := httptest.NewUnstartedServer(router)

	return ts
}
