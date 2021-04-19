package repositoryIndexerService

type jsonSendToServiceWordIsExist struct {
	Word string `json:"word"`
}

type jsonSendFromServiceWordIsExist struct {
	WordIsExist          bool `json:"word_is_exist"`
	DatabaseIsReindexing bool `json:"database_is_reindexing"`
}


type jsonNearestRepository struct {
	RepositoryID          uint             `json:"repository_id"`
	NearestRepositoriesID map[uint]float64 `json:"nearest_repositories_id"`
}

type jsonSendToServiceNearestRepositoriesForRepository struct {
	RepositoryID uint `json:"repository_id"`
}

type jsonSendFromServiceNearestRepositoriesForRepository struct {
	NearestRepositories  []jsonNearestRepository `json:"nearest_repositories"`
	DatabaseIsReindexing bool                    `json:"database_is_reindexing"`
}