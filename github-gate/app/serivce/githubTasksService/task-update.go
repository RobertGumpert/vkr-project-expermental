package githubTasksService

import (
	"encoding/json"
	"errors"
	"github-gate/app/models/dataModel"
	"github-gate/app/models/interapplicationModels/githubCollectorModels"
	"github-gate/pckg/runtimeinfo"
	"github-gate/pckg/textPreprocessing/textClearing"
	"github-gate/pckg/textPreprocessing/textDictionary"
	"github-gate/pckg/textPreprocessing/textVectorized"
	"github.com/aaaton/golem/v4"
	"github.com/aaaton/golem/v4/dicts/en"
	"strings"
)

var (
	lemmatizer, _ = golem.New(en.New())
)

func (service *GithubTasksService) updateTaskRepositoriesDescriptionByURL(updateStateTask *githubCollectorModels.UpdateTaskRepositoriesByURLS) (error, []dataModel.Repository) {
	var (
		taskForCollector *TaskForCollector
		taskKey          = updateStateTask.ExecutionTaskStatus.TaskKey
		executionStatus  = updateStateTask.ExecutionTaskStatus.TaskCompleted
		listRepositories = updateStateTask.Repositories
		repositories     = make([]dataModel.Repository, 0)
	)
	if strings.TrimSpace(taskKey) == "" {
		return errors.New("GETTING FROM GITHUB-COLLECTOR TASK WITH EMPTY KEY. "), nil
	}
	for i := 0; i < len(service.tasksForCollectorsQueue); i++ {
		if service.tasksForCollectorsQueue[i].GetKey() == taskKey {
			taskForCollector = service.tasksForCollectorsQueue[i]
			break
		}
	}
	if taskForCollector == nil {
		return errors.New("GETTING FROM GITHUB-COLLECTOR TASK WITH KEY [" + taskKey + "] ISN'T EXIST. "), nil
	}
	runtimeinfo.LogInfo("GETTING UPDATE (status: ", executionStatus, ") FOR TASK [", taskKey, "] WITH LIST ELEMENTS SIZE OF [", len(listRepositories), "]")
	repositories = service.createRepositoryDataModels(updateStateTask.Repositories)
	err := service.db.AddRepositories(repositories)
	if err != nil {
		return err, nil
	}
	if taskForCollector.GetResult() != nil {
		slice := taskForCollector.GetResult().([]dataModel.Repository)
		pointer := &slice
		*pointer = append(*pointer, repositories...)
	}
	if executionStatus {
		taskForCollector.SetExecutionStatus(true)
		service.completedTasksChannel <- taskForCollector
	} else {
		taskForCollector.SetExecutionStatus(false)
	}
	return nil, repositories
}

func (service *GithubTasksService) updateTaskRepositoryIssues(updateStateTask *githubCollectorModels.UpdateTaskRepositoryIssues) (error, []dataModel.Issue) {
	var (
		taskForCollector *TaskForCollector
		taskKey          = updateStateTask.ExecutionTaskStatus.TaskKey
		executionStatus  = updateStateTask.ExecutionTaskStatus.TaskCompleted
		listIssues       = updateStateTask.Issues
		issues           = make([]dataModel.Issue, 0)
	)
	if strings.TrimSpace(taskKey) == "" {
		return errors.New("GETTING FROM GITHUB-COLLECTOR TASK WITH EMPTY KEY. "), nil
	}
	for i := 0; i < len(service.tasksForCollectorsQueue); i++ {
		if service.tasksForCollectorsQueue[i].GetKey() == taskKey {
			taskForCollector = service.tasksForCollectorsQueue[i]
			break
		}
	}
	if taskForCollector == nil {
		return errors.New("GETTING FROM GITHUB-COLLECTOR TASK WITH KEY [" + taskKey + "] ISN'T EXIST. "), nil
	}
	if taskForCollector.details.GetEntityID() == 0 {
		split := strings.Split(updateStateTask.Issues[0].URL, "/")
		repositoryName := split[len(split)-3]
		repository, err := service.db.GetRepositoryByName(repositoryName)
		if err != nil {
			return err, nil
		}
		taskForCollector.details.SetEntityID(repository.ID)
	}
	runtimeinfo.LogInfo("GETTING UPDATE (status: ", executionStatus, ") FOR TASK [", taskKey, "] WITH LIST ELEMENTS SIZE OF [", len(listIssues), "]")
	issues = service.createIssueDataModels(taskForCollector.details.GetEntityID(), updateStateTask.Issues)
	err := service.db.AddIssues(issues)
	if err != nil {
		return err, nil
	}
	if taskForCollector.GetResult() != nil {
		slice := taskForCollector.GetResult().([]dataModel.Issue)
		pointer := &slice
		*pointer = append(*pointer, issues...)
	}
	if executionStatus {
		taskForCollector.SetExecutionStatus(true)
		service.completedTasksChannel <- taskForCollector
	} else {
		taskForCollector.SetExecutionStatus(false)
	}
	return nil, issues
}

func (service *GithubTasksService) createRepositoryDataModels(updateStateTask []githubCollectorModels.UpdateTaskRepository) []dataModel.Repository {
	var (
		repositories = make([]dataModel.Repository, 0)
	)
	for i := 0; i < len(updateStateTask); i++ {
		repository := updateStateTask[i]
		if repository.Err != nil {
			runtimeinfo.LogError("CREATE REPOSITORY DATA MODELS ERROR: ", repository.Err)
			continue
		}
		split := strings.Split(repository.URL, "/")
		name := split[len(split)-1]
		owner := split[len(split)-2]
		textClearing.ClearASCII(&repository.Description)
		textClearing.ClearSymbols(&repository.Description)
		textClearing.ClearSpecialWord(&repository.Description)
		slice := textClearing.GetLemmas(&repository.Description, false, lemmatizer)
		repository.Description = strings.Join(*slice, " ")
		//
		topics := strings.Join(repository.Topics, " ")
		textClearing.ClearASCII(&topics)
		textClearing.ClearSymbols(&topics)
		repository.Topics = *(textClearing.GetLemmas(&topics, false, lemmatizer))
		//
		repositories = append(
			repositories,
			dataModel.Repository{
				URL:         repository.URL,
				Name:        name,
				Owner:       owner,
				Topics:      repository.Topics,
				Description: repository.Description,
			},
		)
	}
	return repositories
}

func (service *GithubTasksService) createIssueDataModels(repositoryID uint, updateStateTask []githubCollectorModels.UpdateTaskIssue) []dataModel.Issue {
	var (
		issues = make([]dataModel.Issue, 0)
	)
	for i := 0; i < len(updateStateTask); i++ {
		issue := updateStateTask[i]
		if issue.Err != nil {
			runtimeinfo.LogError("CREATE ISSUE DATA MODELS ERROR: ", issue.Err)
			continue
		}
		textClearing.ClearASCII(&issue.Title)
		textClearing.ClearSymbols(&issue.Title)
		slice := textClearing.GetLemmas(&issue.Title, false, lemmatizer)
		issue.Title = strings.Join(*slice, " ")
		//
		dictionary := textDictionary.TextTransformToFeaturesSlice(issue.Title)
		frequency := textVectorized.GetFrequencyMap(dictionary)
		m := make(map[string]float64, 0)
		for item := range frequency.IterBuffered() {
			m[item.Key] = item.Val.(float64)
		}
		frequencyJsonBytes, _ := json.Marshal(&dataModel.TitleFrequencyJSON{Dictionary: m})
		//
		issues = append(
			issues,
			dataModel.Issue{
				RepositoryID:       repositoryID,
				Number:             issue.Number,
				URL:                issue.URL,
				Title:              issue.Title,
				State:              issue.State,
				Body:               issue.Body,
				TitleDictionary:    dictionary,
				TitleFrequencyJSON: frequencyJsonBytes,
			},
		)
	}
	return issues
}
