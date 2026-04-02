package ne

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"nms_lte/internal/infra/netconf"
)

const (
	defaultNetconfPort              = 830
	defaultNetconfUsername          = "admin"
	defaultNetconfPassword          = "admin"
	defaultNetconfReconnectInterval = 5 * time.Second
)

type Connector interface {
	Connect(address string) (Session, error)
}

type Session = netconf.RPCClient

type Option func(*Service)

type netconfConnector struct {
	username    string
	password    string
	defaultPort uint16
	schemaPath  string
}

func WithConnector(connector Connector) Option {
	return func(s *Service) {
		s.connector = connector
	}
}

func WithReconnectInterval(interval time.Duration) Option {
	return func(s *Service) {
		if interval > 0 {
			s.reconnectInterval = interval
		}
	}
}

func NewManagedService(store Store) (*Service, error) {
	if err := netconf.Init(); err != nil {
		return nil, err
	}

	service := NewService(
		store,
		WithConnector(newNetconfConnectorFromEnv()),
		WithReconnectInterval(reconnectIntervalFromEnv()),
	)
	service.destroyRuntime = true

	return service, nil
}

func newNetconfConnectorFromEnv() Connector {
	return netconfConnector{
		username:    envOrDefault("NETCONF_USERNAME", defaultNetconfUsername),
		password:    envOrDefault("NETCONF_PASSWORD", defaultNetconfPassword),
		defaultPort: uint16(envIntOrDefault("NETCONF_PORT", defaultNetconfPort)),
		schemaPath:  os.Getenv("NETCONF_SCHEMA_PATH"),
	}
}

func reconnectIntervalFromEnv() time.Duration {
	raw := strings.TrimSpace(os.Getenv("NETCONF_RECONNECT_INTERVAL"))
	if raw == "" {
		return defaultNetconfReconnectInterval
	}

	interval, err := time.ParseDuration(raw)
	if err != nil || interval <= 0 {
		return defaultNetconfReconnectInterval
	}

	return interval
}

func (c netconfConnector) Connect(address string) (Session, error) {
	host, port, err := splitAddress(address, c.defaultPort)
	if err != nil {
		return nil, err
	}

	return netconf.ConnectSSH(host, port, c.username, c.password, c.schemaPath)
}

func splitAddress(address string, defaultPort uint16) (string, uint16, error) {
	address = strings.TrimSpace(address)
	if address == "" {
		return "", 0, fmt.Errorf("address is required")
	}

	if host, port, err := net.SplitHostPort(address); err == nil {
		parsedPort, convErr := strconv.Atoi(port)
		if convErr != nil || parsedPort <= 0 || parsedPort > 65535 {
			return "", 0, fmt.Errorf("invalid port in address %q", address)
		}
		return strings.Trim(host, "[]"), uint16(parsedPort), nil
	}

	return strings.Trim(address, "[]"), defaultPort, nil
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
