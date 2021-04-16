package issueIndexerService

import (
	"errors"
	"github-gate/app/config"
	"github.com/RobertGumpert/gotasker/itask"
	"github.com/RobertGumpert/vkr-pckg/requests"
	"github.com/gin-gonic/gin"
	concurrentMap "github.com/streamrail/concurrent-map"
	"net/http"
	"strings"
)

type IndexerService struct {
	config                  *config.Config
	client                  *http.Client
	tasks                   concurrentMap.ConcurrentMap
	channelSendToAppService chan itask.ITask
}

func NewService(config *config.Config, channelSendToAppService chan itask.ITask, engine *gin.Engine) *IndexerService {
	service := new(IndexerService)
	service.client = new(http.Client)
	service.config = config
	service.ConcatTheirRestHandlers(engine)
	service.tasks = concurrentMap.New()
	service.channelSendToAppService = channelSendToAppService
	return service
}

func (service *IndexerService) ServiceQueueIsFilled() (err error) {
	response, err := requests.GET(service.client,
		strings.Join([]string{
			service.config.IssueIndexerAddress,
			service.config.IssueIndexerEndpoints.GetState,
		}, "/"),
		nil,
	)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		return errors.New("Service queue is filled. ")
	}
	return nil
}

func (service *IndexerService) CompareGroupRepositories(task itask.ITask) (err error) {
	if err := service.ServiceQueueIsFilled(); err != nil {
		return err
	}
	response, err := requests.POST(service.client,
		strings.Join([]string{
			service.config.IssueIndexerAddress,
			service.config.IssueIndexerEndpoints.CompareForGroupRepositories,
		}, "/"),
		nil,
		task.GetState().GetSendContext().(*JsonSendToIndexerCompareGroup),
	)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		return errors.New("Service queue is filled. ")
	}
	service.tasks.Set(task.GetKey(), task)
	return nil
}
