package protos

type CommonRequest struct {
	From      string  `json:"from"`
	To        string  `json:"to"`
	Amount    float64 `json:"amount"`
	Nonce     uint64  `json:"nonce"`
	Signature string  `json:"signature,omitempty"`
	GasTipCap string  `json:"gasTipCap,omitempty"`
	GasFeeCap string  `json:"gasFeeCap,omitempty"`
	Gas       string  `json:"gas,omitempty"`
}

type CheckAllowanceRequest struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type GetTokenRequest struct {
	PublicAddress string `json:"public_address"`
	Signature     string `json:"signature"`
}

type GetTokenResponse struct {
	PublicAddress string `json:"publicAddress"`
	Token         string `json:"token"`
}

type UpdateUserRequest struct {
	PublicAddress string   `json:"publicAddress"`
	User          *User    `json:"user"`
	UpdateMask    []string `json:"updateMask"`
}

type UpdateProductRequest struct {
	ProductId  string   `json:"product_id"`
	Product    *Product `json:"product"`
	UpdateMask []string `json:"update_mask"`
}

type GetProductListResponse struct {
	Products      []*Product `json:"products"`
	NextPageToken string     `json:"next_page_token"`
}

type UpdateOrderStatusRequest struct {
	OrderId string `json:"order_id"`
	Status  Status `json:"status"`
}

type PayRequest struct {
	OrderId string         `json:"order_id"`
	Pay     *CommonRequest `json:"pay"`
}

type TestRequest struct {
	PrivateKey string `json:"private_key"`
	Signature  string `json:"signature"`
}
