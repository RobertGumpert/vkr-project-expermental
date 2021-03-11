from flask import Flask
from readfiles.repository import read_combined_blocks_about_and_topics
from analysis.clustering.model import Model

model = Model()

# repositories = read_combined_blocks_about_and_topics()
# corpus = model.create_corpus_from_dict(repositories)
# model.create_clusters(corpus)

model.download_models()
cluster = model.predict(['An Efficient ProxyPool with Getter, Tester and Server', 'flask http proxy proxypool redis webspider'])
print(cluster)

#
# -----------------------------------------------------------------------------------------------------------------------
#

app = Flask(__name__)


@app.route('/')
def hello_world():
    return 'Hello World!'


if __name__ == '__main__':
    app.run()
