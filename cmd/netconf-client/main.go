package main

import (
	"fmt"
	"log"
	"nms_lte/internal/infra/netconf"
)

func main() {
	c, err := netconf.ConnectSSH(
		"127.0.0.1",
		10000,
		"admin",
		"admin",
		"/home/nqs/nms_rc/third_party/libnetconf2/modules",
	)
	if err != nil {
		log.Fatalf("Failed to connect to NETCONF server: %v", err)
	}
	defer netconf.Destroy()

	// reply, err := netconf.Rpc(c, 8, "<interfaces xmlns=\"urn:ietf:params:xml:ns:yang:ietf-interfaces\"/>")
	reply, err := netconf.Rpc(c, 2, "<interfaces xmlns=\"urn:ietf:params:xml:ns:yang:ietf-interfaces\"/>")
	if err != nil {
		log.Fatalf("RPC failed: %v", err)
	}

	fmt.Println(reply)

	c.Close()
}

/*
Connection established
Received RPC:
  get-schema
  identifier = "ietf-datastores"
  format = "ietf-netconf-monitoring:yang"
Received RPC:
  get
  filter = "(null)"
    type = "xpath"
    select = "/ietf-yang-library:*"
Received RPC:
  get-config
  running = ""
Received RPC:
  close-session
*/
