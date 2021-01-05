package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_myIP(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"ip":"86.191.129.53"}`)
	}))
	defer ts.Close()
	ip, err := myIP(ts.URL)
	assert.NoError(t, err)
	assert.Equal(t, "86.191.129.53", ip.String())
}
