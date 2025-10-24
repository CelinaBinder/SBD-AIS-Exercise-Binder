package solution

import (
	"ordersystem/model"
	"time"
)

type DatabaseHandler struct {
	// drinks represent all available drinks
	drinks []model.Drink
	// orders serves as order history
	orders []model.Order
}

// todo
func NewDatabaseHandler() *DatabaseHandler {
	// Init the drinks slice with some test data
	drinks := []model.Drink{
		{Name: "Cola", Price: 3.30, Description: "cold fizzy drink"},
		{Name: "Beer", Price: 4.00, Description: "bitter tasting drink"},
		{Name: "Tea", Price: 3.20, Description: "herbal taste"},
	}
	// Init orders slice with some test data
	orders := []model.Order{
		{DrinkID: 1, CreatedAt: time.Now(), Amount: 1},
		{DrinkID: 2, CreatedAt: time.Now(), Amount: 2},
		{DrinkID: 3, CreatedAt: time.Now(), Amount: 1},
	}

	return &DatabaseHandler{
		drinks: drinks,
		orders: orders,
	}
}

func (db *DatabaseHandler) GetDrinks() []model.Drink {
	return db.drinks
}

func (db *DatabaseHandler) GetOrders() []model.Order {
	return db.orders
}

// todo
func (db *DatabaseHandler) GetTotalledOrders() map[uint64]float64 {
	// calculate total orders

	// key = DrinkID, value = Amount of orders
	// totalledOrders map[uint64]uint64
	totalledOrders := make(map[uint64]float64)

	for _, order := range db.orders {
		totalledOrders[order.DrinkID] += (order.Amount) //changing to float64 for map, otherwise mismatched types here
	}

	return totalledOrders
}

func (db *DatabaseHandler) AddOrder(order *model.Order) {
	// todo
	// add order to db.orders slice
	order.CreatedAt = time.Now()
	db.orders = append(db.orders, *order)
}
