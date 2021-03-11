import re
from os import listdir
from os.path import isfile, join
from sklearn.cluster import AffinityPropagation
from sklearn.feature_extraction import text
from sklearn.feature_extraction.text import CountVectorizer
from sklearn.feature_extraction.text import TfidfTransformer
from sklearn.decomposition import TruncatedSVD
import pandas as pd
import nltk
from nltk.stem import WordNetLemmatizer
from sklearn.metrics.pairwise import cosine_similarity

nltk.download('wordnet')
lemmatizer = WordNetLemmatizer()
files_dict = dict()
stop_words = text.ENGLISH_STOP_WORDS.union(
    ['javascript', 'python', 'java', 'kotlin', 'android', 'hacktoberfest', 'js', 'nodejs',
     'windows', 'macos', 'react', 'angular', 'vue', 'linux'])


def clear_text(lines_to_string):
    lines_to_string = re.sub(r"[\]\d%:$\"\';\[&*=<>\}{)(?!/.,_\-^@]", r" ", lines_to_string)
    lines_to_string = re.sub(r"[^\x00-\x7F]+", r" ", lines_to_string)
    lines_to_string = lemmatizer.lemmatize(lines_to_string)
    return lines_to_string


def read_files(files_dict, delimitor):
    path = "C:/VKR/vkr-project-expermental/go-agregator/data/group-by-elements/topics+descriptions/"
    for element in listdir(path):
        repository = element
        if delimitor not in repository:
            continue
        else:
            repository = repository.split(delimitor)[0]
        element_path = join(path, element)
        if isfile(element_path):
            with open(element_path) as file:
                lines = file.readlines()
                lines_to_string = ''.join(lines).lower()
                lines_to_string = clear_text(lines_to_string)
                files_dict[repository] = (lines_to_string, lines)


def create_corpus(files):
    index = 0
    repository_index = dict()
    corpus = []
    for repository, data in files.items():
        # text = data[1][1]
        text = data[0]
        text = text.replace('\n', ' ').strip().lower()
        corpus.append(text)
        repository_index[index] = repository
        index += 1
    return corpus, repository_index


def get_clusters(corpus):
    # , min_df=2, max_features=len(corpus)
    vectorizer = CountVectorizer(stop_words=stop_words)
    bag_of_words_count = vectorizer.fit_transform(corpus)
    tfidf_transformer = TfidfTransformer(use_idf=True).fit(bag_of_words_count)
    vectors = tfidf_transformer.transform(bag_of_words_count)
    clusterizator = AffinityPropagation().fit(vectors.toarray())
    count = len(clusterizator.cluster_centers_)
    clusters = []
    print('Count clusters: ', count)
    for cluster in range(count):
        clusters.append('cluster_' + str(cluster))
    return clusters, vectorizer, vectors, clusterizator


def get_lsa(clusters, corpus, vectors, vectorizer):
    svd = TruncatedSVD(n_components=len(clusters))
    lsa = svd.fit_transform(vectors)
    df = pd.DataFrame(lsa, columns=clusters)
    df['repository'] = corpus
    dictionary = vectorizer.get_feature_names()
    encoding = pd.DataFrame(svd.components_, index=clusters, columns=dictionary).T
    pd.options.display.max_rows = len(dictionary)
    # print(encoding.sort_values(clusters[2], ascending=False))
    return lsa, df, encoding


def calc_distances(vectors, repository_index, c):
    for index, vector in enumerate(vectors):
        main_repository = repository_index[index]
        result = dict()
        print('MAIN: (LSA)', main_repository)
        for i, v in enumerate(vectors):
            second_repository = repository_index[i]
            vec_main = vectors[index]
            vec_second = vectors[i]
            distance = cosine_similarity([vec_main, vec_second])[0][1] * 100
            if distance < 40:
                continue
            result[second_repository] = distance
        sorted_dict = dict()
        sorted_keys = sorted(result, key=result.get)
        for w in sorted_keys:
            sorted_dict[w] = result[w]
        i = 0
        for w in sorted_dict:
            if len(sorted_dict) - i <= c:
                print("\t\t-> ", w, " = ", sorted_dict[w])
            i += 1
    return


read_files(files_dict, "-topics.txt")
corpus, repository_index = create_corpus(files_dict)
clusters, vectorizer, vectors, clusterizator = get_clusters(corpus)
lsa, df, encoding = get_lsa(clusters, corpus, vectors, vectorizer)
calc_distances(lsa, repository_index, len(corpus))
