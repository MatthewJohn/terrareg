
from datetime import datetime
import json


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
        'id': 1,
        'modules': {
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
                'latest_verison': None,
                'repo_base_url_template': 'https://custom-localhost.com/{namespace}/{module}-{provider}',
                'repo_browse_url_template': 'https://custom-localhost.com/{namespace}/{module}-{provider}/browse/{tag}/{path}',
                'repo_clone_url_template': 'ssh://custom-localhost.com/{namespace}/{module}-{provider}',
                'versions': {
                    '10.2.1': {'published': False}
                }
            }},
            'withsecurityissues': {'testprovider': {
                'id': 20,
                'latest_version': '1.0.0',
                'versions': {
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
                                }
                            ]
                        })
                    }
                }
            }}
        },
    },
    'moduledetails': {
        'id': 2,
        'modules': {
            'fullypopulated': {'testprovider': {
                'id': 26,
                'repo_base_url_template': 'https://mp-base-url.com/{namespace}/{module}-{provider}',
                'repo_browse_url_template': 'https://mp-browse-url.com/{namespace}/{module}-{provider}/browse/{tag}/{path}suffix',
                'repo_clone_url_template': 'ssh://mp-clone-url.com/{namespace}/{module}-{provider}',
                'versions': {
                    # Older version
                    '1.2.0': {'published': True},
                    # Newer unpublished version
                    '1.6.0': {},
                    # Newer published beta version
                    '1.6.1-beta': {'published': True, 'beta': True},
                    # Unpublished and beta version
                    '1.0.0-beta': {'published': False, 'beta': True},
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
                        }),
                        'files': {
                            'LICENSE': 'This is a license file\nAll rights are not reserved for this example file content',
                            'CHANGELOG.md': '# Changelog\n## 1.0.0\n * This is an initial release',
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
                                'infracost': json.dumps({
                                    'totalMonthlyCost': '61.536',
                                    'totalHourlyCost': '0.0842958904109589',
                                    'timeGenerated': '2022-08-17T18:39:55.964808023Z',
                                    'currency': 'USD',
                                    'diffTotalHourlyCost':
                                    '0.0842958904109589',
                                    'version': '0.2',
                                    'pastTotalHourlyCost': '0',
                                    'pastTotalMonthlyCost': '0',
                                    'diffTotalMonthlyCost': '61.536',
                                    'summary': {
                                        'totalNoPriceResources': 0,
                                        'unsupportedResourceCounts': {},
                                        'totalUsageBasedResources': 1,
                                        'totalUnsupportedResources': 0,
                                        'totalDetectedResources': 1,
                                        'totalSupportedResources': 1,
                                        'noPriceResourceCounts': {}
                                    },
                                    'projects': [
                                        {
                                            'pastBreakdown': {
                                                'totalMonthlyCost': '0',
                                                'totalHourlyCost': '0',
                                                'resources': []
                                            },
                                            'breakdown': {
                                                'totalMonthlyCost': '61.536',
                                                'totalHourlyCost': '0.0842958904109589',
                                                'resources': [
                                                    {
                                                        'hourlyCost': '0.0842958904109589',
                                                        'name': 'aws_instance.test',
                                                        'monthlyCost': '61.536',
                                                        'costComponents': [
                                                            {
                                                                'hourlyCost': '0.0832',
                                                                'name': 'Instance usage (Linux/UNIX, on-demand, t3.large)',
                                                                'hourlyQuantity': '1',
                                                                'price': '0.0832',
                                                                'monthlyCost': '60.736',
                                                                'monthlyQuantity': '730',
                                                                'unit': 'hours'
                                                            },
                                                            {
                                                                'hourlyCost': '0',
                                                                'name': 'CPU credits',
                                                                'hourlyQuantity': '0',
                                                                'price': '0.05',
                                                                'monthlyCost': '0',
                                                                'monthlyQuantity': '0',
                                                                'unit': 'vCPU-hours'
                                                            }
                                                        ],
                                                        'subresources': [
                                                            {
                                                                'costComponents': [
                                                                    {
                                                                        'hourlyCost': '0.0010958904109589',
                                                                        'name': 'Storage (general purpose SSD, gp2)',
                                                                        'hourlyQuantity': '0.010958904109589',
                                                                        'price': '0.1',
                                                                        'monthlyCost': '0.8',
                                                                        'monthlyQuantity': '8',
                                                                        'unit': 'GB'
                                                                    }
                                                                ],
                                                                'hourlyCost': '0.0010958904109589',
                                                                'monthlyCost': '0.8',
                                                                'name': 'root_block_device',
                                                                'metadata': {}
                                                            }
                                                        ],
                                                        'metadata': {
                                                            'calls': [
                                                                {
                                                                    'filename': 'main.tf',
                                                                    'blockName': 'aws_instance.test'
                                                                }
                                                            ],
                                                            'filename': 'main.tf'
                                                        }
                                                    }
                                                ]
                                            },
                                            'name': '2222/pub/terrareg/example/cost_example',
                                            'summary': {
                                                'totalNoPriceResources': 0,
                                                'unsupportedResourceCounts': {},
                                                'totalUsageBasedResources': 1,
                                                'totalUnsupportedResources': 0,
                                                'totalDetectedResources': 1,
                                                'totalSupportedResources': 1,
                                                'noPriceResourceCounts': {}
                                            },
                                            'diff': {
                                                'totalMonthlyCost': '61.536',
                                                'totalHourlyCost': '0.0842958904109589',
                                                'resources': [
                                                    {
                                                        'hourlyCost': '0.0842958904109589',
                                                        'name': 'aws_instance.test',
                                                        'monthlyCost': '61.536',
                                                        'costComponents': [
                                                            {
                                                                'hourlyCost': '0.0832',
                                                                'name': 'Instance usage (Linux/UNIX, on-demand, t3.large)',
                                                                'hourlyQuantity': '1',
                                                                'price': '0.0832',
                                                                'monthlyCost': '60.736',
                                                                'monthlyQuantity': '730',
                                                                'unit': 'hours'
                                                            },
                                                            {
                                                                'hourlyCost': '0',
                                                                'name': 'CPU credits',
                                                                'hourlyQuantity': '0',
                                                                'price': '0.05',
                                                                'monthlyCost': '0',
                                                                'monthlyQuantity': '0',
                                                                'unit': 'vCPU-hours'
                                                            }
                                                        ],
                                                        'subresources': [
                                                            {
                                                                'costComponents': [
                                                                    {
                                                                        'hourlyCost': '0.0010958904109589',
                                                                        'name': 'Storage (general purpose SSD, gp2)',
                                                                        'hourlyQuantity': '0.010958904109589',
                                                                        'price': '0.1',
                                                                        'monthlyCost': '0.8',
                                                                        'monthlyQuantity': '8',
                                                                        'unit': 'GB'
                                                                    }
                                                                ],
                                                                'hourlyCost': '0.0010958904109589',
                                                                'monthlyCost': '0.8',
                                                                'name': 'root_block_device',
                                                                'metadata': {}
                                                            }
                                                        ],
                                                        'metadata': {}
                                                    }
                                                ]
                                            },
                                            'metadata': {
                                                'path': '.',
                                                'type': 'terraform_dir',
                                                'vcsSubPath': 'example/cost_example'
                                            }
                                        }
                                    ],
                                    'metadata': {
                                        'commitTimestamp': '2022-08-17T06:58:57Z',
                                        'commitMessage': 'Add screenshot of example page to README',
                                        'vcsRepoUrl': 'https://gitlab.dockstudios.co.uk:2222/pub/terrareg.git',
                                        'commitAuthorName': 'Matthew John',
                                        'infracostCommand': 'breakdown',
                                        'branch': '226-investigate-showing-costs-of-each-module-examples',
                                        'commit': '4822f3af904200b26ff0a3399750c76d20007f6b',
                                        'commitAuthorEmail': 'matthew@dockstudios.co.uk'
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
                                    'requirements': [],
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
                                })
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
                                    'requirements': [],
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
                                })
                            }
                        }
                    },
                }
            }},
        }
    },
    'secondtestnamespace': {
        'id': 3,
        'modules': {
            'mockmodule2': { 'secondprovider': {
                'id': 4,
                'latest_version': '3.0.0',
                'versions': {'3.0.0': {}}
            }}
        }
    },
    'smallernamespacelist': {
        'id': 4,
        'modules': {
            'publishedone': {
                'testprovider': {
                    'id': 21,
                    'latest_version': '2.1.1',
                    'versions': {
                        '2.1.1': {
                            'published': True,
                            'description': 'Test description'
                        }
                    }
                },
                'secondnamespace': {
                    'id': 22,
                    'latest_version': '2.2.2',
                    'versions': {
                        '2.2.2': {
                            'published': True,
                            'description': 'Description of second provider in module'
                        }
                    }
                }
            },
            'noversions': {'testprovider': {
                'id': 25,
                'versions': {}
            }},
            'onlybeta': {'testprovider': {
                'id': 23,
                'versions': {
                    '2.2.4-beta': {
                        'published': True,
                        'beta': True,
                        'description': 'Test description'
                    }
                }
            }},
            'onlyunpublished': {'testprovider': {
                'id': 24,
                'versions': {
                    '1.0.0': {
                        'published': False,
                        'description': 'Test description'
                    }
                }
            }}
        }
    },
    'moduleextraction': {
        'id': 5,
        'modules': {
            'test-module': { 'testprovider': {
                'id': 7,
                'repo_clone_url_template': 'ssh://example.com/repo.git'
            }},
            'bitbucketexample': {
                'testprovider': {
                    'id': 27,
                    'repo_clone_url_template': 'ssh://git@localhost:7999/bla/test-module.git',
                    'git_tag_format': 'v{version}',
                    'versions': {}
                },
                'norepourl': {
                    'id': 28,
                    'git_tag_format': 'v{version}',
                    'versions': {}
                }
            },
            'githubexample': {
                'testprovider': {
                    'id': 8,
                    'repo_clone_url_template': 'ssh://git@localhost:7999/bla/test-module.git',
                    'git_tag_format': 'v{version}',
                    'versions': {}
                },
                'norepourl': {
                    'id': 29,
                    'git_tag_format': 'v{version}',
                    'versions': {}
                }
            },
            'gitextraction': {
                'staticrepourl': {
                    'id': 9,
                    'repo_clone_url_template': 'ssh://git@localhost:7999/bla/test-module.git',
                    'git_tag_format': 'v{version}',
                    'versions': {}
                },
                'placeholdercloneurl': {
                    'id': 10,
                    'repo_clone_url_template': 'ssh://git@localhost:7999/{namespace}/{module}-{provider}.git',
                    'git_tag_format': 'v{version}',
                    'versions': {}
                },
                'usesgitprovider': {
                    'id': 11,
                    'git_provider_id': 1,
                    'git_tag_format': 'v{version}',
                    'versions': {}
                },
                'usesgitproviderwithversions': {
                    'id': 19,
                    'git_provider_id': 1,
                    'git_tag_format': 'v{version}',
                    'latest_version': '2.2.2',
                    'versions': {'2.2.2': {'published': True}}
                },
                'nogittagformat': {
                    'id': 30,
                    'git_provider_id': 1,
                    'versions': {}
                },
                'complexgittagformat': {
                    'id': 13,
                    'git_provider_id': 1,
                    'git_tag_format': 'unittest{version}value',
                    'versions': {}
                },
                'norepourl': {
                    'id': 14,
                    'git_tag_format': 'v{version}',
                    'versions': {}
                }
            }
        }
    },
    'emptynamespace': {
        'id': 6,
        'modules': {}
    }
}
