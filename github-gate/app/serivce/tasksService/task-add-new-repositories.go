package tasksService

type AddRepositoriesTaskContext struct {
	repositoriesURLS []string
	//
	repositoriesHaveBeenDownloaded bool
	issuesHaveBeenIndexed          bool
}

func (t *AddRepositoriesTaskContext) IssuesHaveBeenIndexed() bool {
	return t.issuesHaveBeenIndexed
}

func (t *AddRepositoriesTaskContext) RepositoriesHaveBeenDownloaded() bool {
	return t.repositoriesHaveBeenDownloaded
}
