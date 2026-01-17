package entity

type OrderState string

const (
	OrderStatePlaced    OrderState = "placed"
	OrderStateAssigned  OrderState = "assigned"
	OrderStatePickedUp  OrderState = "pickedup"
	OrderStateDelivered OrderState = "delivered"
	OrderStateCanceled  OrderState = "canceled"
)
