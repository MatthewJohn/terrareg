
test_git_providers = {
    1: {
        'name': 'testgitprovider',
        'base_url_template': 'https://localhost.com/{namespace}/{module}-{provider}',
        'browse_url_template': 'https://localhost.com/{namespace}/{module}-{provider}/browse/{tag}/{path}',
        'clone_url_template': 'ssh://localhost.com/{namespace}/{module}-{provider}'
    },
    2: {
        'name': 'second-git-provider',
        'base_url_template': 'https://localhost2.example/{namespace}-{module}-{provider}',
        'browse_url_template': 'https://localhost2.com/{namespace}/{module}-{provider}/browse/{tag}/{path}',
        'clone_url_template': 'ssh://localhost2.com/{namespace}/{module}-{provider}'
    }
}

test_data_full = {
    'testnamespace': {
        'testmodulename': {'testprovider': {
            'repo_base_url_template': 'http://mock.example.com/mockmodule',
            'id': 1,
            'latest_version': '2.4.1',
            'verified': True,
            'versions': {'2.4.1': {'published': True}, '1.0.0': {'published': True}}
        }},
        'lonelymodule': {'testprovider': {
            'id': 2,
            'latest_version': '1.0.0',
            'verified': True,
            'versions': {'1.0.0': {'published': True}}
        }},
        'mock-module': {'testprovider': {
            'id': 3,
            'verified': True,
            'repo_base_url_template': 'http://github.com/{namespace}/{module}',
            'latest_version': '1.2.3',
            'versions': {'1.2.3': {'published': True}}
        }},
        'unverifiedmodule': {'testprovider': {
            'id': 16,
            'verified': False,
            'latest_version': '1.2.3',
            'versions': {'1.2.3': {'published': True}}
        }},
        'internalmodule': {'testprovider': {
            'id': 17,
            'verified': False,
            'latest_version': '5.2.0',
            'versions': {'5.2.0': {'internal': True, 'published': True}}
        }},
        'modulenorepourl': {'testprovider': {
            'id': 5,
            'latest_version': '2.2.4',
            'versions': {'2.2.4': {'published': True}}
        }},
        'onlybeta': {'testprovider': {
            'id': 18,
            'versions': {'2.2.4-beta': {'published': True, 'beta': True}}
        }},
        'modulewithrepourl': {'testprovider': {
            'id': 6,
            'latest_version': '2.1.0',
            'repo_clone_url_template': 'https://github.com/test/test.git',
            'versions': {'2.1.0': {}}
        }},
        'modulenotpublished': {'testprovider': {
            'id': 15,
            'latest_verison': '10.2.1',
            'repo_base_url_template': 'https://custom-localhost.com/{namespace}/{module}-{provider}',
            'repo_browse_url_template': 'https://custom-localhost.com/{namespace}/{module}-{provider}/browse/{tag}/{path}',
            'repo_clone_url_template': 'ssh://custom-localhost.com/{namespace}/{module}-{provider}',
            'versions': {
                '10.2.1': {'published': False}
            }
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
            'repo_clone_url_template': 'ssh://example.com/repo.git'
        }},
        'bitbucketexample': {
            'testprovider': {
                'id': 8,
                'repo_clone_url_template': 'ssh://git@localhost:7999/bla/test-module.git',
                'git_tag_format': 'v{version}',
                'versions': []
            }
        },
        'gitextraction': {
            'staticrepourl': {
                'id': 9,
                'repo_clone_url_template': 'ssh://git@localhost:7999/bla/test-module.git',
                'git_tag_format': 'v{version}',
                'versions': []
            },
            'placeholdercloneurl': {
                'id': 10,
                'repo_clone_url_template': 'ssh://git@localhost:7999/{namespace}/{module}-{provider}.git',
                'git_tag_format': 'v{version}',
                'versions': []
            },
            'usesgitprovider': {
                'id': 11,
                'git_provider_id': 1,
                'git_tag_format': 'v{version}',
                'versions': []
            },
            'nogittagformat': {
                'id': 12,
                'git_provider_id': 1,
                'versions': []
            },
            'complexgittagformat': {
                'id': 13,
                'git_provider_id': 1,
                'git_tag_format': 'unittest{version}value',
                'versions': []
            },
            'norepourl': {
                'id': 14,
                'git_tag_format': 'v{version}',
                'versions': []
            }
        }
    }
}
