package main

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func expect(t *testing.T, actual, expected interface{}) {
	if expected != actual {
		t.Errorf("expected: %v, got: %v", expected, actual)
	}
}

func TestPostMissingContentType(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	server := httptest.NewServer(handler(ctx))

	resp, err := http.Post(server.URL + "/time", "", nil)
	expect(t, err, nil)
	expect(t, resp.StatusCode, 400)
}

func TestPostWrongContentType(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	server := httptest.NewServer(handler(ctx))

	resp, err := http.Post(server.URL + "/time", "application/json", nil)
	expect(t, err, nil)
	expect(t, resp.StatusCode, 400)
}

func TestPostNotANumber(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	server := httptest.NewServer(handler(ctx))

	resp, err := http.Post(server.URL + "/time", "text/plain", strings.NewReader("asdfg"))
	expect(t, err, nil)
	expect(t, resp.StatusCode, 400)
}

func TestPostNumber(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	server := httptest.NewServer(handler(ctx))

	resp, err := http.Post(server.URL + "/time", "text/plain", strings.NewReader("120"))
	expect(t, err, nil)
	expect(t, resp.StatusCode, 200)
}

func TestPostNegativeNumber(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	server := httptest.NewServer(handler(ctx))

	resp, err := http.Post(server.URL + "/time", "text/plain", strings.NewReader("-120"))
	expect(t, err, nil)
	expect(t, resp.StatusCode, 200)
}

func TestGetHasDefaultValue(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	server := httptest.NewServer(handler(ctx))

	resp, err := http.Get(server.URL + "/time")
	expect(t, err, nil)
	expect(t, resp.StatusCode, 200)
	expect(t, resp.Header.Get("Content-Type"), "text/plain")

	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	expect(t, len(body) > 0, true)
}

func TestGetSameAsPost(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	server := httptest.NewServer(handler(ctx))

	resp, err := http.Post(server.URL + "/time", "text/plain", strings.NewReader("120"))
	expect(t, err, nil)

	resp, err = http.Get(server.URL + "/time")
	expect(t, err, nil)

	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	expect(t, string(body), "120")
}
