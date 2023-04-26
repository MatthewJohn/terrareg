
from datetime import datetime
import json


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

integration_test_data = {
    'testnamespace': {
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
    },
    'modulesearch': {
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
    },
    'modulesearch-contributed': {
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
    },
    'modulesearch-trusted': {
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
    },
    'searchbynamespace': {
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
    },
    'searchbynamesp-similar': {
        'searchbymodulename3': {'searchbyprovideraws': {
            'id': 44,
            'versions': {'4.4.1': {'published': True}},
            'verified': True
        }},
        'searchbymodulename4': {'aws': {
            'id': 45,
            'versions': {'5.5.5': {'published': True}}
        }}
    },
    'genericmodules': {
        'modulename': {'providername': {
            'id': 48,
            'versions': {'1.2.3': {'published': True}}
        }}
    },
    'moduledetails': {
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
    },
    'mostrecentextractionincomplete': {
        'modulename': {'providername': {
            'id': 61,
            'versions': {
                # Test with module version somehow getting published True, with extraction complete as False
                '1.2.1': {'published': True, 'extraction_complete': False},

                '1.2.3': {'published': False, 'extraction_complete': False},
                '1.5.3-beta': {'published': True, 'beta': True, 'extraction_complete': False}
            }
        }}
    },
}