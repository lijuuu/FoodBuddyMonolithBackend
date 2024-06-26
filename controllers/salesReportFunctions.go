package controllers

import (
	"errors"
	"foodbuddy/database"
	"foodbuddy/model"
	"time"
)

//total orders  //oi
//total delivered  //oi
//total cancelled //oi
//revenue generated //oi
//coupon deductions //oi
//product offer deductions //oi
//referral rewards given //oi
func TotalOrders(From time.Time, Till time.Time, PaymentStatus string) (model.OrderCount, error) {
    var orders []model.Order

	// From += "T00:00:00Z"

    // Fetch orders within the specified time frame and payment status
    if err := database.DB.Where("ordered_at BETWEEN? AND? AND payment_status =?", From, Till, PaymentStatus).Find(&orders).Error; err!= nil {
        return model.OrderCount{}, errors.New("error fetching orders")
    }

    // Initialize counters map
    var orderStatusCounts = map[string]int64{
        model.OrderStatusProcessing: 0,
        model.OrderStatusInPreparation: 0,
        model.OrderStatusPrepared: 0,
        model.OrderStatusOntheway: 0,
        model.OrderStatusDelivered: 0,
        model.OrderStatusCancelled: 0,
    }

    // Iterate through each order and count order item statuses
    for _, order := range orders {
        var orderItems []model.OrderItem
        if err := database.DB.Where("order_id =?", order.OrderID).Find(&orderItems).Error; err!= nil {
            return model.OrderCount{}, errors.New("error fetching order items")
        }

        for _, status := range []string{
            model.OrderStatusProcessing,
            model.OrderStatusInPreparation,
            model.OrderStatusPrepared,
            model.OrderStatusOntheway,
            model.OrderStatusDelivered,
            model.OrderStatusCancelled,
        } {
            var count int64
            if err := database.DB.Model(&model.OrderItem{}).Where("order_id =? AND order_status =?", order.OrderID, status).Count(&count).Error; err!= nil {
                return model.OrderCount{}, errors.New("failed to query order items")
            }
            // Update the map counters based on the status
            orderStatusCounts[status] += count
        }
    }

    // Sum up the counts of order items across all statuses to get the total order count
    var totalCount int64
    for _, count := range orderStatusCounts {
        totalCount += count
    }

    // Construct and return the final OrderCount
    return model.OrderCount{
        TotalOrder: uint(totalCount),
        TotalProcessing: uint(orderStatusCounts[model.OrderStatusProcessing]),
        TotalInPreparation: uint(orderStatusCounts[model.OrderStatusInPreparation]),
        TotalPrepared: uint(orderStatusCounts[model.OrderStatusPrepared]),
        TotalOnTheWay: uint(orderStatusCounts[model.OrderStatusOntheway]),
        TotalDelivered: uint(orderStatusCounts[model.OrderStatusDelivered]),
        TotalCancelled: uint(orderStatusCounts[model.OrderStatusCancelled]),
    }, nil
}



//get orders based on time period
//make ordercounts map for order transition
//count for status for order items based on all the orders
//nested loop

//check if the order status is valid
// OrderCounts := make(map[string]int64)
// var Order model.Order
// OrderTransition := []string{model.OrderStatusProcessing, model.OrderStatusInPreparation, model.OrderStatusPrepared, model.OrderStatusOntheway, model.OrderStatusDelivered, model.OrderStatusCancelled}
// for _, status := range OrderTransition {
// 	var count int64
// 	if err := database.DB.Model(&model.OrderItem{}).Where("order_status =?", status).Count(&count).Error; err != nil {
// 		// errors.New("failed to query order items")
// 	}
// 	OrderCounts[status] = count
// }
// var TotalCount int64
// for _, v := range OrderCounts {
// 	TotalCount += v
// }
