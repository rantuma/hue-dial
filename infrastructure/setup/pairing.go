package setup

import (
	"fmt"
	"time"

	"github.com/rantuma/hue-dial/infrastructure/hue"
)

const (
	pairingTimeout  = 30 * time.Second
	pairingInterval = 2 * time.Second
)

// PairBridge attempts to register with the bridge by polling CreateUser until
// the user presses the link button or the timeout expires.
func PairBridge(adapter *hue.Adapter, deviceType string) (*hue.CreateUserSuccess, error) {
	deadline := time.Now().Add(pairingTimeout)

	for time.Now().Before(deadline) {
		result, err := adapter.CreateUser(deviceType)
		if err == nil {
			return result, nil
		}

		// The bridge returns an error description when the link button has not
		// been pressed yet — keep retrying until timeout.
		time.Sleep(pairingInterval)
	}

	return nil, fmt.Errorf(
		"pairing timed out after %s — press the link button on the bridge",
		pairingTimeout,
	)
}
