package logging

import "os"

// Struct to hold the filepath string and os.File objects for a log file
// This is needed to simlify the File creation and reference process
type LogFile struct {
	Path string
	File *os.File
}
