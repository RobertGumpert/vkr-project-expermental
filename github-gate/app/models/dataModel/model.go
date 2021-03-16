package dataModel

import (
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Repository struct {
	gorm.Model
	URL         string         `gorm:"not null;"`
	Name        string         `gorm:"not null; index:repository_name,unique;"`
	Owner       string         `gorm:"not null;"`
	Topics      pq.StringArray `gorm:"not null; type:text[];"`
	Description string         `gorm:"not null;"`
	Issues      []Issue        `gorm:"foreignKey:RepositoryID; constraint:OnDelete:CASCADE;"`
	//
	NearestIssuesOtherRepositories         []NearestIssues `gorm:"foreignKey:RepositoryID; constraint:OnDelete:CASCADE;"`
	IssuesNearestToIssuesOtherRepositories []NearestIssues `gorm:"foreignKey:RepositoryIDNearestIssue; constraint:OnDelete:CASCADE;"`
}

type Issue struct {
	gorm.Model
	RepositoryID            uint
	Number                  int             `gorm:"not null;"`
	URL                     string          `gorm:"not null;"`
	Title                   string          `gorm:"not null;"`
	State                   string          `gorm:"not null;"`
	Body                    string          `gorm:"not null;"`
	TitleDictionary         pq.StringArray  `gorm:"not null; type:text[];"`
	TitleFrequencyJSON      []byte          `gorm:"not null;"`
	//
	NearestIssuesOtherRepositories   []NearestIssues `gorm:"foreignKey:IssueID; constraint:OnDelete:CASCADE;"`
	NearestToIssuesOtherRepositories []NearestIssues `gorm:"foreignKey:NearestIssueID; constraint:OnDelete:CASCADE;"`
}

type NearestIssues struct {
	gorm.Model
	RepositoryID uint
	IssueID      uint
	//
	NearestIssueID           uint
	RepositoryIDNearestIssue uint
	//
	CosineDistance float64        `gorm:"not null;"`
	Intersections  pq.StringArray `gorm:"not null; type:text[];"`
}
