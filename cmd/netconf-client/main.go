package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"nms_lte/internal/infra/netconf"
)

const (
	defaultHost     = "127.0.0.1"
	defaultPort     = 10000
	defaultUsername = "admin"
	defaultPassword = "admin"
)

type config struct {
	host             string
	port             int
	username         string
	password         string
	schemaPath       string
	operation        string
	datastore        string
	filter           string
	showCapabilities bool
}

func main() {
	cfg := parseFlags()

	if err := netconf.Init(); err != nil {
		log.Fatalf("initialize NETCONF client: %v", err)
	}
	defer netconf.Destroy()

	client, err := netconf.ConnectSSH(
		cfg.host,
		uint16(cfg.port),
		cfg.username,
		cfg.password,
		cfg.schemaPath,
	)
	if err != nil {
		failNetconf("connect", err)
	}
	defer client.Close()

	if cfg.showCapabilities {
		capabilities, err := client.Capabilities()
		if err != nil {
			failNetconf("read capabilities", err)
		}
		printCapabilities(capabilities)
	}

	switch cfg.operation {
	case "get":
		reply, err := client.Get(cfg.filter)
		if err != nil {
			failNetconf("get", err)
		}
		printReply("GET reply", reply)
	case "get-config":
		reply, err := client.GetConfig(cfg.datastore, cfg.filter)
		if err != nil {
			failNetconf("get-config", err)
		}
		printReply(fmt.Sprintf("GET-CONFIG reply (%s)", cfg.datastore), reply)
	case "both":
		reply, err := client.GetConfig(cfg.datastore, cfg.filter)
		if err != nil {
			failNetconf("get-config", err)
		}
		printReply(fmt.Sprintf("GET-CONFIG reply (%s)", cfg.datastore), reply)

		reply, err = client.Get(cfg.filter)
		if err != nil {
			failNetconf("get", err)
		}
		printReply("GET reply", reply)
	default:
		log.Fatalf("unsupported operation %q, use get, get-config or both", cfg.operation)
	}
}

func parseFlags() config {
	cfg := config{}

	flag.StringVar(&cfg.host, "host", envOrDefault("NETCONF_HOST", defaultHost), "NETCONF host")
	flag.IntVar(&cfg.port, "port", envIntOrDefault("NETCONF_PORT", defaultPort), "NETCONF SSH port")
	flag.StringVar(&cfg.username, "username", envOrDefault("NETCONF_USERNAME", defaultUsername), "NETCONF username")
	flag.StringVar(&cfg.password, "password", envOrDefault("NETCONF_PASSWORD", defaultPassword), "NETCONF password")
	flag.StringVar(&cfg.schemaPath, "schema-path", strings.TrimSpace(os.Getenv("NETCONF_SCHEMA_PATH")), "schema search path; empty uses built-in libnetconf2 default")
	flag.StringVar(&cfg.operation, "op", "both", "operation to execute: get, get-config, both")
	flag.StringVar(&cfg.datastore, "datastore", netconf.DatastoreRunning, "target datastore for get-config")
	flag.StringVar(&cfg.filter, "filter", "", "NETCONF subtree/xpath filter passed as raw XML/string")
	flag.BoolVar(&cfg.showCapabilities, "capabilities", true, "print server capabilities before reading")
	flag.Parse()

	cfg.operation = strings.TrimSpace(strings.ToLower(cfg.operation))
	cfg.datastore = strings.TrimSpace(strings.ToLower(cfg.datastore))

	if cfg.host == "" {
		log.Fatal("host is required")
	}
	if cfg.port <= 0 || cfg.port > 65535 {
		log.Fatalf("invalid port %d", cfg.port)
	}
	if cfg.username == "" {
		log.Fatal("username is required")
	}
	if cfg.password == "" {
		log.Fatal("password is required")
	}

	return cfg
}

func printCapabilities(capabilities []string) {
	fmt.Println("Capabilities:")
	if len(capabilities) == 0 {
		fmt.Println("(none)")
		fmt.Println()
		return
	}

	for _, capability := range capabilities {
		fmt.Printf("- %s\n", capability)
	}
	fmt.Println()
}

func printReply(title string, reply []byte) {
	fmt.Printf("%s:\n", title)
	if len(reply) == 0 {
		fmt.Println("(empty)")
		fmt.Println()
		return
	}

	fmt.Println(string(reply))
	fmt.Println()
}

func failNetconf(operation string, err error) {
	switch {
	case netconf.IsTimeout(err):
		log.Fatalf("%s timed out: %v", operation, err)
	case netconf.IsReadFailed(err):
		log.Fatalf("%s failed while reading reply: %v", operation, err)
	case netconf.IsSendFailed(err):
		log.Fatalf("%s failed while sending rpc: %v", operation, err)
	default:
		log.Fatalf("%s failed: %v", operation, err)
	}
}

func envOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func envIntOrDefault(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}
