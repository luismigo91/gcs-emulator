package emulator_test

import (
	"testing"

	"github.com/luismiguelgilolivert/gcs-emulator/pkg/emulator"
)

func ExampleNewServer() {
	srv, err := emulator.NewServer(emulator.Config{
		Port:           9090,
		DefaultProject: "test-project",
		StorageMode:    "memory",
	})
	if err != nil {
		panic(err)
	}

	if err := srv.Start(); err != nil {
		panic(err)
	}

	defer srv.Stop()

	println("Server URL:", srv.URL())
}

func TestServerStartStop(t *testing.T) {
	srv, err := emulator.NewServer(emulator.Config{
		Port:           0,
		DefaultProject: "test-project",
		StorageMode:    "memory",
	})
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	if err := srv.StartTestServer(); err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}

	defer srv.StopTestServer()

	url := srv.URL()
	if url == "" {
		t.Error("Expected server URL to be non-empty")
	}
}
