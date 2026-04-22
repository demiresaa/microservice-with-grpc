package errors

const (
	CodeOrderNotFound        = "ORDER_001"
	CodeOrderAlreadyExists   = "ORDER_002"
	CodeInvalidOrderStatus   = "ORDER_003"
	CodeEmptyCustomerID      = "ORDER_004"
	CodeEmptyProductID       = "ORDER_005"
	CodeInvalidQuantity      = "ORDER_006"
	CodeInvalidTotalPrice    = "ORDER_007"

	CodePaymentNotFound        = "PAYMENT_001"
	CodeInsufficientBalance    = "PAYMENT_002"
	CodePaymentAlreadyRefunded = "PAYMENT_003"

	CodeInsufficientStock = "INVENTORY_001"
	CodeProductNotFound   = "INVENTORY_002"
)

var (
	ErrOrderNotFound      = New(CodeOrderNotFound, "order not found", nil)
	ErrOrderAlreadyExists = New(CodeOrderAlreadyExists, "order already exists", nil)
	ErrInvalidOrderStatus = New(CodeInvalidOrderStatus, "invalid order status transition", nil)
	ErrEmptyCustomerID    = New(CodeEmptyCustomerID, "customer_id is required", nil)
	ErrEmptyProductID     = New(CodeEmptyProductID, "product_id is required", nil)
	ErrInvalidQuantity    = New(CodeInvalidQuantity, "quantity must be greater than 0", nil)
	ErrInvalidTotalPrice  = New(CodeInvalidTotalPrice, "total_price must be greater than 0", nil)

	ErrPaymentNotFound        = New(CodePaymentNotFound, "payment not found", nil)
	ErrInsufficientBalance    = New(CodeInsufficientBalance, "insufficient balance", nil)
	ErrPaymentAlreadyRefunded = New(CodePaymentAlreadyRefunded, "payment already refunded", nil)

	ErrInsufficientStock = New(CodeInsufficientStock, "insufficient stock", nil)
	ErrProductNotFound   = New(CodeProductNotFound, "product not found", nil)
)

func Wrap(base *AppError, err error) *AppError {
	return New(base.Code, base.Message, err)
}
