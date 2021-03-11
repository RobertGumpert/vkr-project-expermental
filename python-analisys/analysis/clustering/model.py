import json
import pathlib
import pickle
from typing import List, Dict, Tuple
import numpy as np
from sklearn.cluster import AffinityPropagation
from sklearn.decomposition import TruncatedSVD
from sklearn.feature_extraction.text import CountVectorizer, TfidfTransformer

from cleartext.basic import do_clear
from data.viewModels.repository import RepositoryViewModel


class Model:
    vectorizer: CountVectorizer = None
    tf_idf: TfidfTransformer = None
    clustering: AffinityPropagation = None

    def __save_corpus_hash(self, corpus: List[RepositoryViewModel]):
        path = pathlib.Path(__file__).parent.absolute()
        path = str(path) + '/hash/hash_repositories.json'
        with open(path, 'w') as outfile:
            json.dump([ob.__dict__ for ob in corpus], outfile)

    def __save_vectorizer(self):
        path = pathlib.Path(__file__).parent.absolute()
        pickle.dump(self.vectorizer.vocabulary_, open(str(path) + '/hash/vocabulary.pkl', "wb"))
        pickle.dump(self.vectorizer, open(str(path) + '/hash/vectorizer.pkl', "wb"))
        pickle.dump(self.tf_idf, open(str(path) + '/hash/tf_idf.pkl', "wb"))

    def __save_clustering(self):
        path = pathlib.Path(__file__).parent.absolute()
        path = str(path) + '/hash/clustering.pkl'
        pickle.dump(self.clustering, open(path, "wb"))

    def download_models(self):
        path = pathlib.Path(__file__).parent.absolute()
        self.vectorizer = pickle.load(open(str(path) + '/hash/vectorizer.pkl', 'rb'))
        self.tf_idf = pickle.load(open(str(path) + '/hash/tf_idf.pkl', 'rb'))
        self.clustering = pickle.load(open(str(path) + '/hash/clustering.pkl', 'rb'))

    def download_corpus_from_hash(self) -> List[Tuple[str, str]]:
        path = pathlib.Path(__file__).parent.absolute()
        path = str(path) + '/hash/hash_repositories.json'
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
        self.__save_vectorizer()
        self.__save_clustering()
        return

    def predict(self, texts: List[str]) -> int:
        text_inline = ' '.join(texts).lower()
        text_inline = text_inline.replace('\n', ' ')
        text_inline = do_clear(text_inline)
        bag_of_words = self.vectorizer.transform([text_inline])
        print(bag_of_words.toarray())
        vector = self.tf_idf.transform(bag_of_words)
        cluster = self.clustering.predict(vector.toarray())[0]
        return cluster
