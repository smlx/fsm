package fsm_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/smlx/fsm"
)

const (
	_ fsm.Event = iota
	pushOpen
	pullShut
)
const (
	_ fsm.State = iota
	opened
	closed
)

type Door struct {
	fsm.Machine
	mu        sync.Mutex
	increment uint
}

func (d *Door) Occur(e fsm.Event, inc uint) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.increment = inc
	return d.Machine.Occur(e)
}

// TestFSMClosure tests a more complex FSM with arguments passed to Occur().
func TestFSMClosure(t *testing.T) {
	var openCount, closeCount uint
	door := Door{
		Machine: fsm.Machine{
			State: opened,
			Transitions: []fsm.Transition{
				{
					Src:   opened,
					Dst:   closed,
					Event: pullShut,
				},
				{
					Src:   closed,
					Dst:   opened,
					Event: pushOpen,
				},
			},
			IgnoreUnexpectedEvent: true,
		},
	}
	// define the OnEntry/OnExit maps after declaring the variable so that door
	// can be closed over.
	door.OnEntry = map[fsm.State][]fsm.TransitionFunc{
		opened: {
			func(e fsm.Event) error {
				fmt.Println(e)
				openCount += door.increment
				return nil
			},
		},
		closed: {
			func(e fsm.Event) error {
				fmt.Println(e)
				closeCount += door.increment
				return nil
			},
		},
	}
	door.OnExit = map[fsm.State][]fsm.TransitionFunc{
		opened: {
			func(e fsm.Event) error {
				var i uint
				for i = 0; i < door.increment; i++ {
					fmt.Println("Slam!")
				}
				return nil
			},
		},
		closed: {
			func(e fsm.Event) error {
				var i uint
				for i = 0; i < door.increment; i++ {
					fmt.Println("Creaaak!")
				}
				return nil
			},
		},
	}
	// e is a collection of expected state
	type e struct {
		state      fsm.State
		openCount  uint
		closeCount uint
	}
	var steps = []struct {
		event  fsm.Event
		expect e
	}{
		{event: pushOpen, expect: e{state: opened, openCount: 0, closeCount: 0}},
		{event: pullShut, expect: e{state: closed, openCount: 0, closeCount: 3}},
		{event: pullShut, expect: e{state: closed, openCount: 0, closeCount: 3}},
		{event: pushOpen, expect: e{state: opened, openCount: 7, closeCount: 3}},
		{event: pushOpen, expect: e{state: opened, openCount: 7, closeCount: 3}},
	}
	for i, step := range steps {
		err := door.Occur(step.event, uint(1+i*2))
		if err != nil {
			t.Fatalf("step %d: %v", i, err)
		}
		if door.State != step.expect.state {
			t.Fatalf("step %d: expected %v, got %v", i, step.expect.state,
				door.State)
		}
		if openCount != step.expect.openCount {
			t.Fatalf("step %d: expected %v, got %v", i, step.expect.openCount,
				openCount)
		}
		if closeCount != step.expect.closeCount {
			t.Fatalf("step %d: expected %v, got %v", i, step.expect.closeCount,
				closeCount)
		}
	}
}

// TestFSMSimple tests a simple FSM with no extra arguments to Occur().
func TestFSMSimple(t *testing.T) {
	door := fsm.Machine{
		State: opened,
		Transitions: []fsm.Transition{
			{
				Src:   opened,
				Dst:   closed,
				Event: pullShut,
			},
			{
				Src:   closed,
				Dst:   opened,
				Event: pushOpen,
			},
		},
		OnEntry: map[fsm.State][]fsm.TransitionFunc{
			opened: {
				func(e fsm.Event) error {
					fmt.Println(e)
					return nil
				},
			},
			closed: {
				func(e fsm.Event) error {
					fmt.Println(e)
					return nil
				},
			},
		},
		OnExit: map[fsm.State][]fsm.TransitionFunc{
			opened: {
				func(e fsm.Event) error {
					fmt.Println("Slam!")
					return nil
				},
			},
			closed: {
				func(e fsm.Event) error {
					fmt.Println("Creaaak!")
					return nil
				},
			},
		},
		IgnoreUnexpectedEvent: true,
	}
	var steps = []struct {
		event  fsm.Event
		expect fsm.State
	}{
		{event: pushOpen, expect: opened},
		{event: pullShut, expect: closed},
		{event: pullShut, expect: closed},
		{event: pushOpen, expect: opened},
		{event: pushOpen, expect: opened},
	}
	for i, step := range steps {
		err := door.Occur(step.event)
		if err != nil {
			t.Fatalf("step %d: %v", i, err)
		}
		if door.State != step.expect {
			t.Fatalf("step %d: expected %v, got %v", i, step.expect, door.State)
		}
	}
}
