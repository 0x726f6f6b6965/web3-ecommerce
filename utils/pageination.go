package utils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

type PageToken interface {
	// GetToken - get token string
	GetToken() string
	// GetID - get ID
	GetID() int
	// GetSize - get page size
	GetSize() uint64
	// SetID - set ID
	SetID(id int)
	// SetSize - set page size
	SetSize(size uint64)
}

type token struct {
	// Id
	Id int `json:"id"`
	// Size
	Size uint64 `json:"size"`
}

// GetID - get ID
func (t *token) GetID() int {
	return t.Id
}

// GetSize - get page size
func (t *token) GetSize() uint64 {
	return t.Size
}

// GetToken - encodes an token struct as a page token string.
func (t *token) GetToken() string {
	b, _ := json.Marshal(t)
	encodedStr := base64.URLEncoding.EncodeToString(b)
	return encodedStr
}

// SetID - set ID
func (t *token) SetID(id int) {
	t.Id = id
}

// SetSize - set page size
func (t *token) SetSize(size uint64) {
	t.Size = size
}

func NewPageToken(id int, size uint64) PageToken {
	return &token{Id: id, Size: size}
}

// GetPageTokenByString - decodes an encoded page token into an token struct
func GetPageTokenByString(in string) (PageToken, error) {
	b, err := base64.URLEncoding.DecodeString(in)
	if err != nil {
		return nil, fmt.Errorf("decode page token struct: %w", err)
	}
	t := new(token)
	if err = json.Unmarshal(b, t); err != nil {
		return nil, fmt.Errorf("unmarshal error: %w", err)
	}
	return t, nil
}
