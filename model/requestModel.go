package model

type EmailSignupRequest struct {
	Name            string `validate:"required" json:"name"`
	Email           string `validate:"required,email" json:"email"`
	Password        string `validate:"required" json:"password"`
	ConfirmPassword string `validate:"required" json:"confirmpassword"`
}

type EmailLoginRequest struct {
	Email    string `form:"email" validate:"required,email" json:"email"`
	Password string `form:"password" validate:"required" json:"password"`
}

type UpdateUserInformation struct{
	ID             uint   `validate:"required"`
	Name           string `gorm:"column:name;type:varchar(255)" validate:"required" json:"name"`
	PhoneNumber    int `gorm:"column:phone_number;type:varchar(255);unique_index" validate:"number" json:"phone_number"`
	Picture        string `gorm:"column:picture;type:text" json:"picture"`
}

type RestaurantSignupRequest struct {
	Name           string `validate:"required" json:"name"`
	Description    string `gorm:"column:description" validate:"required" json:"description"`
	Address        string `gorm:"column:address" validate:"required" json:"address"`
	Email          string `gorm:"column:email" validate:"required,email" json:"email"`
	Password       string `gorm:"column:password" validate:"required" json:"password"`
	PhoneNumber    string `gorm:"column:phone_number" validate:"required" json:"phone_number"`
	ImageURL       string `gorm:"column:image_url" validate:"required" json:"image_url"`
	CertificateURL string `gorm:"column:certificate_url" validate:"required" json:"certificate_url"`
}

type RestaurantLoginRequest struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required"`
}

type AdminLoginRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type AddToCartReq struct {
	UserID   uint    `gorm:"column:user_id" validate:"required,number" json:"user_id"`
	ProductID uint    `gorm:"column:product_id" validate:"required,number" json:"product_id"`
	Quantity  uint    `validate:"required,number" json:"quantity"`
	CookingRequest string `json:"cooking_request"`// similar to zomato,, requesting restaurant to add or remove specific ingredients etc
}

type RemoveItem struct{
	UserID    uint `validate:"required,number" json:"user_id"`
	ProductID uint `validate:"required,number" json:"product_id"`
}


type PlaceOrder struct{
	UserID    uint `validate:"required,number" json:"user_id"`
	AddressID uint `validate:"required,number" json:"address_id"`
	PaymentMethod string `validate:"required" json:"payment_method"`
	
}

type InitiatePayment struct{
	OrderID string `json:"order_id"`
}

type RazorpayPayment struct {
	PaymentID string `form:"razorpay_payment_id" binding:"required"`
	OrderID   string `form:"razorpay_order_id" binding:"required"`
	Signature string `form:"razorpay_signature" binding:"required"`
}

type OrderHistoryRestaurants struct{
	RestaurantID uint `json:"restaurant_id"`
	OrderStatus string `json:"order_status"`
}

type UserOrderHistory struct{
	UserID uint `json:"user_id"`
	PaymentStatus string `json:"payment_status"`
}

type GetOrderInfoByOrderID struct{
	OrderID string `json:"order_id"`
}

type PaymentDetailsByOrderID struct{
	OrderID string `json:"order_id"`
	PaymentStatus string `json:"payment_status"`
}

type UpdateOrderStatusForRestaurant struct{
	OrderID string `json:"order_id"`
	ProductID uint `json:"product_id"`
}

type CancelOrderedProduct struct{
	OrderID string `json:"order_id"`
	ProductId uint `json:"product_id"`
}