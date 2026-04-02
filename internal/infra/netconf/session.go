package netconf

/*
#cgo CFLAGS: -I${SRCDIR}
#cgo CFLAGS: -I${SRCDIR}/../../../.local/include
#cgo LDFLAGS: -L${SRCDIR}/../../../.local/lib -lnetconf2 -lyang
#include "shim.h"
*/
import "C"

import "strings"

type Session interface {
	Capabilities() ([]string, error)
	IsAlive() bool
	Close()
}

func (c *Client) Capabilities() ([]string, error) {
	if c == nil || c.ptr == nil {
		return nil, errNilClient
	}

	var cCapabilities *C.char
	var cErr *C.char

	if rc := C.ncgo_session_capabilities(c.ptr, &cCapabilities, &cErr); rc != 0 {
		return nil, cgoError(cErr)
	}
	if cCapabilities == nil {
		return nil, nil
	}

	raw := C.GoString(cCapabilities)
	C.ncgo_string_free(cCapabilities)

	lines := strings.Split(raw, "\n")
	capabilities := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		capabilities = append(capabilities, line)
	}

	return capabilities, nil
}

func (c *Client) IsAlive() bool {
	if c == nil || c.ptr == nil {
		return false
	}

	return C.ncgo_session_is_alive(c.ptr) != 0
}

var _ Session = (*Client)(nil)
