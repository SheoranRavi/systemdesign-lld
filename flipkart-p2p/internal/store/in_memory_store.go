package store

import (
	"errors"
	"sync"

	"github.com/google/uuid"
	"github.com/sheoranravi/systemdesign-lld/flipkart-p2p/internal/entity"
)

type InMemoryStore struct {
	customers map[string]*entity.Customer
	orders    map[string]*entity.Order
	drivers   map[string]*entity.Driver
	itemTypes map[entity.ItemType]struct{}
	mu        sync.RWMutex
}

func NewInMemoryStore() *InMemoryStore {
	s := &InMemoryStore{
		orders:    make(map[string]*entity.Order),
		customers: make(map[string]*entity.Customer),
		drivers:   make(map[string]*entity.Driver),
		itemTypes: make(map[entity.ItemType]struct{}),
	}

	// seed reference data
	s.itemTypes["food"] = struct{}{}
	s.itemTypes["electronics"] = struct{}{}
	s.itemTypes["medicine"] = struct{}{}
	s.itemTypes["tools"] = struct{}{}
	s.itemTypes["clothes"] = struct{}{}

	return s
}

func (s *InMemoryStore) SaveOrder(order *entity.Order) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if order.ID == "" {
		order.ID = uuid.NewString()
	}

	s.orders[order.ID] = order
	return order.ID, nil
}

func (s *InMemoryStore) GetOrderByID(id string) (*entity.Order, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	order, ok := s.orders[id]
	if !ok {
		return nil, errors.New("order not found")
	}
	return order, nil
}

func (s *InMemoryStore) ListOrders() ([]*entity.Order, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*entity.Order, 0, len(s.orders))
	for _, o := range s.orders {
		result = append(result, o)
	}
	return result, nil
}

func (s *InMemoryStore) SaveDriver(driver *entity.Driver) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if driver.ID == "" {
		driver.ID = uuid.NewString()
	}

	s.drivers[driver.ID] = driver
	return driver.ID, nil
}

func (s *InMemoryStore) GetDriverByID(id string) (*entity.Driver, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	driver, ok := s.drivers[id]
	if !ok {
		return nil, errors.New("driver not found")
	}
	return driver, nil
}

func (s *InMemoryStore) ListDrivers() ([]*entity.Driver, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*entity.Driver, 0, len(s.drivers))
	for _, o := range s.drivers {
		result = append(result, o)
	}
	return result, nil
}

func (s *InMemoryStore) SaveCustomer(customer *entity.Customer) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if customer.ID == "" {
		customer.ID = uuid.NewString()
	}
	s.customers[customer.ID] = customer
	return customer.ID, nil
}

func (s *InMemoryStore) GetCustomerByID(id string) (*entity.Customer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	customer, ok := s.customers[id]
	if !ok {
		return nil, errors.New("customer not found")
	}
	return customer, nil
}

func (s *InMemoryStore) IsValid(t entity.ItemType) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, ok := s.itemTypes[t]
	return ok
}

func (s *InMemoryStore) ListItemTypes() []entity.ItemType {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]entity.ItemType, 0, len(s.itemTypes))
	for t := range s.itemTypes {
		result = append(result, t)
	}
	return result
}
