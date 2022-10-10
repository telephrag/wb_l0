package cache

import (
	"errors"
	"fmt"
	"l0/service/model"
	"sync"
)

var ErrDuplicate = errors.New("records are immutable and can't be updated")

type OrdersCache struct {
	data map[string]*model.Order
	mu   sync.RWMutex
	cap  int
}

func (oc *OrdersCache) Init(capacity int) (self *OrdersCache) {
	oc.data = make(map[string]*model.Order)
	oc.mu = sync.RWMutex{}
	oc.cap = capacity

	return oc
}

func (oc *OrdersCache) Cap() int {
	return oc.cap
}

func (oc *OrdersCache) Len() int {
	return len(oc.data)
}

func (oc *OrdersCache) Get(k string) (v *model.Order, ok bool) {
	oc.mu.RLock()
	defer oc.mu.RUnlock()

	v, ok = oc.data[k]
	return v, ok
}

func (oc *OrdersCache) Set(k string, v *model.Order) error {
	oc.mu.Lock()
	defer oc.mu.Unlock()

	if _, ok := oc.data[k]; ok {
		return ErrDuplicate
	}

	if len(oc.data) == oc.cap { // delete latest record if cache is full
		var latestorder string
		for k := range oc.data {
			latestorder = k
			break
		}

		for k := range oc.data {
			if oc.data[k].DateCreated.Unix() <= oc.data[latestorder].DateCreated.Unix() {
				latestorder = k
			}
		}
		delete(oc.data, latestorder)
	}

	oc.data[k] = v

	return nil
}

// Sets underlying map to provided value.
// Meant to be used for restoring cache contents on service startup.
func (oc *OrdersCache) SetData(data map[string]*model.Order) {
	oc.mu.Lock()
	defer oc.mu.Unlock()

	oc.data = data
}

func (oc *OrdersCache) PrintIDs() {
	oc.mu.RLock()
	defer oc.mu.RUnlock()

	for _, v := range oc.data {
		fmt.Println(v.OrderUID, v.DateCreated.UnixMicro())
	}
}

// select * from public.order order by public.order.record->'date_created' desc limit 10;
// won't error if table is empty
