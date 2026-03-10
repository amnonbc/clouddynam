package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_myIP(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"ip":"86.191.129.53"}`)
	}))
	defer ts.Close()
	ip, err := myIP(t.Context(), ts.URL)
	assert.NoError(t, err)
	assert.Equal(t, "86.191.129.53", ip.String())
}

func Test_myIP_IPv6(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"ip":"2001:db8::1"}`)
	}))
	defer ts.Close()
	ip, err := myIP(t.Context(), ts.URL)
	assert.NoError(t, err)
	assert.Equal(t, "2001:db8::1", ip.String())
	assert.True(t, ip.Is6())
}

func Test_myIP_ServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}))
	defer ts.Close()
	_, err := myIP(t.Context(), ts.URL)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

func Test_myIP_InvalidJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `not json`)
	}))
	defer ts.Close()
	_, err := myIP(t.Context(), ts.URL)
	assert.Error(t, err)
}

func Test_myIP_InvalidIP(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"ip":"not-an-ip"}`)
	}))
	defer ts.Close()
	_, err := myIP(t.Context(), ts.URL)
	assert.Error(t, err)
}

func Test_loadConfig(t *testing.T) {
	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "config.json")
	err := os.WriteFile(cfgFile, []byte(`{"ApiKey":"mykey","Domains":["example.com","foo.com"]}`), 0600)
	require.NoError(t, err)

	cfg, err := loadConfig(cfgFile)
	assert.NoError(t, err)
	assert.Equal(t, "mykey", cfg.ApiKey)
	assert.Equal(t, []string{"example.com", "foo.com"}, cfg.Domains)
}

func Test_loadConfig_FileNotFound(t *testing.T) {
	_, err := loadConfig("/nonexistent/path/config.json")
	assert.Error(t, err)
}

func Test_loadConfig_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "config.json")
	err := os.WriteFile(cfgFile, []byte(`not json`), 0600)
	require.NoError(t, err)

	_, err = loadConfig(cfgFile)
	assert.Error(t, err)
}

func Test_loadConfig_EmptyDomains(t *testing.T) {
	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "config.json")
	err := os.WriteFile(cfgFile, []byte(`{"ApiKey":"mykey","Domains":[]}`), 0600)
	require.NoError(t, err)

	cfg, err := loadConfig(cfgFile)
	assert.NoError(t, err)
	assert.Equal(t, "mykey", cfg.ApiKey)
	assert.Empty(t, cfg.Domains)
}
