
from datetime import datetime
import unittest.mock

import pytest

from test.integration.terrareg import TerraregIntegrationTest
from test import client, app_context, test_request_context
import terrareg.provider_search
from terrareg.auth import AdminSessionAuthMethod
import terrareg.provider_version_model


class TestApiProviderVersionsGet(TerraregIntegrationTest):
    """Test ApiProviderVersions GET endpoint"""

    def test_endpoint(self, client):
        """Test endpoint."""
        res = client.get('/v1/providers/initial-providers/multiple-versions')
        assert res.status_code == 200
        assert res.json == {
            'alias': None,
            'description': 'Test Multiple Versions',
            'docs': [],
            'downloads': 0,
            'id': 'initial-providers/multiple-versions/2.0.1',
            'logo_url': 'https://git.example.com/initalproviders/terraform-provider-test-initial.png',
            'name': 'multiple-versions',
            'namespace': 'initial-providers',
            'owner': 'initial-providers',
            'published_at': '2023-10-01T12:05:56',
            'source': 'https://github.example.com/initial-providers/terraform-provider-multiple-versions',
            'tag': 'v2.0.1',
            'tier': 'community',
            'version': '2.0.1',
            'versions': [
                '2.0.1',
                '2.0.0',
                '1.5.0',
                '1.1.0',
                '1.1.0-beta',
                '1.0.0'
            ]
        }

    def test_endpoint_with_provider_without_versions(self, client):
        """Test endpoint with provider that doesn't have any versions"""
        res = client.get('/v1/providers/initial-providers/empty-provider-publish')
        assert res.status_code == 404
        assert res.json == {'errors': ['Not Found']}

    def test_endpoint_non_existent_provider(self, client):
        """Test endpoint with non-existent provider"""
        res = client.get('/v1/providers/initial-providers/doesnotexist')
        assert res.status_code == 404
        assert res.json == {'errors': ['Not Found']}

    def test_endpoint_non_existent_namespace(self, client):
        """Test endpoint with non-existent namespace"""
        res = client.get('/v1/providers/doesnotexist/doesnotexist')
        assert res.status_code == 404
        assert res.json == {'errors': ['Not Found']}

    def test_unauthenticated(self, client):
        """Test unauthenticated call to API"""
        def call_endpoint():
            return client.get('/v1/providers/initial-providers/multiple-versions')

        self._test_unauthenticated_terraform_api_endpoint_test(call_endpoint)


class TestApiProviderVersionsPost(TerraregIntegrationTest):
    """Test ApiProviderVersions POST endpoint"""

    def test_non_existent_namespace(self, client):
        """Test endpoint with invalid namespace"""
        with unittest.mock.patch('terrareg.csrf.check_csrf_token', unittest.mock.MagicMock(return_value=True)) as mock_check_csrf, \
                unittest.mock.patch('terrareg.provider_model.Provider.refresh_versions', unittest.mock.MagicMock(return_value=[])) as mock_refresh_versions:

            res = client.post(f"/v1/providers/doesnotexist/non-existent/versions", json={"csrf_token": "test", "version": "1.2.5"})
            assert res.status_code == 404
            assert res.json == {'errors': ['Not Found']}
            mock_check_csrf.assert_called_once_with('test')
            mock_refresh_versions.assert_not_called()

    def test_non_existent_provider(self, client):
        """Test endpoint with invalid namespace"""
        with unittest.mock.patch('terrareg.csrf.check_csrf_token', unittest.mock.MagicMock(return_value=True)) as mock_check_csrf, \
                unittest.mock.patch('terrareg.provider_model.Provider.refresh_versions', unittest.mock.MagicMock(return_value=[])) as mock_refresh_versions:

            res = client.post(f"/v1/providers/initial-providers/does-not-exist/versions", json={"csrf_token": "test", "version": "1.2.5"})
            assert res.status_code == 404
            assert res.json == {'errors': ['Not Found']}
            mock_check_csrf.assert_called_once_with('test')
            mock_refresh_versions.assert_not_called()

    def test_without_permissions(self, client, app_context, test_request_context):
        """Test endpoint without required permissions"""
        mock_auth_method = unittest.mock.MagicMock()
        mock_auth_method.can_publish_module_version.return_value = False
        mock_get_current_auth_method = unittest.mock.MagicMock(return_value=mock_auth_method)

        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', mock_get_current_auth_method), \
                unittest.mock.patch('terrareg.csrf.check_csrf_token', return_value=True) as mock_check_csrf, \
                unittest.mock.patch('terrareg.provider_model.Provider.refresh_versions', unittest.mock.MagicMock(return_value=[])) as mock_refresh_versions:

            res = client.post(f"/v1/providers/initial-providers/multiple-versions/versions", json={"csrf_token": "test", "version": "1.2.5"})
            assert res.json == {'message': "You don't have the permission to access the requested resource. "
                                           'It is either read-protected or not readable by the server.'}
            assert res.status_code == 403

            mock_auth_method.can_publish_module_version.assert_called_once_with(namespace='initial-providers')
            mock_refresh_versions.assert_not_called()
            mock_check_csrf.assert_not_called()

    def test_authenticated_with_admin(self, client):
        """Test Endpoint whilst authenticated with admin session"""
        self._get_current_auth_method_mock.stop()
        mock_get_current_auth_method = unittest.mock.MagicMock(return_value=AdminSessionAuthMethod())

        with unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', mock_get_current_auth_method), \
                unittest.mock.patch('terrareg.csrf.check_csrf_token', return_value=True) as mock_check_csrf, \
                unittest.mock.patch('terrareg.provider_model.Provider.refresh_versions', unittest.mock.MagicMock(
                    return_value=[terrareg.provider_version_model.ProviderVersion(provider=None, version="1.2.5")])) as mock_refresh_versions:

            res = client.post(f"/v1/providers/initial-providers/multiple-versions/versions", json={"csrf_token": "test", "version": "1.2.5"})
            assert res.status_code == 200
            assert res.json == {"versions": ["1.2.5"]}
            mock_check_csrf.assert_called_once_with('test')
            mock_refresh_versions.assert_called_once_with(version="1.2.5")

    def test_refresh_versions_extraction_terrareg_exception(self, client):
        """Test refresh_versions method with Terrareg exception raised when extracting version"""
        mock_get_current_auth_method = unittest.mock.MagicMock(return_value=AdminSessionAuthMethod())

        class UnittestExtractionException(terrareg.errors.TerraregError):
            pass

        def raise_refresh_versions_exception(*args, **kwargs):
            """Raise exception when attempting to refresh versions"""
            raise UnittestExtractionException("Unit test generic exception")

        with unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', mock_get_current_auth_method), \
                unittest.mock.patch('terrareg.csrf.check_csrf_token', unittest.mock.MagicMock(return_value=True)) as mock_check_csrf, \
                unittest.mock.patch('terrareg.provider_model.Provider.refresh_versions', unittest.mock.MagicMock(side_effect=raise_refresh_versions_exception)) as mock_refresh_versions:

            res = client.post(f"/v1/providers/initial-providers/multiple-versions/versions", json={"csrf_token": "test", "version": "1.2.5"})
            assert res.status_code == 500
            assert res.json == {"message": "Unit test generic exception", 'status': 'Error'}

            mock_check_csrf.assert_called_once_with('test')
