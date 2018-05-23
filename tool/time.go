package tool

import "time"

const (
	KMicTimeFormat = "2006/01/02 15:04:05.000000"
	KDateTimeFormat = "2006_01_02_15:04:05"
)


// Get a formatted Microseconds time
func GetMicTimeFormat() string {
	return time.Now().Format(KMicTimeFormat)
}

func GetDateTime() string {
	return time.Now().Format(KDateTimeFormat)
}
