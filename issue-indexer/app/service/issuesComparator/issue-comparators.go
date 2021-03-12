package issuesComparator

import (
	"errors"
	"issue-indexer/app/models/dataModel"
	"issue-indexer/pckg/textPreprocessing"
	"issue-indexer/pckg/textPreprocessing/textDictionary"
	"issue-indexer/pckg/textPreprocessing/textMetrics"
	"issue-indexer/pckg/textPreprocessing/textVectorized"
)

func (comparator *IssuesComparator) CompareOnlyTitles(i, j int, main, second []dataModel.Issue) (dataModel.NearestIssues, error) {
	var (
		corpus = make([]string, 2)
		nearestIssues dataModel.NearestIssues
	)
	corpus[0] = main[i].Title
	corpus[1] = second[j].Title
	dictionary, vectorsOfWords, countFeatures := textDictionary.FullDictionary(
		corpus,
		textPreprocessing.LinearMode,
	)
	if countFeatures == 0 {
		return nearestIssues, errors.New("Count features equal 0. ")
	}
	bagOfWords := textVectorized.FrequencyVectorized(
		vectorsOfWords,
		dictionary,
		textPreprocessing.LinearMode,
	)
	completenessMatrix := textMetrics.CompletenessText(bagOfWords, textPreprocessing.LinearMode)
	if completenessMatrix[0] < comparator.MinimumTextCompletenessThreshold ||
		completenessMatrix[1] < comparator.MinimumTextCompletenessThreshold {
		return nearestIssues, errors.New("Text(s) isn't completeness on dictionary. ")
	}
	cosineDistance, err := textMetrics.CosineDistanceOnPairVectors(bagOfWords)
	if err != nil {
		return nearestIssues, err
	}
	intersection := textMetrics.Intersections(bagOfWords, textPreprocessing.LinearMode)
	intersectionWords := make([]string, 0)
	for item := range dictionary.IterBuffered() {
		for _, index := range intersection[0][1].IntersectionIndices {
			if index == item.Val.(int64) {
				intersectionWords = append(
					intersectionWords,
					item.Key,
				)
			}
		}
	}
	nearestIssues = dataModel.NearestIssues{
		RepositoryID:   main[i].RepositoryID,
		IssueID:        main[i].ID,
		NearestIssueID: second[j].ID,
		CosineDistance: cosineDistance,
		Intersections:  intersectionWords,
	}
	return nearestIssues, nil
}
