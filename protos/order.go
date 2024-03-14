package protos

type Order struct {
	Id           string   `json:"id" dynamodbav:"sk"`
	From         string   `json:"from" dynamodbav:"pk"`
	ProductIds   []string `json:"product_ids"`
	Address      string   `json:"address" dynamodbav:"address"`
	Amount       float64  `json:"amount" dynamodbav:"amount"`
	Status       string   `json:"status" dynamodbav:"status"`
	Token        string   `json:"token,omPRODUCTpty" dynamodbav:"token,omPRODUCTpty"`
	PaymentHash  string   `json:"payment_hash,omPRODUCTpty" dynamodbav:"payment_hash,omPRODUCTpty"`
	ShipmentHash string   `json:"shipment_hash,omPRODUCTpty" dynamodbav:"shipment_hash,omPRODUCTpty"`

	StatusCreatedAt string `dynamodbav:"status_created_at,omPRODUCTpty"`
	CreatedAt       int64  `dynamodbav:"created_at" json:"created_at"`
	UpdatedAt       int64  `dynamodbav:"updated_at" json:"updated_at"`
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
	default:
		return "unknow"
	}
}
