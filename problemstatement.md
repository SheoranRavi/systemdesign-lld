
The system should be able to onboard new customers and drivers.
The list of items that can be delivered is preconfigured in the system and is fixed.
Customers should be able to place an order for the delivery of a parcel and also be able to cancel it. 
One driver can pickup only one order at a time.
Orders should be auto-assigned to drivers based on availability. Even If no driver is available, the system should accept the order and assign it to a driver when the driver becomes free. The number of ongoing orders can exceed the number of drivers.
Once an order is assigned to a driver, the driver should be able to pick up the order and also mark the order as delivered after delivery. 
The system should be able to show the status of orders and drivers.
Canceled orders shouldn’t be assigned to the driver. If an assigned order gets canceled the driver shouldn’t be able to pick up the order, the driver should be available for other orders. 
Once a driver picks up an order the order cannot be canceled by the user nor system.
Assume driver is available 24*7. Ignore the travel time. 
Ensure application is thread safe and all concurrency scenarios.
