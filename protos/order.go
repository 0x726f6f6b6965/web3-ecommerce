package protos

type Order struct {
	Id           string          `json:"id" dynamodbav:"sk"`
	From         string          `json:"from" dynamodbav:"pk"`
	ProductIds   []OrderProducts `json:"product_ids"`
	Address      string          `json:"address" dynamodbav:"address"`
	Amount       float64         `json:"amount" dynamodbav:"amount"`
	Status       Status          `json:"status" dynamodbav:"status"`
	Token        string          `json:"token,omitempty" dynamodbav:"token,omitempty"`
	PaymentHash  string          `json:"payment_hash,omitempty" dynamodbav:"payment_hash,omitempty"`
	ShipmentHash string          `json:"shipment_hash,omitempty" dynamodbav:"shipment_hash,omitempty"`

	StatusCreatedAt string `dynamodbav:"status_created_at,omitempty"`
	CreatedAt       int64  `dynamodbav:"created_at" json:"created_at"`
	UpdatedAt       int64  `dynamodbav:"updated_at" json:"updated_at"`
}

type OrderProducts struct {
	Id       string  `json:"id"`
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
}

type Status int

const (
	StatusUnknow Status = iota
	StatusCreated
	StatusPending
	StatusPaid
	StatusPaidFailed
	StatusShipped
	StatusDelivered
	StatusCancelled
	StatusMonitorFailed
)

func (s Status) String() string {
	switch s {
	case StatusCreated:
		return "created"
	case StatusPending:
		return "pending"
	case StatusPaid:
		return "paid"
	case StatusPaidFailed:
		return "paid_failed"
	case StatusShipped:
		return "shipped"
	case StatusDelivered:
		return "delivered"
	case StatusCancelled:
		return "cancelled"
	case StatusMonitorFailed:
		return "monitor_failed"
	default:
		return "unknow"
	}
}
