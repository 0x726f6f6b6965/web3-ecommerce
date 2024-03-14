package protos

type Product struct {
	Id          string  `json:"id" dynamodbav:"pk"`
	Name        string  `json:"name" dynamodbav:"name"`
	Description string  `json:"description" dynamodbav:"description"`
	Image       string  `json:"image" dynamodbav:"image"`
	Price       float64 `json:"price" dynamodbav:"price"`
	SoftDeleted int     `json:"soft_deleted" dynamodbav:"soft_deleted"`

	CreatedAt int64 `dynamodbav:"created_at" json:"created_at"`
	UpdatedAt int64 `dynamodbav:"updated_at" json:"updated_at"`
}
