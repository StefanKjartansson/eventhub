package eventhub

import (
	"testing"
)

func TestLocalMemoryStore(t *testing.T) {

	d := NewLocalMemoryStore()
	InsertUpdateTest(t, d)
	d.Clear()
	FilterByTest(t, d)
	d.Clear()
	QueryTest(t, d)
}
