package retroActions

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	"main/internal/util"
	"main/pkg/global"
	"main/pkg/types"
)

func GetSignText(
	client *fasthttp.Client,
	accountData types.AccountData,
) string {
	url := fmt.Sprintf("https://api-retro-9000.avax.network/api/auth/get-nonce/%s",
		accountData.AccountAddress.String())

	for {
		req := fasthttp.AcquireRequest()
		req.SetRequestURI(url)
		req.Header.SetMethod("GET")
		req.Header.Set("accept", "application/json, text/plain, */*")
		req.Header.Set("accept-language", "ru,en;q=0.9")
		req.Header.Set("origin", "https://retro9000.avax.network")
		req.Header.Set("referer", "https://retro9000.avax.network/")

		resp := fasthttp.AcquireResponse()

		err := client.Do(req, resp)
		if err != nil {
			log.Printf("%s | Error When Making a Request to Retrieve Sign Text %s",
				accountData.AccountAddress.String(), err)

			fasthttp.ReleaseRequest(req)
			fasthttp.ReleaseResponse(resp)
			continue
		}

		responseData := &getSignTextResponse{}

		if err = json.Unmarshal(resp.Body(), &responseData); err != nil {
			log.Printf("%s | Failed to Parse JSON Response While Retrieving Sign Text: %s, response: %s",
				accountData.AccountAddress.String(), err, string(resp.Body()))

			fasthttp.ReleaseRequest(req)
			fasthttp.ReleaseResponse(resp)
			continue
		}

		if responseData.StatusCode != 200 || responseData.Data.Nonce == "" {
			log.Printf("%s | Failed To Parse Sign Text: %s, response: %s",
				accountData.AccountAddress.String(), string(resp.Body()), string(resp.Body()))

			fasthttp.ReleaseRequest(req)
			fasthttp.ReleaseResponse(resp)
			continue
		}

		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(resp)

		return responseData.Data.Nonce
	}
}

func DoAuth(
	client *fasthttp.Client,
	accountData types.AccountData,
	signedMessage string,
) (string, string, error) {
	url := "https://api-retro-9000.avax.network/api/auth/login"
	payload := map[string]string{
		"walletAddress": accountData.AccountAddress.String(),
		"signature":     signedMessage,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", "", fmt.Errorf("%s | Failed to marshal JSON payload when Logging: %s",
			accountData.AccountAddress.String(), err)
	}

	for {
		req := fasthttp.AcquireRequest()
		req.SetRequestURI(url)
		req.Header.SetMethod("POST")
		req.Header.Set("accept", "application/json, text/plain, */*")
		req.Header.Set("accept-language", "ru,en;q=0.9")
		req.Header.Set("content-type", "application/json")
		req.Header.Set("origin", "https://retro9000.avax.network")
		req.Header.Set("referer", "https://retro9000.avax.network/")
		req.SetBody(payloadBytes)

		resp := fasthttp.AcquireResponse()

		err = client.Do(req, resp)
		if err != nil {
			log.Printf("%s | Error While Making a Login Request %s",
				accountData.AccountAddress.String(), err)

			fasthttp.ReleaseRequest(req)
			fasthttp.ReleaseResponse(resp)
			continue
		}

		responseData := &doLoginResponse{}

		if err = json.Unmarshal(resp.Body(), &responseData); err != nil {
			log.Printf("%s | Failed To Parse JSON Response While Logging In: %s, response: %s",
				accountData.AccountAddress.String(), err, string(resp.Body()))

			fasthttp.ReleaseRequest(req)
			fasthttp.ReleaseResponse(resp)
			continue
		}

		if responseData.StatusCode != 200 {
			log.Printf("%s | Wrong Response While Logging In: %s, response: %s",
				accountData.AccountAddress.String(), string(resp.Body()), string(resp.Body()))

			fasthttp.ReleaseRequest(req)
			fasthttp.ReleaseResponse(resp)
			continue
		}

		accessTokenCookie := resp.Header.PeekCookie("accessToken")
		refreshTokenCookie := resp.Header.PeekCookie("refreshToken")

		if accessTokenCookie == nil || refreshTokenCookie == nil {
			log.Printf("%s | No Cookies In response While Logging In, response: %s", accountData.AccountAddress.String(), string(resp.Body()))

			fasthttp.ReleaseRequest(req)
			fasthttp.ReleaseResponse(resp)
			continue
		}

		accessTokenCookieString := util.ExtractCookieValue(string(accessTokenCookie),
			"accessToken")
		refreshTokenCookieString := util.ExtractCookieValue(string(refreshTokenCookie),
			"refreshToken")

		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(resp)

		return accessTokenCookieString, refreshTokenCookieString, nil
	}
}

func GetVotesAmount(
	client *fasthttp.Client,
	accountData types.AccountData,
	accessToken string,
	refreshToken string,
) (int64, int64, int64) {
	url := fmt.Sprintf("https://api-retro-9000.avax.network/api/vote/rounds/%s/ballot-votes",
		global.Const.RoundID)

	for {
		req := fasthttp.AcquireRequest()
		req.SetRequestURI(url)
		req.Header.SetMethod("GET")
		req.Header.Set("accept", "application/json, text/plain, */*")
		req.Header.Set("accept-language", "ru,en;q=0.9")
		req.Header.Set("origin", "https://retro9000.avax.network")
		req.Header.Set("referer", "https://retro9000.avax.network/")
		req.Header.SetCookie("accessToken", accessToken)
		req.Header.SetCookie("refreshToken", refreshToken)

		resp := fasthttp.AcquireResponse()

		err := client.Do(req, resp)
		if err != nil {
			log.Printf("%s | Error When Parsing Available Votes %s",
				accountData.AccountAddress.String(), err)

			fasthttp.ReleaseRequest(req)
			fasthttp.ReleaseResponse(resp)
			continue
		}

		responseData := &GetVotesResponse{}

		if err = json.Unmarshal(resp.Body(), &responseData); err != nil {
			log.Printf("%s | Failed To Parse JSON Response When Parsing Available Votes: %s, response: %s",
				accountData.AccountAddress.String(), err, string(resp.Body()))

			fasthttp.ReleaseRequest(req)
			fasthttp.ReleaseResponse(resp)
			continue
		}

		if responseData.StatusCode == 404 && responseData.Message == "Ballot not found!" {
			return 0, 0, 0
		}

		if responseData.StatusCode != 200 {
			log.Printf("%s | Wrong Response When Parsing Available Votes: %s, response: %s",
				accountData.AccountAddress.String(), string(resp.Body()), string(resp.Body()))

			fasthttp.ReleaseRequest(req)
			fasthttp.ReleaseResponse(resp)
			continue
		}

		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(resp)

		return responseData.Data.TotalEligibleVotes, responseData.Data.UsedVotes,
			responseData.Data.TotalEligibleVotes - responseData.Data.UsedVotes
	}
}

func GetProjectsList(
	client *fasthttp.Client,
	accountData types.AccountData,
	accessToken string,
	refreshToken string,
) []ProjectData {
	url := fmt.Sprintf("https://api-retro-9000.avax.network/api/rounds/%s/submissions?roundId=%s&page=1&perPage=1000&sortBy=votes&sortOrder=desc&includeField=userVotes",
		global.Const.RoundID, global.Const.RoundID)

	for {
		req := fasthttp.AcquireRequest()
		req.SetRequestURI(url)
		req.Header.SetMethod("GET")
		req.Header.Set("accept", "application/json, text/plain, */*")
		req.Header.Set("accept-language", "ru,en;q=0.9")
		req.Header.Set("origin", "https://retro9000.avax.network")
		req.Header.Set("referer", "https://retro9000.avax.network/")
		req.Header.SetCookie("accessToken", accessToken)
		req.Header.SetCookie("refreshToken", refreshToken)

		resp := fasthttp.AcquireResponse()

		err := client.Do(req, resp)
		if err != nil {
			log.Printf("%s | Error When Parsing Projects List %s",
				accountData.AccountAddress.String(), err)

			fasthttp.ReleaseRequest(req)
			fasthttp.ReleaseResponse(resp)
			continue
		}

		responseData := &getProjectsListResponse{}

		if err = json.Unmarshal(resp.Body(), &responseData); err != nil {
			log.Printf("%s | Failed To Parse JSON Response When Parsing Projects List: %s, response: %s",
				accountData.AccountAddress.String(), err, string(resp.Body()))

			fasthttp.ReleaseRequest(req)
			fasthttp.ReleaseResponse(resp)
			continue
		}

		if responseData.StatusCode != 200 {
			log.Printf("%s | Wrong Response When Parsing Projects List: %s, response: %s",
				accountData.AccountAddress.String(), string(resp.Body()), string(resp.Body()))

			fasthttp.ReleaseRequest(req)
			fasthttp.ReleaseResponse(resp)
			continue
		}

		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(resp)

		return responseData.Data
	}
}

func DoVote(
	client *fasthttp.Client,
	accountData types.AccountData,
	accessToken string,
	refreshToken string,
	projectID string,
	voteCount int64,
) error {
	url := fmt.Sprintf("https://api-retro-9000.avax.network/api/vote/rounds/%s/projects/%s/vote",
		global.Const.RoundID, projectID)

	payload := map[string]int64{
		"voteCount": voteCount,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("%s | Failed to marshal JSON payload when Logging: %s",
			accountData.AccountAddress.String(), err)
	}

	for {
		req := fasthttp.AcquireRequest()
		req.SetRequestURI(url)
		req.Header.SetMethod("POST")
		req.Header.Set("accept", "application/json, text/plain, */*")
		req.Header.Set("accept-language", "ru,en;q=0.9")
		req.Header.Set("origin", "https://retro9000.avax.network")
		req.Header.Set("referer", "https://retro9000.avax.network/")
		req.Header.Set("content-type", "application/json")
		req.Header.SetCookie("accessToken", accessToken)
		req.Header.SetCookie("refreshToken", refreshToken)
		req.SetBody(payloadBytes)

		resp := fasthttp.AcquireResponse()

		err = client.Do(req, resp)
		if err != nil {
			log.Printf("%s | Error When Voting %s",
				accountData.AccountAddress.String(), err)

			fasthttp.ReleaseRequest(req)
			fasthttp.ReleaseResponse(resp)
			continue
		}

		responseData := &doVoteResponse{}

		if err = json.Unmarshal(resp.Body(), &responseData); err != nil {
			log.Printf("%s | Failed To Parse JSON Response When Voting: %s, response: %s",
				accountData.AccountAddress.String(), err, string(resp.Body()))

			fasthttp.ReleaseRequest(req)
			fasthttp.ReleaseResponse(resp)
			continue
		}

		if responseData.StatusCode != 200 || responseData.Message != "Voting successful!" {
			log.Printf("%s | Wrong Response When Voting: %s, response: %s",
				accountData.AccountAddress.String(), string(resp.Body()), string(resp.Body()))

			fasthttp.ReleaseRequest(req)
			fasthttp.ReleaseResponse(resp)
			continue
		}

		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(resp)

		return nil
	}
}

func GetVotes(
	client *fasthttp.Client,
	accountData types.AccountData,
	accessToken string,
	refreshToken string,
) *GetVotesResponse {
	url := fmt.Sprintf("https://api-retro-9000.avax.network/api/vote/rounds/%s/ballot-votes",
		global.Const.RoundID)

	for {
		req := fasthttp.AcquireRequest()
		req.SetRequestURI(url)
		req.Header.SetMethod("GET")
		req.Header.Set("accept", "application/json, text/plain, */*")
		req.Header.Set("accept-language", "ru,en;q=0.9")
		req.Header.Set("origin", "https://retro9000.avax.network")
		req.Header.Set("referer", "https://retro9000.avax.network/")
		req.Header.SetCookie("accessToken", accessToken)
		req.Header.SetCookie("refreshToken", refreshToken)

		resp := fasthttp.AcquireResponse()

		err := client.Do(req, resp)
		if err != nil {
			log.Printf("%s | Error When Parsing Votes %s",
				accountData.AccountAddress.String(), err)

			fasthttp.ReleaseRequest(req)
			fasthttp.ReleaseResponse(resp)
			continue
		}

		responseData := &GetVotesResponse{}

		if err = json.Unmarshal(resp.Body(), &responseData); err != nil {
			log.Printf("%s | Failed To Parse JSON Response When Parsing Votes: %s, response: %s",
				accountData.AccountAddress.String(), err, string(resp.Body()))

			fasthttp.ReleaseRequest(req)
			fasthttp.ReleaseResponse(resp)
			continue
		}

		if responseData.StatusCode == 404 && responseData.Message == "Ballot not found!" {
			return responseData
		}

		if responseData == nil || responseData.StatusCode != 200 {
			log.Printf("%s | Wrong Response When Parsing Votes: %s, response: %s",
				accountData.AccountAddress.String(), string(resp.Body()), string(resp.Body()))

			fasthttp.ReleaseRequest(req)
			fasthttp.ReleaseResponse(resp)
			continue
		}

		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(resp)

		return responseData
	}
}

func ApproveVotes(
	client *fasthttp.Client,
	accountData types.AccountData,
	accessToken string,
	refreshToken string,
	votesIDs []string,
) error {
	payload := map[string][]map[string]string{
		"votes": {},
	}

	for _, id := range votesIDs {
		payload["votes"] = append(payload["votes"], map[string]string{"voteId": id})
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("%s | Error When Marshalling JSON When Approving Votes",
			accountData.AccountAddress.String())
	}

	url := fmt.Sprintf("https://api-retro-9000.avax.network/api/vote/rounds/%s/confirm-votes",
		global.Const.RoundID)

	for {
		req := fasthttp.AcquireRequest()
		req.SetRequestURI(url)
		req.Header.SetMethod("POST")
		req.Header.Set("accept", "application/json, text/plain, */*")
		req.Header.Set("accept-language", "ru,en;q=0.9")
		req.Header.Set("origin", "https://retro9000.avax.network")
		req.Header.Set("referer", "https://retro9000.avax.network/")
		req.Header.Set("content-type", "application/json")
		req.Header.SetCookie("accessToken", accessToken)
		req.Header.SetCookie("refreshToken", refreshToken)
		req.SetBody(payloadBytes)

		resp := fasthttp.AcquireResponse()

		err = client.Do(req, resp)
		if err != nil {
			log.Printf("%s | Error When Approving Votes %s",
				accountData.AccountAddress.String(), err)

			fasthttp.ReleaseRequest(req)
			fasthttp.ReleaseResponse(resp)
			continue
		}

		responseData := &GetVotesResponse{}

		if err = json.Unmarshal(resp.Body(), &responseData); err != nil {
			log.Printf("%s | Failed To Parse JSON Response When Approving Votes: %s, response: %s",
				accountData.AccountAddress.String(), err, string(resp.Body()))

			fasthttp.ReleaseRequest(req)
			fasthttp.ReleaseResponse(resp)
			continue
		}

		if responseData.StatusCode != 200 || responseData.Message != "Votes confirmed!" {
			log.Printf("%s | Wrong Response When Approving Votes: %s, response: %s",
				accountData.AccountAddress.String(), string(resp.Body()), string(resp.Body()))

			fasthttp.ReleaseRequest(req)
			fasthttp.ReleaseResponse(resp)
			continue
		}

		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(resp)

		return nil
	}
}

func DeleteVote(
	client *fasthttp.Client,
	accountData types.AccountData,
	accessToken string,
	refreshToken string,
	voteID string,
) error {
	url := fmt.Sprintf("https://api-retro-9000.avax.network/api/vote/projects/%s/vote",
		voteID)

	for {
		req := fasthttp.AcquireRequest()
		req.SetRequestURI(url)
		req.Header.SetMethod("DELETE")
		req.Header.Set("accept", "application/json, text/plain, */*")
		req.Header.Set("accept-language", "ru,en;q=0.9")
		req.Header.Set("origin", "https://retro9000.avax.network")
		req.Header.Set("referer", "https://retro9000.avax.network/")
		req.Header.Set("content-type", "application/json")
		req.Header.SetCookie("accessToken", accessToken)
		req.Header.SetCookie("refreshToken", refreshToken)

		resp := fasthttp.AcquireResponse()

		err := client.Do(req, resp)
		if err != nil {
			log.Printf("%s | Error When Deleting Votes %s",
				accountData.AccountAddress.String(), err)

			fasthttp.ReleaseRequest(req)
			fasthttp.ReleaseResponse(resp)
			continue
		}

		responseData := &GetVotesResponse{}

		if err = json.Unmarshal(resp.Body(), &responseData); err != nil {
			log.Printf("%s | Failed To Parse JSON Response When Deleting Votes: %s, response: %s",
				accountData.AccountAddress.String(), err, string(resp.Body()))

			fasthttp.ReleaseRequest(req)
			fasthttp.ReleaseResponse(resp)
			continue
		}

		if responseData.StatusCode != 200 || responseData.Message != "Vote deleted!" {
			log.Printf("%s | Wrong Response When Deleting Votes: %s, response: %s",
				accountData.AccountAddress.String(), string(resp.Body()), string(resp.Body()))

			fasthttp.ReleaseRequest(req)
			fasthttp.ReleaseResponse(resp)
			continue
		}

		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(resp)

		return nil
	}
}
