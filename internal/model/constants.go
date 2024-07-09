package model

const (
	LocalHost                  = "localhost"
	ProjectRoot                = "PROJECTROOT"
	EmailLoginMethod           = "email"
	GoogleSSOMethod            = "googlesso"
	VerificationStatusVerified = "verified"
	VerificationStatusPending  = "pending"
	UserRole                   = "user"
	AdminRole                  = "admin"
	RestaurantRole             = "restaurant"
	PasswordEntropy            = 75
	MaxUserQuantity            = 50
	YES                        = "YES"
	NO                         = "NO"

	CashOnDelivery = "COD"
	OnlinePayment  = "ONLINE"

	Razorpay = "RAZORPAY"
	Stripe   = "STRIPE"
	Wallet   = "WALLET"

	WalletIncoming = "INCOMING"
	WalletOutgoing = "OUTGOING"

	WalletTxTypeOrderRefund    = "ORDERREFUND"
	WalletTxTypeReferralReward = "REFERRALREWARD"
	WalletTxTypeOrderPayment   = "ORDERPAYMENT"

	OnlinePaymentPending   = "ONLINE_PENDING"
	OnlinePaymentConfirmed = "ONLINE_CONFIRMED"
	OnlinePaymentFailed    = "ONLINE_FAILED"

	CODStatusPending   = "COD_PENDING"
	CODStatusConfirmed = "COD_CONFIRMED"
	CODStatusFailed    = "COD_FAILED"

	OrderStatusProcessing    = "PROCESSING"
	OrderStatusInPreparation = "PREPARATION"
	OrderStatusPrepared      = "PREPARED"
	OrderStatusOntheway      = "ONTHEWAY"
	OrderStatusDelivered     = "DELIVERED"
	OrderStatusCancelled     = "CANCELLED"

	CouponDiscountPercentageLimit = 50

	ReferralClaimAmount = 30
	ReferralClaimLimit  = 1
)