package lru

import (
	"container/list"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"io"
	"io/ioutil"
	"os"
	"sync"
)

/*
 * @Author: Gpp
 * @File:   main.go
 * @Date:   2021/9/13 3:35 下午
 */

// Cache is an LRU cache. It is not safe for concurrent access.
type Cache struct {
	sync.Mutex
	MaxEntries int
	OnEvicted  func(key Key, value interface{})
	ll         *list.List
	cache      map[[16]byte]*list.Element
	save       []byte
}

// A Key may be any value that is comparable. See http://golang.org/ref/spec#Comparison_operators
type Key [16]byte

type entry struct {
	key   Key
	value interface{}
}

func New(maxEntries int) *Cache {
	return &Cache{
		MaxEntries: maxEntries,
		ll:         list.New(),
		cache:      make(map[[16]byte]*list.Element),
		save:       make([]byte, maxEntries*16),
	}
}

func (c *Cache) Add(key Key, value interface{}) {
	c.Lock()
	defer c.Unlock()
	if c.cache == nil {
		c.cache = make(map[[16]byte]*list.Element)
		c.ll = list.New()
	}
	if ee, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ee)
		ee.Value.(*entry).value = value
		return
	}
	ele := c.ll.PushFront(&entry{key, value})
	c.cache[key] = ele
	if c.MaxEntries != 0 && c.ll.Len() > c.MaxEntries {
		c.RemoveOldest()
	}
}

func (c *Cache) Get(key Key) (value interface{}, ok bool) {
	if c.cache == nil {
		return
	}
	if ele, hit := c.cache[key]; hit {
		c.ll.MoveToFront(ele)
		return ele.Value.(*entry).value, true
	}
	return
}

// Remove removes the provided key from the cache.
func (c *Cache) Remove(key Key) {
	if c.cache == nil {
		return
	}
	if ele, hit := c.cache[key]; hit {
		c.removeElement(ele)
	}
}

// RemoveOldest removes the oldest item from the cache.
func (c *Cache) RemoveOldest() {
	if c.cache == nil {
		return
	}
	ele := c.ll.Back()
	if ele != nil {
		c.removeElement(ele)
	}
}

func (c *Cache) removeElement(e *list.Element) {
	c.ll.Remove(e)
	kv := e.Value.(*entry)
	delete(c.cache, kv.key)
	if c.OnEvicted != nil {
		c.OnEvicted(kv.key, kv.value)
	}
}

func (c *Cache) Len() int {
	if c.cache == nil {
		return 0
	}
	return c.ll.Len()
}

func (c *Cache) Clear() {
	if c.OnEvicted != nil {
		for _, e := range c.cache {
			kv := e.Value.(*entry)
			c.OnEvicted(kv.key, kv.value)
		}
	}
	c.ll = nil
	c.cache = nil
}

func (c *Cache) GetCacheStringSlice() []byte {
	c.Lock()
	defer c.Unlock()
	var i int
	for e := range c.cache {
		copy(c.save[i*16:], e[:])
		i++
	}
	return c.save[:i*16]
}

func (c *Cache) SaveFile(logger log.Logger, path string) error {
	saveCache := c.GetCacheStringSlice()

	err := ioutil.WriteFile(path, saveCache, 0644)
	if err != nil {
		_ = level.Error(logger).Log("err", err.Error())
		return err
	}

	return nil

}

func (c *Cache) LoadFile(logger log.Logger, path string) error {
	var f *os.File
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			f, err = os.Create(path)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	defer func() { _ = f.Close() }()
	n, err := f.Read(c.save)
	if err != nil && err != io.EOF {
		_ = level.Error(logger).Log("msg", err)
		return err
	}
	var k [16]byte
	for i := 0; i < n; i += 16 {
		for j := 0; j < 16; j++ {
			k[j] = c.save[i+j]
		}
		c.Add(k, struct{}{})
	}
	return nil
}
