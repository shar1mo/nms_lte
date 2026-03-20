package netconf

/*
#cgo CFLAGS: -I${SRCDIR}
#cgo CFLAGS: -I${SRCDIR}/../../../.local/include
#cgo LDFLAGS: -L${SRCDIR}/../../../.local/lib -lnetconf2 -lyang
#include "shim.h"
*/
import "C"

const (
	NC_RPC_UNKNOWN = iota
	NC_RPC_ACT_GENERIC

	/* ietf-netconf */
	NC_RPC_GETCONFIG
	NC_RPC_EDIT
	NC_RPC_COPY
	NC_RPC_DELETE
	NC_RPC_LOCK
	NC_RPC_UNLOCK
	NC_RPC_GET
	NC_RPC_KILL
	NC_RPC_COMMIT
	NC_RPC_DISCARD
	NC_RPC_CANCEL
	NC_RPC_VALIDATE

	/* ietf-netconf-monitoring */
	NC_RPC_GETSCHEMA

	/* notifications */
	NC_RPC_SUBSCRIBE

	/* ietf-netconf-nmda */
	NC_RPC_GETDATA
	NC_RPC_EDITDATA

	/* ietf-subscribed-notifications */
	NC_RPC_ESTABLISHSUB
	NC_RPC_MODIFYSUB
	NC_RPC_DELETESUB
	NC_RPC_KILLSUB

	/* ietf-yang-push */
	NC_RPC_ESTABLISHPUSH
	NC_RPC_MODIFYPUSH
	NC_RPC_RESYNCSUB
)

const (
	NC_DATASTORE_ERROR     = iota /**< error state of functions returning the datastore type */
	NC_DATASTORE_CONFIG           /**< value describing that the datastore is set as config */
	NC_DATASTORE_URL              /**< value describing that the datastore data should be given from the URL */
	NC_DATASTORE_RUNNING          /**< base NETCONF's datastore containing the current device configuration */
	NC_DATASTORE_STARTUP          /**< separated startup datastore as defined in Distinct Startup Capability */
	NC_DATASTORE_CANDIDATE        /**< separated working datastore as defined in Candidate Configuration Capability */
)

const (
	DatastoreConfig    = "config"
	DatastoreURL       = "url"
	DatastoreRunning   = "running"
	DatastoreStartup   = "startup"
	DatastoreCandidate = "candidate"
)

func (c *Client) Get(filter string) ([]byte, error) {
	cFilter := CString(filter)
	defer freeCString(cFilter)

	return c.call(func(reply **C.char, callErr **C.char) C.int {
		return C.ncgo_rpc_get(c.ptr, cFilter, reply, callErr)
	})
}

func (c *Client) GetConfig(datastore, filter string) ([]byte, error) {
	cDatastore := CString(datastore)
	cFilter := CString(filter)
	defer freeCString(cDatastore)
	defer freeCString(cFilter)

	return c.call(func(reply **C.char, callErr **C.char) C.int {
		return C.ncgo_rpc_getconfig(c.ptr, cDatastore, cFilter, reply, callErr)
	})
}

func (c *Client) Edit(datastore, editContent string) ([]byte, error) {
	cDatastore := CString(datastore)
	cEditContent := CString(editContent)
	defer freeCString(cDatastore)
	defer freeCString(cEditContent)

	return c.call(func(reply **C.char, callErr **C.char) C.int {
		return C.ncgo_rpc_edit(c.ptr, cDatastore, cEditContent, reply, callErr)
	})
}

func (c *Client) Copy(target, urlTarget, source, urlOrConfigSource string) ([]byte, error) {
	cTarget := CString(target)
	cURLTarget := CString(urlTarget)
	cSource := CString(source)
	cURLOrConfigSource := CString(urlOrConfigSource)
	defer freeCString(cTarget)
	defer freeCString(cURLTarget)
	defer freeCString(cSource)
	defer freeCString(cURLOrConfigSource)

	return c.call(func(reply **C.char, callErr **C.char) C.int {
		return C.ncgo_rpc_copy(
			c.ptr,
			cTarget,
			cURLTarget,
			cSource,
			cURLOrConfigSource,
			reply,
			callErr,
		)
	})
}

func (c *Client) Delete(target, url string) ([]byte, error) {
	cTarget := CString(target)
	cURL := CString(url)
	defer freeCString(cTarget)
	defer freeCString(cURL)

	return c.call(func(reply **C.char, callErr **C.char) C.int {
		return C.ncgo_rpc_delete(c.ptr, cTarget, cURL, reply, callErr)
	})
}

func (c *Client) Lock(datastore string) ([]byte, error) {
	cDatastore := CString(datastore)
	defer freeCString(cDatastore)

	return c.call(func(reply **C.char, callErr **C.char) C.int {
		return C.ncgo_rpc_lock(c.ptr, cDatastore, reply, callErr)
	})
}

func (c *Client) Unlock(datastore string) ([]byte, error) {
	cDatastore := CString(datastore)
	defer freeCString(cDatastore)

	return c.call(func(reply **C.char, callErr **C.char) C.int {
		return C.ncgo_rpc_unlock(c.ptr, cDatastore, reply, callErr)
	})
}

func (c *Client) Commit() ([]byte, error) {
	return c.call(func(reply **C.char, callErr **C.char) C.int {
		return C.ncgo_rpc_commit(c.ptr, reply, callErr)
	})
}

func (c *Client) Discard() ([]byte, error) {
	return c.call(func(reply **C.char, callErr **C.char) C.int {
		return C.ncgo_rpc_discard(c.ptr, reply, callErr)
	})
}

func (c *Client) Cancel(persistID string) ([]byte, error) {
	cPersistID := CString(persistID)
	defer freeCString(cPersistID)

	return c.call(func(reply **C.char, callErr **C.char) C.int {
		return C.ncgo_rpc_cancel(c.ptr, cPersistID, reply, callErr)
	})
}

func (c *Client) Validate(source, urlOrConfig string) ([]byte, error) {
	cSource := CString(source)
	cURLOrConfig := CString(urlOrConfig)
	defer freeCString(cSource)
	defer freeCString(cURLOrConfig)

	return c.call(func(reply **C.char, callErr **C.char) C.int {
		return C.ncgo_rpc_validate(c.ptr, cSource, cURLOrConfig, reply, callErr)
	})
}

func (c *Client) GetSchema(identifier, version, format string) ([]byte, error) {
	cIdentifier := CString(identifier)
	cVersion := CString(version)
	cFormat := CString(format)
	defer freeCString(cIdentifier)
	defer freeCString(cVersion)
	defer freeCString(cFormat)

	return c.call(func(reply **C.char, callErr **C.char) C.int {
		return C.ncgo_rpc_getschema(c.ptr, cIdentifier, cVersion, cFormat, reply, callErr)
	})
}

func (c *Client) Subscribe(streamName, filter, startTime, stopTime string) ([]byte, error) {
	cStreamName := CString(streamName)
	cFilter := CString(filter)
	cStartTime := CString(startTime)
	cStopTime := CString(stopTime)
	defer freeCString(cStreamName)
	defer freeCString(cFilter)
	defer freeCString(cStartTime)
	defer freeCString(cStopTime)

	return c.call(func(reply **C.char, callErr **C.char) C.int {
		return C.ncgo_rpc_subscribe(c.ptr, cStreamName, cFilter, cStartTime, cStopTime, reply, callErr)
	})
}

func (c *Client) GetData(datastore, filter string) ([]byte, error) {
	cDatastore := CString(datastore)
	cFilter := CString(filter)
	defer freeCString(cDatastore)
	defer freeCString(cFilter)

	return c.call(func(reply **C.char, callErr **C.char) C.int {
		return C.ncgo_rpc_getdata(c.ptr, cDatastore, cFilter, reply, callErr)
	})
}

func (c *Client) EditData(datastore, editContent string) ([]byte, error) {
	cDatastore := CString(datastore)
	cEditContent := CString(editContent)
	defer freeCString(cDatastore)
	defer freeCString(cEditContent)

	return c.call(func(reply **C.char, callErr **C.char) C.int {
		return C.ncgo_rpc_editdata(c.ptr, cDatastore, cEditContent, reply, callErr)
	})
}

func (c *Client) Kill(sessionID uint32) ([]byte, error) {
	return c.call(func(reply **C.char, callErr **C.char) C.int {
		return C.ncgo_rpc_kill(c.ptr, C.uint32_t(sessionID), reply, callErr)
	})
}
