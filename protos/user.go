package protos

type User struct {
	PublicAddress string             `json:"public_address" dynamodbav:"pk"`
	Name          string             `json:"name" dynamodbav:"name"`
	Email         string             `json:"email" dynamodbav:"email"`
	Addresses     map[string]Address `json:"addresses" dynamodbav:"addresses"`

	CreatedAt int64 `dynamodbav:"created_at" json:"created_at"`
	UpdatedAt int64 `dynamodbav:"updated_at" json:"updated_at"`
}

type Address struct {
	Title         string `json:"title"`
	StreetAddress string `dynamodbav:"street_address" json:"street_address"`
	PostalCode    int    `dynamodbav:"postal_code" json:"postal_code"`
	CountryCode   string `dynamodbav:"country_code" json:"country_code"`
}

type UserToken struct {
	PublicAddress string `json:"public_address"`
	Nonce         string `json:"nonce"`
	ExpireAt      int64  `json:"expire_at"`
	CreatedAt     int64  `json:"created_at"`
}
