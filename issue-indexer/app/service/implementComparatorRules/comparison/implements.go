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
	"math"
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
	return bagOfWords, (intersectionMatrix[0] + intersectionMatrix[1]) / 2, intersections, nil
}

func (implement *ImplementRules) CompareTitlesWithConditionIntersection(a, b dataModel.IssueModel, rules *issueCompator.CompareRules) (nearest dataModel.NearestIssuesModel, err error) {
	bagOfWords, numberIntersections, intersections, err := implement.compareTitlesByConditionIntersections(
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
	cosineDistance = cosineDistance * 100
	nearest = dataModel.NearestIssuesModel{
		RepositoryID:             a.RepositoryID,
		IssueID:                  a.ID,
		NearestIssueID:           b.ID,
		RepositoryIDNearestIssue: b.RepositoryID,
		TitleNumberIntersections: numberIntersections,
		TitleCosineDistance:      cosineDistance,
		BodyNumberIntersections:  0,
		BodyCosineDistance:       0,
		Rank:                     0,
		CosineDistance:           cosineDistance,
		Intersections:            intersections,
	}
	return nearest, nil
}

func (implement *ImplementRules) CompareBodyAfterCompareTitles(a, b dataModel.IssueModel, rules *issueCompator.CompareRules) (nearest dataModel.NearestIssuesModel, err error) {
	titleBagOfWords, titleNumberIntersections, titleIntersections, err := implement.compareTitlesByConditionIntersections(
		a,
		b,
		rules,
	)
	if err != nil {
		return nearest, err
	}
	titleCosineDistance, err := textMetrics.CosineDistanceOnPairVectors(titleBagOfWords)
	if err != nil {
		return nearest, err
	}
	if math.IsNaN(titleCosineDistance) {
		titleCosineDistance = 0.0
	}
	if math.IsNaN(titleNumberIntersections) {
		titleNumberIntersections = 0.0
	}
	dictionary, vectorOfWords, _ := textDictionary.FullDictionary([]string{a.Body, b.Body}, textPreprocessing.LinearMode)
	bodyBagOfWords := textVectorized.FrequencyVectorized(vectorOfWords, dictionary, textPreprocessing.LinearMode)
	bodyCosineDistance, err := textMetrics.CosineDistanceOnPairVectors(bodyBagOfWords)
	if err != nil {
		return nearest, err
	}
	bodyMatrixIntersections := textMetrics.CompletenessText(bodyBagOfWords, textPreprocessing.LinearMode)
	bodyNumberIntersections := (bodyMatrixIntersections[0] + bodyMatrixIntersections[1]) / 2
	if math.IsNaN(bodyCosineDistance) {
		bodyCosineDistance = 0.0
	}
	if math.IsNaN(bodyNumberIntersections) {
		bodyNumberIntersections = 0.0
	}
	titleCosineDistance = titleCosineDistance * 100
	bodyCosineDistance = bodyCosineDistance * 100
	//
	titleRank := (titleCosineDistance + titleNumberIntersections) / 2
	bodyRank := (bodyCosineDistance + bodyNumberIntersections) / 2
	rank := (titleRank + bodyRank) / 200
	nearest = dataModel.NearestIssuesModel{
		RepositoryID:             a.RepositoryID,
		IssueID:                  a.ID,
		NearestIssueID:           b.ID,
		RepositoryIDNearestIssue: b.RepositoryID,
		TitleNumberIntersections: titleNumberIntersections,
		TitleCosineDistance:      titleCosineDistance,
		BodyNumberIntersections:  bodyNumberIntersections,
		BodyCosineDistance:       bodyCosineDistance,
		Rank:                     rank,
		CosineDistance:           0,
		Intersections:            titleIntersections,
	}
	return nearest, nil
}
