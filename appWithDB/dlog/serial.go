package dlog

import (
	"sync/atomic"
)

var globalDebugSerial uint64

// DebugGetSerialNumber for debug serial creator
func DebugGetSerialNumber() uint64 {
	return atomic.AddUint64(&globalDebugSerial, 1)
}
