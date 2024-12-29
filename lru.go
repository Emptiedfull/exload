package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Node struct {
	key  string
	val  []byte
	ttl  time.Time
	prev *Node
	next *Node
}

type Cache struct {
	head     *Node
	tail     *Node
	items    map[string]*Node
	capacity int

	mu     sync.RWMutex
	hashmu sync.RWMutex
}

func NewCache(cap int) *Cache {
	c := &Cache{
		items:    make(map[string]*Node),
		capacity: cap,
		mu:       sync.RWMutex{},
		hashmu:   sync.RWMutex{},
	}

	return c
}

func (c *Cache) addToFront(n *Node) {
	c.mu.Lock()
	n.next = c.head
	n.prev = nil
	if c.head != nil {
		c.head.prev = n
	}
	c.head = n
	if c.tail == nil {
		c.tail = n
	}
	c.mu.Unlock()

}

func (c *Cache) moveToFront(n *Node) {

	if n == c.head {
		return
	}
	c.mu.Lock()
	if n.prev != nil {
		n.prev.next = n.next
	}
	if n.next != nil {
		n.next.prev = n.prev
	}
	if n == c.tail {
		c.tail = n.prev
	}
	c.mu.Unlock()
	c.addToFront(n)
}

func (c *Cache) delNode(n *Node) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if n.prev != nil {
		n.prev.next = n.next
	} else {
		c.head = n.next
	}

	if n.next != nil {
		n.next.prev = n.prev
	} else {
		c.tail = n.prev
	}

	delete(c.items, n.key)

}

func (c *Cache) put(key string, val []byte, ttl time.Duration) {
	expire := time.Now().Add(ttl)
	c.hashmu.RLock()
	temp := c.items[key]
	c.hashmu.RUnlock()

	if temp != nil {
		temp.val = val
		temp.ttl = expire
		c.moveToFront(temp)
		return
	}
	node := &Node{key: key, val: val, ttl: expire}
	c.hashmu.Lock()
	if len(c.items) >= c.capacity {
		c.hashmu.Unlock()
		c.delNode(c.tail)
		c.hashmu.Lock()
	}
	c.items[key] = node
	c.hashmu.Unlock()
	c.addToFront(node)

}

func (c *Cache) get(key string) (hit bool, val []byte) {
	c.hashmu.RLock()
	temp := c.items[key]
	c.hashmu.RUnlock()

	if temp == nil {
		return false, nil
	}

	if time.Now().After(temp.ttl) {

		defer c.delNode(temp)
		return false, nil
	}
	defer c.moveToFront(temp)

	return true, temp.val
}

// func (c *Cache) print() {
// 	current := c.head
// 	for current != nil {
// 		fmt.Printf("Key: %s, Value: %v\n", current.key, current.val)
// 		current = current.next
// 	}
// }

func WriteThrough(r *http.Request, w http.ResponseWriter, cache *Cache, body []byte) {
	r_str, err := reqToStr(r)
	if err != nil {
		fmt.Println("hii")
	}
	NoCache := w.Header().Get("No-Cache")

	ttl := w.Header().Get("ttl")

	w.Write(body)

	if NoCache != "" {

		return
	}

	var expire time.Duration = 10 * time.Second

	if ttl != "" {
		ex, err := strconv.ParseInt(ttl, 10, 64)
		if err != nil {
			expire = time.Duration(ex) * time.Second
		}
	}

	cache.put(r_str, body, expire)

}

func ReadThrough(r *http.Request, c *Cache) (bool, []byte) {

	if r.Method != "GET" {
		return false, nil
	}

	if r.Header[http.CanonicalHeaderKey("No-Cache")] != nil {
		return false, nil
	}

	req, err := reqToStr(r)
	if err != nil {
		return false, nil
	}

	hit, val := c.get(req)
	if !hit {

		return false, nil
	}

	return true, val

}

func reqToStr(r *http.Request) (string, error) {
	var buf strings.Builder

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return "", err
	}
	// Restore the request body for further use
	r.Body = io.NopCloser(strings.NewReader(string(body)))

	buf.WriteString(r.URL.String())

	buf.WriteString(string(body))

	return buf.String(), nil

}
