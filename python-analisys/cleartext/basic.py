import re

from cleartext import lemmatizer


def do_clear(text_inline):
    text_inline = re.sub(r"[\]\d%:$\"\';\[&*=<>\}{)(?!/.,_\-^@]", r" ", text_inline)
    text_inline = re.sub(r"[^\x00-\x7F]+", r" ", text_inline)
    text_inline = lemmatizer.lemmatize(text_inline)
    return text_inline
