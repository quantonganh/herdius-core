package cache

import (
	"sync"
)

type Item struct {
	Object interface{}
}

type cache struct {
	items     map[string]Item
	mu        sync.RWMutex
	onEvicted func(string, interface{})
}

type Cache struct {
	*cache
}

func (c *cache) Set(k string, x interface{}) {
	c.mu.Lock()
	c.items[k] = Item{
		Object: x,
	}
	c.mu.Unlock()
}

func (c *cache) Get(k string) (interface{}, bool) {
	c.mu.RLock()
	item, found := c.items[k]
	if !found {
		c.mu.RUnlock()
		return nil, false
	}
	c.mu.RUnlock()
	return item.Object, true
}

func (c *cache) Delete(k string) {
	c.mu.Lock()
	delete(c.items, k)
	c.mu.Unlock()
}

func (c *cache) GetAll() map[string]Item {
	return c.items

}
func (c *cache) DeleteAll() {
	c.mu.Lock()
	for k := range c.items {
		delete(c.items, k)
	}
	c.mu.Unlock()
}

func newCache(m map[string]Item) *Cache {
	c := &cache{
		items: m,
	}
	C := &Cache{c}
	return C
}

func New() *Cache {
	items := make(map[string]Item)

	return newCache(items)
}
