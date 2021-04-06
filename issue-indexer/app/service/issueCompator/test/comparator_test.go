package test

import (
	"encoding/json"
	"github.com/RobertGumpert/vkr-pckg/dataModel"
	"github.com/RobertGumpert/vkr-pckg/repository"
	"github.com/RobertGumpert/vkr-pckg/textPreprocessing/textDictionary"
	"github.com/RobertGumpert/vkr-pckg/textPreprocessing/textVectorized"
	"issue-indexer/app/service/implementComparatorRules/comparison"
	"issue-indexer/app/service/implementComparatorRules/sampling"
	"issue-indexer/app/service/issueCompator"
	"log"
	"strconv"
	"testing"
	"time"
)

var (
	countIssues       = 5
	countRepositories = 5
	storageProvider   = repository.SQLCreateConnection(
		repository.TypeStoragePostgres,
		repository.DSNPostgres,
		nil,
		"postgres",
		"toster123",
		"vkr-db",
		"5432",
		"disable",
	)
)

func connect() repository.IRepository {
	sqlRepository := repository.NewSQLRepository(
		storageProvider,
	)
	return sqlRepository
}

func TestTruncate(t *testing.T) {
	_ = connect()
	storageProvider.SqlDB.Exec("TRUNCATE TABLE repositories CASCADE")
	storageProvider.SqlDB.Exec("TRUNCATE TABLE issues CASCADE")
}

func createFakeData(db repository.IRepository) {
	for i := 0; i < countRepositories; i++ {
		err := db.AddRepository(&dataModel.RepositoryModel{
			URL:         "a" + strconv.Itoa(i),
			Name:        "a" + strconv.Itoa(i),
			Owner:       "a" + strconv.Itoa(i),
			Topics:      []string{"a", "a", "a"},
			Description: "a",
		})
		if err != nil {
			log.Fatal(err)
		}
	}
	for i := 0; i < countRepositories; i++ {
		repoID := i + 1
		for j := 0; j < countIssues; j++ {
			title := ""
			if j%2 == 0 {
				title = "injecting body post request into component"
			} else {
				title = "injecting body post"
			}
			dictionary := textDictionary.TextTransformToFeaturesSlice(title)
			frequency := textVectorized.GetFrequencyMap(dictionary)
			m := make(map[string]float64, 0)
			for item := range frequency.IterBuffered() {
				m[item.Key] = item.Val.(float64)
			}
			frequencyJsonBytes, _ := json.Marshal(&dataModel.TitleFrequencyJSON{Dictionary: m})
			err := db.AddIssue(&dataModel.IssueModel{
				RepositoryID:       uint(repoID),
				Number:             j,
				URL:                "a",
				Title:              "a",
				State:              "a",
				Body:               "a",
				TitleDictionary:    []string{"a"},
				TitleFrequencyJSON: frequencyJsonBytes,
			})
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func createRules(db repository.IRepository,
	repositoryID uint, maxCountThreads int64,
	identifier interface{},
	returnResultFromComparator issueCompator.ReturnResult,
	conditionComparison comparison.ConditionIntersections,
	conditionSampling sampling.ConditionIssuesFromGroupRepository) (*issueCompator.CompareRules, *issueCompator.CompareResult) {
	//
	comparisonRules := comparison.NewImplementRules()
	samplingRules := sampling.NewSampler(
		db,
		1*time.Minute,
	)
	rules := issueCompator.NewCompareRules(
		repositoryID,
		maxCountThreads,
		samplingRules.IssuesOnlyFromGroupRepositories,
		comparisonRules.CompareTitlesWithConditionIntersection,
		returnResultFromComparator,
		conditionComparison,
		conditionSampling,
	)
	result := issueCompator.NewCompareResult(identifier)
	return rules, result
}

func TestMigration(t *testing.T) {
	storageProvider.SqlDB.Exec("drop table repository_models cascade")
	storageProvider.SqlDB.Exec("drop table issue_models cascade")
	storageProvider.SqlDB.Exec("drop table nearest_issues_models cascade")
	storageProvider.SqlDB.Exec("drop table nearest_repositories_models cascade")
	storageProvider.SqlDB.Exec("drop table repositories_key_words_models cascade")
	storageProvider.SqlDB.Exec("drop table number_issue_intersections_models cascade")
	_ = connect()
}

func Test(t *testing.T) {
	db := connect()
	createFakeData(db)
	comparator := issueCompator.NewComparator(
		db,
	)
	conditionComparision := comparison.ConditionIntersections{CrossingThreshold: 90.0}
	for i := 0; i < countRepositories; i++ {
		repoID := uint(i) + 1
		conditionSampling := sampling.ConditionIssuesFromGroupRepository{
			RepositoryID:      repoID,
			GroupRepositories: make([]uint, 0),
		}
		for j := 0; j < countRepositories; j++ {
			comRepoID := uint(j) + 1
			if comRepoID == repoID {
				continue
			}
			conditionSampling.GroupRepositories = append(
				conditionSampling.GroupRepositories,
				comRepoID,
			)
		}
		rules, result := createRules(
			db,
			repoID,
			2,
			i,
			returnResult,
			conditionComparision,
			conditionSampling,
		)
		err := comparator.DOCompare(rules, result)
		if err != nil {
			t.Fatal(err)
		}
	}
	time.Sleep(1 * time.Hour)
}

func returnResult(result *issueCompator.CompareResult) {
	if result.GetErr() != nil {
		log.Println(result.GetErr())
		return
	}
	log.Println("COMPLETED FOR: ", result.GetIdentifier())
}
