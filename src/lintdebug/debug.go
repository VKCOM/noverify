package lintdebug

import "fmt"

var (
	callbacks []func(string)
)

// Send a debug message
func Send(msg string, args ...interface{}) {
	formatted := fmt.Sprintf(msg, args...)
	for _, cb := range callbacks {
		cb(formatted)
	}
}

// Register debug events receiver. There must be only one receiver
func Register(cb func(string)) {
	callbacks = append(callbacks, cb)
}
