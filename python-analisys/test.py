from sklearn.cluster import AgglomerativeClustering
from sklearn.cluster import MeanShift
from sklearn.cluster import KMeans
from sklearn.cluster import AffinityPropagation
from sklearn.mixture import GaussianMixture
from sklearn.manifold import TSNE
from matplotlib import pyplot as plt
from sklearn.metrics.pairwise import cosine_similarity
from sklearn.feature_extraction.text import CountVectorizer
from sklearn.feature_extraction.text import TfidfTransformer
from sklearn.decomposition import TruncatedSVD
import pandas as pd
import numpy as np
import pprint
from sklearn.feature_extraction import text
from os import listdir
from os.path import isfile, join
import re
import nltk
from nltk.stem import WordNetLemmatizer

nltk.download('wordnet')
lemmatizer = WordNetLemmatizer()
files_dict = dict()


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


def distances(x, index_to_file, c):
    for index, vector in enumerate(x):
        main_repository = index_to_file[index]
        result = dict()
        print('MAIN (simple): ', main_repository)

        for i, v in enumerate(x):
            indeces = []
            vec_main = x[index]
            vec_second = x[i]
            for wi, word in enumerate(vec_main):
                if vec_main[wi] > 0 and vec_second[wi] > 0:
                    indeces.append(count_vectorizer.get_feature_names()[wi])
            second_repository = index_to_file[i]
            distance = cosine_similarity([vec_main, vec_second])[0][1] * 100
            if distance > 40:
                result[second_repository] = distance
                # print('\t\t', indeces)
            # result[second_repository] = distance
        sorted_dict = dict()
        sorted_keys = sorted(result, key=result.get)
        for w in sorted_keys:
            sorted_dict[w] = result[w]
        i = 0
        for w in sorted_dict:
            if len(sorted_dict) - i <= c:
                print("\t\t-> ", w, " = ", sorted_dict[w])
            i += 1


stop_words = text.ENGLISH_STOP_WORDS.union(
    ['javascript', 'python', 'java', 'kotlin', 'android', 'hacktoberfest', 'js', 'nodejs',
     'windows', 'macos', 'react', 'angular', 'vue', 'linux'])

read_files(files_dict, "-topics.txt")
corpus, repository_index = create_corpus(files_dict)
count_vectorizer = CountVectorizer(stop_words=stop_words, min_df=2, max_features=len(corpus))

bag_of_words_count = count_vectorizer.fit_transform(corpus)
tfidf_transformer = TfidfTransformer(use_idf=True).fit(bag_of_words_count)
bag_of_word_tfidf = tfidf_transformer.transform(bag_of_words_count)
pprint.pprint(dict(zip(count_vectorizer.get_feature_names(), tfidf_transformer.idf_)))

c = len(corpus)
distances(bag_of_words_count.toarray(), repository_index, c)


def indexing_repositories():
    for wi, word in enumerate(count_vectorizer.get_feature_names()):
        print('WORD: ', word)
        indeces = dict()
        vectors = []
        index = 0
        for i, vector in enumerate(bag_of_word_tfidf.toarray()):
            if vector[wi] > 0:
                vectors.append(vector)
                indeces[index] = repository_index[i]
                index += 1
        for i, main in enumerate(vectors):
            print('\t\tMAIN: ', indeces[i], ', WORD: ', word)
            result = dict()
            sort = dict()
            for j, second in enumerate(vectors):
                distance = cosine_similarity([main, second])[0][1] * 100
                if distance > 40:
                    result[distance] = indeces[j] + ' = ' + str(distance)
            for key in sorted(result.keys()):
                sort[key] = result[key]
            for k, v in sort.items():
                print('\t\t\t\t -> ', v)


# indexing_repositories()

#
# topics = []
# for i in range(len(count_vectorizer.get_feature_names()) - 1):
#     topics.append('topic_' + str(i))
# svd = TruncatedSVD(n_components=len(topics))
# lsa = svd.fit_transform(bag_of_word_tfidf)
# df = pd.DataFrame(lsa, columns=topics)
# df['repository'] = corpus
# # display(df)
# #
# dictionary = count_vectorizer.get_feature_names()
# encoding = pd.DataFrame(svd.components_, index=topics, columns=dictionary).T
# pd.options.display.max_rows = len(dictionary)
# # display(encoding)
