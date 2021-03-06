from os import listdir
from os.path import isfile, join
from cleartext.basic import do_clear
from readfiles import DATA_DIR
from data.viewModels.repository import RepositoryViewModel



def read_combined_blocks_about_and_topics():
    repositories = dict()
    delimiter = '-topics.txt'
    path = DATA_DIR + 'group-by-elements/topics+descriptions/'
    index = 0
    for element in listdir(path):
        repository = element
        if delimiter not in repository:
            continue
        else:
            repository = repository.split(delimiter)[0]
        element_path = join(path, element)
        if isfile(element_path):
            with open(element_path) as file:
                lines = file.readlines()
                text_inline = ''.join(lines).lower()
                clear_text_inline = do_clear(text_inline)
                repositories[repository] = RepositoryViewModel(
                    name=repository,
                    url='',
                    repository_id=index,
                    about=lines[0],
                    topics=lines[1],
                    combined=text_inline,
                    clear_about='',
                    clear_combined=clear_text_inline
                )
            index += 1
    return repositories
