// SPDX-License-Identifier: MIT
// Provenance: original

package server

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"syscall"
	"testing"
	"time"
)

func TestListenAddr(t *testing.T) {
	for _, tt := range []struct {
		values []string
		want   string
	}{
		{nil, ":8080"},
		{[]string{"9000"}, ":9000"},
		{[]string{"0"}, ":0"},
		{[]string{"9000", "127.0.0.1"}, "127.0.0.1:9000"},
		{[]string{"9000", "::1"}, "[::1]:9000"},
	} {
		got, err := ListenAddr(tt.values...)
		if err != nil || got != tt.want {
			t.Errorf("ListenAddr(%q) = %q, %v; want %q", tt.values, got, err, tt.want)
		}
	}
	for _, bad := range []string{"", "http", "-1", "65536"} {
		if _, err := ListenAddr(bad); err == nil {
			t.Errorf("ListenAddr(%q): no error", bad)
		}
	}
}

// serveInBackground runs Serve on ln and returns a function that stops it
// and returns Serve's result.
func serveInBackground(t *testing.T, ln net.Listener) func() error {
	t.Helper()
	h, err := New(Config{Render: (&fakeRenderer{}).render})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- Serve(ctx, ln, h, nil) }()
	return func() error {
		cancel()
		select {
		case err := <-done:
			return err
		case <-time.After(10 * time.Second):
			t.Fatal("Serve did not return")
			return nil
		}
	}
}

func checkServerInfo(t *testing.T, url string) {
	t.Helper()
	resp, err := (&http.Client{Timeout: 10 * time.Second}).Get(url + "/serverinfo")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK || string(body) != infoBody {
		t.Errorf("status %d, body %q", resp.StatusCode, body)
	}
}

func TestHTTPServerDefaultPort(t *testing.T) {
	addr, err := ListenAddr()
	if err != nil || addr != ":8080" {
		t.Fatalf("default address %q, %v", addr, err)
	}
	if DefaultPort != 8080 {
		t.Fatalf("DefaultPort %d", DefaultPort)
	}
	// Serve on the default port, on the loopback interface only.
	ln, err := net.Listen("tcp", "127.0.0.1"+addr)
	if errors.Is(err, syscall.EADDRINUSE) {
		t.Skipf("port 8080 is in use: %v", err)
	}
	if err != nil {
		t.Fatal(err)
	}
	stop := serveInBackground(t, ln)
	checkServerInfo(t, "http://127.0.0.1:8080")
	if err := stop(); err != nil {
		t.Errorf("Serve: %v", err)
	}
}

func TestServeOnFreePort(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	stop := serveInBackground(t, ln)
	url := "http://" + ln.Addr().String()
	checkServerInfo(t, url)
	if err := stop(); err != nil {
		t.Errorf("Serve: %v", err)
	}
	if _, err := http.Get(url + "/serverinfo"); err == nil {
		t.Error("the server still answers after the context is done")
	}
}

func TestListenAndServe(t *testing.T) {
	h, err := New(Config{})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	if err := ListenAndServe(ctx, "127.0.0.1:0", h, nil); err != nil {
		t.Errorf("ListenAndServe: %v", err)
	}
	if err := ListenAndServe(context.Background(), "127.0.0.1:http-nope", h, nil); err == nil {
		t.Error("no error for a bad address")
	}
}
