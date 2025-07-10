package controllers

import (
	"bytes"
	"fmt"
	"foodbuddy/internal/database"
	"foodbuddy/internal/model"
	"foodbuddy/internal/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf/v2"
)

func ProductReport(c *gin.Context) {
	productID, err := strconv.Atoi(c.Query("productid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid ProductID",
		})
		return
	}
	report := IndividualProductReport(strconv.Itoa(productID))
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
                         AND (order_items.order_status = 'DELIVERED' OR order_items.order_status IS NULL);`, ProductID)

	var report model.ProductSales
	err := database.DB.Raw(query).Scan(&report).Error
	if err != nil {
		return report
	}
	report.AvgRating = RoundDecimalValue(report.AvgRating)
	return report
}

func GeneratePDFInvoice(order model.Order, orderItems []model.OrderItem, address model.Address, User model.User) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "Tabloid", "")
	// Add a page
	pdf.AddPage()

	// Set font
	pdf.SetFont("Arial", "B", 12)

	// Inserting an image
	pdf.Image("./assets/foodbuddy.png", 10, 10, 50, 0, false, "", 0, "")
	pdf.Ln(20)
	// Title
	pdf.Cell(40, 10, "Order Invoice")
	pdf.Ln(20)

	// Order Information
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(40, 10, fmt.Sprintf("Order ID: %s", order.OrderID))
	pdf.Ln(10)
	pdf.Cell(40, 10, fmt.Sprintf("Name  : %v", User.Name))
	pdf.Ln(10)
	pdf.Cell(40, 10, fmt.Sprintf("Email  : %v", User.Email))
	pdf.Ln(10)
	pdf.Cell(40, 10, fmt.Sprintf("Phone Number  : %v", User.PhoneNumber))
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
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(15, 10, "ID")
	pdf.Cell(60, 10, "Product Name")
	pdf.Cell(25, 10, "Offer")
	pdf.Cell(40, 10, "Restaurant")
	pdf.Cell(20, 10, "Qty")
	pdf.Cell(30, 10, "Amount")
	pdf.Cell(30, 10, "Status")
	pdf.Ln(10)

	// Table body
	pdf.SetFont("Arial", "", 12)
	for _, item := range orderItems {
		var Result model.Restaurant
		var Product model.Product
		//restaurant info
		database.DB.Where("id = ?", item.RestaurantID).First(&Result)
		pdf.Cell(15, 10, fmt.Sprintf("%d", item.ProductID))
		//product info
		database.DB.Where("id = ?", item.ProductID).First(&Product)
		pdf.Cell(60, 10, fmt.Sprintf("%v", Product.Name))

		pdf.Cell(25, 10, fmt.Sprintf("%.2f", item.ProductOfferAmount))
		pdf.Cell(40, 10, fmt.Sprintf("%v", Result.Name))
		pdf.Cell(20, 10, fmt.Sprintf("%d", item.Quantity))
		pdf.Cell(30, 10, fmt.Sprintf("%.2f", item.Amount))
		pdf.Cell(30, 10, fmt.Sprintf("%v", item.OrderStatus))
		pdf.Ln(10)
	}

	//total amount before deduction
	pdf.Ln(10)
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(40, 10, fmt.Sprintf("Gross Amount: %.2f", order.TotalAmount))
	pdf.Ln(10)

	//coupon discount
	pdf.SetTextColor(255, 0, 0) //red for discount
	pdf.Cell(40, 10, fmt.Sprintf("Coupon Discount Amount: %.2f", order.CouponDiscountAmount))
	pdf.Ln(10)

	//product disocunt
	pdf.Cell(40, 10, fmt.Sprintf("Product Discount Amount: %.2f", order.ProductOfferAmount))
	pdf.Ln(10)
	pdf.SetTextColor(0, 0, 0) //reset to black

	//final amount after dicounts
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(40, 10, fmt.Sprintf("Net Amount: %.2f", order.FinalAmount))
	pdf.Ln(10)

	// Reset font settings
	pdf.SetFont("Arial", "", 12)

	var pdfBytes bytes.Buffer
	err := pdf.Output(&pdfBytes)
	if err != nil {
		return nil, err
	}

	return pdfBytes.Bytes(), nil
}

func GetOrderInfoByOrderIDAndGeneratePDF(c *gin.Context) {
	OrderID := c.Query("order_id")

	var Order model.Order
	if err := database.DB.Where("order_id = ?", OrderID).First(&Order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "failed to fetch order information",
		})
		return
	}

	var OrderItems []model.OrderItem
	if err := database.DB.Where("order_id = ?", OrderID).Find(&OrderItems).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to fetch order information",
		})
		return
	}
	var User model.User
	if err := database.DB.Where("id = ?", Order.UserID).First(&User).Error; err != nil {
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

	pdfBytes, err := GeneratePDFInvoice(Order, OrderItems, Address, User)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to generate PDF",
			"error":   err.Error(),
		})
		return
	}

	c.Writer.Header().Set("Content-type", "application/pdf")
	c.Writer.Header().Set("Content-Disposition", "inline; filename=invoice.pdf") //https://blog.devgenius.io/tutorial-creating-an-endpoint-to-download-files-using-golang-and-gin-gonic-27abbcf75940
	c.Writer.Write(pdfBytes)
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

		if err := database.DB.Where("id = ?", product.CategoryID).First(&dbCategory).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"status": false, "message": "failed to get category information"})
			return
		}

		if err := database.DB.Where("id = ?", product.RestaurantID).First(&dbRestaurant).Error; err != nil {
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
			Price:          (product.Price),
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

		if err := database.DB.Where("id = ?", product.CategoryID).First(&dbCategory).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"status": false, "message": "failed to get category information"})
			return
		}

		if err := database.DB.Where("id = ?", product.RestaurantID).First(&dbRestaurant).Error; err != nil {
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
			Price:          (product.Price),
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
			Price:          (product.Price),
			StockLeft:      product.StockLeft,
			AverageRating:  product.AverageRating,
			Veg:            product.Veg,
		})
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "data": response})
}

func PlatformOverallSalesReport(c *gin.Context) {
	var input model.PlatformSalesReportInput

	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.StartDate == "" && input.EndDate == "" && input.Limit == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "please provide start date and end date, or specify the limit as day, week, month, year"})
		return
	}

	// Handle case where Limit is specified
	if input.Limit != "" {
		limits := []string{"day", "week", "month", "year"}
		found := false
		for _, l := range limits {
			if input.Limit == l {
				found = true
				break
			}
		}
		if !found {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit specified, valid options are: day, week, month, year"})
			return
		}

		// Process based on the specified limit
		var startDate, endDate string
		switch input.Limit {
		case "day":
			startDate = time.Now().AddDate(0, 0, -1).Format("2006-01-02")
			endDate = time.Now().Format("2006-01-02")
		case "week":
			today := time.Now()
			startDate = today.AddDate(0, 0, -int(today.Weekday())).Format("2006-01-02") // Start of the week (Sunday)
			endDate = today.AddDate(0, 0, 7-int(today.Weekday())).Format("2006-01-02")  // End of the week (Saturday)
		case "month":
			today := time.Now()
			firstDayOfMonth := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, today.Location())
			lastDayOfMonth := firstDayOfMonth.AddDate(0, 1, -1)
			startDate = firstDayOfMonth.Format("2006-01-02")
			endDate = lastDayOfMonth.Format("2006-01-02")
		case "year":
			endDate = time.Now().Format("2006-01-02")
			startDate = time.Now().AddDate(-1, 0, 0).Format("2006-01-02")
		}

		fmt.Println(startDate)
		fmt.Println(endDate)
		result, amount, err := TotalOrders(startDate, endDate, input.PaymentStatus, 0)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error processing orders"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status":  true,
			"message": "successfully created sales report",
			"result":  result,
			"amount":  amount,
		})
		return
	}

	result, amount, err := TotalOrders(input.StartDate, input.EndDate, input.PaymentStatus, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error processing orders"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "successfully created sales report",
		"result":  result,
		"amount":  amount,
	})
}

func RestaurantOverallSalesReport(c *gin.Context) {
	email, role, err := utils.GetJWTClaim(c)
	if role != model.UserRole || err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "unauthorized request",
		})
		return
	}

	RestaurantID, _ := RestIDfromEmail(email)

	var input model.RestaurantOverallSalesReport

	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.StartDate == "" && input.EndDate == "" && input.Limit == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "please provide start date and end date, or specify the limit as day, week, month, year"})
		return
	}

	// Handle case where Limit is specified
	if input.Limit != "" {
		limits := []string{"day", "week", "month", "year"}
		found := false
		for _, l := range limits {
			if input.Limit == l {
				found = true
				break
			}
		}
		if !found {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit specified, valid options are: day, week, month, year"})
			return
		}

		// Process based on the specified limit
		var startDate, endDate string
		switch input.Limit {
		case "day":
			startDate = time.Now().AddDate(0, 0, -1).Format("2006-01-02")
			endDate = time.Now().Format("2006-01-02")
		case "week":
			today := time.Now()
			startDate = today.AddDate(0, 0, -int(today.Weekday())).Format("2006-01-02") // Start of the week (Sunday)
			endDate = today.AddDate(0, 0, 7-int(today.Weekday())).Format("2006-01-02")  // End of the week (Saturday)
		case "month":
			today := time.Now()
			firstDayOfMonth := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, today.Location())
			lastDayOfMonth := firstDayOfMonth.AddDate(0, 1, -1)
			startDate = firstDayOfMonth.Format("2006-01-02")
			endDate = lastDayOfMonth.Format("2006-01-02")
		case "year":
			endDate = time.Now().Format("2006-01-02")
			startDate = time.Now().AddDate(-1, 0, 0).Format("2006-01-02")
		}

		fmt.Println(startDate)
		fmt.Println(endDate)
		result, amount, err := TotalOrders(startDate, endDate, input.PaymentStatus, RestaurantID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error processing orders"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status":  true,
			"message": "successfully created sales report",
			"result":  result,
			"amount":  amount,
		})
		return
	}

	result, amount, err := TotalOrders(input.StartDate, input.EndDate, input.PaymentStatus, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error processing orders"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "successfully created sales report",
		"result":  result,
		"amount":  amount,
	})
}
