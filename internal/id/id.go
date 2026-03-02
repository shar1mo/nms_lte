package id

import (
	"fmt"
	"sync/atomic"
	"time"
)

var counter uint64

func New(prefix string) string {
	next := atomic.AddUint64(&counter, 1)
	return fmt.Sprintf("%s-%d-%d", prefix, time.Now().UnixNano(), next)
}
