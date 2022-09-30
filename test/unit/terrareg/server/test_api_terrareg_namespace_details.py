
from unittest import mock
from terrareg.module_search import ModuleSearch, ModuleSearchResults

from test import client
from test.unit.terrareg import (
    TerraregUnitTest,
    MockModuleProvider, MockModule, MockNamespace,
    setup_test_data, mocked_server_namespace_fixture
)


class TestApiTerraregNamespaceDetails(TerraregUnitTest):
    """Test ApiTerraregNamespaceDetails resource."""

    @setup_test_data()
    def test_with_non_existent_namespace(self, client, mocked_server_namespace_fixture):
        """Test namespace details with non-existent namespace."""
        res = client.get('/v1/terrareg/namespaces/doesnotexist')

        assert res.status_code == 404
        assert res.json == {'errors': ['Not Found']}

    @setup_test_data()
    def test_with_existing_namespace(self, client, mocked_server_namespace_fixture):
        """Test namespace details with existing namespace."""
        res = client.get('/v1/terrareg/namespaces/testnamespace')

        assert res.status_code == 200
        assert res.json == {'is_auto_verified': False, 'trusted': False}

    @setup_test_data()
    def test_with_trusted_namespace(self, client, mocked_server_namespace_fixture):
        """Test namespace details with trusted namespace."""
        with mock.patch('terrareg.config.Config.TRUSTED_NAMESPACES', ['testnamespace']):
            res = client.get('/v1/terrareg/namespaces/testnamespace')

            assert res.status_code == 200
            assert res.json == {'is_auto_verified': False, 'trusted': True}

    @setup_test_data()
    def test_with_auto_verified_namespace(self, client, mocked_server_namespace_fixture):
        """Test namespace details with auto-verified namespace."""
        with mock.patch('terrareg.config.Config.VERIFIED_MODULE_NAMESPACES', ['testnamespace']):
            res = client.get('/v1/terrareg/namespaces/testnamespace')

            assert res.status_code == 200
            assert res.json == {'is_auto_verified': True, 'trusted': False}
