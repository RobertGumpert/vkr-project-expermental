package sampling

type ConditionIssuesBesidesRepository struct {
	RepositoryID uint
}

type ConditionIssuesFromGroupRepository struct {
	RepositoryID      uint
	GroupRepositories []uint
}
