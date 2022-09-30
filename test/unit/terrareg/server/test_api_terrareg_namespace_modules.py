
from unittest import mock

import pytest
from test.unit.terrareg import (
    MockModuleProvider, MockModule, MockNamespace,
    mocked_server_namespace_fixture,
    setup_test_data, TerraregUnitTest
)
from test import client


class TestApiTerraregNamespaceModules(TerraregUnitTest):
    """Test ApiTerraregNamespaceModules resource."""

    @setup_test_data()
    def test_existing_namespace_with_mixed_modules(self, client, mocked_server_namespace_fixture):
        res = client.get('/v1/terrareg/modules/smallernamespacelist')

        assert res.json == {
            'meta': {
                'current_offset': 0,
                'limit': 10
            },
            'modules': [
                {
                    # Ensure normal published module is shown
                    'description': 'Test description',
                    'downloads': 0,
                    'id': 'smallernamespacelist/publishedone/testprovider/2.1.1',
                    'internal': False,
                    'name': 'publishedone',
                    'namespace': 'smallernamespacelist',
                    'owner': 'Mock Owner',
                    'provider': 'testprovider',
                    'published_at': '2020-01-01T23:18:12',
                    'source': None,
                    'trusted': False,
                    'verified': False,
                    'version': '2.1.1'
                },
                {
                    # Second published module provider in same module
                    'description': 'Description of second provider in module',
                    'downloads': 0,
                    'id': 'smallernamespacelist/publishedone/secondnamespace/2.2.2',
                    'internal': False,
                    'name': 'publishedone',
                    'namespace': 'smallernamespacelist',
                    'owner': 'Mock Owner',
                    'provider': 'secondnamespace',
                    'published_at': '2020-01-01T23:18:12',
                    'source': None,
                    'trusted': False,
                    'verified': False,
                    'version': '2.2.2'
                },
                {
                    # Ensure module provider with no versions is returned correctly
                    'id': 'smallernamespacelist/noversions/testprovider',
                    'name': 'noversions',
                    'namespace': 'smallernamespacelist',
                    'provider': 'testprovider',
                    'trusted': False,
                    'verified': False
                },
                {
                    # Ensure published module provider that contains only a beta version is shown
                    'id': 'smallernamespacelist/onlybeta/testprovider',
                    'name': 'onlybeta',
                    'namespace': 'smallernamespacelist',
                    'provider': 'testprovider',
                    'trusted': False,
                    'verified': False
                },
                {
                    # Ensure published module provider that contains only an unpublished version is shown
                    'id': 'smallernamespacelist/onlyunpublished/testprovider',
                    'name': 'onlyunpublished',
                    'namespace': 'smallernamespacelist',
                    'provider': 'testprovider',
                    'trusted': False,
                    'verified': False
                }
            ]
        }

        assert res.status_code == 200

    def test_non_existent_namespace(self, client, mocked_server_namespace_fixture):
        """Test endpoint with non-existent module"""

        res = client.get('/v1/terrareg/modules/doesnotexist')

        assert res.json == {'errors': ['Not Found']}
        assert res.status_code == 404

    @setup_test_data()
    def test_analytics_token_not_converted(self, client, mocked_server_namespace_fixture):
        """Test endpoint with analytics token and ensure it doesn't convert the analytics token."""

        res = client.get('/v1/terrareg/modules/test_token-name__testnamespace')

        assert res.json == {'errors': ['Not Found']}
        assert res.status_code == 404
