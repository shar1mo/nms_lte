package netconf

import "testing"

func TestWrapNetconfErrorClassifiesTimeout(t *testing.T) {
	err := wrapNetconfError("nc_recv_reply timed out")
	if !IsTimeout(err) {
		t.Fatalf("expected timeout error, got %v", err)
	}
}

func TestWrapNetconfErrorClassifiesReadFailure(t *testing.T) {
	err := wrapNetconfError("nc_recv_reply read failed: socket closed")
	if !IsReadFailed(err) {
		t.Fatalf("expected read failure, got %v", err)
	}
}

func TestWrapNetconfErrorClassifiesSendFailure(t *testing.T) {
	err := wrapNetconfError("nc_send_rpc failed: session busy")
	if !IsSendFailed(err) {
		t.Fatalf("expected send failure, got %v", err)
	}
}
