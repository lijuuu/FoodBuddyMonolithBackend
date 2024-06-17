package controllers

import (
	"fmt"
	"foodbuddy/database"
	"foodbuddy/model"
	"net/http"
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

func GeneratePDFInvoice(c *gin.Context, order model.Order, orderItems []model.OrderItem,address model.Address) error {
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
	if err := database.DB.Where("order_id = ?", Request.OrderID).First(&Order).Error; err != nil{
		c.JSON(http.StatusNotFound, gin.H{
			"status":     false,
			"message":    "failed to fetch order information",
			"error_code": http.StatusNotFound,
		})
		return
	}

	var OrderItems []model.OrderItem
	if err := database.DB.Where("order_id = ?", Request.OrderID).Find(&OrderItems).Error; err != nil{
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "failed to fetch order information",
			"error_code": http.StatusInternalServerError,
		})
		return
	}
	var User model.User
	if err:=database.DB.Where("id =?",Order.UserID).First(&User).Error;err!=nil{
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "failed to get user information",
			"error":   err.Error(),
		})
		return
	}

	var Address model.Address
    if err:=database.DB.Where("address_id = ?",Order.AddressID).First(&Address).Error;err!=nil{
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "failed to get address",
			"error":   err.Error(),
		})
		return
	}

	if err := GeneratePDFInvoice(c, Order, OrderItems,Address); err != nil{
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to generate PDF",
			"error":   err.Error(),
		})
		return
	}

	response := gin.H{
        "order_id":          Order.OrderID,
        "discount_amount":   Order.DiscountAmount,
        "coupon_code":       Order.CouponCode,
        "total_amount":      Order.TotalAmount,
        "final_amount":      Order.FinalAmount,
        "payment_method":    Order.PaymentMethod,
        "payment_status":    Order.PaymentStatus,
        "Ordered_at":        Order.OrderedAt.Format(time.RFC3339),
        
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
		"data":    gin.H{
			"OrderInformation":response,

		},
	})
}


func BestSellingProduct(c *gin.Context)  {
	var BestProduct struct{
		ProductID uint
	}
	
	tx := database.DB.Table("order_items").
	Select("product_id, COUNT(product_id) AS product_count").
	Group("product_id").
	Order("product_count desc").
	First(&BestProduct)

	if tx.Error!=nil{
	   c.JSON(http.StatusInternalServerError,gin.H{
		"status":false,
		"message":"failed to get best selling product",
	   })
	   return
	}

	var Product model.Product
    if err:=database.DB.Where("id = ?",BestProduct.ProductID).First(&Product).Error;err!=nil{
 	   c.JSON(http.StatusInternalServerError,gin.H{
		"status":false,
		"message":"failed to get product information",
	   })
	   return
	}

	c.JSON(http.StatusOK,gin.H{
		"status":true,
		"data":Product,
	   })
}