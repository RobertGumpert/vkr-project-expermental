package repositoryIndexerService

import (
	"app/app_/config"
	"errors"
	"github.com/RobertGumpert/vkr-pckg/requests"
	"net/http"
	"strings"
)

type Service struct {
	config *config.Config
	client *http.Client
}

func NewService(config *config.Config, client *http.Client) *Service {
	return &Service{config: config, client: client}
}

func (service *Service) WordIsExist(word string) (jsonResponseBody *jsonSendFromServiceWordIsExist, err error) {
	var (
		url = strings.Join([]string{
			service.config.RepositoryIndexerAddress,
			service.config.RepositoryIndexerEndpoints.WordIsExist,
		}, "/")
	)
	response, err := requests.POST(
		service.client,
		url,
		nil,
		jsonSendToServiceWordIsExist{Word: word},
	)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.New("Status not 200. ")
	}
	jsonResponseBody = new(jsonSendFromServiceWordIsExist)
	err = requests.Deserialize(jsonResponseBody, response)
	return jsonResponseBody, err
}

func (service *Service) GetNearestRepositories(repositoryId uint) (jsonResponseBody *jsonSendFromServiceNearestRepositoriesForRepository, err error) {
	var (
		url = strings.Join([]string{
			service.config.RepositoryIndexerAddress,
			service.config.RepositoryIndexerEndpoints.GetNearestRepositories,
		}, "/")
	)
	response, err := requests.POST(
		service.client,
		url,
		nil,
		jsonSendToServiceNearestRepositoriesForRepository{RepositoryID:repositoryId},
	)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.New("Status not 200. ")
	}
	jsonResponseBody = new(jsonSendFromServiceNearestRepositoriesForRepository)
	err = requests.Deserialize(jsonResponseBody, response)
	return jsonResponseBody, err
}