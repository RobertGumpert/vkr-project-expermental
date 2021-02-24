from flask import Flask
import analisys

import time
start_time = time.time()

vectors = analisys.read_vectors(analisys.LEARNING_VECTORS)

print("--- %s seconds ---" % (time.time() - start_time))

#learning_model = analisys.learning_mean_shift_clustering(vectors)

#print(learning_model.cluster_centers_)

#print(learning_model.labels_)

#analisys.save_mean_shift_model(learning_model, 'test_model.sav')

#print('Save!')

print()

#
# -----------------------------------------------------------------------------------------------------------------------
#

app = Flask(__name__)


@app.route('/')
def hello_world():
    return 'Hello World!'


if __name__ == '__main__':
    app.run()
