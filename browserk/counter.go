package browserk

import "sync/atomic"

var browserCounter int64

// GetBrowserID a global browser ID 
func GetBrowserID() int64 {
	return atomic.AddInt64(&browserCounter, 1)
}