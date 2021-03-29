package indexerService

type nearestRepository struct {
	id uint
	text string
	nearest map[uint]float64
}

func (nr nearestRepository) GetText() string {
	return nr.text
}

func (nr nearestRepository) GetRepositoryID() uint {
	return nr.id
}

func (nr nearestRepository) GetNearestRepositories() map[uint]float64 {
	return nr.nearest
}