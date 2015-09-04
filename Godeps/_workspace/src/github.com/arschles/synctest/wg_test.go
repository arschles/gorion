package synctest

import (
	"sync"
	"testing"
	"time"
)

func TestWaitGroupChan(t *testing.T) {
	var wg1 sync.WaitGroup
	wg1.Add(1)
	ch := WaitGroupChan(&wg1)
	select {
	case <-ch:
		t.Errorf("chan received when it shouldn't have")
	case <-time.After(100 * time.Millisecond):
	}
	var wg2 sync.WaitGroup
	ch = WaitGroupChan(&wg2)
	select {
	case <-ch:
	case <-time.After(100 * time.Millisecond):
		t.Errorf("chan didn't receive when it should have")
	}
}

func TestWaitGroupTimeout(t *testing.T) {
	var wg1 sync.WaitGroup
	err := WaitGroupTimeout(&wg1, 1*time.Second)
	if err != nil {
		t.Errorf("error not expected")
	}
	var wg2 sync.WaitGroup
	wg2.Add(1)
	err = WaitGroupTimeout(&wg2, 100*time.Millisecond)
	if err == nil {
		t.Errorf("error expected")
	}
}
