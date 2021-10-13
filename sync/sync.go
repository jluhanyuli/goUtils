package sync

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"runtime/debug"
	"sync"
	"time"
)

// 仅仅是对函数本身做Panic Recover，自身不会启动携程，需要在携程中调用
func SafeGo(fun func()) {
	defer func() {
		if err := recover(); err != nil {
			stack := string(debug.Stack())
			logrus.Errorf("[SafeGo] Goroutine Recover: %+v, stack is %v", err, stack)
		}
	}()

	fun()
}

func SafeGoInt32(index int32, fun func(index int32)) {
	defer func() {
		if err := recover(); err != nil {
			stack := string(debug.Stack())
			logrus.Errorf("[SafeGo] Goroutine Recover: %+v, stack is %v", err, stack)
		}
	}()

	fun(index)
}

func JustRecover(ctx context.Context) {
	if err := recover(); err != nil {
		stack := fmt.Sprintf("err: %v \n %s", err, string(debug.Stack()))
		logrus.WithContext(ctx).Errorf("[Recover] panic , backtrace:\n %s", stack)
	}
}

func AsyncDo(ctx context.Context, tag string, f func() error) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				logrus.WithContext(ctx).Warnf("[AsyncDo][%s] panic err: %v", tag, err)
			}
		}()

		if err := f(); err != nil {
			logrus.WithContext(ctx).Warnf("[AsyncDo][%s] AsyncDo err: %v", tag, err)
		}
	}()
}

const (
	WaitGroupWaitCodeTimeout = 1
	WaitGroupWaitCodeNormal  = 2
)

type WaitGroupWithTimeout struct {
	wg      sync.WaitGroup
	timeout time.Duration
}

func NewWaitGroupWithTimeout(t time.Duration) *WaitGroupWithTimeout {
	return &WaitGroupWithTimeout{
		timeout: t,
	}
}

func (wg *WaitGroupWithTimeout) WaitWithCode() int {
	c := make(chan struct{})

	go func() {
		defer func() {
			if err := recover(); err != nil {
				logrus.Warnf("panic err: %v", err)
			}
		}()
		defer close(c)
		wg.wg.Wait()
	}()

	select {
	case <-time.After(wg.timeout): //waitGroup time out
		return WaitGroupWaitCodeTimeout
	case <-c:
		return WaitGroupWaitCodeNormal
	}
}

func (wg *WaitGroupWithTimeout) Done() {
	wg.wg.Done()
}

func (wg *WaitGroupWithTimeout) Add(delta int) {
	wg.wg.Add(delta)
}

func (wg *WaitGroupWithTimeout) Wait() {
	wg.WaitWithCode()
}

type WaitGroupWrapper struct {
	sync.WaitGroup
}

func (wg *WaitGroupWrapper) Wrap(f func()) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		SafeGo(f)
	}()
}

func (wg *WaitGroupWrapper) WrapInt32(index int32, f func(index int32)) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		SafeGoInt32(index, f)
	}()
}
