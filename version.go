package zinc

// Version is the current version of Zinc.
// This should be updated for each release.
const Version = "0.1.2"

// GetVersion returns the current version of Zinc.
func GetVersion() string {
	return Version
}

// GetVersionHeader returns the full version string for HTTP headers.
func GetVersionHeader() string {
	return "Zinc/" + Version
}
