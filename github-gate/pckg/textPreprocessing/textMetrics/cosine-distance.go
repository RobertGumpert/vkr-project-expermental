package textMetrics

import (
	"errors"
	"issue-indexer/pckg/textPreprocessing"
	"math"
	"sync"
)


func CosineDistanceOnPairVectors(bagOfWords [][]float64) (float64, error) {
	return calculateCosineDistanceInPair(bagOfWords[0], bagOfWords[1])
}

func CosineDistance(bagOfWords [][]float64, mode textPreprocessing.ThreadMode) [][]float64 {
	if mode == textPreprocessing.ParallelMode {
		return parallelCalculateCosineDistance(bagOfWords)
	}
	return linearCalculateCosineDistance(bagOfWords)
}

func linearCalculateCosineDistance(bagOfWords [][]float64) [][]float64 {
	var (
		cosineMatrix = make([][]float64, len(bagOfWords))
	)
	for i := 0; i < len(bagOfWords); i++ {
		matrixRow := make([]float64, len(bagOfWords))
		for j := 0; j < len(bagOfWords); j++ {
			cosineDistance, err := calculateCosineDistanceInPair(
				bagOfWords[i],
				bagOfWords[j],
				)
			if err != nil {
				matrixRow[j] = float64(-1)
			} else {
				matrixRow[j] = cosineDistance
			}
		}
		cosineMatrix[i] = matrixRow
	}
	return cosineMatrix
}

func parallelCalculateCosineDistance(bagOfWords [][]float64) [][]float64 {
	var (
		wg = new(sync.WaitGroup)
		cosineMatrix = make([][]float64, len(bagOfWords))
	)
	for i := 0; i < len(bagOfWords); i++ {
		wg.Add(1)
		go func(i int, bagOfWords [][]float64, cosineMatrix [][]float64, wg *sync.WaitGroup) {
			defer wg.Done()
			matrixRow := make([]float64, len(bagOfWords))
			for j := 0; j < len(bagOfWords); j++ {
				cosineDistance, err := calculateCosineDistanceInPair(
					bagOfWords[i],
					bagOfWords[j],
				)
				if err != nil {
					matrixRow[j] = float64(-1)
				} else {
					matrixRow[j] = cosineDistance
				}
			}
			cosineMatrix[i] = matrixRow
			return
		}(i, bagOfWords, cosineMatrix, wg)
	}
	wg.Wait()
	return cosineMatrix
}

func calculateCosineDistanceInPair(vecA, vecB []float64) (float64, error) {
	if len(vecA) != len(vecB) {
		return 0, errors.New("Vectors must be equals size. ")
	}
	var (
		numerator, denumerator = float64(0), float64(0)
		squareA, squareB       = float64(0), float64(0)
	)
	for i := 0; i < len(vecA); i++ {
		ai := switchToFloat64(vecA[i])
		bi := switchToFloat64(vecB[i])
		numerator += ai * bi
		squareA += ai * ai
		squareB += bi * bi
	}
	squareA = math.Sqrt(squareA)
	squareB = math.Sqrt(squareB)
	denumerator = squareA * squareB
	return numerator / denumerator, nil
}
