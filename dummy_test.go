package eventhub

import (
	"testing"
)

func TestDummyBackend(t *testing.T) {

	d := NewDummyBackend()
	InsertUpdateTest(t, d)
	d.Clear()
	FilterByTest(t, d)
}
