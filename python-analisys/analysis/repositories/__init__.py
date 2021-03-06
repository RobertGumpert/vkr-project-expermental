from sklearn.feature_extraction.text import CountVectorizer
from sklearn.feature_extraction.text import TfidfTransformer

VECTORIZER = CountVectorizer(stop_words='english', min_df=2)
