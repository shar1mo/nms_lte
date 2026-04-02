package ne

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"sort"
	"strings"

	"nms_lte/internal/infra/netconf"
	"nms_lte/internal/model"
)

func (s *Service) ApplyTransaction(neID string, changes []model.InventoryObject) (err error) {
	if _, ok := s.store.GetNE(neID); !ok {
		return errors.New("network element not found")
	}
	if len(changes) == 0 {
		return errors.New("changes are required")
	}

	return s.WithRPCClient(neID, func(client netconf.RPCClient) (err error) {
		if _, err = client.Lock(netconf.DatastoreCandidate); err != nil {
			return fmt.Errorf("lock candidate datastore: %w", err)
		}

		dirty := false
		defer func() {
			if err != nil && dirty {
				if discardErr := discardCandidate(client); discardErr != nil {
					err = errors.Join(err, discardErr)
				}
			}

			if unlockErr := unlockCandidate(client); unlockErr != nil {
				err = errors.Join(err, unlockErr)
			}
		}()

		for _, change := range changes {
			payload, buildErr := buildEditPayload(change)
			if buildErr != nil {
				return buildErr
			}

			dirty = true
			if _, err = client.Edit(netconf.DatastoreCandidate, payload); err != nil {
				return fmt.Errorf("edit candidate datastore for %q: %w", change.Class, err)
			}
		}

		if _, err = client.Validate(netconf.DatastoreCandidate, ""); err != nil {
			return fmt.Errorf("validate candidate datastore: %w", err)
		}

		if _, err = client.Commit(); err != nil {
			return fmt.Errorf("commit candidate datastore: %w", err)
		}

		return nil
	})
}

func (s *Service) WithRPCClient(neID string, fn func(netconf.RPCClient) error) error {
	if fn == nil {
		return errors.New("rpc client callback is required")
	}

	conn, err := s.getManagedConn(neID)
	if err != nil {
		return err
	}

	conn.mu.Lock()
	defer conn.mu.Unlock()

	if conn.client == nil || !conn.client.IsAlive() {
		return errors.New("network element connection is not alive")
	}

	return fn(conn.client)
}

func (s *Service) getManagedConn(neID string) (*managedConn, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	conn, ok := s.conns[neID]
	if !ok || conn == nil {
		return nil, errors.New("no active connection to network element")
	}

	return conn, nil
}

func discardCandidate(client netconf.RPCClient) error {
	if _, err := client.Discard(); err != nil {
		return fmt.Errorf("discard candidate datastore: %w", err)
	}
	return nil
}

func unlockCandidate(client netconf.RPCClient) error {
	if _, err := client.Unlock(netconf.DatastoreCandidate); err != nil {
		return fmt.Errorf("unlock candidate datastore: %w", err)
	}
	return nil
}

func buildEditPayload(change model.InventoryObject) (string, error) {
	className := strings.TrimSpace(change.Class)
	if !isSafeXMLName(className) {
		return "", fmt.Errorf("invalid change class %q", change.Class)
	}
	if len(change.Attributes) == 0 {
		return "", errors.New("change attributes are required")
	}

	keys := make([]string, 0, len(change.Attributes))
	for key := range change.Attributes {
		key = strings.TrimSpace(key)
		if !isSafeXMLName(key) {
			return "", fmt.Errorf("invalid attribute name %q", key)
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var b strings.Builder
	b.WriteString("<config>")
	b.WriteString("<")
	b.WriteString(className)
	b.WriteString(">")

	for _, key := range keys {
		b.WriteString("<")
		b.WriteString(key)
		b.WriteString(">")
		b.WriteString(escapeXMLText(change.Attributes[key]))
		b.WriteString("</")
		b.WriteString(key)
		b.WriteString(">")
	}

	b.WriteString("</")
	b.WriteString(className)
	b.WriteString(">")
	b.WriteString("</config>")

	return b.String(), nil
}

func escapeXMLText(value string) string {
	var buf bytes.Buffer
	_ = xml.EscapeText(&buf, []byte(value))
	return buf.String()
}

func isSafeXMLName(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return false
	}
	return !strings.ContainsAny(value, " \t\r\n<>/&\"'")
}
