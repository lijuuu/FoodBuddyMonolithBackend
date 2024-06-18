package controllers

import (
	"fmt"
	"foodbuddy/database"
	"foodbuddy/model"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf/v2"
)

func ProductReport(c *gin.Context) {
	ProductID := c.Param("productid")
	report := IndividualProductReport(ProductID)
	c.JSON(http.StatusOK, gin.H{
		"data": report,
	})
}

func IndividualProductReport(ProductID string) model.ProductSales {
	query := fmt.Sprintf(`SELECT SUM(order_items.amount) AS total_amount,
                        COUNT(order_items.amount) AS total_orders,
                        AVG(order_items.order_rating) AS avg_rating,
                        SUM(order_items.quantity) AS quantity
                         FROM order_items
                         JOIN orders ON order_items.order_id = orders.order_id
                         WHERE order_items.product_id = %s
                         AND orders.ordered_at BETWEEN '2024-06-01' AND '2024-06-18'
                         AND (order_items.order_status = 'PROCESSING' OR order_items.order_status IS NULL);`, ProductID)

	var report model.ProductSales
	err := database.DB.Raw(query).Scan(&report).Error
	if err != nil {
		return report
	}
	return report
}

func GeneratePDFInvoice(c *gin.Context, order model.Order, orderItems []model.OrderItem, address model.Address) error {
	pdf := gofpdf.New("P", "mm", "A4", "")

	// Add a page
	pdf.AddPage()

	// Set font
	pdf.SetFont("Arial", "B", 16)

	// Title
	pdf.Cell(40, 10, "Invoice")

	// Line break
	pdf.Ln(20)

	// Order Information
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(40, 10, fmt.Sprintf("Order ID: %s", order.OrderID))
	pdf.Ln(10)
	pdf.Cell(40, 10, fmt.Sprintf("User ID: %d", order.UserID))
	pdf.Ln(10)
	pdf.Cell(40, 10, fmt.Sprintf("Address Info: %s, %s, %s, %s, %s,%s", address.AddressType, address.StreetName, address.StreetNumber, address.City, address.State, address.PostalCode))
	pdf.Ln(10)
	pdf.Cell(40, 10, fmt.Sprintf("Payment Method: %s", order.PaymentMethod))
	pdf.Ln(10)
	pdf.Cell(40, 10, fmt.Sprintf("Payment Status: %s", order.PaymentStatus))
	pdf.Ln(10)
	pdf.Cell(40, 10, fmt.Sprintf("Ordered At: %s", order.OrderedAt.Format(time.RFC1123)))
	pdf.Ln(20)

	// Table header
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(40, 10, "Product ID")
	pdf.Cell(40, 10, "Quantity")
	pdf.Cell(40, 10, "Amount")
	pdf.Ln(10)

	// Table body
	pdf.SetFont("Arial", "", 12)
	for _, item := range orderItems {
		pdf.Cell(40, 10, fmt.Sprintf("%d", item.ProductID))
		pdf.Cell(40, 10, fmt.Sprintf("%d", item.Quantity))
		pdf.Cell(40, 10, fmt.Sprintf("%d", item.Amount))
		pdf.Ln(10)
	}

	// Total amount
	pdf.Ln(10)
	pdf.Cell(40, 10, fmt.Sprintf("Total Amount: %.2f", order.TotalAmount))
	pdf.Ln(10)
	pdf.Cell(40, 10, fmt.Sprintf("Final Amount: %.2f", order.FinalAmount))

	// Save the PDF to file
	err := pdf.OutputFileAndClose("invoice.pdf")
	if err != nil {
		return err
	}

	return nil
}

func GetOrderInfoByOrderIDAndGeneratePDF(c *gin.Context) {
	var Request model.GetOrderInfoByOrderID
	if err := c.Bind(&Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "failed to bind request",
			"error_code": http.StatusBadRequest,
		})
		return
	}

	var Order model.Order
	if err := database.DB.Where("order_id = ?", Request.OrderID).First(&Order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":     false,
			"message":    "failed to fetch order information",
			"error_code": http.StatusNotFound,
		})
		return
	}

	var OrderItems []model.OrderItem
	if err := database.DB.Where("order_id = ?", Request.OrderID).Find(&OrderItems).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "failed to fetch order information",
			"error_code": http.StatusInternalServerError,
		})
		return
	}
	var User model.User
	if err := database.DB.Where("id =?", Order.UserID).First(&User).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "failed to get user information",
			"error":   err.Error(),
		})
		return
	}

	var Address model.Address
	if err := database.DB.Where("address_id = ?", Order.AddressID).First(&Address).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "failed to get address",
			"error":   err.Error(),
		})
		return
	}

	if err := GeneratePDFInvoice(c, Order, OrderItems, Address); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to generate PDF",
			"error":   err.Error(),
		})
		return
	}

	response := gin.H{
		"order_id":        Order.OrderID,
		"discount_amount": Order.DiscountAmount,
		"coupon_code":     Order.CouponCode,
		"total_amount":    Order.TotalAmount,
		"final_amount":    Order.FinalAmount,
		"payment_method":  Order.PaymentMethod,
		"payment_status":  Order.PaymentStatus,
		"Ordered_at":      Order.OrderedAt.Format(time.RFC3339),

		"address": gin.H{
			"Address_type":  Address.AddressType,
			"street_name":   Address.StreetName,
			"street_number": Address.StreetNumber,
			"city":          Address.City,
			"state":         Address.State,
			"postal_code":   Address.PostalCode,
		},

		"order_item": OrderItems,
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "Successfully generated PDF",
		"data": gin.H{
			"OrderInformation": response,
		},
	})
}

func BestSellingProducts(c *gin.Context) {
	index := c.Query("index")
	indexNum, _ := strconv.Atoi(index)

	var BestProduct []model.BestProduct

	tx := database.DB.Table("order_items").
		Select("product_id, COUNT(product_id) AS TotalSales").
		Group("product_id").
		Order("TotalSales desc").
		Find(&BestProduct)

	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "failed to get best selling product"})
		return
	}

	if indexNum > len(BestProduct) || indexNum == 0 {
		indexNum = len(BestProduct)
	}

	for i, product := range BestProduct {
		var dbProduct model.Product
		if err := database.DB.Where("id =?", product.ProductID).First(&dbProduct).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"status": false, "message": "failed to find product information"})
			return
		}

		var dbCategory model.Category
		if err := database.DB.Where("id =?", dbProduct.CategoryID).First(&dbCategory).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"status": false, "message": "failed to find category information"})
			return
		}

		BestProduct[i].Name = dbProduct.Name
		BestProduct[i].CategoryName = dbCategory.Name
		BestProduct[i].Description = dbProduct.Description
		BestProduct[i].ImageURL = dbProduct.ImageURL
		BestProduct[i].Price = float64(dbProduct.Price)
		BestProduct[i].Rating = dbProduct.AverageRating
	}

	data := BestProduct[:indexNum]
	c.JSON(http.StatusOK, gin.H{"status": true, "data": data})
}

func PriceLowToHigh(c *gin.Context) {
	var Products []model.Product

	tx := database.DB.Table("products").Select("*").Order("price ASC").Find(&Products)
	if tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": false, "message": "failed to get product information"})
		return
	}

	var Response []model.ProductResponse
	for _, product := range Products {
		var dbCategory model.Category
		var dbRestaurant model.Restaurant

		if err := database.DB.Where("id =?", product.CategoryID).First(&dbCategory).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"status": false, "message": "failed to get category information"})
			return
		}

		if err := database.DB.Where("id =?", product.RestaurantID).First(&dbRestaurant).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"status": false, "message": "failed to get restaurant information"})
			return
		}

		Response = append(Response, model.ProductResponse{
			ID:             product.ID,
			RestaurantName: dbRestaurant.Name,
			CategoryName:   dbCategory.Name,
			Name:           product.Name,
			Description:    product.Description,
			ImageURL:       product.ImageURL,
			Price:          product.Price,
			StockLeft:      product.StockLeft,
			AverageRating:  product.AverageRating,
			Veg:            product.Veg,
		})
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "data": Response})
}

func PriceHighToLow(c *gin.Context) {
	var Products []model.Product

	tx := database.DB.Table("products").Select("*").Order("price DESC").Find(&Products)
	if tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": false, "message": "failed to get product information"})
		return
	}

	var Response []model.ProductResponse
	for _, product := range Products {
		var dbCategory model.Category
		var dbRestaurant model.Restaurant

		if err := database.DB.Where("id =?", product.CategoryID).First(&dbCategory).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"status": false, "message": "failed to get category information"})
			return
		}

		if err := database.DB.Where("id =?", product.RestaurantID).First(&dbRestaurant).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"status": false, "message": "failed to get restaurant information"})
			return
		}

		Response = append(Response, model.ProductResponse{
			ID:             product.ID,
			RestaurantName: dbRestaurant.Name,
			CategoryName:   dbCategory.Name,
			Name:           product.Name,
			Description:    product.Description,
			ImageURL:       product.ImageURL,
			Price:          product.Price,
			StockLeft:      product.StockLeft,
			AverageRating:  product.AverageRating,
			Veg:            product.Veg,
		})
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "data": Response})
}

func NewArrivals(c *gin.Context) {
	var products []model.Product
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)

	tx := database.DB.Where("created_at >= ?", sevenDaysAgo).
		Order("price ASC").
		Find(&products)

	if tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "failed to get product information",
		})
		return
	}

	var response []model.ProductResponse
	for _, product := range products {
		var dbCategory model.Category
		var dbRestaurant model.Restaurant

		if err := database.DB.Where("id = ?", product.CategoryID).First(&dbCategory).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"status": false, "message": "failed to get category information"})
			return
		}

		if err := database.DB.Where("id = ?", product.RestaurantID).First(&dbRestaurant).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"status": false, "message": "failed to get restaurant information"})
			return
		}

		response = append(response, model.ProductResponse{
			ID:             product.ID,
			RestaurantName: dbRestaurant.Name,
			CategoryName:   dbCategory.Name,
			Name:           product.Name,
			Description:    product.Description,
			ImageURL:       product.ImageURL,
			Price:          product.Price,
			StockLeft:      product.StockLeft,
			AverageRating:  product.AverageRating,
			Veg:            product.Veg,
		})
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "data": response})
}

