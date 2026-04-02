package inventory

import (
	"errors"
	"fmt"
	"testing"

	"nms_lte/internal/infra/netconf"
	"nms_lte/internal/model"
	"nms_lte/internal/service/ne"
	"nms_lte/internal/store/memory"
)

type fakeRPCProvider struct {
	client netconf.RPCClient
	err    error
	calls  int
}

type fakeRPCClient struct {
	configReply []byte
	stateReply  []byte
	configErr   error
	stateErr    error
	operations  []string
}

func (p *fakeRPCProvider) WithRPCClient(_ string, fn func(netconf.RPCClient) error) error {
	p.calls++
	if p.err != nil {
		return p.err
	}
	if p.client == nil {
		return errors.New("missing fake client")
	}
	return fn(p.client)
}

func (c *fakeRPCClient) Capabilities() ([]string, error) {
	return nil, nil
}

func (c *fakeRPCClient) IsAlive() bool {
	return true
}

func (c *fakeRPCClient) Close() {}

func (c *fakeRPCClient) Get(filter string) ([]byte, error) {
	c.operations = append(c.operations, "get:"+filter)
	if c.stateErr != nil {
		return nil, c.stateErr
	}
	return append([]byte(nil), c.stateReply...), nil
}

func (c *fakeRPCClient) GetConfig(datastore, filter string) ([]byte, error) {
	c.operations = append(c.operations, fmt.Sprintf("get-config:%s:%s", datastore, filter))
	if c.configErr != nil {
		return nil, c.configErr
	}
	return append([]byte(nil), c.configReply...), nil
}

func (c *fakeRPCClient) Edit(string, string) ([]byte, error)                 { return nil, nil }
func (c *fakeRPCClient) Copy(string, string, string, string) ([]byte, error) { return nil, nil }
func (c *fakeRPCClient) Delete(string, string) ([]byte, error)               { return nil, nil }
func (c *fakeRPCClient) Lock(string) ([]byte, error)                         { return nil, nil }
func (c *fakeRPCClient) Unlock(string) ([]byte, error)                       { return nil, nil }
func (c *fakeRPCClient) Commit() ([]byte, error)                             { return nil, nil }
func (c *fakeRPCClient) Discard() ([]byte, error)                            { return nil, nil }
func (c *fakeRPCClient) Cancel(string) ([]byte, error)                       { return nil, nil }
func (c *fakeRPCClient) Validate(string, string) ([]byte, error)             { return nil, nil }
func (c *fakeRPCClient) GetSchema(string, string, string) ([]byte, error)    { return nil, nil }
func (c *fakeRPCClient) Subscribe(string, string, string, string) ([]byte, error) {
	return nil, nil
}
func (c *fakeRPCClient) GetData(string, string) ([]byte, error)  { return nil, nil }
func (c *fakeRPCClient) EditData(string, string) ([]byte, error) { return nil, nil }
func (c *fakeRPCClient) Kill(uint32) ([]byte, error)             { return nil, nil }

func TestSyncUnknownNE(t *testing.T) {
	service := NewService(memory.New(), nil)
	_, err := service.Sync("unknown")
	if err == nil {
		t.Fatal("expected error for unknown network element")
	}
}

func TestSyncReadsRunningConfigAndOperationalData(t *testing.T) {
	store := memory.New()
	neService := ne.NewService(store)

	neItem, err := neService.Register("enb-1", "10.0.0.1", "vendor-a", nil)
	if err != nil {
		t.Fatalf("register ne: %v", err)
	}

	client := &fakeRPCClient{
		configReply: []byte(`
			<rpc-reply xmlns="urn:ietf:params:xml:ns:netconf:base:1.0">
			  <data>
			    <ManagedElement>
			      <id>me-1</id>
			      <name>enb-1</name>
			      <ENBFunction>
			        <id>1</id>
			        <EUtranCellFDD>
			          <id>cell-1</id>
			          <pci>100</pci>
			          <earfcnDL>300</earfcnDL>
			        </EUtranCellFDD>
			      </ENBFunction>
			    </ManagedElement>
			  </data>
			</rpc-reply>`),
		stateReply: []byte(`
			<rpc-reply xmlns="urn:ietf:params:xml:ns:netconf:base:1.0">
			  <data>
			    <ManagedElement>
			      <id>me-1</id>
			      <ENBFunction>
			        <id>1</id>
			        <status>connected</status>
			      </ENBFunction>
			    </ManagedElement>
			  </data>
			</rpc-reply>`),
	}
	provider := &fakeRPCProvider{client: client}
	service := NewService(store, provider)

	snapshot, err := service.Sync(neItem.ID)
	if err != nil {
		t.Fatalf("sync inventory: %v", err)
	}

	if provider.calls != 1 {
		t.Fatalf("expected 1 rpc provider call, got %d", provider.calls)
	}
	if len(snapshot.Objects) != 3 {
		t.Fatalf("expected 3 inventory objects, got %d", len(snapshot.Objects))
	}
	if !equalStrings(client.operations, []string{
		"get-config:running:",
		"get:",
	}) {
		t.Fatalf("unexpected read operations: %#v", client.operations)
	}

	assertInventoryObject(t, snapshot.Objects, "ManagedElement=me-1", "ManagedElement", map[string]string{
		"id":   "me-1",
		"name": "enb-1",
	})
	assertInventoryObject(t, snapshot.Objects, "ManagedElement=me-1,ENBFunction=1", "ENBFunction", map[string]string{
		"id":     "1",
		"status": "connected",
	})
	assertInventoryObject(t, snapshot.Objects, "ManagedElement=me-1,ENBFunction=1,EUtranCellFDD=cell-1", "EUtranCellFDD", map[string]string{
		"id":       "cell-1",
		"pci":      "100",
		"earfcnDL": "300",
	})

	latest, ok := service.GetLatest(neItem.ID)
	if !ok {
		t.Fatal("expected latest snapshot")
	}
	if latest.ID != snapshot.ID {
		t.Fatalf("latest snapshot mismatch: %s != %s", latest.ID, snapshot.ID)
	}
}

func TestSyncReturnsTimeoutError(t *testing.T) {
	store := memory.New()
	neService := ne.NewService(store)

	neItem, err := neService.Register("enb-1", "10.0.0.1", "vendor-a", nil)
	if err != nil {
		t.Fatalf("register ne: %v", err)
	}

	provider := &fakeRPCProvider{
		client: &fakeRPCClient{
			configErr: fmt.Errorf("%w: nc_recv_reply timed out", netconf.ErrTimeout),
		},
	}
	service := NewService(store, provider)

	_, err = service.Sync(neItem.ID)
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if !errors.Is(err, netconf.ErrTimeout) {
		t.Fatalf("expected timeout classification, got %v", err)
	}
}

func TestSyncReturnsReadFailure(t *testing.T) {
	store := memory.New()
	neService := ne.NewService(store)

	neItem, err := neService.Register("enb-1", "10.0.0.1", "vendor-a", nil)
	if err != nil {
		t.Fatalf("register ne: %v", err)
	}

	provider := &fakeRPCProvider{
		client: &fakeRPCClient{
			configReply: []byte(`<data><ManagedElement><id>me-1</id></ManagedElement></data>`),
			stateErr:    fmt.Errorf("%w: nc_recv_reply read failed: socket closed", netconf.ErrReadFailed),
		},
	}
	service := NewService(store, provider)

	_, err = service.Sync(neItem.ID)
	if err == nil {
		t.Fatal("expected read failure")
	}
	if !errors.Is(err, netconf.ErrReadFailed) {
		t.Fatalf("expected read failure classification, got %v", err)
	}
}

func assertInventoryObject(t *testing.T, objects []model.InventoryObject, dn, class string, attrs map[string]string) {
	t.Helper()

	for _, object := range objects {
		if object.DN != dn {
			continue
		}
		if object.Class != class {
			t.Fatalf("unexpected class for %s: %s", dn, object.Class)
		}
		for key, want := range attrs {
			if got := object.Attributes[key]; got != want {
				t.Fatalf("unexpected attribute %s for %s: got %q want %q", key, dn, got, want)
			}
		}
		return
	}

	t.Fatalf("inventory object %s not found", dn)
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
