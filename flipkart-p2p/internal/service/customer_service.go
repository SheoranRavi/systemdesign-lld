package service

import (
	"github.com/sheoranravi/systemdesign-lld/flipkart-p2p/internal/entity"
	"github.com/sheoranravi/systemdesign-lld/flipkart-p2p/internal/store"
)

type CustomerService struct {
	customers store.CustomerStore
}

func NewCustomerService(store store.CustomerStore) *CustomerService {
	return &CustomerService{customers: store}
}

func (cs *CustomerService) Create(name string) (string, error) {
	// ToDo check if customer already exists
	customer := entity.Customer{
		Name:  name,
		Email: "email",
		Phone: "phone",
	}
	id, err := cs.customers.SaveCustomer(&customer)
	return id, err
}
