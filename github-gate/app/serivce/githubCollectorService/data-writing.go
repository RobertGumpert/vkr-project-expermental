package githubCollectorService

import (
	"encoding/json"
	"github.com/RobertGumpert/vkr-pckg/dataModel"
	"github.com/RobertGumpert/vkr-pckg/runtimeinfo"
	"github.com/RobertGumpert/vkr-pckg/textPreprocessing/textClearing"
	"github.com/RobertGumpert/vkr-pckg/textPreprocessing/textDictionary"
	"github.com/RobertGumpert/vkr-pckg/textPreprocessing/textVectorized"
	"github.com/aaaton/golem/v4"
	"github.com/aaaton/golem/v4/dicts/en"
	"strings"
)

var (
	lemmatizer, _ = golem.New(en.New())
)

func (service *CollectorService) getRepositoryNameFromURL(url string) (name, owner string) {
	split := strings.Split(url, "/")
	name = split[len(split)-1]
	owner = split[len(split)-2]
	return name, owner
}

func (service *CollectorService) writeRepositoriesToDB(repositories []repositoryDescription) (models []dataModel.RepositoryModel) {
	models = make([]dataModel.RepositoryModel, 0)
	for _, repository := range repositories {
		if repository.Err != nil {
			runtimeinfo.LogError("CREATE REPOSITORY DATA MODELS ERROR: ", repository.Err)
			continue
		}
		name, owner := service.getRepositoryNameFromURL(repository.URL)
		textClearing.ClearASCII(&repository.Description)
		textClearing.ClearSymbols(&repository.Description)
		textClearing.ClearSpecialWord(&repository.Description)
		slice := textClearing.GetLemmas(&repository.Description, false, lemmatizer)
		repository.Description = strings.Join(*slice, " ")
		topics := strings.Join(repository.Topics, " ")
		textClearing.ClearASCII(&topics)
		textClearing.ClearSymbols(&topics)
		repository.Topics = *(textClearing.GetLemmas(&topics, false, lemmatizer))
		models = append(
			models,
			dataModel.RepositoryModel{
				URL:         repository.URL,
				Name:        name,
				Owner:       owner,
				Topics:      repository.Topics,
				Description: repository.Description,
			},
		)
	}
	err := service.repository.AddRepositories(models)
	if err != nil {
		runtimeinfo.LogError("WRITE REPOSITORIES MODELS ERROR: ", err)
		return models
	}
	return models
}

func (service *CollectorService) writeIssuesToDB(issues []issueDescription, repositoryID uint) (models []dataModel.IssueModel) {
	models = make([]dataModel.IssueModel, 0)
	for _, issue := range issues {
		if issue.Err != nil {
			runtimeinfo.LogError("CREATE REPOSITORY DATA MODELS ERROR: ", issue.Err)
			continue
		}
		textClearing.ClearASCII(&issue.Title)
		textClearing.ClearSymbols(&issue.Title)
		slice := textClearing.GetLemmas(&issue.Title, false, lemmatizer)
		issue.Title = strings.Join(*slice, " ")
		dictionary := textDictionary.TextTransformToFeaturesSlice(issue.Title)
		frequency := textVectorized.GetFrequencyMap(dictionary)
		m := make(map[string]float64, 0)
		for item := range frequency.IterBuffered() {
			m[item.Key] = item.Val.(float64)
		}
		frequencyJsonBytes, _ := json.Marshal(&dataModel.TitleFrequencyJSON{Dictionary: m})
		models = append(
			models,
			dataModel.IssueModel{
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
	err := service.repository.AddIssues(models)
	if err != nil {
		runtimeinfo.LogError("WRITE ISSUES MODELS ERROR: ", err)
		return models
	}
	return models
}