package benchmarks

import (
	"net/http/httptest"
	"testing"
	"time"
)

const (
	focusedRPSConcurrency = 100
	focusedRPSDuration    = time.Second
)

func focusedRPSCases() []benchmarkCase {
	return []benchmarkCase{
		{name: "Zinc", build: buildZincRPSHandler},
		{name: "Chi", build: buildChiRPSHandler},
		{name: "Echo", build: buildEchoRPSHandler},
		{name: "Gin", build: buildGinRPSHandler},
	}
}

func BenchmarkRequestsPerSecondFocused(b *testing.B) {
	for _, bc := range focusedRPSCases() {
		b.Run(bc.name, func(b *testing.B) {
			server := httptest.NewServer(bc.build())
			defer server.Close()
			measureRPS(b, server.URL+"/rps", focusedRPSConcurrency, focusedRPSDuration)
		})
	}
}
