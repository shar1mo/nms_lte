package main

import (
	"fmt"
	"log"

	"nms_lte/internal/infra/netconf"
)

func main() {
	if err := netconf.Init(); err != nil {
		log.Fatalf("Failed to initialize NETCONF client: %v", err)
	}
	defer netconf.Destroy()

	c, err := netconf.ConnectSSH(
		"127.0.0.1",
		10000,
		"admin",
		"admin",
		"../nms_rc/third_party/libnetconf2/modules",
	)
	if err != nil {
		log.Fatalf("Failed to connect to NETCONF server: %v", err)
	}
	defer c.Close()

	reply, err := c.Get("")
	if err != nil {
		log.Fatalf("RPC failed: %v", err)
	}

	res, err := netconf.StringToXml(reply)
	if err != nil {
		log.Fatalf("Failed to parse XML: %v", err)
	}

	fmt.Printf("%+v\n", res)
}
