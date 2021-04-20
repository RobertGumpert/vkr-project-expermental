package githubGateService

type JsonUserRequest struct {
	UserKeyword string `json:"user_keyword"`
	UserName    string `json:"user_name"`
	UserOwner   string `json:"user_owner"`
	UserEmail   string `json:"user_email"`
}

type jsonRepositoryModel struct {
	Name  string `json:"name"`
	Owner string `json:"owner"`
}

type jsonCreateTaskNewRepositoryWithExistKeyword struct {
	UserRequest JsonUserRequest `json:"user_request"`
	//
	Repositories []jsonRepositoryModel `json:"repositories"`
}

type jsonCreateTaskNewRepositoryWithNewKeyword struct {
	UserRequest JsonUserRequest `json:"user_request"`
	//
	Keyword    string              `json:"keyword"`
	Repository jsonRepositoryModel `json:"repository"`
}

type jsonCreateTaskExistRepositoryReindexing struct {
	UserRequest JsonUserRequest `json:"user_request"`
	//
	RepositoryID uint                `json:"repository_id"`
	Repository   jsonRepositoryModel `json:"repository"`
}
