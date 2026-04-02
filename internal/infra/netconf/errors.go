package netconf

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrTimeout    = errors.New("netconf timeout")
	ErrReadFailed = errors.New("netconf read failed")
	ErrSendFailed = errors.New("netconf send failed")
)

func IsTimeout(err error) bool {
	return errors.Is(err, ErrTimeout)
}

func IsReadFailed(err error) bool {
	return errors.Is(err, ErrReadFailed)
}

func IsSendFailed(err error) bool {
	return errors.Is(err, ErrSendFailed)
}

func wrapNetconfError(message string) error {
	message = strings.TrimSpace(message)
	if message == "" {
		return errNetconfCall
	}

	switch {
	case strings.Contains(message, "timed out"):
		return fmt.Errorf("%w: %s", ErrTimeout, message)
	case strings.Contains(message, "nc_recv_reply read failed"):
		return fmt.Errorf("%w: %s", ErrReadFailed, message)
	case strings.Contains(message, "nc_send_rpc failed"):
		return fmt.Errorf("%w: %s", ErrSendFailed, message)
	default:
		return errors.New(message)
	}
}
