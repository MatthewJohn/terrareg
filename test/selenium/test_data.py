
from datetime import datetime
import json

from terrareg.database import Database


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
    },
    4: {
        'name': 'no_browse_url',
        'base_url_template': 'https://base-url.com/{namespace}/{module}-{provider}',
        'browse_url_template': None,
        'clone_url_template': 'ssh://clone-url.com/{namespace}/{module}-{provider}'
    }
}

one_namespace_test_data = {
    'testnamespace': {
        'onemodule': {'testprovider': {
            'id': 1,
            'versions': {'1.5.0': {'published': True}}
        }},
        'module-two': {'testprovider': {
            'id': 2,
            'versions': {'1.8.0': {'published': True}}
        }}
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
        }}
    },
    'onlyunpublished': {
        'betamodule': {'test': {
            'id': 60,
            'versions': {
                '1.5.0': {'published': False}
            }
        }}
    },
    'onlybeta': {
        'betamodule': {'test': {
            'id': 61,
            'versions': {
                '1.4.0-beta': {'beta': True, 'published': True}
            }
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
                    '1.0.0': {'published': True}
                }
            },
            'gcp': {
                'id': 23,
                'versions': {
                    '1.0.0': {'published': True}
                }
            },
            'null': {
                'id': 55,
                'versions': {
                    '1.0.0': {'published': True}
                }
            },
            'datadog': {
                'id': 59,
                'versions': {
                    '1.0.0': {'published': True}
                }
            },
            'doesnotexist': {
                'id': 24,
                'versions': {
                    '1.0.0': {'published': True}
                }
            },
            'consul': {
                'id': 79,
                'versions': {
                    '1.0.0': {'published': True}
                }
            },
            'nomad': {
                'id': 80,
                'versions': {
                    '1.0.0': {'published': True}
                }
            },
            'vagrant': {
                'id': 81,
                'versions': {
                    '1.0.0': {'published': True}
                }
            },
            'vault': {
                'id': 82,
                'versions': {
                    '1.0.0': {'published': True}
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
            'versions': {'1.0.0': {'published': True}}
        }},
        'contributedmodule-multiversion': {'aws': {
            'id': 26,
            'versions': {
                '1.2.3': {'published': True},
                '2.0.0': {'published': True}
            }
        }},
        'contributedmodule-withbetaversion': {'aws': {
            'id': 49,
            'versions': {
                '1.2.3': {'published': True},
                '2.0.0-beta': {'published': True, 'beta': True}
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
                '1.0.0': {}
            }
        }},
        'verifiedmodule-oneversion': {'aws': {
            'verified': True,
            'id': 29,
            'versions': {'1.0.0': {'published': True}}
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
        'mixedsearch-trusted-second-result': {'datadog': {
            'id': 37,
            'versions': {
                '5.2.1': {'published': True},
            }
        }},
        'mixedsearch-trusted-result-multiversion': {'null': {
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
        'mixedsearch-trusted-result-verified': {'gcp': {
            'id': 54,
            'verified': True,
            'versions': {
                '2.0.0': {'published': True}
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
    'trustednamespace': {
        'secondlatestmodule': {'aws': {
            'id': 44,
            'versions': {'4.4.1': {'published': True}}
        }},
        'searchbymodulename4': {'aws': {
            'id': 45,
            'versions': {'5.5.5': {'published': True}}
        }}
    },
    'relevancysearch': {
        'partialmodulenamematch': {
            'partialprovidernamematch': {
                'id': 69,
                'versions': {'1.0.0': {'published': True}}
            },
            'namematch': {
                'id': 78,
                'versions': {'1.0.0': {'published': True}}
            }
        },
        'namematch': {
            'namematch': {
                'id': 70,
                'versions': {'1.0.0': {'published': True}}
            },
            'partialprovidernamematch': {
                'id': 71,
                'versions': {'1.0.0': {'published': True}}
            }
        },
        # This feels unlikely to happen
        'descriptionmatch': {'testprovider': {
            'id': 72,
            'versions': {'1.0.0': {
                'published': True,
                'description': 'namematch'
            }}
        }},
        'partialdescriptionmatch': {'testprovider': {
            'id': 73,
            'versions': {'1.0.0': {
                'published': True,
                'description': 'partialnamematch'
            }}
        }},
        'olddescriptionmatch': {'testprovider': {
            'id': 74,
            'versions': {
                '1.0.0': {
                    'published': True,
                    'description': 'namematch'
                },
                '1.1.0': {
                    'published': True
                }
            }
        }},
        'ownermatch': {'testprovider': {
            'id': 75,
            'versions': {'1.0.0': {
                'published': True,
                'owner': 'namematch'
            }}
        }},
        'partialownermatch': {'testprovider': {
            'id': 76,
            'versions': {'1.0.0': {
                'published': True,
                'owner': 'partialnamematch'
            }}
        }},
        'oldownermatch': {'testprovider': {
            'id': 77,
            'versions': {
                '1.0.0': {
                    'published': True,
                    'owner': 'namematch'
                },
                '1.1.0': {
                    'published': True
                }
            }
        }},
    },
    'withdisplayname': {
        'display_name': 'A Display Name'
    },
    'moduledetails': {
        'fullypopulated': {'testprovider': {
            'id': 56,
            'repo_base_url_template': 'https://mp-base-url.com/{namespace}/{module}-{provider}',
            'repo_browse_url_template': 'https://mp-browse-url.com/{namespace}/{module}-{provider}/browse/{tag}/{path}suffix',
            'repo_clone_url_template': 'ssh://mp-clone-url.com/{namespace}/{module}-{provider}',
            'versions': {
                # Older version
                '1.2.0': {
                    'published': True,
                    'terraform_docs': json.dumps({
                        'header': '',
                        'footer': '',
                        'inputs': [],
                        'modules': [],
                        'outputs': [],
                        'providers': [],
                        'requirements': [
                            {
                                "name": "terraform",
                                "version": ">= 2.1.1, < 2.5.4"
                            },
                            {
                                "name": "someothercompany/unsafe",
                                "version": ">= 4.45"
                            }
                        ],
                        'resources': []
                    }),
                    'examples': {
                        'examples/old-version-example': {
                            'example_files': {
                                'examples/test-example/main.tf': '# Call root module\nmodule "old_version_root_call" {\n  source = "../../"\n}'
                            },
                        }
                    }
                },
                # Newer unpublished version
                '1.6.0': {
                    'examples': {
                        'examples/unpublished-example': {
                            'example_files': {
                                'examples/test-example/main.tf': '# Call root module\nmodule "unpublished_root_call" {\n  source = "../../"\n}'
                            },
                        }
                    }
                },
                # Newer published beta version
                '1.6.1-beta': {'published': True, 'beta': True},
                # Unpublished and beta version
                '1.0.0-beta': {'published': False, 'beta': True},
                '1.7.0-beta': {
                    'published': True,
                    'beta': True,
                    'terraform_docs': json.dumps({
                        'header': '',
                        'footer': '',
                        'inputs': [],
                        'modules': [],
                        'outputs': [],
                        'providers': [],
                        'requirements': [
                            {
                                "name": "terraform",
                                "version": ">= 5.12, < 21.0.0"
                            },
                            {
                                "name": "someothercompany/unsafe",
                                "version": ">= 4.45"
                            }
                        ],
                        'resources': []
                    }),
                    'examples': {
                        'examples/beta-example': {
                            'example_files': {
                                'examples/test-example/main.tf': '# Call root module\nmodule "beta_root_call" {\n  source = "../../"\n}'
                            },
                        }
                    }
                },
                '1.5.0': {
                    'description': 'This is a test module version for tests.',
                    'owner': 'This is the owner of the module',
                    'repo_base_url_template': 'https://link-to.com/source-code-here',
                    'published': True,
                    'beta': False,
                    'internal': False,
                    'published_at': datetime(2022, 1, 5, 22, 53, 12),
                    'readme_content': """
# This is an example README!

Following this example module call:

```
module "test_example_call" {
  source = "../../"

  name = "example-name"
}
```

This should work with all versions > 5.2.0 and <= 6.0.0

```
module "text_ternal_call" {
  source  = "a-public/module"
  version = "> 5.2.0, <= 6.0.0"

  another = "example-external"
}
```
""",
                    # Set to non-latest extraction version
                    'extraction_version': 0,
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
                                'name': 'example_boolean_input',
                                'type': 'bool',
                                'description': 'Override the truthful boolean',
                                'default': True,
                                'required': False
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
                                # Ensure GT and LT are displayed correctly in browser
                                'version': '>= 5.2.1, < 6.0.0'
                            },
                            {
                                'name': 'someothercompany/unsafe',
                                'alias': None,
                                'version': '2.0.0'
                            }
                        ],
                        'requirements': [
                            {
                                "name": "terraform",
                                "version": ">= 1.0, < 2.0.0"
                            },
                            {
                                "name": "someothercompany/unsafe",
                                "version": ">= 4.45"
                            }
                        ],
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
                    }),
                    "terraform_graph": """
digraph {
	compound = "true"
	newrank = "true"
	subgraph "root" {
		"[root] aws_s3_bucket.test_bucket (expand)" [label = "aws_s3_bucket.test_bucket", shape = "box"]
		"[root] aws_s3_object.test_obj_root_module (expand)" [label = "aws_s3_object.test_obj_root_module", shape = "box"]
		"[root] module.count_call.aws_instance.web (expand)" [label = "module.count_call.aws_instance.web", shape = "box"]
		"[root] module.count_call.data.aws_ami.ubuntu (expand)" [label = "module.count_call.data.aws_ami.ubuntu", shape = "box"]
		"[root] module.count_call.module.second-module-call.aws_instance.test_second_instance (expand)" [label = "module.count_call.module.second-module-call.aws_instance.test_second_instance", shape = "box"]
		"[root] module.count_call.module.second-module-call.data.aws_ami.ubuntu (expand)" [label = "module.count_call.module.second-module-call.data.aws_ami.ubuntu", shape = "box"]
		"[root] module.for_each_call.aws_instance.web (expand)" [label = "module.for_each_call.aws_instance.web", shape = "box"]
		"[root] module.for_each_call.data.aws_ami.ubuntu (expand)" [label = "module.for_each_call.data.aws_ami.ubuntu", shape = "box"]
		"[root] module.for_each_call.module.second-module-call.aws_instance.test_second_instance (expand)" [label = "module.for_each_call.module.second-module-call.aws_instance.test_second_instance", shape = "box"]
		"[root] module.for_each_call.module.second-module-call.data.aws_ami.ubuntu (expand)" [label = "module.for_each_call.module.second-module-call.data.aws_ami.ubuntu", shape = "box"]
		"[root] module.submodule-call.aws_instance.web (expand)" [label = "module.submodule-call.aws_instance.web", shape = "box"]
		"[root] module.submodule-call.data.aws_ami.ubuntu (expand)" [label = "module.submodule-call.data.aws_ami.ubuntu", shape = "box"]
		"[root] module.submodule-call.module.second-module-call.aws_instance.test_second_instance (expand)" [label = "module.submodule-call.module.second-module-call.aws_instance.test_second_instance", shape = "box"]
		"[root] module.submodule-call.module.second-module-call.data.aws_ami.ubuntu (expand)" [label = "module.submodule-call.module.second-module-call.data.aws_ami.ubuntu", shape = "box"]
		"[root] output.name" [label = "output.name", shape = "note"]
		"[root] provider[\\"registry.terraform.io/hashicorp/aws\\"]" [label = "provider[\\"registry.terraform.io/hashicorp/aws\\"]", shape = "diamond"]
		"[root] var.name" [label = "var.name", shape = "note"]
		"[root] aws_s3_bucket.test_bucket (expand)" -> "[root] provider[\\"registry.terraform.io/hashicorp/aws\\"]"
		"[root] aws_s3_bucket.test_bucket (expand)" -> "[root] var.name"
		"[root] aws_s3_object.test_obj_root_module (expand)" -> "[root] aws_s3_bucket.test_bucket (expand)"
		"[root] module.count_call (close)" -> "[root] module.count_call.aws_instance.web (expand)"
		"[root] module.count_call (close)" -> "[root] module.count_call.module.second-module-call (close)"
		"[root] module.count_call (close)" -> "[root] module.count_call.var.passing_name (expand)"
		"[root] module.count_call.aws_instance.web (expand)" -> "[root] module.count_call.data.aws_ami.ubuntu (expand)"
		"[root] module.count_call.data.aws_ami.ubuntu (expand)" -> "[root] module.count_call (expand)"
		"[root] module.count_call.data.aws_ami.ubuntu (expand)" -> "[root] provider[\\"registry.terraform.io/hashicorp/aws\\"]"
		"[root] module.count_call.module.second-module-call (close)" -> "[root] module.count_call.module.second-module-call.aws_instance.test_second_instance (expand)"
		"[root] module.count_call.module.second-module-call (expand)" -> "[root] module.count_call (expand)"
		"[root] module.count_call.module.second-module-call.aws_instance.test_second_instance (expand)" -> "[root] module.count_call.module.second-module-call.data.aws_ami.ubuntu (expand)"
		"[root] module.count_call.module.second-module-call.data.aws_ami.ubuntu (expand)" -> "[root] module.count_call.module.second-module-call (expand)"
		"[root] module.count_call.module.second-module-call.data.aws_ami.ubuntu (expand)" -> "[root] provider[\\"registry.terraform.io/hashicorp/aws\\"]"
		"[root] module.count_call.var.passing_name (expand)" -> "[root] module.count_call (expand)"
		"[root] module.for_each_call (close)" -> "[root] module.for_each_call.aws_instance.web (expand)"
		"[root] module.for_each_call (close)" -> "[root] module.for_each_call.module.second-module-call (close)"
		"[root] module.for_each_call (close)" -> "[root] module.for_each_call.var.passing_name (expand)"
		"[root] module.for_each_call.aws_instance.web (expand)" -> "[root] module.for_each_call.data.aws_ami.ubuntu (expand)"
		"[root] module.for_each_call.data.aws_ami.ubuntu (expand)" -> "[root] module.for_each_call (expand)"
		"[root] module.for_each_call.data.aws_ami.ubuntu (expand)" -> "[root] provider[\\"registry.terraform.io/hashicorp/aws\\"]"
		"[root] module.for_each_call.module.second-module-call (close)" -> "[root] module.for_each_call.module.second-module-call.aws_instance.test_second_instance (expand)"
		"[root] module.for_each_call.module.second-module-call (expand)" -> "[root] module.for_each_call (expand)"
		"[root] module.for_each_call.module.second-module-call.aws_instance.test_second_instance (expand)" -> "[root] module.for_each_call.module.second-module-call.data.aws_ami.ubuntu (expand)"
		"[root] module.for_each_call.module.second-module-call.data.aws_ami.ubuntu (expand)" -> "[root] module.for_each_call.module.second-module-call (expand)"
		"[root] module.for_each_call.module.second-module-call.data.aws_ami.ubuntu (expand)" -> "[root] provider[\\"registry.terraform.io/hashicorp/aws\\"]"
		"[root] module.for_each_call.var.passing_name (expand)" -> "[root] module.for_each_call (expand)"
		"[root] module.submodule-call (close)" -> "[root] module.submodule-call.aws_instance.web (expand)"
		"[root] module.submodule-call (close)" -> "[root] module.submodule-call.module.second-module-call (close)"
		"[root] module.submodule-call (close)" -> "[root] module.submodule-call.var.passing_name (expand)"
		"[root] module.submodule-call.aws_instance.web (expand)" -> "[root] module.submodule-call.data.aws_ami.ubuntu (expand)"
		"[root] module.submodule-call.data.aws_ami.ubuntu (expand)" -> "[root] module.submodule-call (expand)"
		"[root] module.submodule-call.data.aws_ami.ubuntu (expand)" -> "[root] provider[\\"registry.terraform.io/hashicorp/aws\\"]"
		"[root] module.submodule-call.module.second-module-call (close)" -> "[root] module.submodule-call.module.second-module-call.aws_instance.test_second_instance (expand)"
		"[root] module.submodule-call.module.second-module-call (expand)" -> "[root] module.submodule-call (expand)"
		"[root] module.submodule-call.module.second-module-call.aws_instance.test_second_instance (expand)" -> "[root] module.submodule-call.module.second-module-call.data.aws_ami.ubuntu (expand)"
		"[root] module.submodule-call.module.second-module-call.data.aws_ami.ubuntu (expand)" -> "[root] module.submodule-call.module.second-module-call (expand)"
		"[root] module.submodule-call.module.second-module-call.data.aws_ami.ubuntu (expand)" -> "[root] provider[\\"registry.terraform.io/hashicorp/aws\\"]"
		"[root] module.submodule-call.var.passing_name (expand)" -> "[root] module.submodule-call (expand)"
		"[root] module.submodule-call.var.passing_name (expand)" -> "[root] var.name"
		"[root] output.name" -> "[root] aws_s3_bucket.test_bucket (expand)"
		"[root] provider[\\"registry.terraform.io/hashicorp/aws\\"] (close)" -> "[root] aws_s3_object.test_obj_root_module (expand)"
		"[root] provider[\\"registry.terraform.io/hashicorp/aws\\"] (close)" -> "[root] module.count_call.aws_instance.web (expand)"
		"[root] provider[\\"registry.terraform.io/hashicorp/aws\\"] (close)" -> "[root] module.count_call.module.second-module-call.aws_instance.test_second_instance (expand)"
		"[root] provider[\\"registry.terraform.io/hashicorp/aws\\"] (close)" -> "[root] module.for_each_call.aws_instance.web (expand)"
		"[root] provider[\\"registry.terraform.io/hashicorp/aws\\"] (close)" -> "[root] module.for_each_call.module.second-module-call.aws_instance.test_second_instance (expand)"
		"[root] provider[\\"registry.terraform.io/hashicorp/aws\\"] (close)" -> "[root] module.submodule-call.aws_instance.web (expand)"
		"[root] provider[\\"registry.terraform.io/hashicorp/aws\\"] (close)" -> "[root] module.submodule-call.module.second-module-call.aws_instance.test_second_instance (expand)"
		"[root] root" -> "[root] module.count_call (close)"
		"[root] root" -> "[root] module.for_each_call (close)"
		"[root] root" -> "[root] module.submodule-call (close)"
		"[root] root" -> "[root] output.name"
		"[root] root" -> "[root] provider[\\"registry.terraform.io/hashicorp/aws\\"] (close)"
	}
}


""",
                    'files': {
                        'LICENSE': """
This is a license file
All rights are not reserved for this example file content
This license > tests
various < characters that could be escaped.
""".strip(),
                        'CHANGELOG.md': """
# Changelog
## 1.0.0
 * This is an initial release

This tests > 2 < 3 escapable characters
""".strip(),
                        'NOT_REFERENCED': 'This file is not referenced by a tab'
                    },
                    'examples': {
                        'examples/test-example': {
                            'example_files': {
                                'examples/test-example/data.tf': '# This contains data objects',
                                'examples/test-example/variables.tf': 'variable "test" {\n  description = "test variable"\n  type = string\n}',
                                'examples/test-example/main.tf': '# Call root module\nmodule "root" {\n  source = "../../"\n}'
                            },
                            'readme_content': '# Example 1 README',
                            'infracost': json.dumps(
                                {
                                    "version": "0.2",
                                    "metadata": {
                                        "infracostCommand": "breakdown",
                                        "branch": "main",
                                        "commit": "1746e56051774c012d5f3de38534c815bba76746",
                                        "commitAuthorName": "Matthew John",
                                        "commitAuthorEmail": "matthew@dockstudios.co.uk",
                                        "commitTimestamp": "2023-01-19T07:49:54Z",
                                        "commitMessage": "Fix bucket and ec2 instance",
                                        "vcsRepoUrl": "https://gitlab.dockstudios.co.uk:2222/pub/terrareg/snippets/5.git"
                                    },
                                    "currency": "USD",
                                    "projects": [
                                        {
                                            "name": "2222/pub/terrareg/snippets/5/examples/basic_usage",
                                            "metadata": {
                                                "path": ".",
                                                "type": "terraform_dir",
                                                "vcsSubPath": "examples/basic_usage"
                                            },
                                            "pastBreakdown": {
                                                "resources": [],
                                                "totalHourlyCost": "0",
                                                "totalMonthlyCost": "0"
                                            },
                                            "breakdown": {
                                                "resources": [
                                                    {
                                                        "name": "module.main_call.aws_s3_bucket.test_bucket",
                                                        "metadata": {
                                                            "calls": [
                                                                {
                                                                    "blockName": "module.main_call",
                                                                    "filename": "main.tf"
                                                                },
                                                                {
                                                                    "blockName": "aws_s3_bucket.test_bucket",
                                                                    "filename": "../../main.tf"
                                                                }
                                                            ],
                                                            "filename": "../../main.tf"
                                                        },
                                                        "hourlyCost": None,
                                                        "monthlyCost": None,
                                                        "subresources": [
                                                            {
                                                                "name": "Standard",
                                                                "metadata": {},
                                                                "hourlyCost": None,
                                                                "monthlyCost": None,
                                                                "costComponents": [
                                                                    {
                                                                        "name": "Storage",
                                                                        "unit": "GB",
                                                                        "hourlyQuantity": None,
                                                                        "monthlyQuantity": None,
                                                                        "price": "0.023",
                                                                        "hourlyCost": None,
                                                                        "monthlyCost": None
                                                                    },
                                                                    {
                                                                        "name": "PUT, COPY, POST, LIST requests",
                                                                        "unit": "1k requests",
                                                                        "hourlyQuantity": None,
                                                                        "monthlyQuantity": None,
                                                                        "price": "0.005",
                                                                        "hourlyCost": None,
                                                                        "monthlyCost": None
                                                                    },
                                                                    {
                                                                        "name": "GET, SELECT, and all other requests",
                                                                        "unit": "1k requests",
                                                                        "hourlyQuantity": None,
                                                                        "monthlyQuantity": None,
                                                                        "price": "0.0004",
                                                                        "hourlyCost": None,
                                                                        "monthlyCost": None
                                                                    },
                                                                    {
                                                                        "name": "Select data scanned",
                                                                        "unit": "GB",
                                                                        "hourlyQuantity": None,
                                                                        "monthlyQuantity": None,
                                                                        "price": "0.002",
                                                                        "hourlyCost": None,
                                                                        "monthlyCost": None
                                                                    },
                                                                    {
                                                                        "name": "Select data returned",
                                                                        "unit": "GB",
                                                                        "hourlyQuantity": None,
                                                                        "monthlyQuantity": None,
                                                                        "price": "0.0007",
                                                                        "hourlyCost": None,
                                                                        "monthlyCost": None
                                                                    }
                                                                ]
                                                            }
                                                        ]
                                                    },
                                                    {
                                                        "name": "module.main_call.module.count_call[0].aws_instance.web",
                                                        "tags": {
                                                            "Name": "HelloWorld"
                                                        },
                                                        "metadata": {
                                                            "calls": [
                                                                {
                                                                    "blockName": "module.main_call",
                                                                    "filename": "main.tf"
                                                                },
                                                                {
                                                                    "blockName": "module.count_call",
                                                                    "filename": "../../main.tf"
                                                                },
                                                                {
                                                                    "blockName": "aws_instance.web",
                                                                    "filename": "../../modules/test-submodule/main.tf"
                                                                }
                                                            ],
                                                            "filename": "../../modules/test-submodule/main.tf"
                                                        },
                                                        "hourlyCost": "0.0114958904109589",
                                                        "monthlyCost": "8.392",
                                                        "costComponents": [
                                                            {
                                                                "name": "Instance usage (Linux/UNIX, on-demand, t3.micro)",
                                                                "unit": "hours",
                                                                "hourlyQuantity": "1",
                                                                "monthlyQuantity": "730",
                                                                "price": "0.0104",
                                                                "hourlyCost": "0.0104",
                                                                "monthlyCost": "7.592"
                                                            },
                                                            {
                                                                "name": "CPU credits",
                                                                "unit": "vCPU-hours",
                                                                "hourlyQuantity": "0",
                                                                "monthlyQuantity": "0",
                                                                "price": "0.05",
                                                                "hourlyCost": "0",
                                                                "monthlyCost": "0"
                                                            }
                                                        ],
                                                        "subresources": [
                                                            {
                                                                "name": "root_block_device",
                                                                "metadata": {},
                                                                "hourlyCost": "0.0010958904109589",
                                                                "monthlyCost": "0.8",
                                                                "costComponents": [
                                                                    {
                                                                        "name": "Storage (general purpose SSD, gp2)",
                                                                        "unit": "GB",
                                                                        "hourlyQuantity": "0.010958904109589",
                                                                        "monthlyQuantity": "8",
                                                                        "price": "0.1",
                                                                        "hourlyCost": "0.0010958904109589",
                                                                        "monthlyCost": "0.8"
                                                                    }
                                                                ]
                                                            }
                                                        ]
                                                    },
                                                    {
                                                        "name": "module.main_call.module.count_call[0].module.second-module-call.aws_instance.test_second_instance",
                                                        "metadata": {
                                                            "calls": [
                                                                {
                                                                    "blockName": "module.main_call",
                                                                    "filename": "main.tf"
                                                                },
                                                                {
                                                                    "blockName": "module.count_call",
                                                                    "filename": "../../main.tf"
                                                                },
                                                                {
                                                                    "blockName": "module.second-module-call",
                                                                    "filename": "../../modules/test-submodule/main.tf"
                                                                },
                                                                {
                                                                    "blockName": "aws_instance.test_second_instance",
                                                                    "filename": "../../modules/second-submodule/main.tf"
                                                                }
                                                            ],
                                                            "filename": "../../modules/second-submodule/main.tf"
                                                        },
                                                        "hourlyCost": "0.0426958904109589",
                                                        "monthlyCost": "31.168",
                                                        "costComponents": [
                                                            {
                                                                "name": "Instance usage (Linux/UNIX, on-demand, t3.medium)",
                                                                "unit": "hours",
                                                                "hourlyQuantity": "1",
                                                                "monthlyQuantity": "730",
                                                                "price": "0.0416",
                                                                "hourlyCost": "0.0416",
                                                                "monthlyCost": "30.368"
                                                            },
                                                            {
                                                                "name": "CPU credits",
                                                                "unit": "vCPU-hours",
                                                                "hourlyQuantity": "0",
                                                                "monthlyQuantity": "0",
                                                                "price": "0.05",
                                                                "hourlyCost": "0",
                                                                "monthlyCost": "0"
                                                            }
                                                        ],
                                                        "subresources": [
                                                            {
                                                                "name": "root_block_device",
                                                                "metadata": {},
                                                                "hourlyCost": "0.0010958904109589",
                                                                "monthlyCost": "0.8",
                                                                "costComponents": [
                                                                    {
                                                                        "name": "Storage (general purpose SSD, gp2)",
                                                                        "unit": "GB",
                                                                        "hourlyQuantity": "0.010958904109589",
                                                                        "monthlyQuantity": "8",
                                                                        "price": "0.1",
                                                                        "hourlyCost": "0.0010958904109589",
                                                                        "monthlyCost": "0.8"
                                                                    }
                                                                ]
                                                            }
                                                        ]
                                                    },
                                                    {
                                                        "name": "module.main_call.module.count_call[1].aws_instance.web",
                                                        "tags": {
                                                            "Name": "HelloWorld"
                                                        },
                                                        "metadata": {
                                                            "calls": [
                                                                {
                                                                    "blockName": "module.main_call",
                                                                    "filename": "main.tf"
                                                                },
                                                                {
                                                                    "blockName": "module.count_call",
                                                                    "filename": "../../main.tf"
                                                                },
                                                                {
                                                                    "blockName": "aws_instance.web",
                                                                    "filename": "../../modules/test-submodule/main.tf"
                                                                }
                                                            ],
                                                            "filename": "../../modules/test-submodule/main.tf"
                                                        },
                                                        "hourlyCost": "0.0114958904109589",
                                                        "monthlyCost": "8.392",
                                                        "costComponents": [
                                                            {
                                                                "name": "Instance usage (Linux/UNIX, on-demand, t3.micro)",
                                                                "unit": "hours",
                                                                "hourlyQuantity": "1",
                                                                "monthlyQuantity": "730",
                                                                "price": "0.0104",
                                                                "hourlyCost": "0.0104",
                                                                "monthlyCost": "7.592"
                                                            },
                                                            {
                                                                "name": "CPU credits",
                                                                "unit": "vCPU-hours",
                                                                "hourlyQuantity": "0",
                                                                "monthlyQuantity": "0",
                                                                "price": "0.05",
                                                                "hourlyCost": "0",
                                                                "monthlyCost": "0"
                                                            }
                                                        ],
                                                        "subresources": [
                                                            {
                                                                "name": "root_block_device",
                                                                "metadata": {},
                                                                "hourlyCost": "0.0010958904109589",
                                                                "monthlyCost": "0.8",
                                                                "costComponents": [
                                                                    {
                                                                        "name": "Storage (general purpose SSD, gp2)",
                                                                        "unit": "GB",
                                                                        "hourlyQuantity": "0.010958904109589",
                                                                        "monthlyQuantity": "8",
                                                                        "price": "0.1",
                                                                        "hourlyCost": "0.0010958904109589",
                                                                        "monthlyCost": "0.8"
                                                                    }
                                                                ]
                                                            }
                                                        ]
                                                    },
                                                    {
                                                        "name": "module.main_call.module.count_call[1].module.second-module-call.aws_instance.test_second_instance",
                                                        "metadata": {
                                                            "calls": [
                                                                {
                                                                    "blockName": "module.main_call",
                                                                    "filename": "main.tf"
                                                                },
                                                                {
                                                                    "blockName": "module.count_call",
                                                                    "filename": "../../main.tf"
                                                                },
                                                                {
                                                                    "blockName": "module.second-module-call",
                                                                    "filename": "../../modules/test-submodule/main.tf"
                                                                },
                                                                {
                                                                    "blockName": "aws_instance.test_second_instance",
                                                                    "filename": "../../modules/second-submodule/main.tf"
                                                                }
                                                            ],
                                                            "filename": "../../modules/second-submodule/main.tf"
                                                        },
                                                        "hourlyCost": "0.0426958904109589",
                                                        "monthlyCost": "31.168",
                                                        "costComponents": [
                                                            {
                                                                "name": "Instance usage (Linux/UNIX, on-demand, t3.medium)",
                                                                "unit": "hours",
                                                                "hourlyQuantity": "1",
                                                                "monthlyQuantity": "730",
                                                                "price": "0.0416",
                                                                "hourlyCost": "0.0416",
                                                                "monthlyCost": "30.368"
                                                            },
                                                            {
                                                                "name": "CPU credits",
                                                                "unit": "vCPU-hours",
                                                                "hourlyQuantity": "0",
                                                                "monthlyQuantity": "0",
                                                                "price": "0.05",
                                                                "hourlyCost": "0",
                                                                "monthlyCost": "0"
                                                            }
                                                        ],
                                                        "subresources": [
                                                            {
                                                                "name": "root_block_device",
                                                                "metadata": {},
                                                                "hourlyCost": "0.0010958904109589",
                                                                "monthlyCost": "0.8",
                                                                "costComponents": [
                                                                    {
                                                                        "name": "Storage (general purpose SSD, gp2)",
                                                                        "unit": "GB",
                                                                        "hourlyQuantity": "0.010958904109589",
                                                                        "monthlyQuantity": "8",
                                                                        "price": "0.1",
                                                                        "hourlyCost": "0.0010958904109589",
                                                                        "monthlyCost": "0.8"
                                                                    }
                                                                ]
                                                            }
                                                        ]
                                                    },
                                                    {
                                                        "name": "module.main_call.module.for_each_call[\"a-value\"].aws_instance.web",
                                                        "tags": {
                                                            "Name": "HelloWorld"
                                                        },
                                                        "metadata": {
                                                            "calls": [
                                                                {
                                                                    "blockName": "module.main_call",
                                                                    "filename": "main.tf"
                                                                },
                                                                {
                                                                    "blockName": "module.for_each_call[\"a-value\"]",
                                                                    "filename": "../../main.tf"
                                                                },
                                                                {
                                                                    "blockName": "aws_instance.web",
                                                                    "filename": "../../modules/test-submodule/main.tf"
                                                                }
                                                            ],
                                                            "filename": "../../modules/test-submodule/main.tf"
                                                        },
                                                        "hourlyCost": "0.0114958904109589",
                                                        "monthlyCost": "8.392",
                                                        "costComponents": [
                                                            {
                                                                "name": "Instance usage (Linux/UNIX, on-demand, t3.micro)",
                                                                "unit": "hours",
                                                                "hourlyQuantity": "1",
                                                                "monthlyQuantity": "730",
                                                                "price": "0.0104",
                                                                "hourlyCost": "0.0104",
                                                                "monthlyCost": "7.592"
                                                            },
                                                            {
                                                                "name": "CPU credits",
                                                                "unit": "vCPU-hours",
                                                                "hourlyQuantity": "0",
                                                                "monthlyQuantity": "0",
                                                                "price": "0.05",
                                                                "hourlyCost": "0",
                                                                "monthlyCost": "0"
                                                            }
                                                        ],
                                                        "subresources": [
                                                            {
                                                                "name": "root_block_device",
                                                                "metadata": {},
                                                                "hourlyCost": "0.0010958904109589",
                                                                "monthlyCost": "0.8",
                                                                "costComponents": [
                                                                    {
                                                                        "name": "Storage (general purpose SSD, gp2)",
                                                                        "unit": "GB",
                                                                        "hourlyQuantity": "0.010958904109589",
                                                                        "monthlyQuantity": "8",
                                                                        "price": "0.1",
                                                                        "hourlyCost": "0.0010958904109589",
                                                                        "monthlyCost": "0.8"
                                                                    }
                                                                ]
                                                            }
                                                        ]
                                                    },
                                                    {
                                                        "name": "module.main_call.module.for_each_call[\"a-value\"].module.second-module-call.aws_instance.test_second_instance",
                                                        "metadata": {
                                                            "calls": [
                                                                {
                                                                    "blockName": "module.main_call",
                                                                    "filename": "main.tf"
                                                                },
                                                                {
                                                                    "blockName": "module.for_each_call[\"a-value\"]",
                                                                    "filename": "../../main.tf"
                                                                },
                                                                {
                                                                    "blockName": "module.second-module-call",
                                                                    "filename": "../../modules/test-submodule/main.tf"
                                                                },
                                                                {
                                                                    "blockName": "aws_instance.test_second_instance",
                                                                    "filename": "../../modules/second-submodule/main.tf"
                                                                }
                                                            ],
                                                            "filename": "../../modules/second-submodule/main.tf"
                                                        },
                                                        "hourlyCost": "0.0426958904109589",
                                                        "monthlyCost": "31.168",
                                                        "costComponents": [
                                                            {
                                                                "name": "Instance usage (Linux/UNIX, on-demand, t3.medium)",
                                                                "unit": "hours",
                                                                "hourlyQuantity": "1",
                                                                "monthlyQuantity": "730",
                                                                "price": "0.0416",
                                                                "hourlyCost": "0.0416",
                                                                "monthlyCost": "30.368"
                                                            },
                                                            {
                                                                "name": "CPU credits",
                                                                "unit": "vCPU-hours",
                                                                "hourlyQuantity": "0",
                                                                "monthlyQuantity": "0",
                                                                "price": "0.05",
                                                                "hourlyCost": "0",
                                                                "monthlyCost": "0"
                                                            }
                                                        ],
                                                        "subresources": [
                                                            {
                                                                "name": "root_block_device",
                                                                "metadata": {},
                                                                "hourlyCost": "0.0010958904109589",
                                                                "monthlyCost": "0.8",
                                                                "costComponents": [
                                                                    {
                                                                        "name": "Storage (general purpose SSD, gp2)",
                                                                        "unit": "GB",
                                                                        "hourlyQuantity": "0.010958904109589",
                                                                        "monthlyQuantity": "8",
                                                                        "price": "0.1",
                                                                        "hourlyCost": "0.0010958904109589",
                                                                        "monthlyCost": "0.8"
                                                                    }
                                                                ]
                                                            }
                                                        ]
                                                    },
                                                    {
                                                        "name": "module.main_call.module.for_each_call[\"second-value\"].aws_instance.web",
                                                        "tags": {
                                                            "Name": "HelloWorld"
                                                        },
                                                        "metadata": {
                                                            "calls": [
                                                                {
                                                                    "blockName": "module.main_call",
                                                                    "filename": "main.tf"
                                                                },
                                                                {
                                                                    "blockName": "module.for_each_call[\"second-value\"]",
                                                                    "filename": "../../main.tf"
                                                                },
                                                                {
                                                                    "blockName": "aws_instance.web",
                                                                    "filename": "../../modules/test-submodule/main.tf"
                                                                }
                                                            ],
                                                            "filename": "../../modules/test-submodule/main.tf"
                                                        },
                                                        "hourlyCost": "0.0114958904109589",
                                                        "monthlyCost": "8.392",
                                                        "costComponents": [
                                                            {
                                                                "name": "Instance usage (Linux/UNIX, on-demand, t3.micro)",
                                                                "unit": "hours",
                                                                "hourlyQuantity": "1",
                                                                "monthlyQuantity": "730",
                                                                "price": "0.0104",
                                                                "hourlyCost": "0.0104",
                                                                "monthlyCost": "7.592"
                                                            },
                                                            {
                                                                "name": "CPU credits",
                                                                "unit": "vCPU-hours",
                                                                "hourlyQuantity": "0",
                                                                "monthlyQuantity": "0",
                                                                "price": "0.05",
                                                                "hourlyCost": "0",
                                                                "monthlyCost": "0"
                                                            }
                                                        ],
                                                        "subresources": [
                                                            {
                                                                "name": "root_block_device",
                                                                "metadata": {},
                                                                "hourlyCost": "0.0010958904109589",
                                                                "monthlyCost": "0.8",
                                                                "costComponents": [
                                                                    {
                                                                        "name": "Storage (general purpose SSD, gp2)",
                                                                        "unit": "GB",
                                                                        "hourlyQuantity": "0.010958904109589",
                                                                        "monthlyQuantity": "8",
                                                                        "price": "0.1",
                                                                        "hourlyCost": "0.0010958904109589",
                                                                        "monthlyCost": "0.8"
                                                                    }
                                                                ]
                                                            }
                                                        ]
                                                    },
                                                    {
                                                        "name": "module.main_call.module.for_each_call[\"second-value\"].module.second-module-call.aws_instance.test_second_instance",
                                                        "metadata": {
                                                            "calls": [
                                                                {
                                                                    "blockName": "module.main_call",
                                                                    "filename": "main.tf"
                                                                },
                                                                {
                                                                    "blockName": "module.for_each_call[\"second-value\"]",
                                                                    "filename": "../../main.tf"
                                                                },
                                                                {
                                                                    "blockName": "module.second-module-call",
                                                                    "filename": "../../modules/test-submodule/main.tf"
                                                                },
                                                                {
                                                                    "blockName": "aws_instance.test_second_instance",
                                                                    "filename": "../../modules/second-submodule/main.tf"
                                                                }
                                                            ],
                                                            "filename": "../../modules/second-submodule/main.tf"
                                                        },
                                                        "hourlyCost": "0.0426958904109589",
                                                        "monthlyCost": "31.168",
                                                        "costComponents": [
                                                            {
                                                                "name": "Instance usage (Linux/UNIX, on-demand, t3.medium)",
                                                                "unit": "hours",
                                                                "hourlyQuantity": "1",
                                                                "monthlyQuantity": "730",
                                                                "price": "0.0416",
                                                                "hourlyCost": "0.0416",
                                                                "monthlyCost": "30.368"
                                                            },
                                                            {
                                                                "name": "CPU credits",
                                                                "unit": "vCPU-hours",
                                                                "hourlyQuantity": "0",
                                                                "monthlyQuantity": "0",
                                                                "price": "0.05",
                                                                "hourlyCost": "0",
                                                                "monthlyCost": "0"
                                                            }
                                                        ],
                                                        "subresources": [
                                                            {
                                                                "name": "root_block_device",
                                                                "metadata": {},
                                                                "hourlyCost": "0.0010958904109589",
                                                                "monthlyCost": "0.8",
                                                                "costComponents": [
                                                                    {
                                                                        "name": "Storage (general purpose SSD, gp2)",
                                                                        "unit": "GB",
                                                                        "hourlyQuantity": "0.010958904109589",
                                                                        "monthlyQuantity": "8",
                                                                        "price": "0.1",
                                                                        "hourlyCost": "0.0010958904109589",
                                                                        "monthlyCost": "0.8"
                                                                    }
                                                                ]
                                                            }
                                                        ]
                                                    },
                                                    {
                                                        "name": "module.main_call.module.submodule-call.aws_instance.web",
                                                        "tags": {
                                                            "Name": "HelloWorld"
                                                        },
                                                        "metadata": {
                                                            "calls": [
                                                                {
                                                                    "blockName": "module.main_call",
                                                                    "filename": "main.tf"
                                                                },
                                                                {
                                                                    "blockName": "module.submodule-call",
                                                                    "filename": "../../main.tf"
                                                                },
                                                                {
                                                                    "blockName": "aws_instance.web",
                                                                    "filename": "../../modules/test-submodule/main.tf"
                                                                }
                                                            ],
                                                            "filename": "../../modules/test-submodule/main.tf"
                                                        },
                                                        "hourlyCost": "0.0114958904109589",
                                                        "monthlyCost": "8.392",
                                                        "costComponents": [
                                                            {
                                                                "name": "Instance usage (Linux/UNIX, on-demand, t3.micro)",
                                                                "unit": "hours",
                                                                "hourlyQuantity": "1",
                                                                "monthlyQuantity": "730",
                                                                "price": "0.0104",
                                                                "hourlyCost": "0.0104",
                                                                "monthlyCost": "7.592"
                                                            },
                                                            {
                                                                "name": "CPU credits",
                                                                "unit": "vCPU-hours",
                                                                "hourlyQuantity": "0",
                                                                "monthlyQuantity": "0",
                                                                "price": "0.05",
                                                                "hourlyCost": "0",
                                                                "monthlyCost": "0"
                                                            }
                                                        ],
                                                        "subresources": [
                                                            {
                                                                "name": "root_block_device",
                                                                "metadata": {},
                                                                "hourlyCost": "0.0010958904109589",
                                                                "monthlyCost": "0.8",
                                                                "costComponents": [
                                                                    {
                                                                        "name": "Storage (general purpose SSD, gp2)",
                                                                        "unit": "GB",
                                                                        "hourlyQuantity": "0.010958904109589",
                                                                        "monthlyQuantity": "8",
                                                                        "price": "0.1",
                                                                        "hourlyCost": "0.0010958904109589",
                                                                        "monthlyCost": "0.8"
                                                                    }
                                                                ]
                                                            }
                                                        ]
                                                    },
                                                    {
                                                        "name": "module.main_call.module.submodule-call.module.second-module-call.aws_instance.test_second_instance",
                                                        "metadata": {
                                                            "calls": [
                                                                {
                                                                    "blockName": "module.main_call",
                                                                    "filename": "main.tf"
                                                                },
                                                                {
                                                                    "blockName": "module.submodule-call",
                                                                    "filename": "../../main.tf"
                                                                },
                                                                {
                                                                    "blockName": "module.second-module-call",
                                                                    "filename": "../../modules/test-submodule/main.tf"
                                                                },
                                                                {
                                                                    "blockName": "aws_instance.test_second_instance",
                                                                    "filename": "../../modules/second-submodule/main.tf"
                                                                }
                                                            ],
                                                            "filename": "../../modules/second-submodule/main.tf"
                                                        },
                                                        "hourlyCost": "0.0426958904109589",
                                                        "monthlyCost": "31.168",
                                                        "costComponents": [
                                                            {
                                                                "name": "Instance usage (Linux/UNIX, on-demand, t3.medium)",
                                                                "unit": "hours",
                                                                "hourlyQuantity": "1",
                                                                "monthlyQuantity": "730",
                                                                "price": "0.0416",
                                                                "hourlyCost": "0.0416",
                                                                "monthlyCost": "30.368"
                                                            },
                                                            {
                                                                "name": "CPU credits",
                                                                "unit": "vCPU-hours",
                                                                "hourlyQuantity": "0",
                                                                "monthlyQuantity": "0",
                                                                "price": "0.05",
                                                                "hourlyCost": "0",
                                                                "monthlyCost": "0"
                                                            }
                                                        ],
                                                        "subresources": [
                                                            {
                                                                "name": "root_block_device",
                                                                "metadata": {},
                                                                "hourlyCost": "0.0010958904109589",
                                                                "monthlyCost": "0.8",
                                                                "costComponents": [
                                                                    {
                                                                        "name": "Storage (general purpose SSD, gp2)",
                                                                        "unit": "GB",
                                                                        "hourlyQuantity": "0.010958904109589",
                                                                        "monthlyQuantity": "8",
                                                                        "price": "0.1",
                                                                        "hourlyCost": "0.0010958904109589",
                                                                        "monthlyCost": "0.8"
                                                                    }
                                                                ]
                                                            }
                                                        ]
                                                    }
                                                ],
                                                "totalHourlyCost": "0.270958904109589",
                                                "totalMonthlyCost": "197.8"
                                            },
                                            "diff": {
                                                "resources": [
                                                    {
                                                        "name": "module.main_call.aws_s3_bucket.test_bucket",
                                                        "metadata": {},
                                                        "hourlyCost": "0",
                                                        "monthlyCost": "0",
                                                        "subresources": [
                                                            {
                                                                "name": "Standard",
                                                                "metadata": {},
                                                                "hourlyCost": "0",
                                                                "monthlyCost": "0",
                                                                "costComponents": [
                                                                    {
                                                                        "name": "Storage",
                                                                        "unit": "GB",
                                                                        "hourlyQuantity": "0",
                                                                        "monthlyQuantity": "0",
                                                                        "price": "0.023",
                                                                        "hourlyCost": "0",
                                                                        "monthlyCost": "0"
                                                                    },
                                                                    {
                                                                        "name": "PUT, COPY, POST, LIST requests",
                                                                        "unit": "1k requests",
                                                                        "hourlyQuantity": "0",
                                                                        "monthlyQuantity": "0",
                                                                        "price": "0.005",
                                                                        "hourlyCost": "0",
                                                                        "monthlyCost": "0"
                                                                    },
                                                                    {
                                                                        "name": "GET, SELECT, and all other requests",
                                                                        "unit": "1k requests",
                                                                        "hourlyQuantity": "0",
                                                                        "monthlyQuantity": "0",
                                                                        "price": "0.0004",
                                                                        "hourlyCost": "0",
                                                                        "monthlyCost": "0"
                                                                    },
                                                                    {
                                                                        "name": "Select data scanned",
                                                                        "unit": "GB",
                                                                        "hourlyQuantity": "0",
                                                                        "monthlyQuantity": "0",
                                                                        "price": "0.002",
                                                                        "hourlyCost": "0",
                                                                        "monthlyCost": "0"
                                                                    },
                                                                    {
                                                                        "name": "Select data returned",
                                                                        "unit": "GB",
                                                                        "hourlyQuantity": "0",
                                                                        "monthlyQuantity": "0",
                                                                        "price": "0.0007",
                                                                        "hourlyCost": "0",
                                                                        "monthlyCost": "0"
                                                                    }
                                                                ]
                                                            }
                                                        ]
                                                    },
                                                    {
                                                        "name": "module.main_call.module.count_call[0].aws_instance.web",
                                                        "tags": {
                                                            "Name": "HelloWorld"
                                                        },
                                                        "metadata": {},
                                                        "hourlyCost": "0.0114958904109589",
                                                        "monthlyCost": "8.392",
                                                        "costComponents": [
                                                            {
                                                                "name": "Instance usage (Linux/UNIX, on-demand, t3.micro)",
                                                                "unit": "hours",
                                                                "hourlyQuantity": "1",
                                                                "monthlyQuantity": "730",
                                                                "price": "0.0104",
                                                                "hourlyCost": "0.0104",
                                                                "monthlyCost": "7.592"
                                                            },
                                                            {
                                                                "name": "CPU credits",
                                                                "unit": "vCPU-hours",
                                                                "hourlyQuantity": "0",
                                                                "monthlyQuantity": "0",
                                                                "price": "0.05",
                                                                "hourlyCost": "0",
                                                                "monthlyCost": "0"
                                                            }
                                                        ],
                                                        "subresources": [
                                                            {
                                                                "name": "root_block_device",
                                                                "metadata": {},
                                                                "hourlyCost": "0.0010958904109589",
                                                                "monthlyCost": "0.8",
                                                                "costComponents": [
                                                                    {
                                                                        "name": "Storage (general purpose SSD, gp2)",
                                                                        "unit": "GB",
                                                                        "hourlyQuantity": "0.010958904109589",
                                                                        "monthlyQuantity": "8",
                                                                        "price": "0.1",
                                                                        "hourlyCost": "0.0010958904109589",
                                                                        "monthlyCost": "0.8"
                                                                    }
                                                                ]
                                                            }
                                                        ]
                                                    },
                                                    {
                                                        "name": "module.main_call.module.count_call[0].module.second-module-call.aws_instance.test_second_instance",
                                                        "metadata": {},
                                                        "hourlyCost": "0.0426958904109589",
                                                        "monthlyCost": "31.168",
                                                        "costComponents": [
                                                            {
                                                                "name": "Instance usage (Linux/UNIX, on-demand, t3.medium)",
                                                                "unit": "hours",
                                                                "hourlyQuantity": "1",
                                                                "monthlyQuantity": "730",
                                                                "price": "0.0416",
                                                                "hourlyCost": "0.0416",
                                                                "monthlyCost": "30.368"
                                                            },
                                                            {
                                                                "name": "CPU credits",
                                                                "unit": "vCPU-hours",
                                                                "hourlyQuantity": "0",
                                                                "monthlyQuantity": "0",
                                                                "price": "0.05",
                                                                "hourlyCost": "0",
                                                                "monthlyCost": "0"
                                                            }
                                                        ],
                                                        "subresources": [
                                                            {
                                                                "name": "root_block_device",
                                                                "metadata": {},
                                                                "hourlyCost": "0.0010958904109589",
                                                                "monthlyCost": "0.8",
                                                                "costComponents": [
                                                                    {
                                                                        "name": "Storage (general purpose SSD, gp2)",
                                                                        "unit": "GB",
                                                                        "hourlyQuantity": "0.010958904109589",
                                                                        "monthlyQuantity": "8",
                                                                        "price": "0.1",
                                                                        "hourlyCost": "0.0010958904109589",
                                                                        "monthlyCost": "0.8"
                                                                    }
                                                                ]
                                                            }
                                                        ]
                                                    },
                                                    {
                                                        "name": "module.main_call.module.count_call[1].aws_instance.web",
                                                        "tags": {
                                                            "Name": "HelloWorld"
                                                        },
                                                        "metadata": {},
                                                        "hourlyCost": "0.0114958904109589",
                                                        "monthlyCost": "8.392",
                                                        "costComponents": [
                                                            {
                                                                "name": "Instance usage (Linux/UNIX, on-demand, t3.micro)",
                                                                "unit": "hours",
                                                                "hourlyQuantity": "1",
                                                                "monthlyQuantity": "730",
                                                                "price": "0.0104",
                                                                "hourlyCost": "0.0104",
                                                                "monthlyCost": "7.592"
                                                            },
                                                            {
                                                                "name": "CPU credits",
                                                                "unit": "vCPU-hours",
                                                                "hourlyQuantity": "0",
                                                                "monthlyQuantity": "0",
                                                                "price": "0.05",
                                                                "hourlyCost": "0",
                                                                "monthlyCost": "0"
                                                            }
                                                        ],
                                                        "subresources": [
                                                            {
                                                                "name": "root_block_device",
                                                                "metadata": {},
                                                                "hourlyCost": "0.0010958904109589",
                                                                "monthlyCost": "0.8",
                                                                "costComponents": [
                                                                    {
                                                                        "name": "Storage (general purpose SSD, gp2)",
                                                                        "unit": "GB",
                                                                        "hourlyQuantity": "0.010958904109589",
                                                                        "monthlyQuantity": "8",
                                                                        "price": "0.1",
                                                                        "hourlyCost": "0.0010958904109589",
                                                                        "monthlyCost": "0.8"
                                                                    }
                                                                ]
                                                            }
                                                        ]
                                                    },
                                                    {
                                                        "name": "module.main_call.module.count_call[1].module.second-module-call.aws_instance.test_second_instance",
                                                        "metadata": {},
                                                        "hourlyCost": "0.0426958904109589",
                                                        "monthlyCost": "31.168",
                                                        "costComponents": [
                                                            {
                                                                "name": "Instance usage (Linux/UNIX, on-demand, t3.medium)",
                                                                "unit": "hours",
                                                                "hourlyQuantity": "1",
                                                                "monthlyQuantity": "730",
                                                                "price": "0.0416",
                                                                "hourlyCost": "0.0416",
                                                                "monthlyCost": "30.368"
                                                            },
                                                            {
                                                                "name": "CPU credits",
                                                                "unit": "vCPU-hours",
                                                                "hourlyQuantity": "0",
                                                                "monthlyQuantity": "0",
                                                                "price": "0.05",
                                                                "hourlyCost": "0",
                                                                "monthlyCost": "0"
                                                            }
                                                        ],
                                                        "subresources": [
                                                            {
                                                                "name": "root_block_device",
                                                                "metadata": {},
                                                                "hourlyCost": "0.0010958904109589",
                                                                "monthlyCost": "0.8",
                                                                "costComponents": [
                                                                    {
                                                                        "name": "Storage (general purpose SSD, gp2)",
                                                                        "unit": "GB",
                                                                        "hourlyQuantity": "0.010958904109589",
                                                                        "monthlyQuantity": "8",
                                                                        "price": "0.1",
                                                                        "hourlyCost": "0.0010958904109589",
                                                                        "monthlyCost": "0.8"
                                                                    }
                                                                ]
                                                            }
                                                        ]
                                                    },
                                                    {
                                                        "name": "module.main_call.module.for_each_call[\"a-value\"].aws_instance.web",
                                                        "tags": {
                                                            "Name": "HelloWorld"
                                                        },
                                                        "metadata": {},
                                                        "hourlyCost": "0.0114958904109589",
                                                        "monthlyCost": "8.392",
                                                        "costComponents": [
                                                            {
                                                                "name": "Instance usage (Linux/UNIX, on-demand, t3.micro)",
                                                                "unit": "hours",
                                                                "hourlyQuantity": "1",
                                                                "monthlyQuantity": "730",
                                                                "price": "0.0104",
                                                                "hourlyCost": "0.0104",
                                                                "monthlyCost": "7.592"
                                                            },
                                                            {
                                                                "name": "CPU credits",
                                                                "unit": "vCPU-hours",
                                                                "hourlyQuantity": "0",
                                                                "monthlyQuantity": "0",
                                                                "price": "0.05",
                                                                "hourlyCost": "0",
                                                                "monthlyCost": "0"
                                                            }
                                                        ],
                                                        "subresources": [
                                                            {
                                                                "name": "root_block_device",
                                                                "metadata": {},
                                                                "hourlyCost": "0.0010958904109589",
                                                                "monthlyCost": "0.8",
                                                                "costComponents": [
                                                                    {
                                                                        "name": "Storage (general purpose SSD, gp2)",
                                                                        "unit": "GB",
                                                                        "hourlyQuantity": "0.010958904109589",
                                                                        "monthlyQuantity": "8",
                                                                        "price": "0.1",
                                                                        "hourlyCost": "0.0010958904109589",
                                                                        "monthlyCost": "0.8"
                                                                    }
                                                                ]
                                                            }
                                                        ]
                                                    },
                                                    {
                                                        "name": "module.main_call.module.for_each_call[\"a-value\"].module.second-module-call.aws_instance.test_second_instance",
                                                        "metadata": {},
                                                        "hourlyCost": "0.0426958904109589",
                                                        "monthlyCost": "31.168",
                                                        "costComponents": [
                                                            {
                                                                "name": "Instance usage (Linux/UNIX, on-demand, t3.medium)",
                                                                "unit": "hours",
                                                                "hourlyQuantity": "1",
                                                                "monthlyQuantity": "730",
                                                                "price": "0.0416",
                                                                "hourlyCost": "0.0416",
                                                                "monthlyCost": "30.368"
                                                            },
                                                            {
                                                                "name": "CPU credits",
                                                                "unit": "vCPU-hours",
                                                                "hourlyQuantity": "0",
                                                                "monthlyQuantity": "0",
                                                                "price": "0.05",
                                                                "hourlyCost": "0",
                                                                "monthlyCost": "0"
                                                            }
                                                        ],
                                                        "subresources": [
                                                            {
                                                                "name": "root_block_device",
                                                                "metadata": {},
                                                                "hourlyCost": "0.0010958904109589",
                                                                "monthlyCost": "0.8",
                                                                "costComponents": [
                                                                    {
                                                                        "name": "Storage (general purpose SSD, gp2)",
                                                                        "unit": "GB",
                                                                        "hourlyQuantity": "0.010958904109589",
                                                                        "monthlyQuantity": "8",
                                                                        "price": "0.1",
                                                                        "hourlyCost": "0.0010958904109589",
                                                                        "monthlyCost": "0.8"
                                                                    }
                                                                ]
                                                            }
                                                        ]
                                                    },
                                                    {
                                                        "name": "module.main_call.module.for_each_call[\"second-value\"].aws_instance.web",
                                                        "tags": {
                                                            "Name": "HelloWorld"
                                                        },
                                                        "metadata": {},
                                                        "hourlyCost": "0.0114958904109589",
                                                        "monthlyCost": "8.392",
                                                        "costComponents": [
                                                            {
                                                                "name": "Instance usage (Linux/UNIX, on-demand, t3.micro)",
                                                                "unit": "hours",
                                                                "hourlyQuantity": "1",
                                                                "monthlyQuantity": "730",
                                                                "price": "0.0104",
                                                                "hourlyCost": "0.0104",
                                                                "monthlyCost": "7.592"
                                                            },
                                                            {
                                                                "name": "CPU credits",
                                                                "unit": "vCPU-hours",
                                                                "hourlyQuantity": "0",
                                                                "monthlyQuantity": "0",
                                                                "price": "0.05",
                                                                "hourlyCost": "0",
                                                                "monthlyCost": "0"
                                                            }
                                                        ],
                                                        "subresources": [
                                                            {
                                                                "name": "root_block_device",
                                                                "metadata": {},
                                                                "hourlyCost": "0.0010958904109589",
                                                                "monthlyCost": "0.8",
                                                                "costComponents": [
                                                                    {
                                                                        "name": "Storage (general purpose SSD, gp2)",
                                                                        "unit": "GB",
                                                                        "hourlyQuantity": "0.010958904109589",
                                                                        "monthlyQuantity": "8",
                                                                        "price": "0.1",
                                                                        "hourlyCost": "0.0010958904109589",
                                                                        "monthlyCost": "0.8"
                                                                    }
                                                                ]
                                                            }
                                                        ]
                                                    },
                                                    {
                                                        "name": "module.main_call.module.for_each_call[\"second-value\"].module.second-module-call.aws_instance.test_second_instance",
                                                        "metadata": {},
                                                        "hourlyCost": "0.0426958904109589",
                                                        "monthlyCost": "31.168",
                                                        "costComponents": [
                                                            {
                                                                "name": "Instance usage (Linux/UNIX, on-demand, t3.medium)",
                                                                "unit": "hours",
                                                                "hourlyQuantity": "1",
                                                                "monthlyQuantity": "730",
                                                                "price": "0.0416",
                                                                "hourlyCost": "0.0416",
                                                                "monthlyCost": "30.368"
                                                            },
                                                            {
                                                                "name": "CPU credits",
                                                                "unit": "vCPU-hours",
                                                                "hourlyQuantity": "0",
                                                                "monthlyQuantity": "0",
                                                                "price": "0.05",
                                                                "hourlyCost": "0",
                                                                "monthlyCost": "0"
                                                            }
                                                        ],
                                                        "subresources": [
                                                            {
                                                                "name": "root_block_device",
                                                                "metadata": {},
                                                                "hourlyCost": "0.0010958904109589",
                                                                "monthlyCost": "0.8",
                                                                "costComponents": [
                                                                    {
                                                                        "name": "Storage (general purpose SSD, gp2)",
                                                                        "unit": "GB",
                                                                        "hourlyQuantity": "0.010958904109589",
                                                                        "monthlyQuantity": "8",
                                                                        "price": "0.1",
                                                                        "hourlyCost": "0.0010958904109589",
                                                                        "monthlyCost": "0.8"
                                                                    }
                                                                ]
                                                            }
                                                        ]
                                                    },
                                                    {
                                                        "name": "module.main_call.module.submodule-call.aws_instance.web",
                                                        "tags": {
                                                            "Name": "HelloWorld"
                                                        },
                                                        "metadata": {},
                                                        "hourlyCost": "0.0114958904109589",
                                                        "monthlyCost": "8.392",
                                                        "costComponents": [
                                                            {
                                                                "name": "Instance usage (Linux/UNIX, on-demand, t3.micro)",
                                                                "unit": "hours",
                                                                "hourlyQuantity": "1",
                                                                "monthlyQuantity": "730",
                                                                "price": "0.0104",
                                                                "hourlyCost": "0.0104",
                                                                "monthlyCost": "7.592"
                                                            },
                                                            {
                                                                "name": "CPU credits",
                                                                "unit": "vCPU-hours",
                                                                "hourlyQuantity": "0",
                                                                "monthlyQuantity": "0",
                                                                "price": "0.05",
                                                                "hourlyCost": "0",
                                                                "monthlyCost": "0"
                                                            }
                                                        ],
                                                        "subresources": [
                                                            {
                                                                "name": "root_block_device",
                                                                "metadata": {},
                                                                "hourlyCost": "0.0010958904109589",
                                                                "monthlyCost": "0.8",
                                                                "costComponents": [
                                                                    {
                                                                        "name": "Storage (general purpose SSD, gp2)",
                                                                        "unit": "GB",
                                                                        "hourlyQuantity": "0.010958904109589",
                                                                        "monthlyQuantity": "8",
                                                                        "price": "0.1",
                                                                        "hourlyCost": "0.0010958904109589",
                                                                        "monthlyCost": "0.8"
                                                                    }
                                                                ]
                                                            }
                                                        ]
                                                    },
                                                    {
                                                        "name": "module.main_call.module.submodule-call.module.second-module-call.aws_instance.test_second_instance",
                                                        "metadata": {},
                                                        "hourlyCost": "0.0426958904109589",
                                                        "monthlyCost": "31.168",
                                                        "costComponents": [
                                                            {
                                                                "name": "Instance usage (Linux/UNIX, on-demand, t3.medium)",
                                                                "unit": "hours",
                                                                "hourlyQuantity": "1",
                                                                "monthlyQuantity": "730",
                                                                "price": "0.0416",
                                                                "hourlyCost": "0.0416",
                                                                "monthlyCost": "30.368"
                                                            },
                                                            {
                                                                "name": "CPU credits",
                                                                "unit": "vCPU-hours",
                                                                "hourlyQuantity": "0",
                                                                "monthlyQuantity": "0",
                                                                "price": "0.05",
                                                                "hourlyCost": "0",
                                                                "monthlyCost": "0"
                                                            }
                                                        ],
                                                        "subresources": [
                                                            {
                                                                "name": "root_block_device",
                                                                "metadata": {},
                                                                "hourlyCost": "0.0010958904109589",
                                                                "monthlyCost": "0.8",
                                                                "costComponents": [
                                                                    {
                                                                        "name": "Storage (general purpose SSD, gp2)",
                                                                        "unit": "GB",
                                                                        "hourlyQuantity": "0.010958904109589",
                                                                        "monthlyQuantity": "8",
                                                                        "price": "0.1",
                                                                        "hourlyCost": "0.0010958904109589",
                                                                        "monthlyCost": "0.8"
                                                                    }
                                                                ]
                                                            }
                                                        ]
                                                    }
                                                ],
                                                "totalHourlyCost": "0.270958904109589",
                                                "totalMonthlyCost": "197.8"
                                            },
                                            "summary": {
                                                "totalDetectedResources": 12,
                                                "totalSupportedResources": 11,
                                                "totalUnsupportedResources": 1,
                                                "totalUsageBasedResources": 11,
                                                "totalNoPriceResources": 0,
                                                "unsupportedResourceCounts": {
                                                    "aws_s3_object": 1
                                                },
                                                "noPriceResourceCounts": {}
                                            }
                                        }
                                    ],
                                    "totalHourlyCost": "0.270958904109589",
                                    "totalMonthlyCost": "197.8",
                                    "pastTotalHourlyCost": "0",
                                    "pastTotalMonthlyCost": "0",
                                    "diffTotalHourlyCost": "0.270958904109589",
                                    "diffTotalMonthlyCost": "197.8",
                                    "timeGenerated": "2023-01-25T07:25:18.01119688Z",
                                    "summary": {
                                        "totalDetectedResources": 12,
                                        "totalSupportedResources": 11,
                                        "totalUnsupportedResources": 1,
                                        "totalUsageBasedResources": 11,
                                        "totalNoPriceResources": 0,
                                        "unsupportedResourceCounts": {
                                            "aws_s3_object": 1
                                        },
                                        "noPriceResourceCounts": {}
                                    }
                                }
                            ),
                            'terraform_docs': json.dumps({
                                'header': '',
                                'footer': '',
                                'inputs': [
                                    {
                                        'name': 'input_for_example',
                                        'type': 'string',
                                        'description': 'Enter the example name',
                                        'default': None,
                                        'required': True
                                    }
                                ],
                                'modules': [],
                                'outputs': [
                                    {
                                        'name': 'example_output',
                                        'description': 'Example name with randomness'
                                    }
                                ],
                                'providers': [
                                    {
                                        'name': 'example_random',
                                        'alias': None,
                                        'version': None
                                    }
                                ],
                                'requirements': [
                                    {
                                        "name": "terraform",
                                        "version": ">= 1.2.1, <= 2.0.0"
                                    },
                                    {
                                        "name": "someothercompany/example_random",
                                        "version": ">= 4.47"
                                    }
                                ],
                                'resources': [
                                    {
                                        'type': 'string',
                                        'name': 'example_random_suffix',
                                        'provider': 'example_random',
                                        'source': 'hashicorp/example_random',
                                        'mode': 'managed',
                                        'version': 'latest',
                                        'description': None
                                    }
                                ]
                            }),
                            "terraform_graph": """
digraph {
	compound = "true"
	newrank = "true"
	subgraph "root" {
		"[root] module.main_call.aws_s3_bucket.test_bucket (expand)" [label = "module.main_call.aws_s3_bucket.test_bucket", shape = "box"]
		"[root] module.main_call.aws_s3_object.test_obj_root_module (expand)" [label = "module.main_call.aws_s3_object.test_obj_root_module", shape = "box"]
		"[root] module.main_call.module.count_call.aws_instance.web (expand)" [label = "module.main_call.module.count_call.aws_instance.web", shape = "box"]
		"[root] module.main_call.module.count_call.data.aws_ami.ubuntu (expand)" [label = "module.main_call.module.count_call.data.aws_ami.ubuntu", shape = "box"]
		"[root] module.main_call.module.count_call.module.second-module-call.aws_instance.test_second_instance (expand)" [label = "module.main_call.module.count_call.module.second-module-call.aws_instance.test_second_instance", shape = "box"]
		"[root] module.main_call.module.count_call.module.second-module-call.data.aws_ami.ubuntu (expand)" [label = "module.main_call.module.count_call.module.second-module-call.data.aws_ami.ubuntu", shape = "box"]
		"[root] module.main_call.module.for_each_call.aws_instance.web (expand)" [label = "module.main_call.module.for_each_call.aws_instance.web", shape = "box"]
		"[root] module.main_call.module.for_each_call.data.aws_ami.ubuntu (expand)" [label = "module.main_call.module.for_each_call.data.aws_ami.ubuntu", shape = "box"]
		"[root] module.main_call.module.for_each_call.module.second-module-call.aws_instance.test_second_instance (expand)" [label = "module.main_call.module.for_each_call.module.second-module-call.aws_instance.test_second_instance", shape = "box"]
		"[root] module.main_call.module.for_each_call.module.second-module-call.data.aws_ami.ubuntu (expand)" [label = "module.main_call.module.for_each_call.module.second-module-call.data.aws_ami.ubuntu", shape = "box"]
		"[root] module.main_call.module.submodule-call.aws_instance.web (expand)" [label = "module.main_call.module.submodule-call.aws_instance.web", shape = "box"]
		"[root] module.main_call.module.submodule-call.data.aws_ami.ubuntu (expand)" [label = "module.main_call.module.submodule-call.data.aws_ami.ubuntu", shape = "box"]
		"[root] module.main_call.module.submodule-call.module.second-module-call.aws_instance.test_second_instance (expand)" [label = "module.main_call.module.submodule-call.module.second-module-call.aws_instance.test_second_instance", shape = "box"]
		"[root] module.main_call.module.submodule-call.module.second-module-call.data.aws_ami.ubuntu (expand)" [label = "module.main_call.module.submodule-call.module.second-module-call.data.aws_ami.ubuntu", shape = "box"]
		"[root] provider[\\"registry.terraform.io/hashicorp/aws\\"]" [label = "provider[\\"registry.terraform.io/hashicorp/aws\\"]", shape = "diamond"]
		"[root] module.main_call (close)" -> "[root] module.main_call.aws_s3_object.test_obj_root_module (expand)"
		"[root] module.main_call (close)" -> "[root] module.main_call.module.count_call (close)"
		"[root] module.main_call (close)" -> "[root] module.main_call.module.for_each_call (close)"
		"[root] module.main_call (close)" -> "[root] module.main_call.module.submodule-call (close)"
		"[root] module.main_call (close)" -> "[root] module.main_call.output.name (expand)"
		"[root] module.main_call.aws_s3_bucket.test_bucket (expand)" -> "[root] module.main_call.var.name (expand)"
		"[root] module.main_call.aws_s3_bucket.test_bucket (expand)" -> "[root] provider[\\"registry.terraform.io/hashicorp/aws\\"]"
		"[root] module.main_call.aws_s3_object.test_obj_root_module (expand)" -> "[root] module.main_call.aws_s3_bucket.test_bucket (expand)"
		"[root] module.main_call.module.count_call (close)" -> "[root] module.main_call.module.count_call.aws_instance.web (expand)"
		"[root] module.main_call.module.count_call (close)" -> "[root] module.main_call.module.count_call.module.second-module-call (close)"
		"[root] module.main_call.module.count_call (close)" -> "[root] module.main_call.module.count_call.var.passing_name (expand)"
		"[root] module.main_call.module.count_call (expand)" -> "[root] module.main_call (expand)"
		"[root] module.main_call.module.count_call.aws_instance.web (expand)" -> "[root] module.main_call.module.count_call.data.aws_ami.ubuntu (expand)"
		"[root] module.main_call.module.count_call.data.aws_ami.ubuntu (expand)" -> "[root] module.main_call.module.count_call (expand)"
		"[root] module.main_call.module.count_call.data.aws_ami.ubuntu (expand)" -> "[root] provider[\\"registry.terraform.io/hashicorp/aws\\"]"
		"[root] module.main_call.module.count_call.module.second-module-call (close)" -> "[root] module.main_call.module.count_call.module.second-module-call.aws_instance.test_second_instance (expand)"
		"[root] module.main_call.module.count_call.module.second-module-call (expand)" -> "[root] module.main_call.module.count_call (expand)"
		"[root] module.main_call.module.count_call.module.second-module-call.aws_instance.test_second_instance (expand)" -> "[root] module.main_call.module.count_call.module.second-module-call.data.aws_ami.ubuntu (expand)"
		"[root] module.main_call.module.count_call.module.second-module-call.data.aws_ami.ubuntu (expand)" -> "[root] module.main_call.module.count_call.module.second-module-call (expand)"
		"[root] module.main_call.module.count_call.module.second-module-call.data.aws_ami.ubuntu (expand)" -> "[root] provider[\\"registry.terraform.io/hashicorp/aws\\"]"
		"[root] module.main_call.module.count_call.var.passing_name (expand)" -> "[root] module.main_call.module.count_call (expand)"
		"[root] module.main_call.module.for_each_call (close)" -> "[root] module.main_call.module.for_each_call.aws_instance.web (expand)"
		"[root] module.main_call.module.for_each_call (close)" -> "[root] module.main_call.module.for_each_call.module.second-module-call (close)"
		"[root] module.main_call.module.for_each_call (close)" -> "[root] module.main_call.module.for_each_call.var.passing_name (expand)"
		"[root] module.main_call.module.for_each_call (expand)" -> "[root] module.main_call (expand)"
		"[root] module.main_call.module.for_each_call.aws_instance.web (expand)" -> "[root] module.main_call.module.for_each_call.data.aws_ami.ubuntu (expand)"
		"[root] module.main_call.module.for_each_call.data.aws_ami.ubuntu (expand)" -> "[root] module.main_call.module.for_each_call (expand)"
		"[root] module.main_call.module.for_each_call.data.aws_ami.ubuntu (expand)" -> "[root] provider[\\"registry.terraform.io/hashicorp/aws\\"]"
		"[root] module.main_call.module.for_each_call.module.second-module-call (close)" -> "[root] module.main_call.module.for_each_call.module.second-module-call.aws_instance.test_second_instance (expand)"
		"[root] module.main_call.module.for_each_call.module.second-module-call (expand)" -> "[root] module.main_call.module.for_each_call (expand)"
		"[root] module.main_call.module.for_each_call.module.second-module-call.aws_instance.test_second_instance (expand)" -> "[root] module.main_call.module.for_each_call.module.second-module-call.data.aws_ami.ubuntu (expand)"
		"[root] module.main_call.module.for_each_call.module.second-module-call.data.aws_ami.ubuntu (expand)" -> "[root] module.main_call.module.for_each_call.module.second-module-call (expand)"
		"[root] module.main_call.module.for_each_call.module.second-module-call.data.aws_ami.ubuntu (expand)" -> "[root] provider[\\"registry.terraform.io/hashicorp/aws\\"]"
		"[root] module.main_call.module.for_each_call.var.passing_name (expand)" -> "[root] module.main_call.module.for_each_call (expand)"
		"[root] module.main_call.module.submodule-call (close)" -> "[root] module.main_call.module.submodule-call.aws_instance.web (expand)"
		"[root] module.main_call.module.submodule-call (close)" -> "[root] module.main_call.module.submodule-call.module.second-module-call (close)"
		"[root] module.main_call.module.submodule-call (close)" -> "[root] module.main_call.module.submodule-call.var.passing_name (expand)"
		"[root] module.main_call.module.submodule-call (expand)" -> "[root] module.main_call (expand)"
		"[root] module.main_call.module.submodule-call.aws_instance.web (expand)" -> "[root] module.main_call.module.submodule-call.data.aws_ami.ubuntu (expand)"
		"[root] module.main_call.module.submodule-call.data.aws_ami.ubuntu (expand)" -> "[root] module.main_call.module.submodule-call (expand)"
		"[root] module.main_call.module.submodule-call.data.aws_ami.ubuntu (expand)" -> "[root] provider[\\"registry.terraform.io/hashicorp/aws\\"]"
		"[root] module.main_call.module.submodule-call.module.second-module-call (close)" -> "[root] module.main_call.module.submodule-call.module.second-module-call.aws_instance.test_second_instance (expand)"
		"[root] module.main_call.module.submodule-call.module.second-module-call (expand)" -> "[root] module.main_call.module.submodule-call (expand)"
		"[root] module.main_call.module.submodule-call.module.second-module-call.aws_instance.test_second_instance (expand)" -> "[root] module.main_call.module.submodule-call.module.second-module-call.data.aws_ami.ubuntu (expand)"
		"[root] module.main_call.module.submodule-call.module.second-module-call.data.aws_ami.ubuntu (expand)" -> "[root] module.main_call.module.submodule-call.module.second-module-call (expand)"
		"[root] module.main_call.module.submodule-call.module.second-module-call.data.aws_ami.ubuntu (expand)" -> "[root] provider[\\"registry.terraform.io/hashicorp/aws\\"]"
		"[root] module.main_call.module.submodule-call.var.passing_name (expand)" -> "[root] module.main_call.module.submodule-call (expand)"
		"[root] module.main_call.module.submodule-call.var.passing_name (expand)" -> "[root] module.main_call.var.name (expand)"
		"[root] module.main_call.output.name (expand)" -> "[root] module.main_call.aws_s3_bucket.test_bucket (expand)"
		"[root] module.main_call.var.name (expand)" -> "[root] module.main_call (expand)"
		"[root] provider[\\"registry.terraform.io/hashicorp/aws\\"] (close)" -> "[root] module.main_call.aws_s3_object.test_obj_root_module (expand)"
		"[root] provider[\\"registry.terraform.io/hashicorp/aws\\"] (close)" -> "[root] module.main_call.module.count_call.aws_instance.web (expand)"
		"[root] provider[\\"registry.terraform.io/hashicorp/aws\\"] (close)" -> "[root] module.main_call.module.count_call.module.second-module-call.aws_instance.test_second_instance (expand)"
		"[root] provider[\\"registry.terraform.io/hashicorp/aws\\"] (close)" -> "[root] module.main_call.module.for_each_call.aws_instance.web (expand)"
		"[root] provider[\\"registry.terraform.io/hashicorp/aws\\"] (close)" -> "[root] module.main_call.module.for_each_call.module.second-module-call.aws_instance.test_second_instance (expand)"
		"[root] provider[\\"registry.terraform.io/hashicorp/aws\\"] (close)" -> "[root] module.main_call.module.submodule-call.aws_instance.web (expand)"
		"[root] provider[\\"registry.terraform.io/hashicorp/aws\\"] (close)" -> "[root] module.main_call.module.submodule-call.module.second-module-call.aws_instance.test_second_instance (expand)"
		"[root] root" -> "[root] module.main_call (close)"
		"[root] root" -> "[root] provider[\\"registry.terraform.io/hashicorp/aws\\"] (close)"
	}
}


"""
                        }
                    },
                    'submodules': {
                        'modules/example-submodule1': {
                            'readme_content': '# Submodule 1 README',
                            'terraform_docs': json.dumps({
                                'header': '',
                                'footer': '',
                                'inputs': [
                                    {
                                        'name': 'input_for_submodule',
                                        'type': 'string',
                                        'description': 'Enter the submodule name',
                                        'default': None,
                                        'required': True
                                    }
                                ],
                                'modules': [],
                                'outputs': [
                                    {
                                        'name': 'submodule_output',
                                        'description': 'Submodule name with randomness'
                                    }
                                ],
                                'providers': [
                                    {
                                        'name': 'submodule_random',
                                        'alias': None,
                                        'version': None
                                    }
                                ],
                                'requirements': [
                                    {
                                        "name": "terraform",
                                        "version": ">= 2.0.0"
                                    },
                                    {
                                        "name": "someothercompany/submodule_random",
                                        "version": ">= 4.49"
                                    }
                                ],
                                'resources': [
                                    {
                                        'type': 'string',
                                        'name': 'submodule_random_suffix',
                                        'provider': 'submodule_random',
                                        'source': 'hashicorp/submodule_random',
                                        'mode': 'managed',
                                        'version': 'latest',
                                        'description': None
                                    }
                                ]
                            }),
                            "terraform_graph": """
digraph {
	compound = "true"
	newrank = "true"
	subgraph "root" {
		"[root] aws_instance.web (expand)" [label = "aws_instance.web", shape = "box"]
		"[root] data.aws_ami.ubuntu (expand)" [label = "data.aws_ami.ubuntu", shape = "box"]
		"[root] module.second-module-call.aws_instance.test_second_instance (expand)" [label = "module.second-module-call.aws_instance.test_second_instance", shape = "box"]
		"[root] module.second-module-call.data.aws_ami.ubuntu (expand)" [label = "module.second-module-call.data.aws_ami.ubuntu", shape = "box"]
		"[root] provider[\\"registry.terraform.io/hashicorp/aws\\"]" [label = "provider[\\"registry.terraform.io/hashicorp/aws\\"]", shape = "diamond"]
		"[root] var.passing_name" [label = "var.passing_name", shape = "note"]
		"[root] aws_instance.web (expand)" -> "[root] data.aws_ami.ubuntu (expand)"
		"[root] data.aws_ami.ubuntu (expand)" -> "[root] provider[\\"registry.terraform.io/hashicorp/aws\\"]"
		"[root] module.second-module-call (close)" -> "[root] module.second-module-call.aws_instance.test_second_instance (expand)"
		"[root] module.second-module-call.aws_instance.test_second_instance (expand)" -> "[root] module.second-module-call.data.aws_ami.ubuntu (expand)"
		"[root] module.second-module-call.data.aws_ami.ubuntu (expand)" -> "[root] module.second-module-call (expand)"
		"[root] module.second-module-call.data.aws_ami.ubuntu (expand)" -> "[root] provider[\\"registry.terraform.io/hashicorp/aws\\"]"
		"[root] provider[\\"registry.terraform.io/hashicorp/aws\\"] (close)" -> "[root] aws_instance.web (expand)"
		"[root] provider[\\"registry.terraform.io/hashicorp/aws\\"] (close)" -> "[root] module.second-module-call.aws_instance.test_second_instance (expand)"
		"[root] root" -> "[root] module.second-module-call (close)"
		"[root] root" -> "[root] provider[\\"registry.terraform.io/hashicorp/aws\\\"] (close)"
		"[root] root" -> "[root] var.passing_name"
	}
}

"""

                        }
                    }
                }
            }
        }},
        'infracost': {'testprovider': {
            'id': 68,
            'versions': {'1.0.0': {
                'published': True,
                'examples': {
                    'examples/with-cost': {
                        'infracost': json.dumps({
                            'totalMonthlyCost': '12.5123',
                        })
                    },
                    'examples/free': {
                        'infracost': json.dumps({
                            'totalMonthlyCost': '0.000',
                        })
                    },
                    'examples/no-infracost-data': {
                        'infracost': None
                    }
                }
            }}
        }},
        'withsecurityissues': {'testprovider': {
            'id': 62,
            'versions': {
                '1.2.0': {
                    'published': True,
                    'submodules': {
                        'modules/withanotherissue': {
                            'tfsec': json.dumps({
                                'results': [
                                    {
                                        'description': 'First security issue.',
                                        'impact': 'First security issue is Medium',
                                        'links': [
                                            'https://example.com/first-issue/'
                                        ],
                                        'location': {
                                            'end_line': 2,
                                            'filename': 'first.tf',
                                            'start_line': 1
                                        },
                                        'long_id': 'first-security-medium-issue',
                                        'resolution': 'Remove first security issue',
                                        'resource': 'aws_s3.first',
                                        'rule_description': 'This type of first issue is Medium',
                                        'rule_id': 'AVD-TRG-001',
                                        'rule_provider': 'aws',
                                        'rule_service': 's3',
                                        'severity': 'MEDIUM',
                                        'status': 0,
                                        'warning': False
                                    },
                                ]
                            })
                        }
                    }
                },
                '1.1.0': {
                    'published': True,
                    'examples': {
                        'examples/withsecissue': {
                            'tfsec': json.dumps({
                                'results': [
                                    {
                                        'description': 'First security issue.',
                                        'impact': 'First security issue is Low',
                                        'links': [
                                            'https://example.com/first-issue/'
                                        ],
                                        'location': {
                                            'end_line': 2,
                                            'filename': 'first.tf',
                                            'start_line': 1
                                        },
                                        'long_id': 'first-security-low-issue',
                                        'resolution': 'Remove first security issue',
                                        'resource': 'aws_s3.first',
                                        'rule_description': 'This type of first issue is Low',
                                        'rule_id': 'AVD-TRG-001',
                                        'rule_provider': 'aws',
                                        'rule_service': 's3',
                                        'severity': 'LOW',
                                        'status': 0,
                                        'warning': False
                                    },
                                    {
                                        'description': 'Second security issue.',
                                        'impact': 'Second security issue is high',
                                        'links': [
                                            'https://example.com/second-issue/'
                                        ],
                                        'location': {
                                            'end_line': 8,
                                            'filename': 'second.tf',
                                            'start_line': 5
                                        },
                                        'long_id': 'second-security-high-issue',
                                        'resolution': 'Remove second security issue',
                                        'resource': 'aws_s3.second',
                                        'rule_description': 'This type of second issue is High',
                                        'rule_id': 'AVD-TRG-002',
                                        'rule_provider': 'aws',
                                        'rule_service': 's3',
                                        'severity': 'HIGH',
                                        'status': 0,
                                        'warning': False
                                    },
                                    {
                                        'description': 'Third security issue.',
                                        'impact': 'Third security issue is medium',
                                        'links': [
                                            'https://example.com/third-issue/',
                                            'https://example.com/third-issue/link2',
                                        ],
                                        'location': {
                                            'end_line': 4,
                                            'filename': 'third.tf',
                                            'start_line': 2
                                        },
                                        'long_id': 'third-security-medium-issue',
                                        'resolution': 'Remove third security issue',
                                        'resource': 'aws_s3.third',
                                        'rule_description': 'This type of third issue is Medium',
                                        'rule_id': 'AVD-TRG-003',
                                        'rule_provider': 'aws',
                                        'rule_service': 's3',
                                        'severity': 'MEDIUM',
                                        'status': 0,
                                        'warning': False
                                    }
                                ]
                            })
                        }
                    }
                },
                '1.0.0': {
                    'published': True,
                    'tfsec': json.dumps({
                        'results': [
                            {
                                'description': 'Secret explicitly uses the default key.',
                                'impact': 'Using AWS managed keys reduces the flexibility and '
                                        'control over the encryption key',
                                'links': [
                                    'https://aquasecurity.github.io/tfsec/v1.26.0/checks/aws/ssm/secret-use-customer-key/',
                                    'https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/secretsmanager_secret#kms_key_id'
                                ],
                                'location': {
                                    'end_line': 4,
                                    'filename': 'main.tf',
                                    'start_line': 2
                                },
                                'long_id': 'aws-ssm-secret-use-customer-key',
                                'resolution': 'Use customer managed keys',
                                'resource': 'aws_secretsmanager_secret.this',
                                'rule_description': 'Secrets Manager should use customer managed '
                                                    'keys',
                                'rule_id': 'AVD-AWS-0098',
                                'rule_provider': 'aws',
                                'rule_service': 'ssm',
                                'severity': 'LOW',
                                'status': 0,
                                'warning': False
                            },
                            {
                                'description': 'Some security issue 2.',
                                'impact': 'Entire project is compromised',
                                'links': [
                                    'https://example.com/issuehere',
                                    'https://example.com/docshere'
                                ],
                                'location': {
                                    'end_line': 1,
                                    'filename': 'main.tf',
                                    'start_line': 6
                                },
                                'long_id': 'dodgy-bad-is-bad',
                                'resolution': 'Do not use bad code',
                                'resource': 'some_data_resource.this',
                                'rule_description': 'Dodgy code should be removed',
                                'rule_id': 'DDG-ANC-001',
                                'rule_provider': 'bad',
                                'rule_service': 'code',
                                'severity': 'HIGH',
                                'status': 0,
                                'warning': False
                            },
                            {
                                'description': 'Some security issue 3.',
                                'impact': 'Entire project is compromised',
                                'links': [
                                    'https://example.com/issuehere',
                                    'https://example.com/docshere'
                                ],
                                'location': {
                                    'end_line': 1,
                                    'filename': 'different.tf',
                                    'start_line': 6
                                },
                                'long_id': 'dodgy-bad-is-bad',
                                'resolution': 'Do not use bad code',
                                'resource': 'some_data_resource.this',
                                'rule_description': 'Dodgy code should be removed',
                                'rule_id': 'DDG-ANC-003',
                                'rule_provider': 'bad',
                                'rule_service': 'code',
                                'severity': 'HIGH',
                                'status': 0,
                                'warning': False
                            },
                            {
                                'description': 'Second high issue.',
                                'impact': 'Entire project is compromised',
                                'links': [
                                    'https://example.com/issuehere',
                                    'https://example.com/docshere'
                                ],
                                'location': {
                                    'end_line': 1,
                                    'filename': 'main.tf',
                                    'start_line': 6
                                },
                                'long_id': 'dodgy-bad-is-bad',
                                'resolution': 'Do not use bad code',
                                'resource': 'some_data_resource.this',
                                'rule_description': 'Dodgy code should be removed',
                                'rule_id': 'DDG-ANC-004',
                                'rule_provider': 'bad',
                                'rule_service': 'code',
                                'severity': 'HIGH',
                                'status': 0,
                                'warning': False
                            },
                            {
                                'description': 'Some security issue 4.',
                                'impact': 'Entire project is compromised',
                                'links': [
                                    'https://example.com/issuehere',
                                    'https://example.com/docshere'
                                ],
                                'location': {
                                    'end_line': 1,
                                    'filename': 'itsfine.tf',
                                    'start_line': 6
                                },
                                'long_id': 'dodgy-bad-is-fine',
                                'resolution': 'Do not use bad code',
                                'resource': 'some_data_resource.this',
                                'rule_description': 'Dodgy code should be removed',
                                'rule_id': 'DDG-ANC-005',
                                'rule_provider': 'bad',
                                'rule_service': 'code',
                                'severity': 'HIGH',
                                'status': 1,
                                'warning': False
                            },
                            {
                                'description': 'Some security issue 5.',
                                'impact': 'Entire project is compromised',
                                'links': [
                                    'https://example.com/issuehere',
                                    'https://example.com/docshere'
                                ],
                                'location': {
                                    'end_line': 1,
                                    'filename': 'ignored.tf',
                                    'start_line': 6
                                },
                                'long_id': 'dodgy-bad-is-ignored',
                                'resolution': 'Do not use bad code',
                                'resource': 'some_data_resource.this',
                                'rule_description': 'Dodgy code should be removed',
                                'rule_id': 'DDG-ANC-006',
                                'rule_provider': 'bad',
                                'rule_service': 'code',
                                'severity': 'HIGH',
                                'status': 2,
                                'warning': False
                            },
                            {
                                'description': 'Some medium issue 6.',
                                'impact': 'This is quite important',
                                'links': [
                                    'https://example.com/issuehere',
                                    'https://example.com/docshere'
                                ],
                                'location': {
                                    'end_line': 1,
                                    'filename': 'ignored.tf',
                                    'start_line': 6
                                },
                                'long_id': 'dodgy-bad-is-important',
                                'resolution': 'Do not use bad code',
                                'resource': 'some_data_resource.this',
                                'rule_description': 'Dodgy code should be removed',
                                'rule_id': 'DDG-ANC-006',
                                'rule_provider': 'bad',
                                'rule_service': 'code',
                                'severity': 'MEDIUM',
                                'status': 0,
                                'warning': False
                            },
                            {
                                'description': 'Some critical issue 7.',
                                'impact': 'This is critical',
                                'links': [
                                    'https://example.com/issuehere',
                                    'https://example.com/docshere'
                                ],
                                'location': {
                                    'end_line': 1,
                                    'filename': 'ignored.tf',
                                    'start_line': 6
                                },
                                'long_id': 'dodgy-bad-is-critical',
                                'resolution': 'Fix critical issue',
                                'resource': 'some_data_resource.this',
                                'rule_description': 'Critical code has an issue',
                                'rule_id': 'DDG-ANC-007',
                                'rule_provider': 'bad',
                                'rule_service': 'code',
                                'severity': 'CRITICAL',
                                'status': 0,
                                'warning': False
                            }
                        ]
                    })
                }
            }
        }},
        'noversion': {'testprovider': {
            'id': 57,
            'versions': {}
        }}
    },

    # Small namespace with module providers
    # with unpublished and beta versions
    'unpublished-beta-version-module-providers': {
        'publishedone': {
            'testprovider': {
                'id': 63,
                'versions': {
                    '2.1.1': {
                        'published': True,
                        'description': 'Test module description for testprovider'
                    }
                }
            },
            'secondprovider': {
                'id': 64,
                'versions': {
                    '2.2.2': {
                        'published': True,
                        'description': 'Description of second provider in module'
                    }
                }
            }
        },
        'noversions': {'testprovider': {
            'id': 67
        }},
        'onlybeta': {'testprovider': {
            'id': 65,
            'versions': {
                '2.2.4-beta': {
                    'published': True,
                    'beta': True,
                    'description': 'Test description'
                }
            }
        }},
        'onlyunpublished': {'testprovider': {
            'id': 66,
            'versions': {
                '1.0.0': {
                    'published': False,
                    'description': 'Test description'
                }
            }
        }}
    },

    'javascriptinjection': {
        'modulename': {'testprovider': {
            'id': 58,
            'versions': {
                '1.5.0': {
                    'description': '<script>var a = document.createElement("div"); a.id = "injectedDescription"; document.body.appendChild(a);</script>',
                    'owner': '<script>var a = document.createElement("div"); a.id = "injectedOwner"; document.body.appendChild(a);</script>',
                    'published': True,
                    'beta': False,
                    'internal': False,
                    'published_at': datetime(2022, 1, 5, 22, 53, 12),
                    'readme_content': '# This is an exaple README!<br /><script>var a = document.createElement("div"); a.id = "injectedReadme"; document.body.appendChild(a);</script>',
                    'variable_template': json.dumps([
                        {
                            'name': '<script>var a = document.createElement("div"); a.id = "injectedVariableTemplateName"; document.body.appendChild(a);</script>',
                            'type': '<script>var a = document.createElement("div"); a.id = "injectedVariableTemplateType"; document.body.appendChild(a);</script>',
                            'quote_value': True,
                            'additional_help': '<script>var a = document.createElement("div"); a.id = "injectedVariableAdditionalHelp"; document.body.appendChild(a);</script>'
                        }

                    ]),
                    'terraform_docs': json.dumps({
                        'header': '',
                        'footer': '',
                        'inputs': [
                            {
                                'name': '<script>var a = document.createElement("div"); a.id = "injectedTerraformDocsInputName"; document.body.appendChild(a);</script>',
                                'type': '<script>var a = document.createElement("div"); a.id = "injectedTerraformDocsInputType"; document.body.appendChild(a);</script>',
                                'description': '<script>var a = document.createElement("div"); a.id = "injectedTerraformDocsInputDescription"; document.body.appendChild(a);</script>',
                                'default': '<script>var a = document.createElement("div"); a.id = "injectedTerraformDocsInputDefault"; document.body.appendChild(a);</script>',
                                'required': True
                            }
                        ],
                        'modules': [],
                        'outputs': [
                            {
                                'name': '<script>var a = document.createElement("div"); a.id = "injectedTerraformDocsOutputName"; document.body.appendChild(a);</script>',
                                'description': '<script>var a = document.createElement("div"); a.id = "injectedTerraformDocsOutputDescription"; document.body.appendChild(a);</script>'
                            }
                        ],
                        'providers': [
                            {
                                'name': '<script>var a = document.createElement("div"); a.id = "injectedTerraformProviderName"; document.body.appendChild(a);</script>',
                                'alias': '<script>var a = document.createElement("div"); a.id = "injectedTerraformDocsProviderAlias"; document.body.appendChild(a);</script>',
                                'version': '<script>var a = document.createElement("div"); a.id = "injectedTerraformDocsProviderVersion"; document.body.appendChild(a);</script>'
                            }
                        ],
                        'requirements': [],
                        'resources': [
                            {
                                'type': '<script>var a = document.createElement("div"); a.id = "injectedTerraformDocsResourceType"; document.body.appendChild(a);</script>',
                                'name': '<script>var a = document.createElement("div"); a.id = "injectedTerraformDocsResourceName"; document.body.appendChild(a);</script>',
                                'provider': '<script>var a = document.createElement("div"); a.id = "injectedTerraformDocsResourceProvider"; document.body.appendChild(a);</script>',
                                'source': '<script>var a = document.createElement("div"); a.id = "injectedTerraformDocsResourceSource"; document.body.appendChild(a);</script>',
                                'mode': '<script>var a = document.createElement("div"); a.id = "injectedTerraformDocsResourceMode"; document.body.appendChild(a);</script>',
                                'version': '<script>var a = document.createElement("div"); a.id = "injectedTerraformDocsResourceVersion"; document.body.appendChild(a);</script>',
                                'description': '<script>var a = document.createElement("div"); a.id = "injectedTerraformDocsResourceDescription"; document.body.appendChild(a);</script>'
                            }
                        ]
                    }),
                    'files': {
                        'LICENSE': '<script>var a = document.createElement("div"); a.id = "injectedAdditionalFilesPlainText"; document.body.appendChild(a);</script>',
                        'CHANGELOG.md': '<script>var a = document.createElement("div"); a.id = "injectedAdditionalMarkDown"; document.body.appendChild(a);</script>'
                    },
                    'examples': {
                        'examples/test-example': {
                            'example_files': {
                                'examples/test-example/data.tf': '<script>var a = document.createElement("div"); a.id = "injectedExampleFileContent"; document.body.appendChild(a);</script>',
                            },
                            'readme_content': '# Example 1 README<script>var a = document.createElement("div"); a.id = "injectedExampleReadme"; document.body.appendChild(a);</script>',
                            'infracost': None
                        },
                        'examples/heredoc-tags': {
                            'example_files': {
                                'examples/heredoc-tags/main.tf': """
module "test" {
  input = <<EOF
Test heredoc content
EOF
}
""",
                            },
                            'infracost': None
                        }
                    },
                    'submodules': {
                        'modules/example-submodule1': {
                            'readme_content': '# Submodule 1 README\n<script>var a = document.createElement("div"); a.id = "injectedSubemoduleFileContent"; document.body.appendChild(a);</script>'
                        }
                    }
                },
            }
        }},
    },

    ## THESE MUST BE AT THE BOTTOM
    'mostrecent': {
        'modulename': {'providername': {
            'id': 48,
            'versions': {'1.2.3': {'published': True}}
        }}
    },
    'mostrecentunpublished': {
        'modulename': {'providername': {
            'id': 53,
            'versions': {
                '1.2.3': {'published': False},
                '1.5.3-beta': {'published': True, 'beta': True}
            }
        }}
    },
    'mostrecentextractionincomplete': {
        'modulename': {'providername': {
            'id': 83,
            'versions': {
                # Test with module version somehow getting published True, with extraction complete as False
                '1.2.1': {'published': True, 'extraction_complete': False},

                '1.2.3': {'published': False, 'extraction_complete': False},
                '1.5.3-beta': {'published': True, 'beta': True, 'extraction_complete': False}
            }
        }}
    },
    'testmodulecreation': {},
    'emptynamespace': {}
}

two_empty_namespaces = {
    'firstnamespace': {
    },
    'second-namespace': {
    }
}

selenium_user_group_data = {
    'nopermissions': {
    },
    'siteadmin': {
        'site_admin': True
    },
    'moduledetailsmodify': {
        'namespace_permissions': {
            'moduledetails': 'MODIFY'
        }
    },
    'moduledetailsfull': {
        'namespace_permissions': {
            'moduledetails': 'FULL'
        }
    },
    'multiplenamespaces': {
        'namespace_permissions': {
            'moduledetails': 'FULL',
            'trustednamespace': 'FULL',
            'testnamespace': 'MODIFY',
            'moduleextraction': 'MODIFY'
        }
    }
}
