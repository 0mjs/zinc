package zinc

import (
	"testing"
)

func TestVersion(t *testing.T) {
	// Test version constant
	if Version != "0.063" {
		t.Errorf("Expected Version to be %q, got %q", "0.063", Version)
	}

	// Test GetVersion function
	if v := GetVersion(); v != Version {
		t.Errorf("Expected GetVersion() to return %q, got %q", Version, v)
	}

	// Test GetVersionHeader function
	expected := "Zinc/" + Version
	if v := GetVersionHeader(); v != expected {
		t.Errorf("Expected GetVersionHeader() to return %q, got %q", expected, v)
	}

	// Test version through App
	app := New()
	if v := app.Version(); v != Version {
		t.Errorf("Expected app.Version() to return %q, got %q", Version, v)
	}

	// Test version in DefaultConfig
	if DefaultConfig.ServerHeader != "Zinc/"+Version {
		t.Errorf("Expected DefaultConfig.ServerHeader to be %q, got %q", "Zinc/"+Version, DefaultConfig.ServerHeader)
	}

	if DefaultConfig.AppVersion != Version {
		t.Errorf("Expected DefaultConfig.AppVersion to be %q, got %q", Version, DefaultConfig.AppVersion)
	}
}
