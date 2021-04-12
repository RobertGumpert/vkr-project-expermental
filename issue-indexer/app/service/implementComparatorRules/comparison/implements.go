package comparison

import (
	"encoding/json"
	"errors"
	"github.com/RobertGumpert/vkr-pckg/dataModel"
	"github.com/RobertGumpert/vkr-pckg/textPreprocessing"
	"github.com/RobertGumpert/vkr-pckg/textPreprocessing/textDictionary"
	"github.com/RobertGumpert/vkr-pckg/textPreprocessing/textMetrics"
	"github.com/RobertGumpert/vkr-pckg/textPreprocessing/textVectorized"
	concurrentMap "github.com/streamrail/concurrent-map"
	"issue-indexer/app/service/issueCompator"
	"strings"
)

type ImplementRules struct {
	stopWords map[string]int
}

func NewImplementRules() *ImplementRules {
	return &ImplementRules{
		stopWords: map[string]int{
			"readme":        0,
			"pull request":  0,
			"md":            0,
			"merge request": 0,
			"issue":         0,
		},
	}
}

func (implement *ImplementRules) compareTitlesByConditionIntersections(a, b dataModel.IssueModel, rules *issueCompator.CompareRules) (bagOfWords [][]float64, numberIntersections float64, intersections []string, err error) {
	var (
		intersectionCondition = rules.GetComparisonCondition().(*ConditionIntersections)
		frequencyIssueA       dataModel.TitleFrequencyJSON
		frequencyIssueB       dataModel.TitleFrequencyJSON
		convertToConcurrent   = func(m map[string]float64) concurrentMap.ConcurrentMap {
			dictionary := concurrentMap.New()
			for key, val := range m {
				dictionary.Set(key, val)
			}
			return dictionary
		}
	)
	for stop, _ := range implement.stopWords {
		if strings.Contains(a.Title, stop) || strings.Contains(b.Title, stop) {
			return nil, 0.0, nil, errors.New("Text(s) contains stop words. ")
		}
	}
	if len(a.TitleDictionary) < 3 || len(b.TitleDictionary) < 3 {
		return nil, 0.0, nil, errors.New("Text(s) contains stop words. ")
	}
	if err := json.Unmarshal(a.TitleFrequencyJSON, &frequencyIssueA); err != nil {
		return nil, 0.0, nil, err
	}
	if err := json.Unmarshal(b.TitleFrequencyJSON, &frequencyIssueB); err != nil {
		return nil, 0.0, nil, err
	}
	dictionaryIssueA := convertToConcurrent(frequencyIssueA.Dictionary)
	dictionaryIssueB := convertToConcurrent(frequencyIssueB.Dictionary)
	bagOfWords, _, intersections = textVectorized.VectorizedPairDictionaries(dictionaryIssueA, dictionaryIssueB)
	if len(intersections) == 0 {
		return nil, 0.0, nil, errors.New("Text(s) isn't completeness on dictionary. ")
	}
	intersectionMatrix := textMetrics.CompletenessText(bagOfWords, textPreprocessing.LinearMode)
	if intersectionMatrix[0] < intersectionCondition.CrossingThreshold ||
		intersectionMatrix[1] < intersectionCondition.CrossingThreshold {
		return nil, 0.0, nil, errors.New("Text(s) isn't completeness on dictionary. ")
	}
	return bagOfWords, intersectionMatrix[0], intersections, nil
}

func (implement *ImplementRules) CompareTitlesWithConditionIntersection(a, b dataModel.IssueModel, rules *issueCompator.CompareRules) (nearest dataModel.NearestIssuesModel, err error) {
	bagOfWords, _, intersections, err := implement.compareTitlesByConditionIntersections(
		a,
		b,
		rules,
	)
	if err != nil {
		return nearest, err
	}
	cosineDistance, err := textMetrics.CosineDistanceOnPairVectors(bagOfWords)
	if err != nil {
		return nearest, err
	}
	nearest = dataModel.NearestIssuesModel{
		RepositoryID:             a.RepositoryID,
		IssueID:                  a.ID,
		NearestIssueID:           b.ID,
		RepositoryIDNearestIssue: b.RepositoryID,
		CosineDistance:           cosineDistance * 100,
		Intersections:            intersections,
	}
	return nearest, nil
}

func (implement *ImplementRules) CompareBodyAfterCompareTitles(a, b dataModel.IssueModel, rules *issueCompator.CompareRules) (nearest dataModel.NearestIssuesModel, err error) {
	_, numberIntersections, intersections, err := implement.compareTitlesByConditionIntersections(
		a,
		b,
		rules,
	)
	if err != nil {
		return nearest, err
	}
	dictionary, vectorOfWords, _ := textDictionary.FullDictionary([]string{a.Body, b.Body}, textPreprocessing.LinearMode)
	bagOfWords := textVectorized.FrequencyVectorized(vectorOfWords, dictionary, textPreprocessing.LinearMode)
	cosineDistance, err := textMetrics.CosineDistanceOnPairVectors(bagOfWords)
	if err != nil {
		return nearest, err
	}
	cosineDistance = (cosineDistance*100 + numberIntersections) / 200
	nearest = dataModel.NearestIssuesModel{
		RepositoryID:             a.RepositoryID,
		IssueID:                  a.ID,
		NearestIssueID:           b.ID,
		RepositoryIDNearestIssue: b.RepositoryID,
		CosineDistance:           cosineDistance * 100,
		Intersections:            intersections,
	}
	return nearest, nil
}
