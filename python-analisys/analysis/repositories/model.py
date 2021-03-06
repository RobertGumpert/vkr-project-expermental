import json
import pathlib
import pickle
from typing import List, Dict, Tuple
import numpy as np
from sklearn.cluster import AffinityPropagation
from sklearn.decomposition import TruncatedSVD
from sklearn.feature_extraction.text import CountVectorizer, TfidfTransformer

from data.viewModels.repository import RepositoryViewModel


class Model:
    vectorizer: CountVectorizer
    tf_idf: TfidfTransformer
    clustering: AffinityPropagation

    def __save_corpus_hash(self, corpus: List[RepositoryViewModel]):
        path = pathlib.Path(__file__).parent.absolute()
        path = str(path) + '/hash/hash.json'
        with open(path, 'w') as outfile:
            json.dump([ob.__dict__ for ob in corpus], outfile)

    def __save_vocabulary(self):
        path = pathlib.Path(__file__).parent.absolute()
        path = str(path) + '/hash/vocabulary.pkl'
        pickle.dump(self.vectorizer.vocabulary_, open(path, "wb"))

    def __save_clustering(self):
        path = pathlib.Path(__file__).parent.absolute()
        path = str(path) + '/hash/clustering.pkl'
        pickle.dump(self.clustering, open(path, "wb"))

    def download_models(self):
        path = pathlib.Path(__file__).parent.absolute()
        self.vectorizer = CountVectorizer(
            stop_words='english',
            min_df=2,
            vocabulary=pickle.load(open(str(path) + '/hash/vocabulary.pkl', "rb"))
        )
        self.tf_idf = TfidfTransformer(use_idf=True)
        self.clustering = pickle.load(open(str(path) + '/hash/clustering.pkl', 'rb'))

    def download_corpus_from_hash(self) -> List[Tuple[str, str]]:
        path = pathlib.Path(__file__).parent.absolute()
        path = str(path) + '/hash/hash.json'
        corpus: List[Tuple[str, str]] = []
        with open(path, 'r') as outfile:
            data = json.load(outfile)
            for element in data:
                obj: RepositoryViewModel = RepositoryViewModel()
                obj.deserialize(element)
                corpus.append((obj.name, obj.clear_combined))
                del obj
        return corpus

    def create_corpus_from_list(self, list_repositories: List[RepositoryViewModel]) -> List[Tuple[str, str]]:
        corpus: List[Tuple[str, str]] = []
        corpus_to_hash: List[RepositoryViewModel] = []
        for repository in list_repositories:
            clear_text = repository.clear_combined.replace('\n', ' ').strip()
            repository.clear_combined = clear_text
            corpus.append((repository.name, clear_text))
            corpus_to_hash.append(repository)
        self.__save_corpus_hash(corpus_to_hash)
        del corpus_to_hash
        return corpus

    def create_corpus_from_dict(self, dict_repositories: Dict[str, RepositoryViewModel]) -> List[Tuple[str, str]]:
        corpus: List[Tuple[str, str]] = []
        corpus_to_hash: List[RepositoryViewModel] = []
        for key, repository in dict_repositories.items():
            clear_text = repository.clear_combined.replace('\n', ' ').strip()
            repository.clear_combined = clear_text
            corpus.append((repository.name, clear_text))
            corpus_to_hash.append(repository)
        self.__save_corpus_hash(corpus_to_hash)
        return corpus

    def create_clusters(self, corpus: List[Tuple[str, str]]):
        if self.vectorizer is None or self.clustering is None or self.tf_idf is None:
            self.vectorizer = CountVectorizer(stop_words='english', min_df=2)
            self.tf_idf = TfidfTransformer(use_idf=True)
            self.clustering = AffinityPropagation()
        text_list = []
        clusters = []
        for text_tuple in corpus:
            text_list.append(text_tuple[1])
        bag_of_words = self.vectorizer.fit_transform(text_list)
        self.tf_idf = self.tf_idf.fit(bag_of_words)
        vectors = self.tf_idf.transform(bag_of_words)
        self.clustering = self.clustering.fit(vectors.toarray())
        count = len(self.clustering.cluster_centers_)
        for cluster in range(count):
            clusters.append('cluster_' + str(cluster))
        svd = TruncatedSVD(n_components=len(clusters))
        lsa = svd.fit_transform(vectors)
        #
        self.__save_vocabulary()
        self.__save_clustering()
        return

    # def create_clusters(self, corpus: List[Tuple[str, str]]):
    #     text_list = []
    #     clusters = []
    #     for text_tuple in corpus:
    #         text_list.append(text_tuple[1])
    #     self.vectorizer = CountVectorizer(stop_words='english', min_df=2)
    #     self.__save_vocabulary()
    #     bag_of_words = self.vectorizer.fit_transform(text_list)
    #     self.tf_idf = TfidfTransformer(use_idf=True).fit(bag_of_words)
    #     vectors = self.tf_idf.transform(bag_of_words)
    #     clustering = AffinityPropagation().fit(vectors.toarray())
    #     count = len(clustering.cluster_centers_)
    #     for cluster in range(count):
    #         clusters.append('cluster_' + str(cluster))
    #     svd = TruncatedSVD(n_components=len(clusters))
    #     lsa = svd.fit_transform(vectors)

    # def select_train_texts_from_corpus(self, corpus: List[Tuple[str, str]]) -> (
    #         List[Tuple[str, str]], List[Tuple[str, str]]):
    #     train_corpus: List[Tuple[str, str]] = []
    #     test_corpus: List[Tuple[str, str]] = []
    #     for text_tuple in corpus:
    #         vectorizer = CountVectorizer(stop_words='english')
    #         try:
    #             vector = vectorizer.fit_transform([text_tuple[1]])
    #         except ValueError:
    #             test_corpus.append(text_tuple)
    #             continue
    #         max_freq = np.max(vector.toarray())
    #         if max_freq < 2:
    #             test_corpus.append(text_tuple)
    #         else:
    #             train_corpus.append(text_tuple)
    #     if len(test_corpus) < (len(corpus) / 5):
    #         for text_tuple in test_corpus:
    #             train_corpus.append(text_tuple)
    #         test_corpus = []
    #     return train_corpus, test_corpus
