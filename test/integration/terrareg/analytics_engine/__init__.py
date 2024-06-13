

from terrareg.models import Module, ModuleProvider, ModuleVersion, Namespace
from terrareg.analytics import AnalyticsEngine
from test.integration.terrareg import TerraregIntegrationTest


class AnalyticsIntegrationTest(TerraregIntegrationTest):
    """Base class for analytics integration tests."""

    _TEST_DATA = {
        'testnamespace': {
            'modules': {
                'publishedmodule': {
                    'testprovider': {
                        'id': 1,
                        'versions': {
                            '0.9.0': {'published': True},
                            '0.9.1': {'published': True},
                            '0.9.2': {'published': True},
                            '1.3.0': {'published': True},
                            '1.4.0': {'published': True},
                            '1.5.0': {'published': True},
                            '1.6.0-beta': {'published': True, 'beta': True},
                            '2.0.0': {'published': True},
                            '2.1.5': {'published': True},
                        }
                    },
                    # Ensure multiple providers in the same module as treated independently
                    'secondprovider': {
                        'id': 2,
                        'versions': {
                            '1.0.0': {'published': True}
                        }
                    }
                },
                'secondmodule': {'testprovider': {
                    'id': 3,
                    'versions': {'1.1.1': {'published': True}}
                }},
                # Ensure a module with no analytics is not displayed
                'unusedmodule': {'testprovider': {
                    'id': 4,
                    'versions': {'1.2.0': {'published': True}}
                }},
                # Ensure module with only beta versino is not included in module count
                'onlybeta': {'testprovider': {
                    'id': 5,
                    'versions': {
                        '1.0.0-beta': {'published': True, 'beta': True}
                    }
                }},
                'noanalyticstoken': {'testprovider': {
                    'id': 6,
                    'versions': {'2.2.2': {'published': True}}
                }},
                # Ensure module with no versions is not included in count
                'noversions': {'testprovider': {
                    'id': 7,
                    'versions': {}
                }},
                # Ensure module with unpublished version is not included in count
                'unpublishedversion': {'testprovider': {
                    'id': 8,
                    'versions': {'2.45.2': {'published': False}}
                }}
            }
        },
        # Ensure a second namespace is displayed
        'secondnamespace': {
            'modules': {
                'othernamespacemodule': {'anotherprovider': {
                    'id': 9,
                    'versions': {
                        '2.0.4': {'published': True}
                    }
                }}
            }
        }
    }

    _TEST_ANALYTICS_DATA = {
        'testnamespace/publishedmodule/testprovider/1.4.0': [
            ['duplicate-application', 'prod-key', '0.12.5'],
            ['application-using-old-version', 'dev-key', '0.23.23']
        ],
        'testnamespace/publishedmodule/testprovider/1.5.0': [
            # Usage without analytics key
            ['test-application', None, '0.11.31'],
            # Multiple downloads by same analytics token
            ['test-application', 'dev-key', '0.11.31'],
            ['test-application', 'dev-key', '0.11.31'],
            ['test-application', 'dev-key', '0.11.31'],
            ['second-application', 'dev-key', '0.11.31'],
            ['second-application', 'prod-key', '0.12.5'],
            ['duplicate-application', 'prod-key', '0.12.5'],
            ['without-analytics-key', None, '2.2.2']
        ],
        'testnamespace/publishedmodule/testprovider/1.6.0-beta': [
            ['test-application', 'dev-key', '0.12.31'],
            ['onlyusedbeta', 'dev-key', '0.23.21']
        ],
        'testnamespace/publishedmodule/secondprovider/1.0.0': [
            ['test-app-using-second-module', 'dev-key', '0.31.2'],
            ['duplicate-application', 'dev-key', '0.12.2']
        ],
        'testnamespace/secondmodule/testprovider/1.1.1': [
            ['test-app-using-second-module', 'dev-key', '0.31.2'],
            ['duplicate-application', 'prod-key', '0.2.2']
        ],
        'testnamespace/onlybeta/testprovider/1.0.0-beta': [
            ['onlyusedbeta', 'dev-key', '0.23.21']
        ],
        'testnamespace/noanalyticstoken/testprovider/2.2.2': [
            ['withoutanalytics', None, '0.2.2']
        ],
        'secondnamespace/othernamespacemodule/anotherprovider/2.0.4': [
            ['duplicate-application', 'dev-key', '2.1.23']
        ]
    }

    def _import_test_analytics(self, download_data):
        """Import test analytics for each module"""
        for module_key in download_data:
            namespace, module, provider, version = module_key.split('/')
            module_version = ModuleVersion.get(ModuleProvider.get(Module(Namespace(namespace), module), provider), version)
            for analytics_token, auth_token, terraform_version in download_data[module_key]:
                print('recording: ', namespace, module, provider, version, analytics_token, auth_token, terraform_version)
                AnalyticsEngine.record_module_version_download(
                    namespace_name=namespace, module_name=module, provider_name=provider,
                    module_version=module_version, terraform_version=terraform_version,
                    analytics_token=analytics_token, user_agent="Terraform/{}".format(terraform_version),
                    auth_token=auth_token)
