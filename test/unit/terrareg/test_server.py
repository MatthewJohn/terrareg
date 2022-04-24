
import unittest.mock
import datetime
import json

import pytest
import werkzeug.exceptions

from terrareg.models import Namespace, Module
from terrareg.module_search import ModuleSearch
from terrareg.filters import NamespaceTrustFilter
from terrareg.analytics import AnalyticsEngine
import terrareg.errors
from test.unit.terrareg import (
    MockModuleProvider, MockModuleVersion, MockModule, MockNamespace,
    client, mocked_server_namespace_fixture,
    test_request_context, app_context,
    setup_test_data, SERVER
)
from terrareg.server import (
    require_admin_authentication, AuthenticationType,
    get_current_authentication_type,
    check_csrf_token
)


@pytest.fixture
def mocked_search_module_providers(request):
    """Create mocked instance of search_module_providers method."""
    unmocked_search_module_providers = ModuleSearch.search_module_providers
    def cleanup_mocked_search_provider():
        ModuleSearch.search_module_providers = unmocked_search_module_providers
    request.addfinalizer(cleanup_mocked_search_provider)

    ModuleSearch.search_module_providers = unittest.mock.MagicMock(return_value=[])


@pytest.fixture
def mock_record_module_version_download(request):
    """Mock record_module_version_download function of AnalyticsEngine class."""
    magic_mock = unittest.mock.MagicMock(return_value=None)
    mock = unittest.mock.patch('terrareg.server.AnalyticsEngine.record_module_version_download', magic_mock)

    def cleanup_mocked_record_module_version_download():
        mock.stop()
    request.addfinalizer(cleanup_mocked_record_module_version_download)
    mock.start()


@pytest.fixture
def mock_server_get_module_provider_download_stats(request):
    """Mock get_module_provider_download_stats function of AnalyticsEngine class."""
    magic_mock = unittest.mock.MagicMock(return_value={
        'week': 10,
        'month': 58,
        'year': 127,
        'total': 226
    })
    mock = unittest.mock.patch('terrareg.server.AnalyticsEngine.get_module_provider_download_stats', magic_mock)

    def cleanup_mocked_record_module_version_download():
        mock.stop()
    request.addfinalizer(cleanup_mocked_record_module_version_download)
    mock.start()

class TestTerraformWellKnown:
    """Test TerraformWellKnown resource."""

    def test_with_no_params(self, client):
        """Test endpoint without query parameters"""
        res = client.get('/.well-known/terraform.json')
        assert res.status_code == 200
        assert res.json == {
            'modules.v1': '/v1/modules/'
        }

    def test_with_post(self, client):
        """Test invalid post request"""
        res = client.post('/.well-known/terraform.json')
        assert res.status_code == 405


class TestApiModuleList:

    def test_with_no_params(self, client, mocked_search_module_providers):
        """Call with no parameters"""
        res = client.get('/v1/modules')

        assert res.status_code == 200
        assert res.json == {
            'meta': {'current_offset': 0, 'limit': 10, 'next_offset': 10, 'prev_offset': 0}, 'modules': []
        }

        ModuleSearch.search_module_providers.assert_called_with(provider=None, verified=False, offset=0, limit=10)

    def test_with_limit_offset(self, client, mocked_search_module_providers):
        """Call with limit and offset"""
        res = client.get('/v1/modules?offset=23&limit=12')

        assert res.status_code == 200
        assert res.json == {
            'meta': {'current_offset': 23, 'limit': 12, 'next_offset': 35, 'prev_offset': 11}, 'modules': []
        }

        ModuleSearch.search_module_providers.assert_called_with(provider=None, verified=False, offset=23, limit=12)

    def test_with_max_limit(self, client, mocked_search_module_providers):
        """Call with limit higher than max"""
        res = client.get('/v1/modules?offset=65&limit=55')

        assert res.status_code == 200
        assert res.json == {
            'meta': {'current_offset': 65, 'limit': 50, 'next_offset': 115, 'prev_offset': 15}, 'modules': []
        }

        ModuleSearch.search_module_providers.assert_called_with(provider=None, verified=False, offset=65, limit=50)

    def test_with_provider_filter(self, client, mocked_search_module_providers):
        """Call with provider limit"""
        res = client.get('/v1/modules?provider=testprovider')

        assert res.status_code == 200
        assert res.json == {
            'meta': {'current_offset': 0, 'limit': 10, 'next_offset': 10, 'prev_offset': 0}, 'modules': []
        }

        ModuleSearch.search_module_providers.assert_called_with(provider='testprovider', verified=False, offset=0, limit=10)

    def test_with_verified_false(self, client, mocked_search_module_providers):
        """Call with verified flag as false"""
        res = client.get('/v1/modules?verified=false')

        assert res.status_code == 200
        assert res.json == {
            'meta': {'current_offset': 0, 'limit': 10, 'next_offset': 10, 'prev_offset': 0}, 'modules': []
        }
        ModuleSearch.search_module_providers.assert_called_with(provider=None, verified=False, offset=0, limit=10)


    def test_with_verified_true(self, client, mocked_search_module_providers):
        """Call with verified flag as true"""
        res = client.get('/v1/modules?verified=true')

        assert res.status_code == 200
        assert res.json == {
            'meta': {'current_offset': 0, 'limit': 10, 'next_offset': 10, 'prev_offset': 0}, 'modules': []
        }
        ModuleSearch.search_module_providers.assert_called_with(provider=None, verified=True, offset=0, limit=10)

    @setup_test_data()
    def test_with_module_response(self, client, mocked_search_module_providers):
        """Test return of single module module"""
        namespace = MockNamespace(name='testnamespace')
        module = MockModule(namespace=namespace, name='mock-module')
        mock_module_provider = MockModuleProvider(module=module, name='testprovider')
        ModuleSearch.search_module_providers.return_value = [mock_module_provider]

        res = client.get('/v1/modules?offset=0&limit=1')

        assert res.status_code == 200
        assert res.json == {
            'meta': {'current_offset': 0, 'limit': 1, 'next_offset': 1, 'prev_offset': 0}, 'modules': [
                {'id': 'testnamespace/mock-module/testprovider/1.2.3', 'owner': 'Mock Owner',
                'namespace': 'testnamespace', 'name': 'mock-module',
                'version': '1.2.3', 'provider': 'testprovider',
                'description': 'Mock description', 'source': 'http://mock.example.com/mockmodule',
                'published_at': '2020-01-01T23:18:12', 'downloads': 0, 'verified': True}
            ]
        }

    @setup_test_data()
    def test_with_multiple_modules_response(self, client, mocked_search_module_providers):
        """Test multiple modules in results"""
        namespace = MockNamespace(name='testnamespace')
        module = MockModule(namespace=namespace, name='mock-module')
        mock_module_provider = MockModuleProvider(module=module, name='testprovider')
        mock_module_provider.MOCK_LATEST_VERSION_NUMBER = '1.2.3'
        mock_namespace_2 = MockNamespace(name='secondtestnamespace')
        mock_module_2 = MockModule(namespace=mock_namespace_2, name='mockmodule2')
        mock_module_provider_2 = MockModuleProvider(module=mock_module_2, name='secondprovider')
        mock_module_provider_2.MOCK_LATEST_VERSION_NUMBER = '3.0.0'
        ModuleSearch.search_module_providers.return_value = [mock_module_provider_2, mock_module_provider]

        res = client.get('/v1/modules?offset=0&limit=2')

        assert res.status_code == 200
        assert res.json == {
            'meta': {'current_offset': 0, 'limit': 2, 'next_offset': 2, 'prev_offset': 0}, 'modules': [
                mock_module_provider_2.get_latest_version().get_api_outline(),
                mock_module_provider.get_latest_version().get_api_outline(),
            ]
        }


class TestApiModuleSearch:

    def test_with_no_params(self, client, mocked_search_module_providers):
        """Test ApiModuleSearch with no params"""
        res = client.get('/v1/modules/search')
        assert res.status_code == 400
        ModuleSearch.search_module_providers.assert_not_called()

    def test_with_query_string(self, client, mocked_search_module_providers):
        """Call with query param"""
        res = client.get('/v1/modules/search?q=unittestteststring')

        assert res.status_code == 200
        assert res.json == {
            'meta': {'current_offset': 0, 'limit': 10, 'next_offset': 10, 'prev_offset': 0}, 'modules': []
        }
        ModuleSearch.search_module_providers.assert_called_with(
            query='unittestteststring', namespace=None, provider=None, verified=False,
            namespace_trust_filters=NamespaceTrustFilter.UNSPECIFIED,
            offset=0, limit=10)

    def test_with_limit_offset(self, client, mocked_search_module_providers):
        """Call with limit and offset"""
        res = client.get('/v1/modules/search?q=test&offset=23&limit=12')

        assert res.status_code == 200
        assert res.json == {
            'meta': {'current_offset': 23, 'limit': 12, 'next_offset': 35, 'prev_offset': 11}, 'modules': []
        }
        ModuleSearch.search_module_providers.assert_called_with(
            query='test', namespace=None, provider=None, verified=False,
            namespace_trust_filters=NamespaceTrustFilter.UNSPECIFIED,
            offset=23, limit=12)

    def test_with_max_limit(self, client, mocked_search_module_providers):
        """Call with limit higher than max"""
        res = client.get('/v1/modules/search?q=test&offset=65&limit=55')

        assert res.status_code == 200
        assert res.json == {
            'meta': {'current_offset': 65, 'limit': 50, 'next_offset': 115, 'prev_offset': 15}, 'modules': []
        }
        ModuleSearch.search_module_providers.assert_called_with(
            query='test', namespace=None, provider=None, verified=False,
            namespace_trust_filters=NamespaceTrustFilter.UNSPECIFIED,
            offset=65, limit=50)

    def test_with_provider(self, client, mocked_search_module_providers):
        """Call with provider filter"""
        res = client.get('/v1/modules/search?q=test&provider=testprovider')

        assert res.status_code == 200
        assert res.json == {
            'meta': {'current_offset': 0, 'limit': 10, 'next_offset': 10, 'prev_offset': 0}, 'modules': []
        }
        ModuleSearch.search_module_providers.assert_called_with(
            query='test', namespace=None, provider='testprovider', verified=False,
            namespace_trust_filters=NamespaceTrustFilter.UNSPECIFIED,
            offset=0, limit=10)

    def test_with_namespace(self, client, mocked_search_module_providers):
        """Call with namespace filter"""
        res = client.get('/v1/modules/search?q=test&namespace=testnamespace')

        assert res.status_code == 200
        assert res.json == {
            'meta': {'current_offset': 0, 'limit': 10, 'next_offset': 10, 'prev_offset': 0}, 'modules': []
        }
        ModuleSearch.search_module_providers.assert_called_with(
            query='test', namespace='testnamespace', provider=None, verified=False,
            namespace_trust_filters=NamespaceTrustFilter.UNSPECIFIED,
            offset=0, limit=10)

    def test_with_namespace_trust_filters(self, client, mocked_search_module_providers):
        """Call with trusted namespace/contributed filters"""
        for namespace_filter in [['&trusted_namespaces=false', []],
                                ['&trusted_namespaces=true', [NamespaceTrustFilter.TRUSTED_NAMESPACES]],
                                ['&contributed=false', []],
                                ['&contributed=true', [NamespaceTrustFilter.CONTRIBUTED]],
                                ['&trusted_namespaces=false&contributed=false', []],
                                ['&trusted_namespaces=true&contributed=false', [NamespaceTrustFilter.TRUSTED_NAMESPACES]],
                                ['&trusted_namespaces=false&contributed=true', [NamespaceTrustFilter.CONTRIBUTED]],
                                ['&trusted_namespaces=true&contributed=true', [NamespaceTrustFilter.TRUSTED_NAMESPACES, NamespaceTrustFilter.CONTRIBUTED]]]:

            res = client.get('/v1/modules/search?q=test{0}'.format(namespace_filter[0]))

            assert res.status_code == 200
            assert res.json == {
                'meta': {'current_offset': 0, 'limit': 10, 'next_offset': 10, 'prev_offset': 0}, 'modules': []
            }
            ModuleSearch.search_module_providers.assert_called_with(
                query='test', namespace=None, provider=None, verified=False,
                namespace_trust_filters=namespace_filter[1],
                offset=0, limit=10)

    def test_with_verified_false(self, client, mocked_search_module_providers):
        """Call with verified flag as false"""
        res = client.get('/v1/modules/search?q=test&verified=false')

        assert res.status_code == 200
        assert res.json == {
            'meta': {'current_offset': 0, 'limit': 10, 'next_offset': 10, 'prev_offset': 0}, 'modules': []
        }
        ModuleSearch.search_module_providers.assert_called_with(
            query='test', namespace=None, provider=None, verified=False,
            namespace_trust_filters=NamespaceTrustFilter.UNSPECIFIED,
            offset=0, limit=10)

    def test_with_verified_true(self, client, mocked_search_module_providers):
        """Test call with verified as true"""
        res = client.get('/v1/modules/search?q=test&verified=true')

        assert res.status_code == 200
        assert res.json == {
            'meta': {'current_offset': 0, 'limit': 10, 'next_offset': 10, 'prev_offset': 0}, 'modules': []
        }
        ModuleSearch.search_module_providers.assert_called_with(
            query='test', namespace=None, provider=None, verified=True,
            namespace_trust_filters=NamespaceTrustFilter.UNSPECIFIED,
            offset=0, limit=10)

    @setup_test_data()
    def test_with_single_module_response(self, client, mocked_search_module_providers):
        """Test return of single module module"""
        namespace = MockNamespace(name='testnamespace')
        module = MockModule(namespace=namespace, name='mock-module')
        mock_module_provider = MockModuleProvider(module=module, name='testprovider')
        mock_module_provider.MOCK_LATEST_VERSION_NUMBER = '1.2.3'
        ModuleSearch.search_module_providers.return_value = [mock_module_provider]

        res = client.get('/v1/modules/search?q=test&offset=0&limit=1')

        assert res.status_code == 200
        assert res.json == {
            'meta': {'current_offset': 0, 'limit': 1, 'next_offset': 1, 'prev_offset': 0}, 'modules': [
                {'id': 'testnamespace/mock-module/testprovider/1.2.3', 'owner': 'Mock Owner',
                'namespace': 'testnamespace', 'name': 'mock-module',
                'version': '1.2.3', 'provider': 'testprovider',
                'description': 'Mock description', 'source': 'http://mock.example.com/mockmodule',
                'published_at': '2020-01-01T23:18:12', 'downloads': 0, 'verified': True}
            ]
        }

    @setup_test_data()
    def test_with_multiple_module_response(self, client, mocked_search_module_providers):
        """Test multiple modules in results"""
        namespace = MockNamespace(name='testnamespace')
        module = MockModule(namespace=namespace, name='mock-module')
        mock_module_provider = MockModuleProvider(module=module, name='testprovider')
        mock_module_provider.MOCK_LATEST_VERSION_NUMBER = '1.2.3'
        mock_namespace_2 = MockNamespace(name='secondtestnamespace')
        mock_module_2 = MockModule(namespace=mock_namespace_2, name='mockmodule2')
        mock_module_provider_2 = MockModuleProvider(module=mock_module_2, name='secondprovider')
        mock_module_provider_2.MOCK_LATEST_VERSION_NUMBER = '3.0.0'
        ModuleSearch.search_module_providers.return_value = [mock_module_provider_2, mock_module_provider]

        res = client.get('/v1/modules/search?q=test&offset=0&limit=2')

        assert res.status_code == 200
        assert res.json == {
            'meta': {'current_offset': 0, 'limit': 2, 'next_offset': 2, 'prev_offset': 0}, 'modules': [
                mock_module_provider_2.get_latest_version().get_api_outline(),
                mock_module_provider.get_latest_version().get_api_outline(),
            ]
        }


class TestApiModuleDetails:
    """Test ApiModuleDetails resource."""

    @setup_test_data()
    def test_existing_module(self, client, mocked_server_namespace_fixture):
        """Test endpoint with existing module"""

        res = client.get('/v1/modules/testnamespace/lonelymodule')

        assert res.json == {
            'meta': {'limit': 5, 'offset': 0}, 'modules': [
                {'id': 'testnamespace/lonelymodule/testprovider/1.0.0', 'owner': 'Mock Owner',
                'namespace': 'testnamespace', 'name': 'lonelymodule', 'version': '1.0.0',
                'provider': 'testprovider', 'description': 'Mock description',
                'source': 'http://mock.example.com/mockmodule',
                'published_at': '2020-01-01T23:18:12', 'downloads': 0, 'verified': True}
            ]
        }
        assert res.status_code == 200

    def test_non_existent_module(self, client, mocked_server_namespace_fixture):
        """Test endpoint with non-existent module"""

        res = client.get('/v1/modules/doesnotexist/unittestdoesnotexist')

        assert res.json == {'errors': ['Not Found']}
        assert res.status_code == 404

    @setup_test_data()
    def test_analytics_token(self, client, mocked_server_namespace_fixture):
        """Test endpoint with analytics token"""

        res = client.get('/v1/modules/test_token-name__testnamespace/lonelymodule')

        assert res.json == {
            'meta': {'limit': 5, 'offset': 0}, 'modules': [
                {'id': 'testnamespace/lonelymodule/testprovider/1.0.0', 'owner': 'Mock Owner',
                'namespace': 'testnamespace', 'name': 'lonelymodule', 'version': '1.0.0',
                'provider': 'testprovider', 'description': 'Mock description',
                'source': 'http://mock.example.com/mockmodule',
                'published_at': '2020-01-01T23:18:12', 'downloads': 0, 'verified': True}
            ]
        }
        assert res.status_code == 200


class TestApiModuleProviderDetails:
    """Test ApiModuleProviderDetails resource."""

    @setup_test_data()
    def test_existing_module_provider(self, client, mocked_server_namespace_fixture):
        res = client.get('/v1/modules/testnamespace/mock-module/testprovider')

        assert res.json == {
            'id': 'testnamespace/mock-module/testprovider/1.2.3', 'owner': 'Mock Owner',
            'namespace': 'testnamespace', 'name': 'mock-module',
            'version': '1.2.3', 'provider': 'testprovider',
            'description': 'Mock description',
            'source': 'http://mock.example.com/mockmodule',
            'published_at': '2020-01-01T23:18:12',
            'downloads': 0, 'verified': True,
            'root': {
                'path': '', 'readme': 'Mock module README file',
                'empty': False, 'inputs': [], 'outputs': [], 'dependencies': [],
                'provider_dependencies': [], 'resources': []
            },
            'submodules': [], 'providers': ['testprovider'], 'versions': []
        }

        assert res.status_code == 200

    def test_non_existent_module_provider(self, client, mocked_server_namespace_fixture):
        """Test endpoint with non-existent module"""

        res = client.get('/v1/modules/doesnotexist/unittestdoesnotexist/unittestproviderdoesnotexist')

        assert res.json == {'errors': ['Not Found']}
        assert res.status_code == 404

    @setup_test_data()
    def test_analytics_token(self, client, mocked_server_namespace_fixture):
        """Test endpoint with analytics token"""

        res = client.get('/v1/modules/test_token-name__testnamespace/testmodulename/testprovider')

        test_namespace = MockNamespace(name='testnamespace')
        test_module = MockModule(namespace=test_namespace, name='testmodulename')
        test_module_provider = MockModuleProvider(module=test_module, name='testprovider')

        assert res.json == test_module_provider.get_latest_version().get_api_details()
        assert res.status_code == 200


class TestApiModuleVersionDetails:
    """Test ApiModuleVersionDetails resource."""

    @setup_test_data()
    def test_existing_module_version(self, client, mocked_server_namespace_fixture):
        res = client.get('/v1/modules/testnamespace/testmodulename/testprovider/1.0.0')

        assert res.json == {
            'id': 'testnamespace/testmodulename/testprovider/1.0.0', 'owner': 'Mock Owner',
            'namespace': 'testnamespace', 'name': 'testmodulename',
            'version': '1.0.0', 'provider': 'testprovider',
            'description': 'Mock description',
            'source': 'http://mock.example.com/mockmodule',
            'published_at': '2020-01-01T23:18:12',
            'downloads': 0, 'verified': True,
            'root': {
                'path': '', 'readme': 'Mock module README file',
                'empty': False, 'inputs': [], 'outputs': [], 'dependencies': [],
                'provider_dependencies': [], 'resources': []
            },
            'submodules': [], 'providers': ['testprovider'], 'versions': []
        }

        assert res.status_code == 200

    @setup_test_data()
    def test_non_existent_module_version(self, client, mocked_server_namespace_fixture):
        """Test endpoint with non-existent module"""

        res = client.get('/v1/modules/namespacename/modulename/providername/0.1.2')

        assert res.json == {'errors': ['Not Found']}
        assert res.status_code == 404

    @setup_test_data()
    def test_analytics_token(self, client, mocked_server_namespace_fixture):
        """Test endpoint with analytics token"""

        res = client.get('/v1/modules/test_token-name__testnamespace/testmodulename/testprovider/2.4.1')

        test_namespace = MockNamespace(name='testnamespace')
        test_module = MockModule(namespace=test_namespace, name='testmodulename')
        test_module_provider = MockModuleProvider(module=test_module, name='testprovider')
        test_module_version = MockModuleVersion(module_provider=test_module_provider, version='2.4.1')

        assert res.json == test_module_version.get_api_details()
        assert res.status_code == 200


class TestApiModuleVersionDownload:
    """Test ApiModuleVersionDownload resource."""

    @setup_test_data()
    def test_existing_module_version_without_alaytics_token(self, client, mocked_server_namespace_fixture):
        res = client.get('/v1/modules/testnamespace/testmodulename/testprovider/1.0.0/download')
        assert res.status_code == 401
        assert res.data == b'\nAn analytics token must be provided.\nPlease update module source to include analytics token.\n\nFor example:\n  source = "localhost/my-tf-application__testnamespace/testmodulename/testprovider"'

    @setup_test_data()
    def test_non_existent_module_version(self, client, mocked_server_namespace_fixture):
        """Test endpoint with non-existent module"""

        res = client.get('/v1/modules/namespacename/modulename/testprovider/0.1.2/download')

        assert res.json == {'errors': ['Not Found']}
        assert res.status_code == 404

    @setup_test_data()
    def test_existing_module_internal_download(self, client, mocked_server_namespace_fixture, mock_record_module_version_download):
        """Test endpoint with analytics token"""

        res = client.get(
            '/v1/modules/test_token-name__testnamespace/testmodulename/testprovider/2.4.1/download',
            headers={'X-Terraform-Version': 'TestTerraformVersion',
                     'User-Agent': 'TestUserAgent'}
        )

        test_namespace = Namespace(name='testnamespace')
        test_module = MockModule(namespace=test_namespace, name='testmodulename')
        test_module_provider = MockModuleProvider(module=test_module, name='testprovider')
        test_module_version = MockModuleVersion(module_provider=test_module_provider, version='2.4.1')

        assert res.headers['X-Terraform-Get'] == '/v1/terrareg/modules/testnamespace/testmodulename/testprovider/2.4.1/source.zip'
        assert res.status_code == 204

        AnalyticsEngine.record_module_version_download.assert_called_with(
            module_version=unittest.mock.ANY,
            analytics_token='test_token-name',
            terraform_version='TestTerraformVersion',
            user_agent='TestUserAgent',
            auth_token=None
        )
        assert AnalyticsEngine.record_module_version_download.isinstance(
            AnalyticsEngine.record_module_version_download.call_args.kwargs['module_version'],
            MockModuleVersion
        )
        assert AnalyticsEngine.record_module_version_download.call_args.kwargs['module_version'].id == test_module_version.id

    @setup_test_data()
    def test_existing_module_internal_download_with_auth_token(
        self, client, mocked_server_namespace_fixture,
        mock_record_module_version_download):
        """Test endpoint with analytics token and auth token"""

        res = client.get(
            '/v1/modules/test_token-name__testnamespace/testmodulename/testprovider/2.4.1/download',
            headers={'X-Terraform-Version': 'TestTerraformVersion',
                     'User-Agent': 'TestUserAgent',
                     'Authorization': 'Bearer test123-authorization-token'}
        )

        test_namespace = Namespace(name='testnamespace')
        test_module = MockModule(namespace=test_namespace, name='testmodulename')
        test_module_provider = MockModuleProvider(module=test_module, name='testprovider')
        test_module_version = MockModuleVersion(module_provider=test_module_provider, version='2.4.1')

        assert res.headers['X-Terraform-Get'] == '/v1/terrareg/modules/testnamespace/testmodulename/testprovider/2.4.1/source.zip'
        assert res.status_code == 204

        AnalyticsEngine.record_module_version_download.assert_called_with(
            module_version=unittest.mock.ANY,
            analytics_token='test_token-name',
            terraform_version='TestTerraformVersion',
            user_agent='TestUserAgent',
            auth_token='test123-authorization-token'
        )
        assert AnalyticsEngine.record_module_version_download.isinstance(
            AnalyticsEngine.record_module_version_download.call_args.kwargs['module_version'],
            MockModuleVersion
        )
        assert AnalyticsEngine.record_module_version_download.call_args.kwargs['module_version'].id == test_module_version.id

    @setup_test_data()
    def test_existing_module_internal_download_with_invalid_auth_token_header(
        self, client, mocked_server_namespace_fixture,
        mock_record_module_version_download):
        """Test endpoint with analytics token and auth token"""

        res = client.get(
            '/v1/modules/test_token-name__testnamespace/testmodulename/testprovider/2.4.1/download',
            headers={'X-Terraform-Version': 'TestTerraformVersion',
                     'User-Agent': 'TestUserAgent',
                     'Authorization': 'This is invalid'}
        )

        test_namespace = Namespace(name='testnamespace')
        test_module = MockModule(namespace=test_namespace, name='testmodulename')
        test_module_provider = MockModuleProvider(module=test_module, name='testprovider')
        test_module_version = MockModuleVersion(module_provider=test_module_provider, version='2.4.1')

        assert res.headers['X-Terraform-Get'] == '/v1/terrareg/modules/testnamespace/testmodulename/testprovider/2.4.1/source.zip'
        assert res.status_code == 204

        AnalyticsEngine.record_module_version_download.assert_called_with(
            module_version=unittest.mock.ANY,
            analytics_token='test_token-name',
            terraform_version='TestTerraformVersion',
            user_agent='TestUserAgent',
            auth_token=None
        )
        assert AnalyticsEngine.record_module_version_download.isinstance(
            AnalyticsEngine.record_module_version_download.call_args.kwargs['module_version'],
            MockModuleVersion
        )
        assert AnalyticsEngine.record_module_version_download.call_args.kwargs['module_version'].id == test_module_version.id


class TestApiModuleProviderDownloadsSummary:
 
    def test_existing_module(self, client, mocked_server_namespace_fixture, mock_server_get_module_provider_download_stats):
        """Test endpoint with existing module"""
        res = client.get('/v1/modules/testnamespace/testmodule/testprovider/downloads/summary')
        assert res.status_code == 200
        assert res.json == {
            'data': {
                'attributes': {'month': 58, 'total': 226, 'week': 10, 'year': 127},
                'id': 'testnamespace/testmodule/testprovider',
                'type': 'module-downloads-summary'
            }
        }


class TestApiModuleVersionCreate:
    """Test module version creation resource."""

    @setup_test_data()
    def test_creation_with_no_module_provider_repository_url(self, client, mocked_server_namespace_fixture):
        """Test creating a module version without who's module provider does not contain a repository URL."""
        with unittest.mock.patch(
                    'terrareg.models.ModuleVersion.prepare_module') as mocked_prepare_module, \
                unittest.mock.patch(
                    'terrareg.module_extractor.GitModuleExtractor.process_upload') as mocked_process_upload:
            res = client.post(
                '/v1/terrareg/modules/testnamespace/modulenorepourl/testprovider/5.5.4/import')
            assert res.status_code == 400
            assert res.json == {'message': 'Module provider is not configured with a repository'}

            mocked_prepare_module.assert_not_called()
            mocked_process_upload.assert_not_called()

    @setup_test_data()
    def test_creation_with_valid_repository_url(self, client, mocked_server_namespace_fixture):
        """Test creating a module version without who's module provider does not contain a repository URL."""
        with unittest.mock.patch(
                    'terrareg.models.ModuleVersion.prepare_module') as mocked_prepare_module, \
                unittest.mock.patch(
                    'terrareg.module_extractor.GitModuleExtractor.process_upload') as mocked_process_upload:
            res = client.post(
                '/v1/terrareg/modules/testnamespace/modulewithrepourl/testprovider/5.5.4/import')
            assert res.status_code == 200
            assert res.json == {'status': 'Success'}

            mocked_prepare_module.assert_called_once()
            mocked_process_upload.assert_called_once()

    @setup_test_data()
    def test_creation_with_non_existent_module_provider(self, client, mocked_server_namespace_fixture):
        """Test creating a module version without who's module provider does not contain a repository URL."""
        with unittest.mock.patch(
                    'terrareg.models.ModuleVersion.prepare_module') as mocked_prepare_module, \
                unittest.mock.patch(
                    'terrareg.module_extractor.GitModuleExtractor.process_upload') as mocked_process_upload:
            res = client.post(
                '/v1/terrareg/modules/testnamespace/moduledoesnotexist/testprovider/5.5.4/import')
            assert res.status_code == 400
            assert res.json == {'message': 'Module provider is not configured with a repository'}

            mocked_prepare_module.assert_not_called()
            mocked_process_upload.assert_not_called()


class TestRequireAdminAuthenticationWrapper:
    """Test require_admin_authentication wrapper"""

    def _mock_function(self, x, y):
        """Test method to wrap to check arg/kwargs"""
        return x, y

    def _run_authentication_test(
        self,
        app_context,
        test_request_context,
        config_secret_key,
        config_admin_authentication_token,
        expect_fail,
        expected_authentication_type=None,
        mock_headers=None,
        mock_session=None):
        """Perform authentication test."""
        with test_request_context, \
                app_context, \
                unittest.mock.patch('terrareg.config.SECRET_KEY', config_secret_key), \
                unittest.mock.patch('terrareg.config.ADMIN_AUTHENTICATION_TOKEN', config_admin_authentication_token):

            # Fake mock_headers and mock_session
            if mock_headers:
                test_request_context.request.headers = mock_headers
            if mock_session:
                test_request_context.session = mock_session

            wrapped_mock = require_admin_authentication(self._mock_function)

            # Ensure before calling authentication, that current authentication
            # type is shown as not checked
            assert get_current_authentication_type() is AuthenticationType.NOT_CHECKED

            if expect_fail:
                with pytest.raises(werkzeug.exceptions.Unauthorized):
                    wrapped_mock()
            else:
                assert wrapped_mock('x-value', y='y-value') == ('x-value', 'y-value')

                # Check authentication_type has been set correctly. 
                if expected_authentication_type:
                    assert get_current_authentication_type() is expected_authentication_type

    def test_unauthenticated(self, app_context, test_request_context):
        """Ensure 401 without an API key or mock_session."""
        self._run_authentication_test(
            app_context=app_context,
            test_request_context=test_request_context,
            config_secret_key='asecrethere',
            config_admin_authentication_token='testpassword',
            expect_fail=True
        )

    def test_mock_session_authentication_with_no_app_secret(self, app_context, test_request_context):
        """Ensure 401 with valid authentication without an APP SECRET."""
        self._run_authentication_test(
            app_context=app_context,
            test_request_context=test_request_context,
            config_secret_key='',
            config_admin_authentication_token='testpassword',
            expect_fail=True,
            mock_session={
                'is_admin_authenticated': True,
                'expires': datetime.datetime.now() + datetime.timedelta(hours=5)
            }
        )

    def test_401_with_expired_mock_session(self, app_context, test_request_context):
        """Ensure resource is called with valid mock_session."""
        self._run_authentication_test(
            app_context=app_context,
            test_request_context=test_request_context,
            config_secret_key='testsecret',
            config_admin_authentication_token='testpassword',
            expect_fail=True,
            mock_session={
                'is_admin_authenticated': True,
                'expires': datetime.datetime.now() - datetime.timedelta(minutes=1)
            }
        )

    def test_invalid_authentication_with_empty_api_key(self, app_context, test_request_context):
        """Ensure resource is called with valid mock_session."""
        self._run_authentication_test(
            app_context=app_context,
            test_request_context=test_request_context,
            config_secret_key='testsecret',
            config_admin_authentication_token='',
            expect_fail=True,
            mock_headers={
                'Host': 'localhost',
                'X-Terrareg-ApiKey': ''
            }
        )

    def test_authentication_with_mock_session(self, app_context, test_request_context):
        """Ensure resource is called with valid mock_session."""
        self._run_authentication_test(
            app_context=app_context,
            test_request_context=test_request_context,
            config_secret_key='testsecret',
            config_admin_authentication_token='testpassword',
            expect_fail=False,
            expected_authentication_type=AuthenticationType.SESSION,
            mock_session={
                'is_admin_authenticated': True,
                'expires': datetime.datetime.now() + datetime.timedelta(hours=5)
            }
        )

    def test_authentication_with_api_key(self, app_context, test_request_context):
        """Ensure resource is called with an API key."""
        self._run_authentication_test(
            app_context=app_context,
            test_request_context=test_request_context,
            config_secret_key='testsecret',
            config_admin_authentication_token='testpassword',
            expect_fail=False,
            expected_authentication_type=AuthenticationType.AUTHENTICATION_TOKEN,
            mock_headers={
                'Host': 'localhost',
                'X-Terrareg-ApiKey': 'testpassword'
            }
        )


class TestApiTerraregIsAuthenticated:

    def test_authenticated(self, client):
        """Test endpoint when user is authenticated."""
        with unittest.mock.patch('terrareg.server.check_admin_authentication') as mock_admin_authentication:
            mock_admin_authentication.return_value = True
            res = client.get('/v1/terrareg/auth/admin/is_authenticated')
            assert res.status_code == 200
            assert res.json == {'authenticated': True}

    def test_unauthenticated(self, client):
        """Test endpoint when user is authenticated."""
        with unittest.mock.patch('terrareg.server.check_admin_authentication') as mock_admin_authentication:
            mock_admin_authentication.return_value = False
            res = client.get('/v1/terrareg/auth/admin/is_authenticated')
            assert res.status_code == 401

class TestApiTerraregAdminAuthenticate:

    def test_authenticated(self, client):
        """Test endpoint when user is authenticated."""
        cookie_expiry_mins = 5
        with unittest.mock.patch('terrareg.server.check_admin_authentication') as mock_admin_authentication:
            with unittest.mock.patch('terrareg.config.SECRET_KEY', 'averysecretkey'):
                with unittest.mock.patch('terrareg.config.ADMIN_SESSION_EXPIRY_MINS', cookie_expiry_mins):
                    # Update real app secret key
                    SERVER._app.secret_key = 'averysecretkey'

                    mock_admin_authentication.return_value = True

                    res = client.post('/v1/terrareg/auth/admin/login')
                    expected_cookie_expiry = datetime.datetime.now() - datetime.timedelta(minutes=cookie_expiry_mins)

                    assert res.status_code == 200
                    assert res.json == {'authenticated': True}
                    with client.session_transaction() as session:
                        # Assert that the session expiry is within 2 seconds of the expected expiry
                        assert (
                            (expected_cookie_expiry.timestamp() - session['expires'].timestamp()) < 2
                        )
                        assert session['is_admin_authenticated'] == True
                        assert len(session['csrf_token']) == 40

    def test_authenticated_without_secret_key(self, client):
        """Test endpoint and ensure session is not provided"""
        with unittest.mock.patch('terrareg.server.check_admin_authentication') as mock_admin_authentication:
            with unittest.mock.patch('terrareg.config.SECRET_KEY', ''):
                # Update real app secret key with fake value,
                # otherwise an error would be received when checking the session.
                SERVER._app.secret_key = 'test'

                mock_admin_authentication.return_value = True

                res = client.post('/v1/terrareg/auth/admin/login')

                assert res.status_code == 403
                assert res.json == {'message': 'Sessions not enabled in configuration'}
                with client.session_transaction() as session:
                    # Assert that no session cookies were provided
                    assert 'expires' not in session
                    assert 'is_admin_authenticated' not in session
                    assert 'csrf_token' not in session

                # Update server secret to empty value and ensure a 403 is still received.
                # The session cannot be checked
                SERVER._app.secret_key = ''
                res = client.post('/v1/terrareg/auth/admin/login')

                assert res.status_code == 403
                assert res.json == {'message': 'Sessions not enabled in configuration'}

    def test_unauthenticated(self, client):
        """Test endpoint when user is authenticated."""
        with unittest.mock.patch('terrareg.server.check_admin_authentication') as mock_admin_authentication:

            mock_admin_authentication.return_value = False

            res = client.post('/v1/terrareg/auth/admin/login')

            assert res.status_code == 401


class TestCSRFFunctions:
    """Test CSRF functions."""

    def test_valid_csrf_with_session(self, app_context, test_request_context, client):
        """Test checking a valid CSRF token with a session."""
        SERVER._app.secret_key = 'averysecretkey'
        with app_context, test_request_context:

            # Create fake session
            test_request_context.session['csrf_token'] = 'testcsrftoken'
            test_request_context.session['is_authenticated'] = True
            test_request_context.session['expires'] = datetime.datetime.now() + datetime.timedelta(minutes=1)
            test_request_context.session.modified = True

            assert check_csrf_token('testcsrftoken') == True

    def test_incorrect_csrf_with_session(self, app_context, test_request_context, client):
        """Test checking an incorrect CSRF token with a session."""
        SERVER._app.secret_key = 'averysecretkey'
        with app_context, test_request_context:

            # Create fake session
            test_request_context.session['csrf_token'] = 'testcsrftoken'
            test_request_context.session['is_authenticated'] = True
            test_request_context.session['expires'] = datetime.datetime.now() + datetime.timedelta(minutes=1)
            test_request_context.session.modified = True

            with pytest.raises(terrareg.errors.IncorrectCSRFTokenError):
                check_csrf_token('doesnotmatch')

    def test_empty_csrf_with_session(self, app_context, test_request_context, client):
        """Test checking an incorrect CSRF token with a session."""
        SERVER._app.secret_key = 'averysecretkey'
        with app_context, test_request_context:

            # Create fake session
            test_request_context.session['csrf_token'] = ''
            test_request_context.session['is_authenticated'] = True
            test_request_context.session['expires'] = datetime.datetime.now() + datetime.timedelta(minutes=1)
            test_request_context.session.modified = True

            with pytest.raises(terrareg.errors.NoSessionSetError):
                check_csrf_token('')

    def test_invalid_csrf_without_session(self, app_context, test_request_context, client):
        """Test checking a invalid CSRF token with a session is not established."""
        SERVER._app.secret_key = 'averysecretkey'
        with app_context, test_request_context:

            with pytest.raises(terrareg.errors.NoSessionSetError):
                check_csrf_token('doesnotmatter')

    def test_csrf_ignored_with_authentication_token(self, app_context, test_request_context, client):
        """Test checking a CSRF token is ignored when using authentication token."""
        SERVER._app.secret_key = 'averysecretkey'
        with app_context, test_request_context:

            # Set global context as authentication token
            app_context.g.authentication_type = AuthenticationType.AUTHENTICATION_TOKEN

            assert check_csrf_token(None) == False

    @pytest.mark.parametrize('authentication_type', [
        (AuthenticationType.NOT_AUTHENTICATED,),
        (AuthenticationType.NOT_CHECKED, ),
        (AuthenticationType.SESSION,)]
    )
    def test_csrf_not_ignored_with_non_authentication_token(self, authentication_type, app_context, test_request_context, client):
        """Test that all authentication types throw errors when CSRF is not passed."""
        SERVER._app.secret_key = 'averysecretkey'

        # Test that no session is thrown when no session is present
        with app_context, test_request_context:

            app_context.g.authentication_type = authentication_type

            with pytest.raises(terrareg.errors.NoSessionSetError):
                check_csrf_token(None)

        # Test that incorrect CSRF token is thrown, when incorrect token is provided
        with app_context, test_request_context:

            app_context.g.authentication_type = authentication_type

            # Create fake session
            test_request_context.session['csrf_token'] = 'iscorrect'
            test_request_context.session['is_admin_authenticated'] = True
            test_request_context.session['expires'] = datetime.datetime.now() + datetime.timedelta(minutes=1)
            test_request_context.session.modified = True

            with pytest.raises(terrareg.errors.IncorrectCSRFTokenError):
                check_csrf_token(None)


class TestApiTerraregModuleProviderSettings:
    """Test module provider settings endpoint"""

    @pytest.mark.parametrize('repository_url', [
        'https://unittest.com/module.git',
        'http://unittest.com/module.git',
        'ssh://unittest.com/module.git'
    ])
    @setup_test_data()
    def test_update_repository_url(
            self, repository_url, app_context,
            test_request_context, mocked_server_namespace_fixture,
            client
        ):
        """Test update of repository URL."""
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.server.check_admin_authentication', return_value=True) as mocked_check_admin_authentication, \
                unittest.mock.patch('terrareg.server.check_csrf_token', return_value=True) as mock_check_csrf, \
                unittest.mock.patch('terrareg.models.ModuleProvider.update_repository_url') as mock_update_repository_url:

            print(repository_url)
            res = client.post(
                '/v1/terrareg/modules/testnamespace/testmodulename/testprovider/settings',
                json={
                    'repository_url': repository_url,
                    'csrf_token': 'unittestcsrf'
                }
            )
            assert res.json == {}
            assert res.status_code == 200

            # Ensure required checks are called
            mock_check_csrf.assert_called_once_with('unittestcsrf')
            mocked_check_admin_authentication.assert_called()
            mock_update_repository_url.assert_called_once_with(
                repository_url=repository_url)

    @setup_test_data()
    def test_update_repository_url_invalid_protocol(self, app_context, test_request_context, mocked_server_namespace_fixture, client):
        """Test update of repository URL with invalid protocol."""
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.server.check_admin_authentication', return_value=True) as mocked_check_admin_authentication, \
                unittest.mock.patch('terrareg.server.check_csrf_token', return_value=True) as mock_check_csrf, \
                unittest.mock.patch('terrareg.models.ModuleProvider.update_repository_url') as mock_update_repository_url:

            res = client.post(
                '/v1/terrareg/modules/testnamespace/testmodulename/testprovider/settings',
                json={
                    'repository_url': 'nope://unittest.com/module.git',
                    'csrf_token': 'unittestcsrf'
                }
            )
            assert res.json == {'message': 'Repository URL contains an unknown scheme (e.g. https/git/http)'}
            assert res.status_code == 400

            # Ensure required checks are called
            mock_check_csrf.assert_called_once_with('unittestcsrf')
            mocked_check_admin_authentication.assert_called()
            mock_update_repository_url.assert_not_called()

    @setup_test_data()
    def test_update_repository_url_invalid_domain(self, app_context, test_request_context, mocked_server_namespace_fixture, client):
        """Test update of repository URL with an invalid domain."""
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.server.check_admin_authentication', return_value=True) as mocked_check_admin_authentication, \
                unittest.mock.patch('terrareg.server.check_csrf_token', return_value=True) as mock_check_csrf, \
                unittest.mock.patch('terrareg.models.ModuleProvider.update_repository_url') as mock_update_repository_url:

            res = client.post(
                '/v1/terrareg/modules/testnamespace/testmodulename/testprovider/settings',
                json={
                    'repository_url': 'https:///module.git',
                    'csrf_token': 'unittestcsrf'
                }
            )
            assert res.json == {'message': 'Repository URL does not contain a host/domain'}
            assert res.status_code == 400

            # Ensure required checks are called
            mock_check_csrf.assert_called_once_with('unittestcsrf')
            mocked_check_admin_authentication.assert_called()
            mock_update_repository_url.assert_not_called()

    @setup_test_data()
    def test_update_repository_url_without_path(self, app_context, test_request_context, mocked_server_namespace_fixture, client):
        """Test update of repository URL without a path."""
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.server.check_admin_authentication', return_value=True) as mocked_check_admin_authentication, \
                unittest.mock.patch('terrareg.server.check_csrf_token', return_value=True) as mock_check_csrf, \
                unittest.mock.patch('terrareg.models.ModuleProvider.update_repository_url') as mock_update_repository_url:

            res = client.post(
                '/v1/terrareg/modules/testnamespace/testmodulename/testprovider/settings',
                json={
                    'repository_url': 'https://example.com',
                    'csrf_token': 'unittestcsrf'
                }
            )
            assert res.json == {'message': 'Repository URL does not contain a path'}
            assert res.status_code == 400

            # Ensure required checks are called
            mock_check_csrf.assert_called_once_with('unittestcsrf')
            mocked_check_admin_authentication.assert_called()
            mock_update_repository_url.assert_not_called()

    @setup_test_data()
    def test_update_repository_without_csrf(self, app_context, test_request_context, mocked_server_namespace_fixture, client):
        """Test update of repository URL without a CSRF token."""
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.server.check_admin_authentication', return_value=True) as mocked_check_admin_authentication, \
                unittest.mock.patch('terrareg.server.check_csrf_token', return_value=True) as mock_check_csrf, \
                unittest.mock.patch('terrareg.models.ModuleProvider.update_repository_url') as mock_update_repository_url:

            res = client.post(
                '/v1/terrareg/modules/testnamespace/testmodulename/testprovider/settings',
                json={
                    'repository_url': 'https://example.com/test.git'
                }
            )
            assert res.json == {}
            assert res.status_code == 200

            # Ensure required checks are called
            mock_check_csrf.assert_called_once_with(None)
            mocked_check_admin_authentication.assert_called()
            mock_update_repository_url.assert_called_once_with(
                repository_url='https://example.com/test.git')


class TestApiModuleVersionCreateBitBucketHook:
    """Test TestApiModuleVersionCreateBitBucketHook resource."""

    @setup_test_data()
    def test_hook_with_full_payload_single_change(self, client, mocked_server_namespace_fixture):
        """Test hook call with no changes."""
        with unittest.mock.patch(
                    'terrareg.models.ModuleVersion.prepare_module') as mocked_prepare_module, \
                unittest.mock.patch(
                    'terrareg.module_extractor.GitModuleExtractor.process_upload') as mocked_process_upload:

            res = client.post(
                '/v1/terrareg/modules/moduleextraction/bitbucketexample/testprovider/hooks/bitbucket',
                json={
                    "eventKey": "repo:refs_changed", "date": "2022-04-23T21:21:46+0000",
                    "actor": {
                        "name": "admin", "emailAddress": "admin@localhost",
                        "id": 2, "displayName": "admin", "active": True,
                        "slug": "admin", "type": "NORMAL",
                        "links": {"self": [ {"href": "http://localhost:7990/users/admin"}]
                        }
                    },
                    "repository": {
                        "slug": "test-module", "id": 1, "name": "test-module", "hierarchyId": "34098b9e0f8011fcfb25",
                        "scmId": "git", "state": "AVAILABLE", "statusMessage": "Available", "forkable": True,
                        "project": {
                            "key": "BLA", "id": 1, "name": "bla", "public": True, "type": "NORMAL",
                            "links": {"self": [{"href": "http://localhost:7990/projects/BLA"}]}
                        },
                        "public": True,
                        "links": {
                            "clone": [
                                {"href": "ssh://git@localhost:7999/bla/test-module.git", "name": "ssh"},
                                {"href": "http://localhost:7990/scm/bla/test-module.git", "name": "http"}
                            ],
                            "self": [{"href": "http://localhost:7990/projects/BLA/repos/test-module/browse"}]
                        }
                    },
                    "changes": [
                        {
                            "ref": {
                                "id": "refs/tags/v4.0.6",
                                "displayId": "v4.0.6",
                                "type": "TAG"
                            },
                            "refId": "refs/tags/v4.0.6",
                            "fromHash": "0000000000000000000000000000000000000000",
                            "toHash": "1097d939669e3209ff33e6dfe982d84c204f6087",
                            "type": "ADD"
                        }
                    ]
                }
            )

            assert res.status_code == 200
            assert res.json == {'status': 'Success', 'message': 'Imported all provided tags', 'tags': {'4.0.6': {'status': 'Success'}}}

            mocked_prepare_module.assert_called_once()
            mocked_process_upload.assert_called_once()
