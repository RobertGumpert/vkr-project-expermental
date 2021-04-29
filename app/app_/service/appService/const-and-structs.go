package appService

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"sort"
)

var (
	ErrorGateQueueIsFilled                   = errors.New("Error Gate Queue Is Filled. ")
	ErrorRequestReceivedLater                = errors.New("Error Request Received Later. ")
	ErrorRepositoryDoesntNearestRepositories = errors.New("Error Repository Doesnt Nearest Repositories. ")
)

type JsonCreateTaskFindNearestRepositories struct {
	Keyword string `json:"keyword"`
	Name    string `json:"name"`
	Owner   string `json:"owner"`
	Email   string `json:"email"`
}

//
//
//

type JsonUserRequest struct {
	UserKeyword string `json:"user_keyword"`
	UserName    string `json:"user_name"`
	UserOwner   string `json:"user_owner"`
	UserEmail   string `json:"user_email"`
}

type JsonFromGetNearestRepositories struct {
	UserRequest JsonUserRequest `json:"user_request"`
	//
	Repositories map[uint]float64 `json:"repositories"`
}

//
//----------------------------------------------------------------------------------------------------------------------
//

type JsonUserRepository struct {
	URL   string `json:"url"`
	Name  string `json:"name"`
	Owner string `json:"owner"`
	//
	Topics      []string `json:"topics"`
	Description string   `json:"description"`
}

type JsonStateTask struct {
	IsDefer  bool   `json:"is_defer"`
	Endpoint string `json:"endpoint"`
}

type JsonResultTaskFindNearestRepositories struct {
	Defer bool `json:"defer"`
	//
	TaskState *JsonStateTask `json:"task_state"`
	//
	UserRequest *JsonUserRequest `json:"user_request"`
	//
	UserRepository *JsonUserRepository     `json:"user_repository"`
	Top            []JsonNearestRepository `json:"top"`
}

type JsonNearestRepository struct {
	URL   string `json:"url"`
	Name  string `json:"name"`
	Owner string `json:"owner"`
	//
	Topics      []string `json:"topics"`
	Description string   `json:"description"`
	//
	TopicsIntersections      []string `json:"topics_intersections"`
	DescriptionIntersections []string `json:"description_intersections"`
	//
	DescriptionDistance     float64 `json:"description_distance"`
	NumberPairIntersections float64 `json:"number_pair_intersections"`
}

func (find *JsonResultTaskFindNearestRepositories) makeTop() {
	sort.Slice(find.Top, func(i, j int) bool {
		return find.Top[i].NumberPairIntersections > find.Top[j].NumberPairIntersections
	})
	if len(find.Top) > 10 {
		find.Top = find.Top[:10]
	}
}

func (find *JsonResultTaskFindNearestRepositories) encodeHash() (hash string, err error) {
	bts, err := json.Marshal(find)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bts), nil
}

func (find *JsonResultTaskFindNearestRepositories) decodeHash(hash string) (err error) {
	bts, err := base64.URLEncoding.DecodeString(hash)
	if err != nil {
		return err
	}
	f := new(JsonResultTaskFindNearestRepositories)
	err = json.Unmarshal(bts, f)
	if err != nil {
		return err
	}
	find.Top = f.Top
	find.Defer = f.Defer
	find.UserRepository = f.UserRepository
	find.UserRequest = f.UserRequest
	return nil
}

//
//----------------------------------------------------------------------------------------------------------------------
//

type JsonNearestIssue struct {
	UserRepositoryName       string `json:"user_repository_name"`
	ComparableRepositoryName string `json:"comparable_repository_name"`
	//
	UserRepositoryTitle       string `json:"user_repository_title"`
	ComparableRepositoryTitle string `json:"comparable_repository_title"`
	//
	UserRepositoryURL       string `json:"user_repository_url"`
	ComparableRepositoryURL string `json:"comparable_repository_url"`
	//
	Rank        float64 `json:"rank"`
	TitleCosine float64 `json:"title_cosine"`
	BodyCosine  float64 `json:"body_cosine"`
	//
	Intersections []string `json:"intersections"`
}

type JsonNearestIssues struct {
	UserRepositoryName       string `json:"user_repository_name"`
	ComparableRepositoryName string `json:"comparable_repository_name"`
	//
	Top []JsonNearestIssue `json:"top"`
}

func (find *JsonNearestIssues) makeTop() {
	sort.Slice(find.Top, func(i, j int) bool {
		return find.Top[i].Rank > find.Top[j].Rank
	})
	if len(find.Top) > 20 {
		find.Top = find.Top[:20]
	}
}
