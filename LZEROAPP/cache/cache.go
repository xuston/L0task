package cache

import (
	"sync"

	"main.go/models"
)

type Cache struct {
	mu    sync.RWMutex
	store map[string]models.Orders
}

func NewCache() *Cache {
	return &Cache{
		store: make(map[string]models.Orders),
	}
}

// Схраняет информацию о заказе
func (c *Cache) Save(order models.Orders) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store[order.Order_uid] = order
}

// Получает заказ из кэша по ID
func (c *Cache) Get(order_uid string) (models.Orders, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	order, exists := c.store[order_uid]
	return order, exists
}

// Добавляет заказ в кэш
func (c *Cache) Set(key string, value models.Orders) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store[key] = value
}

// Загружает начальные данные в кэш из базы данных
func (c *Cache) LoadFromDB(orders []models.Orders) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, order := range orders {
		c.store[order.Order_uid] = order
	}
}

var (
	cache = make(map[string]models.Orders)
	mu    sync.RWMutex
)
