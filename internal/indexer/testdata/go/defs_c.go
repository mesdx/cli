package sample

import "models"

// CriticalUserModelCache caches CriticalUserModel instances.
type CriticalUserModelCache struct {
	entries map[int]models.CriticalUserModel
}

// Get retrieves a CriticalUserModel from the cache.
func (c *CriticalUserModelCache) Get(id int) (models.CriticalUserModel, bool) {
	m, ok := c.entries[id]
	return m, ok
}

// Put stores a CriticalUserModel in the cache.
func (c *CriticalUserModelCache) Put(id int, m models.CriticalUserModel) {
	if c.entries == nil {
		c.entries = make(map[int]models.CriticalUserModel)
	}
	c.entries[id] = m
}
