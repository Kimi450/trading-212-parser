package config

import "time"

// Get Date Time prefix for a file
func GetDateTimePrefixForFile() string {
	return time.Now().Format("temp")
	return time.Now().Format("20060102-150405")
}
