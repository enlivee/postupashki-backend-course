package once

import (
	"sync/atomic"

	"primitives/internal/futex"
)

type Once struct {
	state uint32
}

const (
	notStarted = 0 // не начато
	running    = 1 // кто-то выполняет f()
	done       = 2 // готово
)

func (o *Once) Do(f func()) {
	for {
		state := atomic.LoadUint32(&o.state)
		switch state {
		case done: // fast
			return // готово и выходим
		case running:
			futex.Wait(&o.state, state) // ждем пока выполняют f()
		case notStarted:
			if atomic.CompareAndSwapUint32(&o.state, notStarted, running) { // победитель
				defer futex.WakeAll(&o.state) // будим тоже в конце, так как все готово
				defer atomic.StoreUint32(&o.state, done)
				// если у нас получилось занять местечко для выполнения f()
				// то в конце мы переведем в состояние 2 для статуса готовности
				f() // сама функия
				// defer выполнятся снизу вверх в порядке lifo типа стек вызовов
				return
			}
		}
	}
}

func (o *Once) Done() bool {
	return atomic.LoadUint32(&o.state) == done // готово !!
}
