package tests

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/Mathious6/httpkit"
	"github.com/Mathious6/httpkit/profiles"
	http "github.com/bogdanfinn/fhttp"
	"github.com/stretchr/testify/assert"
)

func TestClient_RandomExtensionOrderChrome(t *testing.T) {
	options := []httpkit.HttpClientOption{
		httpkit.WithClientProfile(profiles.Chrome_107),
		httpkit.WithRandomTLSExtensionOrder(),
	}

	client, err := httpkit.NewHttpClient(nil, options...)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodGet, peetApiEndpoint, nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	defer resp.Body.Close()

	readBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	tlsApiResponse := TlsApiResponse{}
	if err := json.Unmarshal(readBytes, &tlsApiResponse); err != nil {
		t.Fatal(err)
	}

	// All Extensions have to occur in random order. Grease and Padding are staying in place
	extensions := strings.Split("5-0-35-16-18-10-23-65281-43-51-27-17513-45-13-11-21", "-")

	ja3String := tlsApiResponse.TLS.Ja3
	ja3StringParts := strings.Split(ja3String, ",")

	returnedExtensions := ja3StringParts[2]

	for _, extension := range extensions {
		assert.Contains(t, returnedExtensions, extension, fmt.Sprintf("extension %s is not part of %s", extension, returnedExtensions))
	}

	returnedExtensionParts := strings.Split(returnedExtensions, "-")

	assert.Equal(t, "21", returnedExtensionParts[len(returnedExtensionParts)-1])
}

func TestClient_RandomExtensionOrderCustom(t *testing.T) {
	options := []httpkit.HttpClientOption{
		httpkit.WithClientProfile(profiles.CloudflareCustom),
		httpkit.WithRandomTLSExtensionOrder(),
	}

	client, err := httpkit.NewHttpClient(nil, options...)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodGet, peetApiEndpoint, nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	defer resp.Body.Close()

	readBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	tlsApiResponse := TlsApiResponse{}
	if err := json.Unmarshal(readBytes, &tlsApiResponse); err != nil {
		t.Fatal(err)
	}

	// All Extensions have to occur in random order. Grease and Padding are staying in place
	extensions := strings.Split("0-11-10-35-16-22-23-13", "-")

	ja3String := tlsApiResponse.TLS.Ja3
	ja3StringParts := strings.Split(ja3String, ",")

	returnedExtensions := ja3StringParts[2]

	for _, extension := range extensions {
		assert.Contains(t, returnedExtensions, extension, fmt.Sprintf("extension %s is not part of %s", extension, returnedExtensions))
	}
}
