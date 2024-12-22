package retroActions

type getSignTextResponse struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
	Data       struct {
		Nonce string `json:"nonce"`
	} `json:"data"`
	Metadata interface{} `json:"metadata"`
	Error    interface{} `json:"error"`
}

type doLoginResponse struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
	Data       struct {
		TotalReferralPoints interface{} `json:"totalReferralPoints"`
		User                struct {
			ChillFactor   int64  `json:"chill_factor"`
			ReferralCode  string `json:"referral_code"`
			WalletAddress string `json:"wallet_address"`
		} `json:"user"`
	} `json:"data"`
	Metadata interface{} `json:"metadata"`
	Error    interface{} `json:"error"`
}

type GetVotesResponse struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
	Data       struct {
		Id                 string `json:"id"`
		TotalEligibleVotes int64  `json:"total_eligible_votes"`
		UsedVotes          int64  `json:"used_votes"`
		Votes              []struct {
			Id          string `json:"id"`
			IsConfirmed bool   `json:"is_confirmed"`
			Project     struct {
				BannerUrl   interface{} `json:"banner_url"`
				Description string      `json:"desription"`
				Id          string      `json:"id"`
				LogoURL     string      `json:"logo_url"`
				Name        string      `json:"name"`
				Status      string      `json:"status"`
				TotalVotes  int64       `json:"total_votes"`
			} `json:"project"`
			VoteCount int64 `json:"vote_count"`
		} `json:"votes"`
	} `json:"data"`
	Metadata interface{} `json:"metadata"`
	Error    interface{} `json:"error"`
}

type getProjectsListResponse struct {
	StatusCode int           `json:"statusCode"`
	Message    string        `json:"message"`
	Data       []ProjectData `json:"data"`
	Metadata   struct {
		Total       int  `json:"total"`
		LastPage    int  `json:"lastPage"`
		CurrentPage int  `json:"currentPage"`
		PerPage     int  `json:"perPage"`
		Prev        *int `json:"prev"`
		Next        *int `json:"next"`
	} `json:"metadata"`
	Error *string `json:"error"`
}

type ProjectData struct {
	ID                      string      `json:"id"`
	RoundID                 string      `json:"round_id"`
	CreatorID               string      `json:"creator_id"`
	Name                    string      `json:"name"`
	Description             string      `json:"description"`
	WebsiteURL              string      `json:"website_url"`
	LogoURL                 string      `json:"logo_url"`
	BannerURL               *string     `json:"banner_url"`
	TeamSize                int         `json:"team_size"`
	DeployerAddress         *string     `json:"deployer_address"`
	Categories              []string    `json:"categories"`
	Links                   []string    `json:"links"`
	Status                  string      `json:"status"`
	ApprovedAt              *string     `json:"approved_at"`
	ReviewNote              *string     `json:"review_note"`
	MetricsGHLastCommit     *string     `json:"metrics_gh_last_commit"`
	MetricsGHStars          interface{} `json:"metrics_gh_stars"`
	ProjectRank             int         `json:"project_rank"`
	UniqueVoters            int         `json:"unique_voters"`
	TotalVotes              int64       `json:"total_votes"`
	TotalVotesLastUpdatedAt string      `json:"total_votes_last_updated_at"`
	CreatedAt               string      `json:"created_at"`
	UpdatedAt               string      `json:"updated_at"`
	IsDeleted               bool        `json:"is_deleted"`
	DeletedAt               *string     `json:"deleted_at"`
}

type doVoteResponse struct {
	StatusCode int         `json:"statusCode"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data"`
	Metadata   interface{} `json:"metadata"`
	Error      *string     `json:"error"`
}
