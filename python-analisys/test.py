from gensim.summarization.summarizer import summarize
from gensim.summarization import keywords
from sklearn.feature_extraction.text import CountVectorizer
from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.metrics.pairwise import cosine_similarity
import numpy as np
from sklearn.cluster import AgglomerativeClustering
from sklearn.cluster import MeanShift
import re
import nltk
from nltk.stem import WordNetLemmatizer

nltk.download('wordnet')
lemmatizer = WordNetLemmatizer()
files_dict = dict()


def clear_text(lines_to_string):
    lines_to_string = lines_to_string.lower()
    lines_to_string = re.sub(r"[\]\d%:$\"\';\[&*=<>\}{)(?!/.,_\-^@]", r" ", lines_to_string)
    lines_to_string = re.sub(r"[^\x00-\x7F]+", r" ", lines_to_string)
    lines_to_string = lemmatizer.lemmatize(lines_to_string)
    return lines_to_string


def read_file(file_name):
    with open('C:/VKR/vkr-project-expermental/go-agregator/data/group-by-elements/titles/' + file_name + '-titles.txt',
              encoding='utf-8') as file:
        lines = file.readlines()
        lines_to_string = ''.join(lines).lower()
        lines_to_string = clear_text(lines_to_string)
        return lines_to_string, lines



files_dict['react'] = read_file('react')
files_dict['vue'] = read_file('vue')
files_dict['angular'] = read_file('angular')
files_dict['gin'] = read_file('gin')
files_dict['okhttp'] = read_file('okhttp')
files_dict['flask'] = read_file('flask')
files_dict['terminal'] = read_file('terminal')
files_dict['hyper'] = read_file('hyper')
files_dict['alacritty'] = read_file('alacritty')


def get_disctance(distances_dict, vectorizer, key_1, key_2, key_vect):
    simple_vect_x = vectorizer.fit_transform([files_dict[key_1][0], files_dict[key_2][0]])
    distance = cosine_similarity(simple_vect_x)[0][1] * 100
    distances_dict[key_vect + '=' + key_1 + ':' + key_2] = distance
    return


def iterate_files_and_calculate_distance(distances_dict, vectorizer, key_vect):
    for k1, v1 in files_dict.items():
        for k2, v2 in files_dict.items():
            get_disctance(distances_dict, vectorizer, k1, k2, key_vect)
    main = ''
    for key, value in distances_dict.items():
        r = key.split(':')[0].split('=')[1]
        if r != main:
            main = r
            print('MAIN: ', r)
        if value > 50.0:
            print("\t\t\t-> (1) ", key, " = ", value)
        if 45.0 < value < 50.0:
            print("\t\t\t---> (2) ", key, " = ", value)
        if value < 45.0:
            print("\t\t\t------> (3) ", key, " = ", value)
    # pprint.pprint(distances_dict)
    return


tfidf_vectorizer = TfidfVectorizer(stop_words='english')
count_vectorizer = CountVectorizer(stop_words='english')

print("Count Vectorized...")
distances_dict = dict()
iterate_files_and_calculate_distance(distances_dict, count_vectorizer, "count")

print("\nTF-IDF...")
distances_dict = dict()
iterate_files_and_calculate_distance(distances_dict, tfidf_vectorizer, "tf-idf")


def get_vectors(vectorizer):
    count_rows = 0
    for k, v in files_dict.items():
        count_rows += len(v[1])
    matrix = []
    for k, v in files_dict.items():
        for i, s in enumerate(v[1]):
            if i > 500:
                break
            matrix.append(clear_text(s))
    simple_vect_x = vectorizer.fit_transform(matrix)
    return simple_vect_x


# x = get_vectors(count_vectorizer)
# print(type(x))
# clustering = AgglomerativeClustering().fit(x.todense())
# print(clustering.labels_)

# clustering = MeanShift().fit(x.toarray())
# print(clustering.labels_)


tfidf_vectorizer = TfidfVectorizer()
count_vectorizer = CountVectorizer(stop_words='english')


def calculate_distance_pair_titles(vectorizer, threshold, bigger_slice, smaller_slice):
    result = "\n"
    for i, i_title in enumerate(bigger_slice):
        for y, y_title in enumerate(smaller_slice):
            print("--------->", i , " : ", y)
            try:
                x = vectorizer.fit_transform([i_title, y_title])
            except ValueError:
                continue
            x_arr = x.toarray()
            if len(x_arr[0]) < 3 or len(x_arr[1]) < 3:
                continue
            count_nulls_main = 0
            count_nulls_second = 0
            for x_i in range(len(x_arr[0])):
                if x_arr[0][x_i] == 0:
                    count_nulls_main += 1
                if x_arr[1][x_i] == 0:
                    count_nulls_second += 1
            percent_crossing_main = (1 - (count_nulls_main / len(x_arr[0]))) * 100
            percent_crossing_second = (1 - (count_nulls_second / len(x_arr[1]))) * 100
            if percent_crossing_main < threshold or percent_crossing_second < threshold:
                continue
            distance = cosine_similarity(x_arr)[0][1] * 100
            result += "\t\t\t\tTitle: [" + i_title + "]\n\t\t\t\tTitle: [" + y_title + "]\n\t\t\t\tDistance: [" + str(
                distance) + "]\n\n"
    return result


def titles_distance(vectorizer, threshold, files_dict):
    result = ""
    for key_main_file, value_main_file in files_dict.items():
        print(key_main_file)
        file_title = "\n\n\nMain: " + key_main_file + "-----------------------------\n"
        distances_result = ""
        for key_second_file, value_second_file in files_dict.items():
            if key_main_file == key_second_file:
                continue
            print("-->", key_second_file)
            file_title += "\t\tSecond: " + key_second_file + "\n"
            slice_main = value_main_file[1]
            slice_second = value_second_file[1]
            bigger_slice = None
            smaller_slice = None
            if len(slice_main) >= len(slice_second):
                bigger_slice = slice_main
                smaller_slice = slice_second
            if len(slice_second) >= len(slice_main):
                bigger_slice = slice_second
                smaller_slice = slice_main
            distances_result += calculate_distance_pair_titles(vectorizer, threshold, bigger_slice, smaller_slice)
        pair_repositories_result = file_title + distances_result
        result += pair_repositories_result
    return result


print('Start caculate...')
result = titles_distance(count_vectorizer, 70, files_dict)
print('Finish caculate...')
