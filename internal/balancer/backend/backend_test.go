package backend

import (
	"net/http/httputil"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBackend(t *testing.T) {
	url, err := url.Parse("http://localhost:8080")
	require.NoError(t, err)

	proxy := &httputil.ReverseProxy{}
	backend := NewBackend(url, true, proxy)

	assert.Equal(t, url, backend.URL, "URL should match")
	assert.True(t, backend.IsAlive(), "Backend should be alive")
	assert.Equal(t, proxy, backend.ReverseProxy, "ReverseProxy should match")
}

func TestSetAlive(t *testing.T) {
	url, err := url.Parse("http://localhost:8080")
	require.NoError(t, err)

	backend := NewBackend(url, true, &httputil.ReverseProxy{})

	backend.SetAlive(false)
	assert.False(t, backend.IsAlive(), "Backend should not be alive")

	backend.SetAlive(true)
	assert.True(t, backend.IsAlive(), "Backend should be alive")
}

func TestIsBackendAlive(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "Valid URL",
			url:     "http://google.com:80",
			wantErr: false,
		},
		{
			name:    "Invalid URL",
			url:     "http://nonexistent.local:12345",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, err := url.Parse(tt.url)
			require.NoError(t, err)

			backend := NewBackend(url, true, &httputil.ReverseProxy{})
			err = backend.IsBackendAlive()

			if tt.wantErr {
				assert.Error(t, err, "Expected error for invalid URL")
			} else {
				assert.NoError(t, err, "Expected no error for valid URL")
			}
		})
	}
}

func TestIsAlive(t *testing.T) {
	url, err := url.Parse("http://localhost:8080")
	require.NoError(t, err)

	backend := NewBackend(url, true, &httputil.ReverseProxy{})

	assert.True(t, backend.IsAlive(), "Backend should be alive initially")

	backend.SetAlive(false)
	assert.False(t, backend.IsAlive(), "Backend should not be alive after SetAlive(false)")
}
