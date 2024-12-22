package util

import (
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	log "github.com/sirupsen/logrus"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
	"main/pkg/types"
)

func validateAndLog(valid bool, err error, inputType string, input string) bool {
	if !valid {
		log.Warnf("%s is not a valid %s: %v", input, inputType, err)
		return false
	}
	return true
}

func isMnemonic(input string) (bool, *ecdsa.PrivateKey, common.Address, error) {
	if !bip39.IsMnemonicValid(input) {
		return false, nil, common.Address{}, errors.New("invalid mnemonic phrase")
	}

	seed, err := bip39.NewSeedWithErrorChecking(input, "")
	if err != nil {
		return false, nil, common.Address{}, err
	}

	masterKey, err := bip32.NewMasterKey(seed)
	if err != nil {
		return false, nil, common.Address{}, err
	}

	addressKey, err := deriveAddressKey(masterKey)
	if err != nil {
		return false, nil, common.Address{}, err
	}

	privateKey, err := crypto.ToECDSA(addressKey.Key)
	if err != nil {
		return false, nil, common.Address{}, err
	}

	address := crypto.PubkeyToAddress(privateKey.PublicKey)
	return true, privateKey, address, nil
}

func isPrivateKey(input string) (bool, *ecdsa.PrivateKey, common.Address, error) {
	input = RemoveHexPrefix(input)

	if len(input) != 64 {
		return false, nil, common.Address{}, errors.New("private key must be 64 characters")
	}

	privateKeyBytes, err := hex.DecodeString(input)
	if err != nil {
		return false, nil, common.Address{}, err
	}

	privateKey, err := crypto.ToECDSA(privateKeyBytes)
	if err != nil {
		return false, nil, common.Address{}, err
	}

	address := crypto.PubkeyToAddress(privateKey.PublicKey)
	return true, privateKey, address, nil
}

func deriveAddressKey(masterKey *bip32.Key) (*bip32.Key, error) {
	purpose, err := masterKey.NewChildKey(bip32.FirstHardenedChild + 44)
	if err != nil {
		return nil, err
	}
	coinType, err := purpose.NewChildKey(bip32.FirstHardenedChild + 60)
	if err != nil {
		return nil, err
	}
	account, err := coinType.NewChildKey(bip32.FirstHardenedChild)
	if err != nil {
		return nil, err
	}
	change, err := account.NewChildKey(0)
	if err != nil {
		return nil, err
	}
	addressKey, err := change.NewChildKey(0)
	if err != nil {
		return nil, err
	}
	return addressKey, nil
}

func GetAccounts(accountsListString []string) ([]types.AccountData, error) {
	var accounts []types.AccountData

	for _, currentAccountData := range accountsListString {
		var valid bool
		var privateKey *ecdsa.PrivateKey
		var accountAddress common.Address
		var err error

		// Проверяем, является ли это мнемонической фразой
		valid, privateKey, accountAddress, err = isMnemonic(currentAccountData)
		if !valid {
			// Если не является мнемонической фразой, проверяем, является ли это приватным ключом
			valid, privateKey, accountAddress, err = isPrivateKey(currentAccountData)
		}

		if !valid {
			// Если данные не валидны ни как мнемоническая фраза, ни как приватный ключ, выводим предупреждение
			log.Warnf("%s | Not a valid mnemonic or private key: %v", currentAccountData, err)
			continue
		}

		// Если валидно, добавляем аккаунт в список
		accounts = append(accounts, types.AccountData{
			PrivateKeyHex:  hex.EncodeToString(crypto.FromECDSA(privateKey)),
			PrivateKey:     privateKey,
			AccountAddress: accountAddress,
		})
	}

	return accounts, nil
}
