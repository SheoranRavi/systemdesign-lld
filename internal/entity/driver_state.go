package entity

type DriverState string

const (
	DriverStateFree     DriverState = "free"
	DriverStateAssigned DriverState = "assigned"
	DriverStatePickedUp DriverState = "pickedup"
)
