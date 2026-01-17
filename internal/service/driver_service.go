package service

import (
	"errors"

	"github.com/sheoranravi/flipkart-p2p/internal/entity"
	"github.com/sheoranravi/flipkart-p2p/internal/store"
)

type DriverService struct {
	drivers store.DriverStore
	orders  store.OrderStore
}

func NewDriverService(store store.DriverStore, orderStore store.OrderStore) *DriverService {
	return &DriverService{drivers: store, orders: orderStore}
}

func (s *DriverService) Create(name, email, phone string) (string, error) {
	driver := &entity.Driver{
		Name:          name,
		State:         entity.DriverStateFree,
		ActiveOrderId: "",
	}
	id, err := s.drivers.SaveDriver(driver)
	return id, err
}

func (s *DriverService) ListFreeDrivers() ([]*entity.Driver, error) {
	all, err := s.drivers.ListDrivers()
	if err != nil {
		return nil, err
	}

	free := make([]*entity.Driver, 0)
	for _, d := range all {
		if d.State == entity.DriverStateFree {
			free = append(free, d)
		}
	}
	return free, nil
}

func (s *DriverService) AssignOrder(driver *entity.Driver, orderId string) error {
	if driver.State != entity.DriverStateFree {
		return errors.New("driver not free")
	}

	driver.State = entity.DriverStateAssigned
	driver.ActiveOrderId = orderId
	_, err := s.drivers.SaveDriver(driver)
	return err
}

func (s *DriverService) UnassignOrder(driverId string) error {
	d, err := s.drivers.GetDriverByID(driverId)
	if err != nil {
		return err
	}
	d.ActiveOrderId = ""
	d.State = entity.DriverStateFree
	_, err = s.drivers.SaveDriver(d)
	return err
}

func (s *DriverService) PickUpOrder(driverId string) error {
	driver, err := s.drivers.GetDriverByID(driverId)
	if err != nil {
		return err
	}

	if driver.State != entity.DriverStateAssigned {
		return errors.New("driver not assigned to any order")
	}

	order, err := s.orders.GetOrderByID(driver.ActiveOrderId)
	if err != nil {
		return err
	}

	if order.State == entity.OrderStateCanceled {
		// auto-release driver
		driver.State = entity.DriverStateFree
		driver.ActiveOrderId = ""
		_, _ = s.drivers.SaveDriver(driver)

		return errors.New("order was canceled")
	}

	if order.State != entity.OrderStateAssigned {
		return errors.New("order not in assigned state")
	}

	order.State = entity.OrderStatePickedUp

	if _, err := s.orders.SaveOrder(order); err != nil {
		return err
	}

	return nil
}

func (s *DriverService) MarkOrderDelivered(driverId string) error {
	// fetch driver
	driver, err := s.drivers.GetDriverByID(driverId)
	if err != nil {
		return err
	}

	if driver.State != entity.DriverStateAssigned {
		return errors.New("driver has no active order")
	}

	if driver.ActiveOrderId == "" {
		return errors.New("driver has no active order id")
	}

	// fetch order
	order, err := s.orders.GetOrderByID(driver.ActiveOrderId)
	if err != nil {
		return err
	}

	// enforce state rules
	if order.State != entity.OrderStatePickedUp {
		return errors.New("order is not picked up, cannot complete")
	}

	// transition order
	order.State = entity.OrderStateDelivered
	if _, err := s.orders.SaveOrder(order); err != nil {
		return err
	}

	// release driver
	driver.State = entity.DriverStateFree
	driver.ActiveOrderId = ""
	if _, err := s.drivers.SaveDriver(driver); err != nil {
		return err
	}

	return nil
}

func (s *DriverService) GetByID(driverId string) (*entity.Driver, error) {
	return s.drivers.GetDriverByID(driverId)
}
