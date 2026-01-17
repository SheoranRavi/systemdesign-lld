package entity

type Driver struct {
	ID            string
	State         DriverState
	Name          string
	ActiveOrderId string
}
