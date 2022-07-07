

from terrareg.analytics import AnalyticsEngine
from terrareg.models import Module, ModuleProvider, ModuleVersion, Namespace
from test.integration.terrareg import TerraregIntegrationTest


class TestGetGlobalModuleUsage(TerraregIntegrationTest):

    _TEST_DATA = {
        'testnamespace': {
            'publishedmodule': {
                'testprovider': {
                    'id': 1,
                    'versions': {
                        '1.4.0': {'published': True},
                        '1.5.0': {'published': True},
                        '1.6.0-beta': {'published': True, 'beta': True}
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
            # Unsure a module with no analytics is not displayed
            'unusedmodule': {'testprovider': {
                'id': 4,
                'versions': {'1.2.0': {'published': True}}
            }},
            'onlybeta': {'testprovider': {
                'id': 5,
                'versions': {
                    '1.0.0-beta': {'published': True, 'beta': True}
                }
            }},
            'noanalyticstoken': {'testprovider': {
                'id': 6,
                'versions': {'2.2.2': {'published': True}}
            }}
        },
        # Ensure a second namespace is displayed
        'secondnamespace': {
            'othernamespacemodule': {'anotherprovider': {
                'id': 7,
                'versions': {
                    '2.0.4': {'published': True}
                }
            }}
        }
    }

    _TEST_ANALYTICS_DATA = {
        'testnamespace/publishedmodule/testprovider/1.4.0': [
            ['duplicate-application', 'prod-key', '0.12.5'],
            ['application-using-old-version', 'dev-key', '0.23.23']
        ],
        'testnamespace/publishedmodule/testprovider/1.5.0': [
            # Usage without analaytics key
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
            ['withoutanalaytics', None, '0.2.2']
        ],
        'secondnamespace/othernamespacemodule/anotherprovider/2.0.4': [
            ['duplicate-application', 'dev-key', '2.1.23']
        ]
    }

    def _import_test_analaytics(self, download_data):
        """Import test analaytics for each module"""
        for module_key in download_data:
            namespace, module, provider, version = module_key.split('/')
            module_version = ModuleVersion.get(ModuleProvider.get(Module(Namespace(namespace), module), provider), version)
            for analytics_token, auth_token, terraform_version in download_data[module_key]:
                print('recording: ', namespace, module, provider, version, analytics_token, auth_token, terraform_version)
                AnalyticsEngine.record_module_version_download(
                    module_version=module_version, terraform_version=terraform_version,
                    analytics_token=analytics_token, user_agent=None, auth_token=auth_token)

    def test_get_global_module_usage_with_no_analytics(self):
        """Test function with no analytics recorded."""
        assert AnalyticsEngine.get_global_module_usage() == {}

    def test_get_global_module_usage_excluding_no_environment(self):
        """Test function with default functionality, excluding stats for analytics without an API token"""
        self._import_test_analaytics(self._TEST_ANALYTICS_DATA)

        assert AnalyticsEngine.get_global_module_usage() == {
            'testnamespace/publishedmodule/testprovider': 4,
            'testnamespace/publishedmodule/secondprovider': 2,
            'testnamespace/secondmodule/testprovider': 2,
            'secondnamespace/othernamespacemodule/anotherprovider': 1
        }

    def test_get_global_module_usage_including_empty_auth_token(self):
        """Test function including stats for analytics without an auth token"""
        self._import_test_analaytics(self._TEST_ANALYTICS_DATA)

        assert AnalyticsEngine.get_global_module_usage(include_empty_auth_token=True) == {
            'testnamespace/publishedmodule/testprovider': 5,
            'testnamespace/publishedmodule/secondprovider': 2,
            'testnamespace/secondmodule/testprovider': 2,
            'secondnamespace/othernamespacemodule/anotherprovider': 1,
            'testnamespace/noanalyticstoken/testprovider': 1
        }


