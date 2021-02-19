package main

import (
	"fmt"
	"go-agregator/pckg/runtimeinfo"
	issuesCollector "go-agregator/pckg/scratching/github-api/issues-collector"
	mapper "go-agregator/pckg/scratching/github-api/mapper"
	repositoriesCollector "go-agregator/pckg/scratching/github-api/repositories-collector"
	textPreprocessing "go-agregator/pckg/scratching/text-preprocessing"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

//
// "./test.txt"
//
func ReadBytes(path string) (string, []byte) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
		return "", nil
	}
	return string(content), content
}

func deserialize(file string) issuesCollector.List {
	absPath, err := filepath.Abs(fmt.Sprintf("../go-agregator/data/tests/%s", file))
	if err != nil {
		fmt.Println(runtimeinfo.Runtime(1), ", ERROR: ", err)
		return nil
	}
	_, content := ReadBytes(absPath)
	return issuesCollector.DeserializeToList(content)
}


func getRepos() {
	list, err := repositoriesCollector.CustomizableSearchCollect(
		repositoriesCollector.NewConfiguration(
			"frontends",
			repositoriesCollector.SetPage(0),
			repositoriesCollector.SetCountPagesAll(),
			repositoriesCollector.SetLanguage(mapper.JavaScript),
			repositoriesCollector.SetTextIn("interface"),
		),
		repositoriesCollector.NewConfiguration(
			"http-clients",
			repositoriesCollector.SetPage(0),
			repositoriesCollector.SetCountPagesAll(),
			repositoriesCollector.SetTextIn("http client"),
		),
		repositoriesCollector.NewConfiguration(
			"terminals",
			repositoriesCollector.SetPage(0),
			repositoriesCollector.SetCountPagesAll(),
			repositoriesCollector.SetTextIn("terminal"),
		),
	)
	if err != nil {
		fmt.Println(err)
	} else {
		wg := new(sync.WaitGroup)
		keys := list.GetKeys()
		for _, key := range keys {
			wg.Add(1)
			go func(key string, list *repositoriesCollector.List, wg *sync.WaitGroup) {
				defer wg.Done()
				topics := []string{
					"\n**\nTopics\n\n",
				}
				descriptions := []string{
					"\n**\nDescriptions\n\n",
				}
				repos := list.GetRepositories(key)
				for _, repo := range repos {
					if len(repo.Topics) == 0 && strings.TrimSpace(repo.Description) == "" {
						continue
					}
					if len(repo.Topics) != 0 {
						topics = append(topics, repo.Topics...)
					}
					if strings.TrimSpace(repo.Description) != "" {
						descriptions = append(descriptions, repo.Description)
					}
				}
				name := key + ".txt"
				content := strings.Join(append(topics, descriptions...), "\n")
				absPath, err := filepath.Abs(fmt.Sprintf("../go-agregator/data/tests/%s", name))
				if err != nil {
					fmt.Println(runtimeinfo.Runtime(1), ", ERROR: ", err)
				}
				f, err := os.Create(absPath)
				if err != nil {
					fmt.Println(runtimeinfo.Runtime(1), ", ERROR: ", err)
				}
				defer f.Close()
				_, err = f.WriteString(content)
				if err != nil {
					fmt.Println(runtimeinfo.Runtime(1), ", ERROR: ", err)
				}

			}(key, &list, wg)
		}
		wg.Wait()
	}
}

//
// main := "alacritty.txt"
// fmt.Println(
//		fmt.Sprintf(
//			"%s <-> %s = %.05f",
//			main,
//			"flask.txt",
//			distanceTitles(main, "flask.txt"),
//		),
//	)
func distanceTitles(mainFile, secondFile string) float64 {
	var (
		read = func(file string) *textPreprocessing.TextPreprocessor {
			absPath, err := filepath.Abs(fmt.Sprintf("../go-agregator/data/tests/%s", file))
			if err != nil {
				fmt.Println(runtimeinfo.Runtime(1), ", ERROR: ", err)
				return nil
			}
			content, _ := ReadBytes(absPath)
			titles := strings.Split(strings.Split(content, "\n**\nBodies:\n")[0], "\n**\nTitles:\n")
			preprocessor := textPreprocessing.NewTextPreprocessor(titles[1]).DOPullThread(5)
			return preprocessor
		}
	)
	mainFilePP := read(mainFile)
	secondFilePP := read(secondFile)
	return textPreprocessing.CosineDistance(mainFilePP.LemmasMI, secondFilePP.LemmasMI)
}

//repos := []string{"react.txt", "angular.txt", "vue.txt", "gin.txt", "flask.txt", "okhttp.txt", "hyper.txt", "terminal.txt", "alacritty.txt"}
//fmt.Println("S,Wab,Vab,Vcab,Cd")
//for _, main := range repos {
//fmt.Println("Main: ", main)
//for _, repo := range repos {
//d := distanceTopicsDescription(main, repo)
//if d == -1 {
//continue
//}
//fmt.Println()
////fmt.Println(
////	fmt.Sprintf(
////		"%s <-> %s = %.05f",
////		main,
////		repo,
////		d,
////	),
////)
//}
//fmt.Println("\n------------------------------------------------------------------------------------------\n")
//}
func distanceTopicsDescription(mainFile, secondFile string) float64 {
	var (
		read = func(file string) *textPreprocessing.TextPreprocessor {
			absPath, err := filepath.Abs(fmt.Sprintf("../go-agregator/data/tests/%s", file))
			if err != nil {
				fmt.Println(runtimeinfo.Runtime(1), ", ERROR: ", err)
				return nil
			}
			var (
				description = ""
				topics      = ""
				linux       = func(s string) string { return fmt.Sprintf("\r\n**\r\n%s:\r\n", s) }
				windows     = func(s string) string { return fmt.Sprintf("\n**\n%s:\n", s) }
			)
			content, _ := ReadBytes(absPath)
			if strings.Contains(content, linux("Description")) && strings.Contains(content, linux("Topics")) {
				description = strings.Split(
					strings.Split(content, linux("Description"))[1],
					linux("Topics"),
				)[0]
				topics = strings.Split(
					strings.Split(content, linux("Topics"))[1],
					linux("Titles"),
				)[0]
			} else {
				description = strings.Split(
					strings.Split(content, windows("Description"))[1],
					windows("Topics"),
				)[0]
				topics = strings.Split(
					strings.Split(content, windows("Topics"))[1],
					windows("Titles"),
				)[0]
			}
			result := ""
			if len(topics) != 0 && len(description) != 0 {
				result = strings.Join([]string{description, topics}, " ")
			}
			if len(topics) != 0 && len(description) == 0 {
				result = topics
			}
			if len(description) != 0 && len(topics) == 0 {
				result = description
			}
			preprocessor := textPreprocessing.NewTextPreprocessor(result).DO()
			return preprocessor
		}
	)
	mainFilePP := read(mainFile)
	secondFilePP := read(secondFile)
	sk := float64(0)
	//
	fmt.Println(
		fmt.Sprintf(
			"%s <-> %s",
			mainFile, secondFile,
		),
	)
	_, _, crossingA, crossingB, divergenceA, divergenceB := textPreprocessing.GetCrossing(mainFilePP.LemmasFrequency, secondFilePP.LemmasFrequency)
	kab, _ := textPreprocessing.MethodSetOperations2(crossingA, crossingB, divergenceA, divergenceB)
	//if ab == 0 {
	//	return 0
	//}
	if kab == 0 {
		fmt.Println("\t\tNone")
		return 0
	}

	//fmt.Println(
	//	fmt.Sprintf(
	//		"AB = %d, NotAB = %d, Kab = %.05f, NotKab = %.05f \t\t %s <-> %s",
	//		ab, notAB, kab, knotab, mainFile, secondFile,
	//	),
	//)

	//s, wab, vab, vcab, cd := textPreprocessing.ExperimentalMethod(mainFilePP.LemmasFrequency, secondFilePP.LemmasFrequency)
	////
	////s, _, _, _, _ := textPreprocessing.ExperimentalMethod(mainFilePP.LemmasFrequency, secondFilePP.LemmasFrequency)
	//if s == 0 {
	//	return -1
	//}
	////if wab > 0.50 {
	////	return -1
	////}
	//if vab == 0 {
	//	fmt.Print(fmt.Sprintf("S = %.02f,  Wab = %.02f,  Vab = %.02f, Vcab = %.02f, Cd = %.02f \t\t", s, wab, vab, vcab, cd))
	//	return 1
	//}
	//sk := float64(0)
	//sk = 1 - s*wab*cd / vab
	////
	//fmt.Print(fmt.Sprintf("S = %.02f,  Wab = %.02f,  Vab = %.02f, Vcab = %.02f, Cd = %.02f \t\t", s, wab, vab, vcab, cd))
	////
	////fmt.Print(fmt.Sprintf("S = %.02f \t\t", s))
	//// fmt.Println(fmt.Sprintf("[%.02f,%.02f,%.02f,%.02f,%.02f],", s, wab, vab, vcab, cd))
	return sk
}

func mostFreqLemmas(file string) {
	absPath, err := filepath.Abs(fmt.Sprintf("../go-agregator/data/tests/%s", file))
	if err != nil {
		fmt.Println(runtimeinfo.Runtime(1), ", ERROR: ", err)
		return
	}
	content, _ := ReadBytes(absPath)
	titles := strings.Split(strings.Split(content, "\n**\nBodies:\n")[0], "\n**\nTitles:\n")[1]
	preprocessor := textPreprocessing.NewTextPreprocessor(titles).DOPullThread(5)
	fmt.Println(preprocessor)
	for mi := range preprocessor.LemmasFrequency.IterBuffered() {
		val := mi.Val
		key := mi.Key
		if val.(int64) > 60 {
			fmt.Println(fmt.Sprintf("%s = %d", key, val.(int64)))
		}
	}
}

//
//go func() {
//	fmt.Println("Downloading is start...")
//	code := downloadIssues(
//		[]string{
//			"https://api.github.com/repos/alacritty/alacritty",
//			"https://api.github.com/repos/angular/angular",
//			"https://api.github.com/repos/pallets/flask",
//		},
//		"angular_alacritty_flask.json",
//	)
//	fmt.Println("Downloading endless with code ", code, "...")
//}()
//
// "https://api.github.com/repos/facebook/react"
//
func downloadIssues(repos []string, jsonFile string) int {
	list, _ := issuesCollector.Collect(
		repos[0],
		repos[1:]...,
	)
	bts, str, _ := list.Serialize()
	absPath, err := filepath.Abs(fmt.Sprintf("../go-agregator/data/tests/%s", jsonFile))
	if err != nil {
		fmt.Println(runtimeinfo.Runtime(1), ", ERROR: ", err)
		fmt.Println(str)
		return 0
	}
	err = ioutil.WriteFile(absPath, bts, 0644)
	if err != nil {
		fmt.Println(runtimeinfo.Runtime(1), ", ERROR: ", err)
		fmt.Println(str)
		return 0
	}
	wg := new(sync.WaitGroup)
	for _, repo := range repos {
		wg.Add(1)
		go func(repo string, list issuesCollector.List, wg *sync.WaitGroup) {
			defer wg.Done()
			issues := list.GetIssues(repo)
			if issues != nil {
				var (
					titles = []string{
						"\n**\nTitles:\n",
					}
					bodies = []string{
						"\n**\nBodies:\n",
					}
				)
				for i := 0; i < len(issues); i++ {
					titles = append(titles, issues[i].Title)
					bodies = append(bodies, issues[i].Body)
				}
				split := strings.Split(repo, "/")
				name := split[len(split)-1] + ".txt"
				content := strings.Join(append(titles, bodies...), "\n")
				absPath, err := filepath.Abs(fmt.Sprintf("../go-agregator/data/tests/%s", name))
				if err != nil {
					fmt.Println(runtimeinfo.Runtime(1), ", ERROR: ", err)
					return
				}
				f, err := os.Create(absPath)
				if err != nil {
					fmt.Println(runtimeinfo.Runtime(1), ", ERROR: ", err)
					return
				}
				defer f.Close()
				_, err = f.WriteString(content)
				if err != nil {
					fmt.Println(runtimeinfo.Runtime(1), ", ERROR: ", err)
					return
				}
			} else {
				fmt.Println("In repo ", repo, " empty issues list. ")
			}
		}(repo, list, wg)
	}
	wg.Wait()
	return 1
}
