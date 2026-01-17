package store

import "github.com/sheoranravi/systemdesign-lld/flipkart-p2p/internal/entity"

type DriverStore interface {
	GetDriverByID(id string) (*entity.Driver, error)
	SaveDriver(*entity.Driver) (string, error)
	ListDrivers() ([]*entity.Driver, error)
}

type OrderStore interface {
	GetOrderByID(id string) (*entity.Order, error)
	SaveOrder(*entity.Order) (string, error)
	ListOrders() ([]*entity.Order, error)
}

type CustomerStore interface {
	GetCustomerByID(id string) (*entity.Customer, error)
	SaveCustomer(*entity.Customer) (string, error)
}

type ItemTypeStore interface {
	IsValid(entity.ItemType) bool
	ListItemTypes() []entity.ItemType
}
