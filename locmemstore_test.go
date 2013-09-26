package eventhub

import (
	"testing"
)

func TestDummyBackend(t *testing.T) {

	d := NewLocalMemoryStore()
	InsertUpdateTest(t, d)
	d.Clear()
	FilterByTest(t, d)
}
