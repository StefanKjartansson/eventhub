package straumur

import "testing"

func TestProcessorList(t *testing.T) {

	const tag = "foo"

	pl := NewProcessorList()
	pl.Register("app.user*", func(e *Event, errchan chan error) {
		e.Tags = append(e.Tags, tag)
	})

	e := NewEvent(
		"app.user.login",
		nil,
		nil,
		"",
		3,
		"app",
		[]string{"ns/foo", "ns/moo"},
		nil,
		nil,
		nil)

	errs := make(chan error)

	pl.Process(e, errs)

	if !stringInSlice(tag, e.Tags) {
		t.Fatalf("%s should be in %+v", tag, e.Tags)
	}
}
