package voter

import (
	"fmt"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	log "github.com/sirupsen/logrus"
	"main/internal/retroActions"
	"main/internal/util"
	"main/pkg/types"
	"math/rand"
)

func generateDistribution(
	projects []retroActions.ProjectData,
	totalVotes int64,
) []DistributionData {
	if len(projects) == 0 || totalVotes <= 0 {
		return nil
	}

	// Генерация случайного количества проектов от 5 до 14
	randomMax := rand.Intn(10) + 5
	numProjects := randomMax
	if len(projects) < randomMax {
		numProjects = len(projects)
	}

	if numProjects == 0 {
		numProjects = 1
	}

	selectedProjects := rand.Perm(len(projects))[:numProjects]
	distribution := make([]DistributionData, 0, numProjects)
	remainingVotes := totalVotes

	// Распределяем голоса
	for i, projectIndex := range selectedProjects {
		var votes int64
		if i == numProjects-1 {
			// Последний проект получает все оставшиеся голоса
			votes = remainingVotes
		} else {
			// Генерируем случайное число голосов, не превышающее оставшихся
			maxVotes := remainingVotes / int64(numProjects-i)
			if maxVotes > 1 {
				votes = rand.Int63n(maxVotes) + 1
			} else {
				votes = 1
			}
		}

		// Проверяем, чтобы не перерасходовать
		if votes > remainingVotes {
			votes = remainingVotes
		}
		remainingVotes -= votes

		distribution = append(distribution, DistributionData{
			ProjectID:   projects[projectIndex].ID,
			VotesAmount: votes,
		})
	}

	// Проверка, если оставшиеся голоса не равны нулю (подстраховка)
	if remainingVotes != 0 {
		for i := range distribution {
			if remainingVotes == 0 {
				break
			}
			distribution[i].VotesAmount++
			remainingVotes--
		}
	}

	return distribution
}

func DoVotes(
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
		return fmt.Errorf("%s | No Available Votes", accountData.AccountAddress.String())
	}

	eligibleVotes := votesData.Data.TotalEligibleVotes
	usedVotes := votesData.Data.UsedVotes
	availableVotes := eligibleVotes - usedVotes

	log.Printf("%s | Eligible Votes: %d | Already Used Votes: %d | Available Votes: %d",
		accountData.AccountAddress.String(), eligibleVotes, usedVotes, availableVotes)

	if availableVotes <= 0 {
		return fmt.Errorf("%s | No Available Votes", accountData.AccountAddress.String())
	}

	projectsList := retroActions.GetProjectsList(client, accountData, accessToken, refreshToken)
	distribution := generateDistribution(projectsList, availableVotes)

	for i, data := range distribution {
		err = retroActions.DoVote(client, accountData, accessToken, refreshToken,
			data.ProjectID, data.VotesAmount)

		if err != nil {
			log.Printf("%v", err)
		} else {
			log.Printf("%s | [%d/%d] | Successfully Voted to %s: %d Votes", accountData.AccountAddress.String(),
				i+1, len(distribution), data.ProjectID, data.VotesAmount)
		}
	}

	var notConfirmedVotes []string
	votesData = retroActions.GetVotes(client, accountData, accessToken, refreshToken)

	for _, voteData := range votesData.Data.Votes {
		if !voteData.IsConfirmed {
			notConfirmedVotes = append(notConfirmedVotes, voteData.Id)
		}
	}

	if notConfirmedVotes == nil {
		return fmt.Errorf("%s | No Not Confirmed Votes", accountData.AccountAddress.String())
	}

	err = retroActions.ApproveVotes(client, accountData, accessToken, refreshToken, notConfirmedVotes)

	if err != nil {
		return fmt.Errorf("%s | Failed to approve votes: %s", accountData.AccountAddress.String(), err)
	}

	log.Printf("%s | Successfully Approved", accountData.AccountAddress.String())

	return nil
}
