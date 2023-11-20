
from datetime import datetime
import json

import terrareg.provider_source_type
import terrareg.provider_documentation_type


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
    },
    3: {
        'name': 'repo_url_tests_uri_encoded',
        'base_url_template': 'https://base-url.com/{namespace}/{module}-{provider}',
        'browse_url_template': 'https://browse-url.com/{namespace}/{module}-{provider}/browse/{path}?at={tag_uri_encoded}',
        'clone_url_template': 'ssh://clone-url.com/{namespace}/{module}-{provider}'
    }
}

integration_provider_categories = [
    {
        "id": 523,
        "name": "Visible Monitoring",
        "slug": "visible-monitoring",
        "user-selectable": True
    },
    {
        "id": 54,
        "name": "Second Visible Cloud",
        "slug": "second-visible-cloud",
        "user-selectable": True
    },
    {
        "id": 55,
        "name": "Default Visible Test",
        "slug": "default-visible-test"
    },
    {
        "id": 99,
        "name": "Hidden Database",
        "slug": "hidden-database",
        "user-selectable": False
    },
    {
        "id": 100,
        "name": "No Slug Provided!",
        "user-selectable": True
    },
]

integration_provider_sources = [
    {
        "name": "Test Github Autogenerate",
        "type": terrareg.provider_source_type.ProviderSourceType.GITHUB.value,
        "base_url": "https://github.example.com",
        "api_url": "https://api.github.example.com",
        "client_id": "unittest-client-id",
        "client_secret": "unittest-client-secret",
        "login_button_text": "Login via Github using this unit test",
        "private_key_path": "./path/to/key.pem",
        "app_id": "1234appid",
        "default_access_token": "pa-test-personal-access-token",
        "default_installation_id": "ut-default-installation-id-here",
        "auto_generate_github_organisation_namespaces": True
    },
    {
        "name": "Test Github No Autogenerate",
        "type": terrareg.provider_source_type.ProviderSourceType.GITHUB.value,
        "base_url": "https://github.example.com",
        "api_url": "https://api.github.example.com",
        "client_id": "unittest-client-id",
        "client_secret": "unittest-client-secret",
        "login_button_text": "Login via Github using this unit test",
        "private_key_path": "./path/to/key.pem",
        "app_id": "1234appid",
        "default_access_token": "pa-test-personal-access-token",
        "default_installation_id": "ut-default-installation-id-here",
        "auto_generate_github_organisation_namespaces": False
    }
]

integration_test_data = {
    'testnamespace': {
        'modules': {
            'wrongversionorder': {'testprovider': {
                'id': 17,
                'versions': {
                    '1.5.4': {'published': True}, '2.1.0': {'published': True}, '0.1.1': {'published': True},
                    '10.23.0': {'published': True}, '0.1.10': {'published': True}, '0.0.9': {'published': True},
                    '0.1.09': {'published': True}, '0.1.8': {'published': True},
                    '23.2.3-beta': {'published': True, 'beta': True}, '5.21.2': {}
                }
            }},
            'noversions': {'testprovider': {
                'id': 53
            }},
            'onlyunpublished': {'testprovider': {
                'id': 54,
                'versions': {
                    '0.1.8': {'published': False}
                }
            }},
            'onlybeta': {'testprovider': {
                'id': 55,
                'versions': {
                    '2.5.0-beta': {'published': True, 'beta': True}
                }
            }},
        }
    },
    'moduleextraction': {
        'modules': {
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
    },
    'real_providers': {
        'modules': {
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
                'null': {
                    'id': 56,
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
        }
    },
    'repo_url_tests': {
        'modules': {
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
            'no-git-provider-uri-encoded': {'test': {
                'id': 46,
                'versions': {
                    '1.4.0': {
                        'repo_base_url_template': 'https://mv-base-url.com/{namespace}/{module}-{provider}',
                        'repo_browse_url_template': 'https://mv-browse-url.com/{namespace}/{module}-{provider}/browse/{path}?at={tag_uri_encoded}',
                        'repo_clone_url_template': 'ssh://mv-clone-url.com/{namespace}/{module}-{provider}'
                    }
                },
                'git_tag_format': 'release@test/{version}/'
            }},
            'git-provider-uri-encoded': {'test': {
                'id': 47,
                'versions': {
                    '1.4.0': {}
                },
                'git_provider_id': 3,
                'git_tag_format': 'release@test/{version}/'
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
    },
    'modulesearch': {
        'modules': {
            'contributedmodule-oneversion': {'aws': {
                'id': 25,
                'versions': {'1.0.0': {
                    'published': True,
                    'description': 'DESCRIPTION-Search-PUBLISHED'
                }}
            }},
            'contributedmodule-multiversion': {'aws': {
                'id': 26,
                'versions': {
                    '1.2.3': {'published': True, 'description': 'DESCRIPTION-Search-OLDVERSION'},
                    '2.0.0': {'published': True}
                }
            }},
            'contributedmodule-withbetaversion': {'aws': {
                'id': 49,
                'versions': {
                    '1.2.3': {'published': True},
                    '2.0.0-beta': {'published': True, 'beta': True, 'description': 'DESCRIPTION-Search-BETAVERSION'}
                }
            }},
            'contributedmodule-onlybeta': {'aws': {
                'id': 50,
                'versions': {
                    '2.5.0-beta': {'published': True, 'beta': True}
                }
            }},
            'contributedmodule-differentprovider': {'gcp': {
                'id': 27,
                'versions': {
                    '1.2.3': {'published': True}
                }
            }},
            'contributedmodule-unpublished': {'aws': {
                'id': 28,
                'versions': {
                    '1.0.0': {'description': 'DESCRIPTION-Search-UNPUBLISHED'}
                }
            }},
            'verifiedmodule-oneversion': {'aws': {
                'verified': True,
                'id': 29,
                'versions': {'1.0.0': {'published': True}}
            }},
            'verifiedmodule-multiversion': {'aws': {
                'verified': True,
                'id': 30,
                'versions': {
                    '1.2.3': {'published': True},
                    '2.0.0': {'published': True}
                }
            }},
            'verifiedmodule-withbetaversion': {'aws': {
                'id': 51,
                'versions': {
                    '1.2.3': {'published': True},
                    '2.0.0-beta': {'published': True, 'beta': True}
                }
            }},
            'verifiedmodule-onybeta': {'aws': {
                'id': 52,
                'versions': {
                    '2.0.0-beta': {'published': True, 'beta': True}
                }
            }},
            'verifiedmodule-differentprovider': {'gcp': {
                'verified': True,
                'id': 31,
                'versions': {
                    '1.2.3': {'published': True}
                }
            }},
            'verifiedmodule-unpublished': {'aws': {
                'verified': True,
                'id': 32,
                'versions': {
                    '1.0.0': {}
                }
            }}
        }
    },
    'modulesearch-contributed': {
        'modules': {
            'mixedsearch-result': {'aws': {
                'id': 33,
                'versions': {'1.0.0': {'published': True}}
            }},
            'mixedsearch-result-multiversion': {'aws': {
                'id': 34,
                'versions': {
                    '1.2.3': {'published': True},
                    '2.0.0': {'published': True}
                }
            }},
            'mixedsearch-result-unpublished': {'aws': {
                'id': 35,
                'versions': {
                    '1.2.3': {},
                    '2.0.0': {}
                }
            }},
        }
    },
    'modulesearch-trusted': {
        'modules': {
            'mixedsearch-trusted-result': {'aws': {
                'id': 36,
                'versions': {'1.0.0': {'published': True}}
            }},
            'mixedsearch-trusted-second-result': {'aws': {
                'id': 37,
                'versions': {
                    '5.2.1': {'published': True},
                }
            }},
            'mixedsearch-trusted-result-multiversion': {'aws': {
                'id': 38,
                'versions': {
                    '1.2.3': {'published': True},
                    '2.0.0': {'published': True}
                }
            }},
            'mixedsearch-trusted-result-unpublished': {'aws': {
                'id': 39,
                'versions': {
                    '1.2.3': {},
                    '2.0.0': {}
                }
            }},
        }
    },
    'searchbynamespace': {
        'modules': {
            'searchbymodulename1': {
                'searchbyprovideraws': {
                    'id': 40,
                    'versions': {
                        '1.2.3': {'published': True}
                    },
                    'verified': True
                },
                'searchbyprovidergcp': {
                    'id': 41,
                    'versions': {
                        '2.0.0': {'published': True}
                    }
                }
            },
            'searchbymodulename2': {
                'notpublished': {
                    'id': 42,
                    'versions': {'1.2.3': {}}
                },
                'published': {
                    'id': 43,
                    'versions': {
                        '3.1.6': {'published': True}
                    }
                }
            }
        }
    },
    'searchbynamesp-similar': {
        'modules': {
            'searchbymodulename3': {'searchbyprovideraws': {
                'id': 44,
                'versions': {'4.4.1': {'published': True}},
                'verified': True
            }},
            'searchbymodulename4': {'aws': {
                'id': 45,
                'versions': {'5.5.5': {'published': True}}
            }}
        }
    },
    'genericmodules': {
        'modules': {
            'modulename': {'providername': {
                'id': 48,
                'versions': {'1.2.3': {'published': True}}
            }}
        }
    },
    'moduledetails': {
        'modules': {
            'withterraformdocs': {'testprovider': {
                'id': 57,
                'versions': {
                    '1.5.0': {
                        'description': 'This is a test module version for tests.',
                        'owner': 'This is the owner of the module',
                        'repo_base_url_template': 'https://link-to.com/source-code-here',
                        'published': True,
                        'beta': False,
                        'internal': False,
                        'published_at': datetime(2022, 1, 5, 22, 53, 12),
                        'readme_content': '# This is an exaple README!',
                        'variable_template': json.dumps([
                            {
                                'name': 'name_of_application',
                                'type': 'text',
                                'quote_value': True,
                                'additional_help': 'Provide the name of the application'
                            }

                        ]),
                        'terraform_docs': json.dumps({
                            'header': '',
                            'footer': '',
                            'inputs': [
                                {
                                    'name': 'name_of_application',
                                    'type': 'string',
                                    'description': 'Enter the application name',
                                    'default': None,
                                    'required': True
                                },
                                {
                                    'name': 'string_with_default_value',
                                    'type': 'string',
                                    'description': 'Override the default string',
                                    'default': 'this is the default',
                                    'required': False
                                },
                                {
                                    'name': 'undocumented_required_variable',
                                    'type': 'string',
                                    'description': 'Override the default string',
                                    'default': None,
                                    'required': True
                                },
                                {
                                    'name': 'example_boolean_input',
                                    'type': 'bool',
                                    'description': 'required boolean variable',
                                    'default': None,
                                    'required': True
                                },
                                {
                                    'name': 'required_list_variable',
                                    'type': 'list(string)',
                                    'description': 'A required list',
                                    'default': None,
                                    'required': True
                                },
                                {
                                    'name': 'example_list_input',
                                    'type': 'list',
                                    'description': 'Override the stringy list',
                                    'default': ['value 1', 'value 2'],
                                    'required': False
                                }
                            ],
                            'modules': [],
                            'outputs': [
                                {
                                    'name': 'generated_name',
                                    'description': 'Name with randomness'
                                },
                                {
                                    'name': 'no_desc_output',
                                    'description': None
                                }
                            ],
                            'providers': [
                                {
                                    'name': 'random',
                                    'alias': 'random-alias',
                                    'version': '5.2.1'
                                },
                                {
                                    'name': 'someothercompany/unsafe',
                                    'alias': None,
                                    'version': '2.0.0'
                                }
                            ],
                            'requirements': [],
                            'resources': [
                                {
                                    'type': 'string',
                                    'name': 'random_suffix',
                                    'provider': 'random',
                                    'source': 'hashicorp/random',
                                    'mode': 'managed',
                                    'version': 'latest',
                                    'description': None
                                }
                            ]
                        })
                    }
                }
            }},
            'readme-tests': { 'provider': {
                'id': 58,
                'versions': {'1.0.0': {
                    'published': True,
                    'readme_content': None,
                    'examples': {
                        'examples/testreadmeexample': {
                            'example_files': {
                                'examples/testreadmeexample/main.tf': ''
                            }
                        }
                    }
                }}
            }},
            'git-path': {'provider': {
                'id': 59,
                'versions': {'1.0.0': { }}
            }},
            "graph-test": {"provider": {
                "id": 60,
                "versions": {"1.0.0": {
                    "terraform_graph": """
digraph {
\tcompound = "true"
\tnewrank = "true"
\tsubgraph "root" {
\t\t"[root] aws_s3_bucket.test_bucket (expand)" [label = "aws_s3_bucket.test_bucket", shape = "box"]
\t\t"[root] aws_s3_object.test_obj_root_module (expand)" [label = "aws_s3_object.test_obj_root_module", shape = "box"]
\t\t"[root] module.submodule-call.aws_ec2_instance.test_instance (expand)" [label = "module.submodule-call.aws_ec2_instance.test_instance", shape = "box"]
\t\t"[root] output.name" [label = "output.name", shape = "note"]
\t\t"[root] provider[\\"registry.terraform.io/hashicorp/aws\\"]" [label = "provider[\\"registry.terraform.io/hashicorp/aws\\"]", shape = "diamond"]
\t\t"[root] var.name" [label = "var.name", shape = "note"]
\t\t"[root] aws_s3_bucket.test_bucket (expand)" -> "[root] provider[\\"registry.terraform.io/hashicorp/aws\\"]"
\t\t"[root] aws_s3_object.test_obj_root_module (expand)" -> "[root] aws_s3_bucket.test_bucket (expand)"
\t\t"[root] module.submodule-call (close)" -> "[root] module.submodule-call.aws_ec2_instance.test_instance (expand)"
\t\t"[root] module.submodule-call (close)" -> "[root] module.submodule-call.var.passing_name (expand)"
\t\t"[root] module.submodule-call.aws_ec2_instance.test_instance (expand)" -> "[root] module.submodule-call (expand)"
\t\t"[root] module.submodule-call.aws_ec2_instance.test_instance (expand)" -> "[root] provider[\\"registry.terraform.io/hashicorp/aws\\"]"
\t\t"[root] module.submodule-call.var.passing_name (expand)" -> "[root] module.submodule-call (expand)"
\t\t"[root] module.submodule-call.var.passing_name (expand)" -> "[root] var.name"
\t\t"[root] provider[\\"registry.terraform.io/hashicorp/aws\\"] (close)" -> "[root] aws_s3_object.test_obj_root_module (expand)"
\t\t"[root] provider[\\"registry.terraform.io/hashicorp/aws\\"] (close)" -> "[root] module.submodule-call.aws_ec2_instance.test_instance (expand)"
\t\t"[root] root" -> "[root] module.submodule-call (close)"
\t\t"[root] root" -> "[root] output.name"
\t\t"[root] root" -> "[root] provider[\\"registry.terraform.io/hashicorp/aws\\"] (close)"
\t}
}

""",
                    "submodules": {
                        "modules/test-submodule": {
                            "terraform_graph": """
digraph {
\tcompound = "true"
\tnewrank = "true"
\tsubgraph "root" {
\t\t"[root] aws_ec2_instance.test_instance (expand)" [label = "aws_ec2_instance.test_instance", shape = "box"]
\t\t"[root] provider[\\"registry.terraform.io/hashicorp/aws\\"]" [label = "provider[\\"registry.terraform.io/hashicorp/aws\\"]", shape = "diamond"]
\t\t"[root] var.passing_name" [label = "var.passing_name", shape = "note"]
\t\t"[root] aws_ec2_instance.test_instance (expand)" -> "[root] provider[\\"registry.terraform.io/hashicorp/aws\\"]"
\t\t"[root] provider[\\"registry.terraform.io/hashicorp/aws\\"] (close)" -> "[root] aws_ec2_instance.test_instance (expand)"
\t\t"[root] root" -> "[root] provider[\\"registry.terraform.io/hashicorp/aws\\"] (close)"
\t\t"[root] root" -> "[root] var.passing_name"
\t}
}

"""
                        }
                    },
                    'examples': {
                        'examples/testreadmeexample': {
                            "terraform_graph": """
digraph {
\tcompound = "true"
\tnewrank = "true"
\tsubgraph "root" {
\t\t"[root] module.main_call.aws_s3_bucket.test_bucket (expand)" [label = "module.main_call.aws_s3_bucket.test_bucket", shape = "box"]
\t\t"[root] module.main_call.aws_s3_object.test_obj_root_module (expand)" [label = "module.main_call.aws_s3_object.test_obj_root_module", shape = "box"]
\t\t"[root] module.main_call.module.submodule-call.aws_ec2_instance.test_instance (expand)" [label = "module.main_call.module.submodule-call.aws_ec2_instance.test_instance", shape = "box"]
\t\t"[root] provider[\\"registry.terraform.io/hashicorp/aws\\"]" [label = "provider[\\"registry.terraform.io/hashicorp/aws\\"]", shape = "diamond"]
\t\t"[root] module.main_call (close)" -> "[root] module.main_call.aws_s3_object.test_obj_root_module (expand)"
\t\t"[root] module.main_call (close)" -> "[root] module.main_call.module.submodule-call (close)"
\t\t"[root] module.main_call (close)" -> "[root] module.main_call.output.name (expand)"
\t\t"[root] module.main_call.aws_s3_bucket.test_bucket (expand)" -> "[root] module.main_call (expand)"
\t\t"[root] module.main_call.aws_s3_bucket.test_bucket (expand)" -> "[root] provider[\\"registry.terraform.io/hashicorp/aws\\"]"
\t\t"[root] module.main_call.aws_s3_object.test_obj_root_module (expand)" -> "[root] module.main_call.aws_s3_bucket.test_bucket (expand)"
\t\t"[root] module.main_call.module.submodule-call (close)" -> "[root] module.main_call.module.submodule-call.aws_ec2_instance.test_instance (expand)"
\t\t"[root] module.main_call.module.submodule-call (close)" -> "[root] module.main_call.module.submodule-call.var.passing_name (expand)"
\t\t"[root] module.main_call.module.submodule-call (expand)" -> "[root] module.main_call (expand)"
\t\t"[root] module.main_call.module.submodule-call.aws_ec2_instance.test_instance (expand)" -> "[root] module.main_call.module.submodule-call (expand)"
\t\t"[root] module.main_call.module.submodule-call.aws_ec2_instance.test_instance (expand)" -> "[root] provider[\\"registry.terraform.io/hashicorp/aws\\"]"
\t\t"[root] module.main_call.module.submodule-call.var.passing_name (expand)" -> "[root] module.main_call.module.submodule-call (expand)"
\t\t"[root] module.main_call.module.submodule-call.var.passing_name (expand)" -> "[root] module.main_call.var.name (expand)"
\t\t"[root] module.main_call.output.name (expand)" -> "[root] module.main_call (expand)"
\t\t"[root] module.main_call.var.name (expand)" -> "[root] module.main_call (expand)"
\t\t"[root] provider[\\"registry.terraform.io/hashicorp/aws\\"] (close)" -> "[root] module.main_call.aws_s3_object.test_obj_root_module (expand)"
\t\t"[root] provider[\\"registry.terraform.io/hashicorp/aws\\"] (close)" -> "[root] module.main_call.module.submodule-call.aws_ec2_instance.test_instance (expand)"
\t\t"[root] root" -> "[root] module.main_call (close)"
\t\t"[root] root" -> "[root] provider[\\"registry.terraform.io/hashicorp/aws\\"] (close)"
\t}
}


"""
                        }
                    }
                }}
            }}
        }
    },

    "initial-providers": {
        "gpg_keys": [
            {
                "ascii_armor": """
-----BEGIN PGP PUBLIC KEY BLOCK-----

mI0EZUHt7QEEAKgSXXCkqShvE54omLsE0Gzu/Es2Nelwnps8ETlcHPKag0VlZch/
0HPyF3hGsdZM7GB1il7fGCGw6Urkmci7XkRj2M09QtAvE2YPOqfNfMvHQrLIAkBV
lP/4xIBnGMmsUYVMAeo0DiDdFf3Q3pIbWDhd7+OCPKh80F/pYM1Rm4qnABEBAAG0
UVRlc3QgVGVycmFyZWcgVGVzdHMgKFRlc3QgS2V5IGZvciB0ZXJyYXJlZyBUZXN0
cykgPHRlcnJhcmVnLXRlc3RzQGNvbGFtYWlsLmNvLnVrPojOBBMBCgA4FiEEIadO
Tj/f5DhTK9WENN43SsNkDNsFAmVB7e0CGwMFCwkIBwIGFQoJCAsCBBYCAwECHgEC
F4AACgkQNN43SsNkDNtkywP/SR8U/c3gzAY4w0KF3ZG5sBJqrBfdA2d2R//Bsjvz
jRCpGdaXVBJG2FFyfl5QLLhC56rS6nsX6vcXkrRGQtYG6Bhroo6eWjVnyT1RMM+A
wD5uwCijPlSdl82q91aFQk3jwqNoe4/gr9ERHagx3MAgMTEhIzPaKpGHtL7TPM+B
nOi4jQRlQe3tAQQAxCeKNhBAv13aXeSvPI1JKW9pcg5g9Hfd4s/qj82/0hE/Kfgt
4u7RGOEe7q1WgKirtoiv/XSpwKMSlXtt9AH8lbgkveiJ3V+DqJxdzCm42Zlyvg9Z
9sqLz6XOAyMkv44U1x182KMipuuethRmSemN8jthc4Bh5iEM/l7460IyRk8AEQEA
AYi2BBgBCgAgFiEEIadOTj/f5DhTK9WENN43SsNkDNsFAmVB7e0CGwwACgkQNN43
SsNkDNtn+AP+Pm3+u+if0BExYTMKJ0/dU4ICWBkyuuMDkQlz8oOn9/w9EYvkqR/r
QypRou1K0KbLxBCz0vqAM7KLXe0rKwZZ3eWSThiwTJkFlkJsUgwMqqROteYmWm3S
MK0hMLszB/mfN0Q2DW4U0tWslehdEA+aaccwA5PVFKdkA12ImK500TY=
=EL4W
-----END PGP PUBLIC KEY BLOCK-----
""".strip()
            },
            {
                "ascii_armor": """
-----BEGIN PGP PUBLIC KEY BLOCK-----

mI0EZVD0zwEEAJtjkOHz5pFnNw80L4qtKU98+/IVvyEEvQyOreGHdB+E5E6rtVFk
buaF7FrzJzaRj+I4hL6QB8ApkdwRdc+gaZL9KsrY6RI5WyYr8jJ/pANoxFkkIwd0
5Q2U6rkxI2SlWuHYuEmtjhJ8rFGPDRnpkTuQMxgUkUxFoHWMFIiprqmnABEBAAG0
LFRlc3QgR1BHIEtleSAyIChUZXN0KSA8dGVzdGdwZzJAZXhhbXBsZS5jb20+iM4E
EwEKADgWIQSUynK3ovRgamwYIRrpSk8q1ijZJgUCZVD0zwIbAwULCQgHAgYVCgkI
CwIEFgIDAQIeAQIXgAAKCRDpSk8q1ijZJkluBACKoMBoW4QO0d6H/h+8Ucx6/eHj
h5c9R/e7IxSJwB6lKxJGc/YkmHniP742O9opwovbxso7CrzHvdoiEoqdUJApwkk6
k2F6FxWgcZGUpFQVPTFc6iueumXsFu24gHHHiCE+106zN8YW72/lORFulVwLfo2d
Gux4McQ/g3qsP2X217iNBGVQ9M8BBADDrdRUG4mZ2cGLfhEAKDQo8f5ezuAIM2Ja
61m9jjAdRkMYwhrq5+tiVmSrVoqueaxE8cbj5C5XoOomfFOMsD4GVkzHE3t/LPdw
A0iu1usXu0rjImNnlMCVaMpIQGFJrf/EtgUPqVMGSQdNHb8ezeztodPP4gqKDB+f
2O2W0j1cxwARAQABiLYEGAEKACAWIQSUynK3ovRgamwYIRrpSk8q1ijZJgUCZVD0
zwIbDAAKCRDpSk8q1ijZJvtVA/978o0EI/lPuSoUO7EuhFyHpX2xVRL9lNwGXsyk
JDGiJXTZ7vi8pwC5GNknF2eOH1ZPXeeJxJZFh3GAa/Zk0C/BuuugnFwd6j3vxKAq
g122uE3VRxyt2bk9hDQg4rI0Y6nqexCt939GLG+Y3sG4/aEGvk3fd9a/fqn2CayP
Olm9bg==
=/0jI
-----END PGP PUBLIC KEY BLOCK-----
""".strip()
            },
            {
                # provider_extrator tests key
                "ascii_armor": """
-----BEGIN PGP PUBLIC KEY BLOCK-----

mI0EZVsJLQEEALa46EUcwh9n1hxpoGx70Hn1Vpy5XNDog0yp1rjyNu27Myjxb3H1
xSI3KtbJPKbXD2DfzfyH+6ULjfjcU3wKJebf6hgwgmdLIth9QycaFQd2oLWc9vnB
ac1cwA1YM7YwHQofMjt+TekOH8r/v+nt2SEb1nQFeWtmmv39Z4lMn/FbABEBAAG0
OlRlcnJhcmVnIEludGVncmF0aW9uIFRlc3RzIDxpbnRlZ3JhdGlvbi10ZXN0c0Bl
eGFtcGxlLmNvbT6IzgQTAQoAOBYhBKD8Qxmrr5wooWgh308wcuWNFv9tBQJlWwkt
AhsDBQsJCAcCBhUKCQgLAgQWAgMBAh4BAheAAAoJEE8wcuWNFv9tklAEAIHMrvj5
9zCg+hRHwhoAwcJwP4GGQFlqlYisXY+2VRywWM9a9oP/MVbp9dQjn+5oijcfY/i8
SwYV9haADWxg64o2MghWB5f+ts/NT0DHjfnaBwqRNhtePoFH0r8Gs6Tghu+UyAM+
Oz4/KdpCJ8mpm0uBnbF1A9gFlPu8SciBPNBTuI0EZVsJLQEEAMjFXCPRXc8u3NSi
iemeJd5PQz1pKO2x1O2Z9eixcBsc2JJZqs8WXRurB8dUcWd9f5nPxyfdOGHPcsGq
HOaURUDmSS/7DbGhdMdSfVlE6pprSbPKrKqpUHcefcpr9WBhHZJvwwYDeS1Uya7E
Pux7TFg9y3qJ3em0WFmHg32uBwATABEBAAGItgQYAQoAIBYhBKD8Qxmrr5wooWgh
308wcuWNFv9tBQJlWwktAhsMAAoJEE8wcuWNFv9tQqwD/RJhnrkzph2/xiM4YpWH
FzeZ2DYeHJciTlyRCIy3Tz0LZ4usTz+3aXZr1SctcVeqCcCE7M2cQaKjejfrI9CT
OlfYU9tX4x1SQV5cCxqJ3Gjrks0nzPhgvsGJHC45FPS1w7kfOqSVXRBdOpqquvj9
dr8xU1lT4+RN4VY9WoNIERpT
=v7Lw
-----END PGP PUBLIC KEY BLOCK-----
""".strip(),
                "private_key": """
-----BEGIN PGP PRIVATE KEY BLOCK-----

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
WoNIERpT
=kScl
-----END PGP PRIVATE KEY BLOCK-----
"""
            }
        ],
        "providers": {
            "test-initial": {
                "repository": {
                    "provider_source": "Test Github Autogenerate",
                    "provider_id": "initial-providers/terraform-provider-test-initial",
                    "name": "terraform-provider-test-initial",
                    "description": "Test Initial Provider",
                    "owner": "initial-providers",
                    "clone_url": "https://git.example.com/initalproviders/terraform-provider-test-initial.git",
                    "logo_url": "https://git.example.com/initalproviders/terraform-provider-test-initial.png"
                },
                "category_slug": "visible-monitoring",
                "use_default_provider_source_auth": True,
                "tier": "community",
                "versions": {
                    "1.5.0": {
                        "git_tag": "v1.5.0",
                        "gpg_key_fingerprint": "21A74E4E3FDFE438532BD58434DE374AC3640CDB"
                    }
                },
            },
            "to-delete": {
                "repository": {
                    "provider_source": "Test Github Autogenerate",
                    "provider_id": "initial-providers/terraform-provider-to-delete",
                    "name": "terraform-provider-to-delete",
                    "description": "Test Multiple Versions",
                    "owner": "initial-providers",
                    "clone_url": "https://git.example.com/initalproviders/terraform-provider-to-delete.git",
                    "logo_url": "https://git.example.com/initalproviders/terraform-provider-test-initial.png"
                },
                "category_slug": "visible-monitoring",
                "use_default_provider_source_auth": True,
                "tier": "community",
                "versions": {
                    "1.0.0": {
                        "git_tag": "v1.0.0",
                        "gpg_key_fingerprint": "21A74E4E3FDFE438532BD58434DE374AC3640CDB"
                    }
                }
            },
            "update-attributes": {
                "repository": {
                    "provider_source": "Test Github Autogenerate",
                    "provider_id": "initial-providers/terraform-provider-update-attributes",
                    "name": "terraform-provider-update-attributes",
                    "description": "Empty Provider Publish",
                    "owner": "initial-providers",
                    "clone_url": "https://git.example.com/initalproviders/terraform-provider-update-attributes.git",
                    "logo_url": "https://git.example.com/initalproviders/terraform-provider-test-initial.png"
                },
                "category_slug": "visible-monitoring",
                "use_default_provider_source_auth": True,
                "tier": "community",
                "versions": {
                    "1.0.0": {
                        "git_tag": "v1.0.0",
                        "gpg_key_fingerprint": "21A74E4E3FDFE438532BD58434DE374AC3640CDB"
                    }
                }
            },
            "empty-provider-publish": {
                "repository": {
                    "provider_source": "Test Github Autogenerate",
                    "provider_id": "initial-providers/terraform-provider-empty-provider-publish",
                    "name": "terraform-provider-empty-provider-publish",
                    "description": "Empty Provider Publish",
                    "owner": "initial-providers",
                    "clone_url": "https://git.example.com/initalproviders/terraform-provider-empty-provider-publish.git",
                    "logo_url": "https://git.example.com/initalproviders/terraform-provider-test-initial.png"
                },
                "category_slug": "visible-monitoring",
                "use_default_provider_source_auth": True,
                "tier": "community",
                "versions": {
                }
            },
            "multiple-versions": {
                "repository": {
                    "provider_source": "Test Github Autogenerate",
                    "provider_id": "initial-providers/terraform-provider-multiple-versions",
                    "name": "terraform-provider-multiple-versions",
                    "description": "Test Multiple Versions",
                    "owner": "initial-providers",
                    "clone_url": "https://git.example.com/initalproviders/terraform-provider-multiple-versions.git",
                    "logo_url": "https://git.example.com/initalproviders/terraform-provider-test-initial.png"
                },
                "category_slug": "visible-monitoring",
                "use_default_provider_source_auth": True,
                "tier": "community",
                "versions": {
                    "1.0.0": {
                        "git_tag": "v1.0.0",
                        "gpg_key_fingerprint": "21A74E4E3FDFE438532BD58434DE374AC3640CDB"
                    },
                    "1.1.0": {
                        "git_tag": "v1.1.0",
                        "gpg_key_fingerprint": "21A74E4E3FDFE438532BD58434DE374AC3640CDB",
                        "binaries": {
                            "terraform-provider-multiple-versions_1.1.0_linux_amd64.zip": {
                                "content": b"Some old linux content",
                                "checksum": "a268d9b6def5fc8f85e158b5dd8436fe2f9eba023190f9dfab9df6e6208360b3"
                            },
                        },
                        "documentation": {
                            ("overview", "hcl"): {
                                "type": terrareg.provider_documentation_type.ProviderDocumentationType.OVERVIEW,
                                "title": "Overview",
                                "description": "Overview of provider old version",
                                "filename": "index.md",
                                "subcategory": None,
                                "content": "This is an old overview of the module!",
                            },
                            ("some_old_resource", "hcl"): {
                                "type": terrareg.provider_documentation_type.ProviderDocumentationType.RESOURCE,
                                "title": "multiple_versions_thing",
                                "description": "Inital thing for multiple versions provider",
                                "filename": "data-sources/thing.md",
                                "subcategory": "some-subcategory",
                                "content": "Documentation for generating an old thing!",
                            },
                            ("some_resource", "hcl"): {
                                "type": terrareg.provider_documentation_type.ProviderDocumentationType.RESOURCE,
                                "title": "multiple_versions_thing",
                                "description": "Inital thing for multiple versions provider",
                                "filename": "data-sources/thing.md",
                                "subcategory": "some-subcategory",
                                "content": "Documentation for generating an old version of thing!",
                            },
                        }
                    },
                    "1.1.0-beta": {
                        "git_tag": "v1.1.0-beta",
                        "gpg_key_fingerprint": "21A74E4E3FDFE438532BD58434DE374AC3640CDB",
                    },
                    "1.5.0": {
                        "git_tag": "v1.5.0",
                        "gpg_key_fingerprint": "21A74E4E3FDFE438532BD58434DE374AC3640CDB",
                        "binaries": {
                            "terraform-provider-multiple-versions_1.5.0_linux_amd64.zip": {
                                "content": b"Some test linux content",
                                "checksum": "a26d0401981bf2749c129ab23b3037e82bd200582ff7489e0da2a967b50daa98"
                            },
                            "terraform-provider-multiple-versions_1.5.0_linux_arm64.zip": {
                                "content": b"Test linux ARM content",
                                "checksum": "bda5d57cf68ab142f5d0c9a5a0739577e24444d4e8fe4a096ab9f4935bec9e9a"
                            },
                            "terraform-provider-multiple-versions_1.5.0_windows_amd64.zip": {
                                "content": b"Windows AMD64",
                                "checksum": "c2d859efacc3fbe1662bef92c80ce32c966834597625914592891eaf564af4bf"
                            },
                            "terraform-provider-multiple-versions_1.5.0_darwin_amd64.zip": {
                                "content": b"Darwin AMD64",
                                "checksum": "e8bc51e741c45feed8d9d7eb1133ac0107152cab3c1db12e74495d4b4ec75a0c"
                            }
                        },
                        "documentation": {
                            ("overview", "hcl"): {
                                "type": terrareg.provider_documentation_type.ProviderDocumentationType.OVERVIEW,
                                "title": "Overview",
                                "description": "Overview of provider",
                                "filename": "index.md",
                                "subcategory": None,
                                "content": "This is an overview of the module!",
                                "id": 6344
                            },
                            ("some_resource", "hcl"): {
                                "type": terrareg.provider_documentation_type.ProviderDocumentationType.RESOURCE,
                                "title": "multiple_versions_thing",
                                "description": "Inital thing for multiple versions provider",
                                "filename": "data-sources/thing.md",
                                "subcategory": "some-subcategory",
                                "content": "Documentation for generating a thing!",
                                "id": 6345
                            },
                            ("some_resource", "python"): {
                                "type": terrareg.provider_documentation_type.ProviderDocumentationType.RESOURCE,
                                "title": "multiple_versions_thing",
                                "description": "Inital thing for multiple versions provider written for python",
                                "filename": "data-sources/thing.md",
                                "subcategory": "some-subcategory",
                                "content": "Documentation for generating a thing in python!",
                                "id": 6346
                            },
                            ("some_new_resource", "hcl"): {
                                "type": terrareg.provider_documentation_type.ProviderDocumentationType.RESOURCE,
                                "title": "multiple_versions_thing_new",
                                "description": "Description for new resource",
                                "filename": "resources/new-thing.md",
                                "subcategory": "some-second-subcategory",
                                "content": """
# Some Title!

## Second title

This module:

 * Creates something
 * Does something else

and it _really_ *does* work!
""",
                                "id": 6347
                            },
                        }
                    },
                    "2.0.0": {
                        "git_tag": "v2.0.0",
                        "gpg_key_fingerprint": "21A74E4E3FDFE438532BD58434DE374AC3640CDB"
                    },
                    "2.0.1": {
                        "git_tag": "v2.0.1",
                        "gpg_key_fingerprint": "94CA72B7A2F4606A6C18211AE94A4F2AD628D926"
                    }
                }
            },
        }
    }
}