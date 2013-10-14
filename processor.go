package straumur

import (
	"container/list"
	"regexp"
)

// Processor is a function which modifies the contents
// of the event and emits any errors to the error channel
type Processor func(*Event, chan error)

// processorRoute is a struct containing a regular expression
// used to determine whether a received event is of interest
// or not.
type processorRoute struct {
	match *regexp.Regexp
	f     Processor
}

// Wrapper around a collection of processorRoute structs
type processorList struct {
	inner *list.List
}

// Returns a new processor list
func NewProcessorList() *processorList {
	return &processorList{inner: list.New()}
}

// Registers a new processor for given pattern
func (p *processorList) Register(pattern string, f Processor) error {
	m, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}
	p.inner.PushBack(processorRoute{m, f})
	return nil
}

// Applies the Processor functions in the order they were registered.
func (p *processorList) Process(e *Event, errchan chan error) {
	for x := p.inner.Front(); x != nil; x = x.Next() {
		// asProcessor, ok := maybe?
		asProcessor := x.Value.(processorRoute)
		if asProcessor.match.MatchString(e.Key) {
			asProcessor.f(e, errchan)
		}
	}
}
