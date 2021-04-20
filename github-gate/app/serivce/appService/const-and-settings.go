package appService

import (
	"errors"
	"github.com/RobertGumpert/gotasker/itask"
)

var (
	ErrorEmptyOrIncompleteJSONData = errors.New("Empty Or Incomplete JSON Data. ")
)

const (
	TaskTypeDownloadRepositoryByName           itask.Type = 10
	TaskTypeDownloadRepositoryByKeyWord        itask.Type = 11
	TaskTypeRepositoryAndRepositoriesByKeyWord itask.Type = 12
	TaskTypeNewRepositoryWithExistKeyword      itask.Type = 100
	TaskTypeNewRepositoryWithNewKeyword        itask.Type = 101
	TaskTypeExistRepository                    itask.Type = 102
)

//
// JSON
//

type JsonUserRequest struct {
	UserKeyword string `json:"user_keyword"`
	UserName    string `json:"user_name"`
	UserOwner   string `json:"user_owner"`
	UserEmail   string `json:"user_email"`
}

type JsonRepository struct {
	Name  string `json:"name"`
	Owner string `json:"owner"`
}

type JsonNewRepositoryWithExistKeyword struct {
	UserRequest JsonUserRequest `json:"user_request"`
	//
	Repositories []JsonRepository `json:"repositories"`
}

type JsonNewRepositoryWithNewKeyword struct {
	UserRequest JsonUserRequest `json:"user_request"`
	//
	Keyword    string         `json:"keyword"`
	Repository JsonRepository `json:"repository"`
}

type JsonExistRepository struct {
	UserRequest JsonUserRequest `json:"user_request"`
	//
	RepositoryID uint           `json:"repository_id"`
	Repository   JsonRepository `json:"repository"`
}

//
//
//

type JsonSendToAppNearestRepositories struct {
	UserRequest JsonUserRequest `json:"user_request"`
	//
	Repositories map[uint]float64 `json:"repositories"`
}
