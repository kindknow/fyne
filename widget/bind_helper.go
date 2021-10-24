package widget

import (
	"sync"

	"fyne.io/fyne/v2/data/binding"
)

// basicBinder stores a DataItem and a function to be called when it changes.
// It provides a convenient way to replace data and callback independently.
type basicBinder struct {
	callbackLock         sync.RWMutex
	callback             func(binding.DataItem) // access guarded by callbackLock
	dataListenerPairLock sync.RWMutex
	dataListenerPair     annotatedListener // access guarded by dataListenerPairLock
}

// SetCallback replaces the function to be called when the data changes.
func (binder *basicBinder) SetCallback(f func(data binding.DataItem)) {
	binder.callbackLock.Lock()
	binder.callback = f
	binder.callbackLock.Unlock()
}

// Bind replaces the data item whose changes are tracked by the callback function.
func (binder *basicBinder) Bind(data binding.DataItem) {
	listener := binding.NewDataListener(func() { // NB: listener captures `data` but always calls the up-to-date callback
		binder.callbackLock.RLock()
		f := binder.callback
		binder.callbackLock.RUnlock()
		if f != nil {
			f(data)
		}
	})
	data.AddListener(listener)
	listenerInfo := annotatedListener{
		data:     data,
		listener: listener,
	}

	binder.dataListenerPairLock.Lock()
	binder.unbindLocked()
	binder.dataListenerPair = listenerInfo
	binder.dataListenerPairLock.Unlock()
}

// Unbind requests the callback to be no longer called when the previously bound
// data item changes.
func (binder *basicBinder) Unbind() {
	binder.dataListenerPairLock.Lock()
	binder.unbindLocked()
	binder.dataListenerPairLock.Unlock()
}

// unbindLocked expects the caller to hold dataListenerPairLock.
func (binder *basicBinder) unbindLocked() {
	previousListener := binder.dataListenerPair
	binder.dataListenerPair = annotatedListener{nil, nil}

	if previousListener.listener == nil || previousListener.data == nil {
		return
	}
	previousListener.data.RemoveListener(previousListener.listener)
}

type annotatedListener struct {
	data     binding.DataItem
	listener binding.DataListener
}
