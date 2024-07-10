package controllers

import (
	"errors"
	"fmt"
	"foodbuddy/internal/database"
	"foodbuddy/internal/model"
	"time"
)

// total orders
// total delivered
// total cancelled
// revenue generated
// coupon deductions
// product offer deductions
// referral rewards given
func TotalOrders(From string, Till string, PaymentStatus string) (model.OrderCount, model.AmountInformation, error) {
	var orders []model.Order

	parsedFrom, err := time.Parse("2006-01-02", From)
	if err != nil {
		return model.OrderCount{}, model.AmountInformation{}, fmt.Errorf("error parsing From time: %v", err)
	}

	parsedTill, err := time.Parse("2006-01-02", Till)
	if err != nil {
		return model.OrderCount{}, model.AmountInformation{}, fmt.Errorf("error parsing Till time: %v", err)
	}

	fFrom := time.Date(parsedFrom.Year(), parsedFrom.Month(), parsedFrom.Day(), 0, 0, 0, 0, time.UTC)
	fTill := time.Date(parsedTill.Year(), parsedTill.Month(), parsedTill.Day(), 23, 59, 59, 999999999, time.UTC)

	startTime := fFrom.Format("2006-01-02T15:04:05Z")
	endDate := fTill.Format("2006-01-02T15:04:05Z")

	// Fetch orders within the specified time frame and payment status
	if err := database.DB.Where("ordered_at BETWEEN? AND? AND payment_status =?", startTime, endDate, PaymentStatus).Find(&orders).Error; err != nil {
		return model.OrderCount{}, model.AmountInformation{}, errors.New("error fetching orders")
	}

	var orderStatusCounts = map[string]int64{
		model.OrderStatusProcessing:    0,
		model.OrderStatusInPreparation: 0,
		model.OrderStatusPrepared:      0,
		model.OrderStatusOntheway:      0,
		model.OrderStatusDelivered:     0,
		model.OrderStatusCancelled:     0,
	}

	var AccountInformation model.AmountInformation
	for _, order := range orders {
		AccountInformation.TotalCouponDeduction += RoundDecimalValue(order.CouponDiscountAmount)
		AccountInformation.TotalProductOfferDeduction += RoundDecimalValue(order.ProductOfferAmount)
		AccountInformation.TotalAmountBeforeDeduction += RoundDecimalValue(order.TotalAmount)
		AccountInformation.TotalAmountAfterDeduction += RoundDecimalValue(order.FinalAmount)
		var orderItems []model.OrderItem
		if err := database.DB.Where("order_id =?", order.OrderID).Find(&orderItems).Error; err != nil {
			return model.OrderCount{}, model.AmountInformation{}, errors.New("error fetching order items")
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
			if err := database.DB.Model(&model.OrderItem{}).Where("order_id =? AND order_status =?", order.OrderID, status).Count(&count).Error; err != nil {
				return model.OrderCount{}, model.AmountInformation{}, errors.New("failed to query order items")
			}

			orderStatusCounts[status] += count
		}
	}
	var totalCount int64
	for _, count := range orderStatusCounts {
		totalCount += count
	}

	return model.OrderCount{
		TotalOrder:         uint(totalCount),
		TotalProcessing:    uint(orderStatusCounts[model.OrderStatusProcessing]),
		TotalInPreparation: uint(orderStatusCounts[model.OrderStatusInPreparation]),
		TotalPrepared:      uint(orderStatusCounts[model.OrderStatusPrepared]),
		TotalOnTheWay:      uint(orderStatusCounts[model.OrderStatusOntheway]),
		TotalDelivered:     uint(orderStatusCounts[model.OrderStatusDelivered]),
		TotalCancelled:     uint(orderStatusCounts[model.OrderStatusCancelled]),
	}, AccountInformation, nil
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
