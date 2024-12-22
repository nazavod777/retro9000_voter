package types

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/common"
)

type AccountData struct {
	PrivateKeyHex  string
	PrivateKey     *ecdsa.PrivateKey
	AccountAddress common.Address
}

type ConstStruct struct {
	RoundID string `json:"round_id"`
}
