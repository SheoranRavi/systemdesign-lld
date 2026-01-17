package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/sheoranravi/flipkart-p2p/internal/entity"
	"github.com/sheoranravi/flipkart-p2p/internal/service"
	"github.com/sheoranravi/flipkart-p2p/internal/store"
)

func main() {
	// ---- wiring ----
	memStore := store.NewInMemoryStore()

	customerService := service.NewCustomerService(memStore)
	orderService := service.NewOrderService(memStore)
	driverService := service.NewDriverService(memStore, memStore)

	assignmentService := service.NewAssignmentService(orderService, driverService)
	assignmentService.Start()
	defer assignmentService.Stop()

	fmt.Println("System started. Waiting for commands.")
	allItemTypes := memStore.ListItemTypes()
	fmt.Println("Select item type from these:", allItemTypes)

	// ---- input loop ----
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		handleCommand(
			line,
			customerService,
			driverService,
			orderService,
		)
	}
}

func handleCommand(
	input string,
	customerService *service.CustomerService,
	driverService *service.DriverService,
	orderService *service.OrderService,
) {
	parts := strings.Split(input, " - ")
	fmt.Println(parts)
	if len(parts) != 2 {
		fmt.Println("invalid command format")
		return
	}

	command := strings.TrimSpace(parts[0])
	args := strings.Split(parts[1], ",")

	for i := range args {
		args[i] = strings.TrimSpace(args[i])
	}

	switch command {

	case "onboard customer":
		if len(args) != 1 {
			fmt.Println("usage: onboard customer - name")
			return
		}
		id, err := customerService.Create(args[0])
		printResult(err, id)

	case "onboard driver":
		if len(args) != 1 {
			fmt.Println("usage: onboard driver - name")
			return
		}
		id, err := driverService.Create(args[0], "", "")
		printResult(err, id)

	case "create order":
		if len(args) != 2 {
			fmt.Println("usage: create order - customer_id, item_id")
			return
		}
		id, err := orderService.Create(entity.ItemType(args[1]), args[0], 1)
		printResult(err, id)

	case "cancel order":
		if len(args) != 1 {
			fmt.Println("usage: cancel order - order_id")
			return
		}

		order, err := orderService.CancelOrder(args[0])
		// unassign the driver too
		if order != nil {
			driverService.UnassignOrder(order.DriverId)
		}
		printResult(err, "")

	case "show order status":
		if len(args) != 1 {
			fmt.Println("usage: show order status - order_id")
			return
		}
		order, err := orderService.GetByID(args[0])
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("order %s state: %s\n", order.ID, order.State)

	case "show driver status":
		if len(args) != 1 {
			fmt.Println("usage: show driver status - order_id")
			return
		}

		order, err := orderService.GetByID(args[0])
		if err != nil {
			fmt.Println(err)
			return
		}

		if order.DriverId == "" {
			fmt.Println("order has no driver assigned yet")
			return
		}

		driver, err := driverService.GetByID(order.DriverId)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Printf(
			"order %s -> driver %s | state: %s\n",
			order.ID,
			driver.ID,
			driver.State,
		)

	case "pick up order":
		if len(args) != 2 {
			fmt.Println("usage: pick up order - driver_id, order_id")
			return
		}
		err := driverService.PickUpOrder(args[0])
		printResult(err, "")

	case "complete order":
		if len(args) != 2 {
			fmt.Println("usage: complete order - driver_id, order_id")
			return
		}
		err := driverService.MarkOrderDelivered(args[0])
		printResult(err, "")

	default:
		fmt.Println("unknown command")
	}
}

func printResult(err error, id string) {
	if err != nil {
		fmt.Println("error:", err)
	} else {
		fmt.Println("ok")
		if id != "" {
			fmt.Println("id:", id)
		}
	}
}
