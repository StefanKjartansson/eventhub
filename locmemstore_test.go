package eventhub

import (
	"testing"
)

func TestLocalMemoryStore(t *testing.T) {

	d := NewLocalMemoryStore()
	c := func() {
		d.Clear()
	}
	RunDataBackendSuite(t, d, c)
}
