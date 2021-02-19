package github_api

import text_preprocessing "go-agregator/pckg/scratching/text-preprocessing"

type Language string

const (
	//
	// 5000 / 60 = 83.3 -> 80
	//
	CoreRequestPerMinute   = 80
	SearchRequestPerMinute = 30
)

const (
	JavaScript Language = "javascript"
	Python     Language = "python"
	Java       Language = "java"
	TypeScript Language = "typescript"
	CSharp     Language = "c#"
	Php        Language = "php"
	CPP        Language = "c++"
	Ruby       Language = "ruby"
	Erlang     Language = "erlang"
	GoLang     Language = "go"
	Rust       Language = "rust"
	Kotlin     Language = "kotlin"
	Swift      Language = "swift"
	Html       Language = "html"
	Css        Language = "css"
)

func GetAllLanguageAlias() []Language {
	return []Language{
		JavaScript,
		Python,
		Java,
		TypeScript,
		CSharp,
		Php,
		CPP,
		Ruby,
		Erlang,
		GoLang,
		Rust,
		Kotlin,
		Swift,
		Html,
		Css,
	}
}

// ---------------------------------------------------------------------------------------------------------------------
// GET REPOS:
// 		Получаем по https://api.github.com/search/repositories?q=language:<lang_name>
// 		список всех репозиториев. Значение поля total_count делим на 100,
//  	получаем кол-во 100-элементных страниц total_page (number/100=total_page).
//  	Получив кол-во страниц total_page, вытаскиваем циклом все страницы
// 		https://api.github.com/search/repositories?q=language:<lang_name>&page=total_page&per_page=100.
//
// 	ERROR:
// 		GitHub не позволяет вытягивать более тысячи первых результатов,
// 		поэтому поиск будет прерван с ошибкой 422, как только общее кол-во репозиториев,
// 		с 0 страницы по текущую страницу в цикле запросов (0 до total_page), превысит 1000 элементов.
//
// ---------------------------------------------------------------------------------------------------------------------
// GET ISSUES:
// 		Получаем по https://api.github.com/repos/<user>/<name>/issues?state=all
//  	список последних issues. Из самого первого в списке issue берем поле number.
//  	Значение поля number делим на 100, получаем кол-во 100-элементных страниц total_page (number/100=total_page).
//  	Получив кол-во страниц total_page, вытаскиваем циклом все страницы
//  	https://api.github.com/repos/<user>/<name>/issues?state=all&page=total_page&per_page=100.
//
// EXAMPLE:
// 		"https://api.github.com/search/issues?q=use different middleware for different routes+language:go"
// 		-> находит https://github.com/gin-gonic/gin/issues/2612
//
// ---------------------------------------------------------------------------------------------------------------------

const AUTHToken = "token aef3219befb0b5a71ebfcf5876dd8c8d9eeb0077"

type JSONRepositoriesList struct {
	Count        int64         `json:"total_count"`
	Repositories []*Repository `json:"items"`
}

type Repository struct {
	//
	// GENERAL
	//
	ID          uint64   `json:"id"`
	NodeID      string   `json:"node_id"`
	Name        string   `json:"name"`
	FullName    string   `json:"full_name"`
	Description string   `json:"description"`
	Language    string   `json:"language"`
	Topics      []string `json:"topics"`
	//
	// FLAG'S
	//
	FlagHasIssues   bool `json:"has_issues"`
	FlagHasProjects bool `json:"has_projects"`
	FlagHasWiki     bool `json:"has_wiki"`
	//
	// URL'S
	//
	URLRepository string `json:"url"`
	URLIssues     string `json:"issues_url"`
	URLLanguages  string `json:"languages_url"`
	URLHomepage   string `json:"homepage_url"`
	//
	// Count's
	//
	CountOpenIssues int64 `json:"count_open_issues"`
	//
	// Text preprossing
	//
	DescriptionPreprossecing *text_preprocessing.TextPreprocessor `json:"description_preprossecing"`
}

type JSONIssuesList struct {
	Issues []*Issue
}

type Issue struct {
	//
	// GENERAL
	//
	ID     uint64 `json:"id"`
	Number uint64 `json:"number"`
	NodeID string `json:"node_id"`
	State  string `json:"state"`
	Body   string `json:"body"`
	Title  string `json:"title"`
}
