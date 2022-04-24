
test_data_full = {
    'testnamespace': {
        'testmodulename': {'testprovider': {
            'id': 1,
            'latest_version': '2.4.1',
            'versions': {'2.4.1': {}, '1.0.0': {}}
        }},
        'lonelymodule': {'testprovider': {
            'id': 2,
            'latest_version': '1.0.0',
            'versions': {'1.0.0': {}}
        }},
        'mock-module': {'testprovider': {
            'id': 3,
            'latest_version': '1.2.3',
            'versions': {'1.2.3': {}}
        }},
        'modulenorepourl': {'testprovider': {
            'id': 5,
            'latest_version': '2.2.4',
            'versions': {'2.2.4': {}}
        }},
        'modulewithrepourl': {'testprovider': {
            'id': 6,
            'latest_version': '2.1.0',
            'repository_url': 'https://github.com/test/test.git',
            'versions': {'2.1.0': {}}
        }}
    },
    'secondtestnamespace': {
        'mockmodule2': { 'secondprovider': {
            'id': 4,
            'latest_version': '3.0.0',
            'versions': {'3.0.0': {}}
        }}
    },
    'moduleextraction': {
        'test-module': { 'testprovider': {
            'id': 7,
            'repository_url': 'ssh://example.com/repo.git'
        }},
        'bitbucketexample': { 'testprovider': {
            'id': 8,
            'repository_url': 'ssh://git@localhost:7999/bla/test-module.git',
            'git_tag_format': 'v{version}',
            'versions': []
        }}
    }
}
