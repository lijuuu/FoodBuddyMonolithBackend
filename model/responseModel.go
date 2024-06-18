package model

type GoogleResponse struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

type SuccessResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Status  bool   `json:"status" example:"false"`
	Message string `json:"message" example:"Error message"`
}

type ProductSales struct {
	TotalAmount int     `json:"total_amount"`
	TotalOrders int     `json:"total_orders"`
	AvgRating   float64 `json:"avg_rating"` // Pointer to allow for NULL values
	Quantity    int     `json:"quantity"`
}

type BestProduct struct {
	ProductID    uint
	Name         string
	CategoryName string
	Description  string
	ImageURL     string
	Price        float64
	Rating       float64
	TotalSales   uint `json:"TotalSales"`
}

type ProductResponse struct {
	ID             uint    `json:"product_id"`
	RestaurantName string  `json:"restaurant_name"`
	CategoryName   string  `json:"category_name"`
	Name           string  `json:"product_name"`
	Description    string  `json:"description"`
	ImageURL       string  `json:"image_url"`
	Price          float64    `json:"price"`
	StockLeft      uint    `json:"stock_left"`
	AverageRating  float64 `json:"average_rating"`
	Veg            string    `json:"veg"`
}
