package buildinfo

// Global build variables - populated at build time via -ldflags
var (
	BuildVersion string
	BuildDate    string
	BuildCommit  string
)

// GetVersion returns build version or "N/A" if not set
func GetVersion() string {
	if BuildVersion == "" {
		return "N/A"
	}
	return BuildVersion
}

// GetDate returns build date or "N/A" if not set
func GetDate() string {
	if BuildDate == "" {
		return "N/A"
	}
	return BuildDate
}

// GetCommit returns build commit or "N/A" if not set
func GetCommit() string {
	if BuildCommit == "" {
		return "N/A"
	}
	return BuildCommit
}
