package protos

type CreateMonitorRequest struct {
	OrderId   string   `json:"order_id"`
	Table     string   `json:"table"`
	Contract  string   `json:"contract"`
	Topics    []string `json:"topics"`
	From      string   `json:"from"`
	FromBlock uint64   `json:"from_block"`
	TxHash    string   `json:"tx_hash" dynamodbav:"payment_hash,omitempty"`
}

type UpdateTrans struct {
	OrderId string `json:"order_id"`
	Table   string `json:"table"`
	TxHash  string `json:"tx_hash"`
	From    string `json:"from"`
	Status  Status `json:"status"`
}
