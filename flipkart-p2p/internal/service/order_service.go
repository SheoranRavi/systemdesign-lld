package service

import (
	"errors"

	"github.com/sheoranravi/systemdesign-lld/flipkart-p2p/internal/entity"
	"github.com/sheoranravi/systemdesign-lld/flipkart-p2p/internal/store"
)

type OrderService struct {
	orders    store.OrderStore
	itemTypes store.ItemTypeStore
}

func NewOrderService(store store.OrderStore, itemStore store.ItemTypeStore) *OrderService {
	return &OrderService{orders: store, itemTypes: itemStore}
}

func (s *OrderService) Create(
	itemType entity.ItemType,
	customerId string,
	numItems int,
) (string, error) {
	isValid := s.itemTypes.IsValid(itemType)
	if !isValid {
		return "", errors.New("This item type is not allowed")
	}
	order := &entity.Order{
		Type:       itemType,
		CustomerId: customerId,
		NumItems:   numItems,
		State:      entity.OrderStatePlaced,
	}
	id, err := s.orders.SaveOrder(order)
	return id, err
}

func (s *OrderService) ListUnassignedOrders() ([]*entity.Order, error) {
	all, err := s.orders.ListOrders()
	if err != nil {
		return nil, err
	}

	pending := make([]*entity.Order, 0)
	for _, o := range all {
		if o.State == entity.OrderStatePlaced {
			pending = append(pending, o)
		}
	}
	return pending, nil
}

func (s *OrderService) AssignDriver(order *entity.Order, driverId string) error {
	if order.State != entity.OrderStatePlaced {
		return errors.New("order not assignable")
	}

	order.State = entity.OrderStateAssigned
	order.DriverId = driverId
	_, err := s.orders.SaveOrder(order)
	return err
}

func (s *OrderService) CancelOrder(orderId string) (*entity.Order, error) {
	order, err := s.orders.GetOrderByID(orderId)
	if err != nil {
		return order, err
	}

	if order.State == entity.OrderStatePickedUp {
		return order, errors.New("cannot cancel picked up order")
	}

	order.State = entity.OrderStateCanceled
	_, err = s.orders.SaveOrder(order)

	return order, err
}

func (s *OrderService) GetByID(orderId string) (*entity.Order, error) {
	return s.orders.GetOrderByID(orderId)
}
