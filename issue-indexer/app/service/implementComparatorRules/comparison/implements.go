package comparison

import (
	"encoding/json"
	"errors"
	"github.com/RobertGumpert/vkr-pckg/dataModel"
	"github.com/RobertGumpert/vkr-pckg/textPreprocessing"
	"github.com/RobertGumpert/vkr-pckg/textPreprocessing/textMetrics"
	"github.com/RobertGumpert/vkr-pckg/textPreprocessing/textVectorized"
	concurrentMap "github.com/streamrail/concurrent-map"
	"issue-indexer/app/service/issueCompator"
)

type ImplementRules struct{

}

func NewImplementRules() *ImplementRules {
	return &ImplementRules{}
}

func (implement *ImplementRules) CompareTitlesWithConditionIntersection(a, b dataModel.IssueModel, rules *issueCompator.CompareRules) (nearest dataModel.NearestIssuesModel, err error) {
	var (
		intersectionCondition = rules.GetComparisonCondition().(ConditionIntersections)
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
	if err := json.Unmarshal(a.TitleFrequencyJSON, &frequencyIssueA); err != nil {
		return nearest, err
	}
	if err := json.Unmarshal(b.TitleFrequencyJSON, &frequencyIssueB); err != nil {
		return nearest, err
	}
	dictionaryIssueA := convertToConcurrent(frequencyIssueA.Dictionary)
	dictionaryIssueB := convertToConcurrent(frequencyIssueB.Dictionary)
	bagOfWords, _, intersections := textVectorized.VectorizedPairDictionaries(dictionaryIssueA, dictionaryIssueB)
	intersectionMatrix := textMetrics.CompletenessText(bagOfWords, textPreprocessing.LinearMode)
	if intersectionMatrix[0] < intersectionCondition.CrossingThreshold ||
		intersectionMatrix[1] < intersectionCondition.CrossingThreshold {
		return nearest, errors.New("Text(s) isn't completeness on dictionary. ")
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
		CosineDistance:           cosineDistance,
		Intersections:            intersections,
	}
	return nearest, nil
}
