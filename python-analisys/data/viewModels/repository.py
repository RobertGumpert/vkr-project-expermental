class RepositoryViewModel:
    repository_id = -1
    url = ''
    name = ''
    about = ''
    topics = []
    combined = ''
    clear_about = ''
    clear_combined = ''

    def __init__(self, name='', url='', repository_id=-1, about='', topics='', combined='', clear_about='', clear_combined=''):
        self.name = name
        self.url = url
        self.repository_id = repository_id
        self.about = about
        self.topics = topics
        self.combined = combined
        self.clear_about = clear_about
        self.clear_combined = clear_combined

    def deserialize(self, dictionary):
        for key in dictionary:
            setattr(self, key, dictionary[key])
