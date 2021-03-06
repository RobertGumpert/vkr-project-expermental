from flask import Flask
from readfiles.repository import read_combined_blocks_about_and_topics
from analysis.repositories.model import Model

# repositories = read_combined_blocks_about_and_topics()
model = Model()
corpus = model.download_corpus_from_hash()
model.select_train_texts_from_corpus(corpus)
#
# -----------------------------------------------------------------------------------------------------------------------
#

app = Flask(__name__)


@app.route('/')
def hello_world():
    return 'Hello World!'


if __name__ == '__main__':
    app.run()
