//go:build integration
// +build integration

package testutil

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func JSONRequest(t *testing.T, method, url, token string, payload interface{}) *http.Request {
	t.Helper()

	var body []byte
	var err error
	if payload != nil {
		body, err = json.Marshal(payload)
		require.NoError(t, err)
	}

	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return req
}

func DecodeJSON(t *testing.T, raw []byte, out interface{}) {
	t.Helper()
	require.NoError(t, json.Unmarshal(raw, out))
}
