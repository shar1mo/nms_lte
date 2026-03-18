package netconf

/*
#cgo CFLAGS: -I${SRCDIR}
#cgo CFLAGS: -I${SRCDIR}/../../../.local/include
#cgo LDFLAGS: -L${SRCDIR}/../../../.local/lib -lnetconf2 -lyang
#include "shim.h"
*/
import "C"

import (
	"fmt"
	"unsafe"
)

type client struct {
	ptr *C.ncgo_client_t
}

//go:uintptrescapes
//go:noinline
func Init() error {
	if rc := C.ncgo_client_init(); rc != 0 {
		return fmt.Errorf("ncgo_client_init failed")
	}
	return nil
}

//go:uintptrescapes
//go:noinline
func Destroy() {
	C.ncgo_client_destroy()
}

//go:uintptrescapes
//go:noinline
func ConnectSSH(host string, port uint16, user, pass, schemaPath string) (*client, error) {
	cHost := C.CString(host)
	cUser := C.CString(user)
	cPass := C.CString(pass)
	cSchema := C.CString(schemaPath)
	defer C.free(unsafe.Pointer(cHost))
	defer C.free(unsafe.Pointer(cUser))
	defer C.free(unsafe.Pointer(cPass))
	defer C.free(unsafe.Pointer(cSchema))

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
		defer C.ncgo_string_free(cErr)
		return nil, fmt.Errorf(C.GoString(cErr))
	}

	return &client{ptr: cClient}, nil
}

//go:uintptrescapes
//go:noinline
func Rpc(c *client, rpcType int, rpcContent string) (string, error) {
	cRpcContent := C.CString(rpcContent)
	defer C.free(unsafe.Pointer(cRpcContent))

	cType := C.NC_RPC_TYPE(rpcType)

	var cReply *C.char
	var cErr *C.char

	rc := C.ncgo_rpc(
		c.ptr,
		cType,
		cRpcContent,
		&cReply,
		&cErr,
	)
	if rc != 0 {
		defer C.ncgo_string_free(cErr)
		return "", fmt.Errorf(C.GoString(cErr))
	}

	reply := C.GoString(cReply)
	C.ncgo_string_free(cReply)

	return reply, nil
}

func (c *client) Close() {
	// C.ncgo_client_close(c.ptr)
}
