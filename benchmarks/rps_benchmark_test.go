package benchmarks

import (
	"net/http/httptest"
	"strconv"
	"testing"
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
	for _, concurrency := range throughputConcurrencyLevels {
		concurrency := concurrency
		b.Run("Concurrency"+strconv.Itoa(concurrency), func(b *testing.B) {
			for _, bc := range focusedRPSCases() {
				b.Run(bc.name, func(b *testing.B) {
					server := httptest.NewServer(bc.build())
					defer server.Close()
					measureRPS(b, server.URL+"/rps", concurrency, throughputDuration)
				})
			}
		})
	}
}
