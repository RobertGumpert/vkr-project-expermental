package dataModel

import (
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Repository struct {
	gorm.Model
	URL              string          `gorm:"not null;"`
	Name             string          `gorm:"not null; index:repository_name,unique;"`
	Owner            string          `gorm:"not null;"`
	Topics           pq.StringArray  `gorm:"not null; type:text[];"`
	Description      string          `gorm:"not null;"`
	AllIssues        []Issue         `gorm:"foreignKey:RepositoryID"`
	IssuesTurnInPair []NearestIssues `gorm:"foreignKey:RepositoryID"`
}

type Issue struct {
	gorm.Model
	RepositoryID  uint
	Number        int             `gorm:"not null;"`
	URL           string          `gorm:"not null;"`
	Title         string          `gorm:"not null;"`
	State         string          `gorm:"not null;"`
	Body          string          `gorm:"not null;"`
	NearestIssues []NearestIssues `gorm:"foreignKey:IssueID"`
	TurnIn        []NearestIssues `gorm:"foreignKey:NearestIssueID"`
}

type NearestIssues struct {
	gorm.Model
	RepositoryID   uint
	IssueID        uint
	NearestIssueID uint
	CosineDistance float64        `gorm:"not null;"`
	Intersections  pq.StringArray `gorm:"not null; type:text[];"`
}

