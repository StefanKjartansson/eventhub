package eventhub

import (
	"testing"
)

func TestLocalMemoryStore(t *testing.T) {

	d := NewLocalMemoryStore()
	InsertUpdateTest(t, d)
	d.Clear()
	QueryByTest(t, d)
	d.Clear()
	QueryTest(t, d)
	d.Clear()
	AggregateTypeTest(t, d)
}
