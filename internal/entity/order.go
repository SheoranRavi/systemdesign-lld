package entity

type Order struct {
	ID         string
	Type       ItemType
	State      OrderState
	NumItems   int
	CustomerId string
	DriverId   string
}
