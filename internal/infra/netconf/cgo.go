package netconf

/*
#cgo CFLAGS: -I${SRCDIR}
#cgo CFLAGS: -I${SRCDIR}/../../../.local/include
#cgo LDFLAGS: -L${SRCDIR}/../../../.local/lib -lnetconf2 -lyang
#include <stdlib.h>
#include "shim.h"
*/
import "C"

import (
	"errors"
	"unsafe"
)

var (
	errNilClient   = errors.New("nil client")
	errNetconfCall = errors.New("netconf call failed")
)

type Client struct {
	ptr *C.ncgo_client_t
}

func Init() error {
	if rc := C.ncgo_client_init(); rc != 0 {
		return errors.New("ncgo_client_init failed")
	}

	return nil
}

func Destroy() {
	C.ncgo_client_destroy()
}

func ConnectSSH(host string, port uint16, user, pass, schemaPath string) (*Client, error) {
	cHost := CString(host)
	cUser := CString(user)
	cPass := CString(pass)
	cSchema := CString(schemaPath)
	defer freeCString(cHost)
	defer freeCString(cUser)
	defer freeCString(cPass)
	defer freeCString(cSchema)

	var cClient *C.ncgo_client_t
	var cErr *C.char

	rc := C.ncgo_connect_ssh(
		cHost,
		C.uint16_t(port),
		cUser,
		cPass,
		cSchema,
		&cClient,
		&cErr,
	)
	if rc != 0 {
		return nil, cgoError(cErr)
	}

	return &Client{ptr: cClient}, nil
}

func (c *Client) Rpc(rpcType int, rpcContent string) ([]byte, error) {
	cContent := CString(rpcContent)
	defer freeCString(cContent)

	return c.call(func(reply **C.char, callErr **C.char) C.int {
		return C.ncgo_rpc(
			c.ptr,
			C.NC_RPC_TYPE(rpcType),
			cContent,
			reply,
			callErr,
		)
	})
}

func (c *Client) Close() {
	if c == nil || c.ptr == nil {
		return
	}

	C.ncgo_close(c.ptr)
	c.ptr = nil
}

func CString(value string) *C.char {
	if value == "" {
		return nil
	}

	return C.CString(value)
}

func freeCString(value *C.char) {
	if value != nil {
		C.free(unsafe.Pointer(value))
	}
}

func cgoError(cErr *C.char) error {
	if cErr == nil {
		return errNetconfCall
	}

	defer C.ncgo_string_free(cErr)
	return wrapNetconfError(C.GoString(cErr))
}

func (c *Client) call(fn func(reply **C.char, callErr **C.char) C.int) ([]byte, error) {
	if c == nil || c.ptr == nil {
		return nil, errNilClient
	}

	var cReply *C.char
	var cErr *C.char

	if rc := fn(&cReply, &cErr); rc != 0 {
		return nil, cgoError(cErr)
	}

	if cReply == nil {
		return nil, nil
	}

	reply := []byte(C.GoString(cReply))
	C.ncgo_string_free(cReply)

	return reply, nil
}
