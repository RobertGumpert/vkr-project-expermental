import csv
from sklearn.cluster import MeanShift
import numpy as np
import pickle

MODELS_DIR = "C:/VKR/vkr-project-expermental/python-analisys/models/"
LEARNING_VECTORS = "C:/VKR/vkr-project-expermental/go-agregator/data/tests/learning-vectors.csv"


def read_vectors(path):
    with open(path, newline='') as f:
        reader = csv.reader(f)
        data = list(reader)
        print(len(data))
        output = [[]]
        for i, row in enumerate(data):
            sum = 0
            row_output = []
            del row[0]
            for j, item in enumerate(row):
                if item == "":
                    continue
                output_item = int(row[j])
                sum += output_item
                row_output.append(output_item)
            if sum != 0:
                output.append(row_output)
            else:
                continue
        del output[0]
        print(len(output))
        return output


def learning_mean_shift_clustering(vectors):
    print("Start...")
    clustering = MeanShift(bandwidth=1).fit(vectors)
    return clustering


def save_mean_shift_model(clustering, filename):
    path = MODELS_DIR + filename
    pickle.dump(clustering, open(path, 'wb'))


def download_mean_shift_model(filename):
    path = MODELS_DIR + filename
    clustering = pickle.load(open(path, 'rb'))
    return clustering
