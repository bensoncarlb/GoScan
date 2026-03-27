package server_test

import (
	"net/http"
	"testing"

	"github.com/bensoncarlb/GoScan/internal/server"
)

func TestPing(t *testing.T) {
	svr := server.Server{}

	err := svr.New()

	if err != nil {
		t.Fatalf("Setup failed: %s", err)
	}

	err = svr.Start()

	if err != nil {
		t.Fatalf("Start failed: %s", err)
	}

	defer svr.Stop()

	res, err := http.Get("http://localhost:8090/ping")

	if err != nil {
		t.Fatalf("Ping failed: %s", err)
	} else if res.StatusCode != http.StatusOK {
		t.Fatalf("Invalid respnse code: %v", res.StatusCode)
	}
}
