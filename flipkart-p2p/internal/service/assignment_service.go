package service

import (
	"log"
	"sync"
	"time"

	"github.com/sheoranravi/systemdesign-lld/flipkart-p2p/internal/entity"
)

type AssignmentService struct {
	orderService  *OrderService
	driverService *DriverService

	orderQueue  []*entity.Order
	driverQueue []*entity.Driver

	mu     sync.Mutex
	ticker *time.Ticker
	stopCh chan struct{}
}

func NewAssignmentService(
	orderService *OrderService,
	driverService *DriverService,
) *AssignmentService {
	return &AssignmentService{
		orderService:  orderService,
		driverService: driverService,
		orderQueue:    make([]*entity.Order, 0),
		driverQueue:   make([]*entity.Driver, 0),
		stopCh:        make(chan struct{}),
	}
}

func (s *AssignmentService) Start() {
	s.ticker = time.NewTicker(5 * time.Second)

	go func() {
		for {
			select {
			case <-s.ticker.C:
				s.runOnce()
			case <-s.stopCh:
				s.ticker.Stop()
				return
			}
		}
	}()
}

func (s *AssignmentService) Stop() {
	close(s.stopCh)
}
func (s *AssignmentService) runOnce() {
	s.refreshQueues()
	s.assignFIFO()
}

func (s *AssignmentService) refreshQueues() {
	s.mu.Lock()
	defer s.mu.Unlock()

	orders, err := s.orderService.ListUnassignedOrders()
	if err != nil {
		log.Println("failed to fetch orders:", err)
		return
	}

	drivers, err := s.driverService.ListFreeDrivers()
	if err != nil {
		log.Println("failed to fetch drivers:", err)
		return
	}

	// FIFO snapshots
	s.orderQueue = orders
	s.driverQueue = drivers
}

func (s *AssignmentService) assignFIFO() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for len(s.orderQueue) > 0 && len(s.driverQueue) > 0 {
		order := s.orderQueue[0]
		driver := s.driverQueue[0]

		// pop FIFO
		s.orderQueue = s.orderQueue[1:]
		s.driverQueue = s.driverQueue[1:]

		err := s.driverService.AssignOrder(driver, order.ID)
		if err != nil {
			log.Println("driver assignment failed:", err)
			continue
		}

		err = s.orderService.AssignDriver(order, driver.ID)
		if err != nil {
			log.Println("order assignment failed:", err)
			continue
		}

		log.Printf(
			"assignment success: order=%s driver=%s\n",
			order.ID,
			driver.ID,
		)
	}
}
