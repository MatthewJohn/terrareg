
integration_git_providers = {
    1: {
        'name': 'testgitprovider',
        'base_url_template': 'https://localhost.com/{namespace}/{module}-{provider}',
        'browse_url_template': 'https://localhost.com/{namespace}/{module}-{provider}/browse/{tag}/{path}',
        'clone_url_template': 'ssh://localhost.com/{namespace}/{module}-{provider}'
    },
    2: {
        'name': 'repo_url_tests',
        'base_url_template': 'https://base-url.com/{namespace}/{module}-{provider}',
        'browse_url_template': 'https://browse-url.com/{namespace}/{module}-{provider}/browse/{tag}/{path}suffix',
        'clone_url_template': 'ssh://clone-url.com/{namespace}/{module}-{provider}'
    }
}

integration_test_data = {
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
            'versions': {'1.0.0': {}}
        }},
        'mock-module': {'testprovider': {
            'id': 3,
            'verified': True,
            'repo_base_url_template': 'http://github.com/{namespace}/{module}',
            'latest_version': '1.2.3',
            'versions': {'1.2.3': {}}
        }},
        'unverifiedmodule': {'testprovider': {
            'id': 16,
            'verified': False,
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
            'repo_clone_url_template': 'https://github.com/test/test.git',
            'versions': {'2.1.0': {}}
        }},
        'modulenotpublished': {'testprovider': {
            'id': 15,
            'latest_version': '10.2.1',
            'versions': {
                '10.2.1': {'published': False}
            }
        }},
        'wrongversionorder': {'testprovider': {
            'id': 17,
            'versions': {
                '1.5.4': {'published': True}, '2.1.0': {'published': True}, '0.1.1': {'published': True},
                '10.23.0': {'published': True}, '0.1.10': {'published': True}, '0.0.9': {'published': True},
                '0.1.09': {'published': True}, '0.1.8': {'published': True}
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
        },
    },
    'real_providers': {
        'test-module': {
            'aws': {
                'id': 22,
                'versions': {
                    '1.0.0': {}
                }
            },
            'gcp': {
                'id': 23,
                'versions': {
                    '1.0.0': {}
                }
            },
            'doesnotexist': {
                'id': 24,
                'versions': {
                    '1.0.0': {}
                }
            }
        }
    },
    'repo_url_tests': {
        'no-git-provider': {'test': {
            'id': 18,
            'versions': {
                '1.0.0': {},
                '1.4.0': {
                    'repo_base_url_template': 'https://mv-base-url.com/{namespace}/{module}-{provider}',
                    'repo_browse_url_template': 'https://mv-browse-url.com/{namespace}/{module}-{provider}/browse/{tag}/{path}suffix',
                    'repo_clone_url_template': 'ssh://mv-clone-url.com/{namespace}/{module}-{provider}'
                }
            }
        }},
        'git-provider-urls': {'test': {
            'id': 19,
            'versions': {
                '1.1.0': {},
                '1.4.0': {
                    'repo_base_url_template': 'https://mv-base-url.com/{namespace}/{module}-{provider}',
                    'repo_browse_url_template': 'https://mv-browse-url.com/{namespace}/{module}-{provider}/browse/{tag}/{path}suffix',
                    'repo_clone_url_template': 'ssh://mv-clone-url.com/{namespace}/{module}-{provider}'
                }
            },
            'git_provider_id': 2
        }},
        'module-provider-urls': { 'test': {
            'id': 20,
            'versions': {
                '1.2.0': {},
                '1.4.0': {
                    'repo_base_url_template': 'https://mv-base-url.com/{namespace}/{module}-{provider}',
                    'repo_browse_url_template': 'https://mv-browse-url.com/{namespace}/{module}-{provider}/browse/{tag}/{path}suffix',
                    'repo_clone_url_template': 'ssh://mv-clone-url.com/{namespace}/{module}-{provider}'
                }
            },
            'repo_base_url_template': 'https://mp-base-url.com/{namespace}/{module}-{provider}',
            'repo_browse_url_template': 'https://mp-browse-url.com/{namespace}/{module}-{provider}/browse/{tag}/{path}suffix',
            'repo_clone_url_template': 'ssh://mp-clone-url.com/{namespace}/{module}-{provider}'
        }},
        'module-provider-override-git-provider': { 'test': {
            'id': 21,
            'versions': {
                '1.3.0': {},
                '1.4.0': {
                    'repo_base_url_template': 'https://mv-base-url.com/{namespace}/{module}-{provider}',
                    'repo_browse_url_template': 'https://mv-browse-url.com/{namespace}/{module}-{provider}/browse/{tag}/{path}suffix',
                    'repo_clone_url_template': 'ssh://mv-clone-url.com/{namespace}/{module}-{provider}'
                }
            },
            'repo_base_url_template': 'https://mp-base-url.com/{namespace}/{module}-{provider}',
            'repo_browse_url_template': 'https://mp-browse-url.com/{namespace}/{module}-{provider}/browse/{tag}/{path}suffix',
            'repo_clone_url_template': 'ssh://mp-clone-url.com/{namespace}/{module}-{provider}',
            'git_provider_id': 2
        }}
    }
}