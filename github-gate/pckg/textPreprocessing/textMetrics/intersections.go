package textMetrics

import (
	"errors"
	"issue-indexer/pckg/textPreprocessing"
	"sync"
)

type intersection struct {
	IntersectionPercent float64
	IntersectionIndices []int64
}

func Intersections(bagOfWords [][]float64, mode textPreprocessing.ThreadMode) [][]intersection {
	if mode == textPreprocessing.ParallelMode {
		return parallelCalculateIntersections(bagOfWords)
	}
	return linearCalculateIntersections(bagOfWords)
}

func linearCalculateIntersections(bagOfWords [][]float64) [][]intersection {
	var (
		intersectionsMatrix = make([][]intersection, len(bagOfWords))
	)
	for i := 0; i < len(bagOfWords); i++ {
		matrixRow := make([]intersection, len(bagOfWords))
		for j := 0; j < len(bagOfWords); j++ {
			intersections, _ := calculatePairIntersections(bagOfWords[i], bagOfWords[j])
			matrixRow[j] = intersections
		}
		intersectionsMatrix[i] = matrixRow
	}
	return intersectionsMatrix
}

func parallelCalculateIntersections(bagOfWords [][]float64) [][]intersection {
	var (
		wg                  = new(sync.WaitGroup)
		intersectionsMatrix = make([][]intersection, len(bagOfWords))
	)
	for i := 0; i < len(bagOfWords); i++ {
		wg.Add(1)
		go func(i int, bagOfWords [][]float64, intersectionsMatrix [][]intersection, wg *sync.WaitGroup) {
			defer wg.Done()
			matrixRow := make([]intersection, len(bagOfWords))
			for j := 0; j < len(bagOfWords); j++ {
				intersections, _ := calculatePairIntersections(bagOfWords[i], bagOfWords[j])
				matrixRow[j] = intersections
			}
			intersectionsMatrix[i] = matrixRow
			return
		}(i, bagOfWords, intersectionsMatrix, wg)
	}
	wg.Wait()
	return intersectionsMatrix
}

func calculatePairIntersections(vecA, vecB []float64) (intersection, error) {
	if len(vecA) != len(vecB) {
		return intersection{}, errors.New("Vectors must be equals size. ")
	}
	var (
		intersection = intersection{
			IntersectionPercent: float64(0),
			IntersectionIndices: make([]int64, 0),
		}
	)
	for i := 0; i < len(vecA); i++ {
		ai := switchToFloat64(vecA[i])
		bi := switchToFloat64(vecB[i])
		if ai > 0 && bi > 0 {
			intersection.IntersectionPercent++
			intersection.IntersectionIndices = append(
				intersection.IntersectionIndices,
				int64(i),
			)
		}
		if ai == 0 && bi == 0 {
			intersection.IntersectionPercent++
		}
	}
	intersection.IntersectionPercent = (intersection.IntersectionPercent / float64(len(vecA))) * 100
	return intersection, nil
}
