package voterParser

import (
	"fmt"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	log "github.com/sirupsen/logrus"
	"main/internal/retroActions"
	"main/internal/util"
	"main/pkg/types"
	util2 "main/pkg/util"
)

func ParseVotes(
	accountData types.AccountData,
	accountProxy string,
) error {
	client := util.GetClient(accountProxy)
	signText := retroActions.GetSignText(client, accountData)
	signature, err := crypto.Sign(accounts.TextHash([]byte(signText)), accountData.PrivateKey)

	if err != nil {
		return fmt.Errorf("%s | Failed to sign auth message: %s", accountData.AccountAddress.String(), err)
	}

	signature[64] += 27
	signedMessage := hexutil.Encode(signature)

	accessToken, refreshToken, err := retroActions.DoAuth(client, accountData, signedMessage)

	if err != nil {
		return err
	}

	log.Printf("%s | Successfully Authorized", accountData.AccountAddress.String())

	votesData := retroActions.GetVotes(client, accountData, accessToken, refreshToken)

	eligibleVotes := votesData.Data.TotalEligibleVotes
	usedVotes := votesData.Data.UsedVotes
	availableVotes := eligibleVotes - usedVotes

	log.Printf("%s | Eligible Votes: %d | Already Used Votes: %d | Available Votes: %d",
		accountData.AccountAddress.String(), eligibleVotes, usedVotes, availableVotes)

	if eligibleVotes > 0 {
		util2.AppendFile("accounts_with_votes.txt",
			fmt.Sprintf("%s\n", accountData.PrivateKeyHex))
	}

	return nil
}
