package voterDeleter

import (
	"fmt"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	log "github.com/sirupsen/logrus"
	"main/internal/retroActions"
	"main/internal/util"
	"main/pkg/types"
)

func DeleteVotes(
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

	if votesData == nil {
		log.Printf("%s | No Available Votes", accountData.AccountAddress.String())
		return nil
	}

	for i, currentVote := range votesData.Data.Votes {
		err = retroActions.DeleteVote(client, accountData, accessToken, refreshToken, currentVote.Project.Id)

		if err != nil {
			log.Printf("%v", err)
		} else {
			log.Printf("%s | [%d/%d] Successfully Deleted Vote To %s",
				accountData.AccountAddress.String(), i+1, len(votesData.Data.Votes), currentVote.Project.Id)
		}
	}

	return nil
}
