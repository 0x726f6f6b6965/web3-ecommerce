package once

import (
	"sync"
	"sync/atomic"
)

type Once struct {
	sync.Mutex
	done uint32
}

func (o *Once) Do(f func() error) {
	if atomic.LoadUint32(&o.done) == 0 {
		o.doSlow(f)
	}
}

func (o *Once) doSlow(f func() error) {
	o.Lock()
	defer o.Unlock()
	if o.done == 0 {
		if err := f(); err == nil {
			atomic.StoreUint32(&o.done, 1)
		}
	}
}
