package e2e

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	startupTimeout = 3 * time.Minute
	healthTimeout  = 5 * time.Second
)

func TestMain(m *testing.M) {
	log.Println("running e2e tests")

	manageStack := false
	if _, ok := os.LookupEnv("CI"); !ok {
		manageStack = true

		if err := exec.Command("docker", "compose", "up", "-d", "--build").Run(); err != nil {
			log.Fatal(fmt.Errorf("docker-compose up -d: %w", err))
		}
	}

	// Without m.Run() the process exits zero and reports a green run with no tests executed,
	// so a stack that never becomes available has to exit non-zero.
	code := 1
	if waitForServices() {
		log.Println("e2e tests started")

		code = m.Run()
	} else {
		log.Printf("timeout: services were not available within %s", startupTimeout)
	}

	if manageStack {
		if err := exec.Command("docker", "compose", "down", "-v").Run(); err != nil {
			log.Println(fmt.Errorf("docker-compose down -v: %w", err))
		}
		log.Println("e2e tests finished")
	}

	os.Exit(code)
}

func waitForServices() bool {
	deadline := time.Now().Add(startupTimeout)

	for time.Now().Before(deadline) {
		if allHealthy() {
			return true
		}

		time.Sleep(1 * time.Second)
	}

	return false
}

func allHealthy() bool {
	for _, url := range []string{PublicURL, PrivateURL} {
		res, err := resty.New().
			// The readiness probe is registered on the root router, not under the public API path.
			SetBaseURL(strings.TrimSuffix(url, "/api")).
			SetTimeout(healthTimeout).
			R().
			Get("/health")

		if err != nil {
			log.Println("waiting for health", url, err)
			return false
		}

		// resty returns no error for non-2xx responses and /health is the readiness probe, which
		// answers 503 while the database is still unreachable, so the status has to be inspected.
		if res.StatusCode() != http.StatusOK {
			log.Println("waiting for health", url, res.StatusCode())
			return false
		}
	}

	return true
}
