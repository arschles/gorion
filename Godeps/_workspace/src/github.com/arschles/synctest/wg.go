package synctest

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrWaitTimeout = errors.New("wait timed out")
)

// WaitGroupWaitChan closes the returned chan immediately after wg.Wait returns. Do not use this func in long-running
// production code. It uses a goroutine to call wg.Wait, so it will never return if wg.Wait never returns.
//
// Intended for use in tests. Example usage:
//
//  var wg sync.WaitGroup
//  wg.Add(1)
//  go func() {
//    time.Sleep(1 * time.Second)
//    wg.Done()
//  }()
//  ch := WaitGroupChan(wg)
//  select {
//  case <-ch:
//    return nil
//  case <- time.After(3 * time.Second):
//    return ErrWaitTimeout
//  }
func WaitGroupChan(wg *sync.WaitGroup) <-chan struct{} {
	ch := make(chan struct{})
	go func() {
		wg.Wait()
		close(ch)
	}()
	return ch
}

// WaitGroupTimeout is a convenience function for WaitGroupChan. Returns ErrWaitTimeout if wg.Wait()
// didn't return before timeout, nil otherwise
func WaitGroupTimeout(wg *sync.WaitGroup, timeout time.Duration) error {
	ch := WaitGroupChan(wg)
	var err error
	select {
	case <-ch:
		err = nil
	case <-time.After(timeout):
		err = ErrWaitTimeout
	}
	return err
}
