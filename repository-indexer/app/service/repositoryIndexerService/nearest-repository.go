package repositoryIndexerService

type nearestRepository struct {
	name, text string
	nearest map[string]float64
}

func (nr nearestRepository) GetText() string {
	return nr.text
}

func (nr nearestRepository) GetRepositoryName() string {
	return nr.name
}

func (nr nearestRepository) GetNearestRepositories() map[string]float64 {
	return nr.nearest
}