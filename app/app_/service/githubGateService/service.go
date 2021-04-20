package githubGateService

import (
	"app/app_/config"
	"errors"
	"github.com/RobertGumpert/vkr-pckg/requests"
	"net/http"
	"strings"
)

type Service struct {
	client *http.Client
	config *config.Config
}

func NewService(client *http.Client, config *config.Config) *Service {
	return &Service{client: client, config: config}
}

func (service *Service) CreateTaskNewRepositoryWithNewKeyword(name, owner, keyword string, userRequest JsonUserRequest) (err error) {
	var (
		url = strings.Join([]string{
			service.config.GithubGateAddress,
			service.config.GithubGateEndpoints.NewRepositoryNewKeyword,
		}, "/")
	)
	response, err := requests.POST(
		service.client,
		url,
		nil,
		jsonCreateTaskNewRepositoryWithNewKeyword{
			UserRequest: userRequest,
			Keyword:     keyword,
			Repository: jsonRepositoryModel{
				Name:  name,
				Owner: owner,
			},
		},
	)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		return errors.New("Status not 200. ")
	}
	return nil
}

func (service *Service) CreateTaskNewRepositoryWithExistKeyword(name, owner string, userRequest JsonUserRequest) (err error) {
	var (
		url = strings.Join([]string{
			service.config.GithubGateAddress,
			service.config.GithubGateEndpoints.NewRepositoryExistKeyword,
		}, "/")
	)
	response, err := requests.POST(
		service.client,
		url,
		nil,
		jsonCreateTaskNewRepositoryWithExistKeyword{
			UserRequest: userRequest,
			Repositories: []jsonRepositoryModel{
				{
					Name:  name,
					Owner: owner,
				},
			},
		},
	)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		return errors.New("Status not 200. ")
	}
	return nil
}

func (service *Service) CreateTaskExistRepositoryReindexing(name, owner string, userRequest JsonUserRequest) (err error) {
	var (
		url = strings.Join([]string{
			service.config.GithubGateAddress,
			service.config.GithubGateEndpoints.ExistRepositoryUpdateNearest,
		}, "/")
	)
	response, err := requests.POST(
		service.client,
		url,
		nil,
		jsonCreateTaskExistRepositoryReindexing{
			UserRequest:  userRequest,
			RepositoryID: 0,
			Repository: jsonRepositoryModel{
				Name:  name,
				Owner: owner,
			},
		},
	)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		return errors.New("Status not 200. ")
	}
	return nil
}
