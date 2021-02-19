package repositories_collector

//
//type List concurrentmap.ConcurrentMap
//
//func (list List) Get(key string) []mapper.Repository {
//	element, ok := concurrentmap.ConcurrentMap(list).Get(key)
//	if !ok {
//		return nil
//	}
//	return element.([]mapper.Repository)
//}
//
//func (list List) Serialize() ([]byte, string, error) {
//	bs, err := json.Marshal(concurrentmap.ConcurrentMap(list))
//	if err != nil {
//		fmt.Println(err)
//		return nil, "", err
//	}
//	return bs, string(bs), nil
//}
//
//func LanguagesCollect(maxCountPages, countElementsOnPage int64, languages ...mapper.Language) List {
//	if languages == nil || len(languages) == 0 {
//		return nil
//	}
//	var (
//		url            = "https://api.github.com/search/repositories?q=language:%s"
//		concurrentMaps = concurrentmap.New()
//		iterators      = make([]pagesiterator.Configurator, 0)
//		wg             = new(sync.WaitGroup)
//	)
//	for _, language := range languages {
//		iteratorConstructor := pagesiterator.NewConfiguration(
//			pagesiterator.SetDefault(
//				fmt.Sprintf(url, string(language)),
//				string(language),
//				0,
//			),
//			pagesiterator.SetSizeBuffer(
//				maxCountPages,
//				countElementsOnPage,
//			),
//			pagesiterator.SetHttpHeaders(
//				map[string]string{
//					"Accept": "application/vnd.github.mercy-preview+json",
//				},
//			),
//		)
//		iterators = append(iterators, iteratorConstructor)
//	}
//	pagesIterator := pagesiterator.NewPagesIterator(
//		10,
//		iterators...,
//	)
//	start := time.Now()
//	pagesIterator.DO()
//	for language := range pagesIterator.Iterators {
//		wg.Add(1)
//		go func(language string, pagesIterator *pagesiterator.PagesIterator, wg *sync.WaitGroup) {
//			defer wg.Done()
//			iterator := pagesIterator.Get(language)
//			repositories := iterateHttpResponsesChannel(iterator.Responses)
//			concurrentMaps.Set(language, *repositories)
//		}(language, pagesIterator, wg)
//	}
//	wg.Wait()
//	duration := time.Since(start)
//	fmt.Println(duration)
//	return List(concurrentMaps)
//}
//
//func iterateHttpResponsesChannel(responses chan *http.Response) *[]*mapper.Repository {
//	var (
//		mx = new(sync.Mutex)
//		wg = new(sync.WaitGroup)
//		repositories = make([]*mapper.Repository, 0)
//	)
//	start := time.Now()
//	for response := range responses {
//		wg.Add(1)
//		go func(repositories *[]*mapper.Repository, response *http.Response, wg *sync.WaitGroup, mx *sync.Mutex) {
//			defer wg.Done()
//			list := new(mapper.JSONRepositoriesList)
//			err := json.NewDecoder(response.Body).Decode(list)
//			if err != nil {
//				fmt.Println(runtimeinfo.Runtime(1), "; error:", err)
//			} else {
//				for i := 0; i < len(list.Repositories); i++ {
//					var (
//						description string
//						repository  = list.Repositories[i]
//					)
//					if strings.TrimSpace((*repository).Description) == "" && len((*repository).Topics) == 0 {
//						continue
//					}
//					if len((*repository).Topics) != 0 && strings.TrimSpace((*repository).Description) != "" {
//						description = strings.Join([]string{
//							(*repository).Description,
//							strings.Join((*repository).Topics, " "),
//						}, " ")
//					} else {
//						if strings.TrimSpace((*repository).Description) != "" {
//							description = (*repository).Description
//						}
//						if len((*repository).Topics) != 0 {
//							description = strings.Join([]string{
//								strings.Join((*repository).Topics, " "),
//							}, " ")
//						}
//					}
//					(*repository).DescriptionPreprossecing = textPreprocessing.NewTextPreprocessor(description).DO()
//				}
//				mx.Lock()
//				*repositories = append(*repositories, list.Repositories...)
//				mx.Unlock()
//			}
//			response.Body.Close()
//			return
//		}(&repositories, response, wg, mx)
//	}
//	wg.Wait()
//	duration := time.Since(start)
//	fmt.Println(duration)
//	return &repositories
//}
