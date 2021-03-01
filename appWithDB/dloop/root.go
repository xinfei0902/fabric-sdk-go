package dloop

import (
	"time"
)

func init() {
	initContainer()
	initService()
}

// RegisterHandle into time service
func RegisterHandle(n time.Duration, cb TimeHandle) (err error) {
	getContainer().push(n, cb)
	return nil
}

// Start time service
func Start() error {
	return getService().Start(getContainer().get())
}

// Quit the service
func Quit() error {
	return getService().Quit()
}

// Wait the service quit
func Wait() error {
	return getService().Wait()
}

// Clear resource
func Clear() {
	initContainer()
	initService()
}
