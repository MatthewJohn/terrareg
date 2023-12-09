
from contextlib import contextmanager
from datetime import datetime
import unittest.mock

import pytest

import terrareg.namespace_type
from terrareg.provider_category_model import ProviderCategory
from terrareg.provider_tier import ProviderTier
from terrareg.user_group_namespace_permission_type import UserGroupNamespacePermissionType
from test.integration.terrareg import TerraregIntegrationTest
from test import client, app_context, test_request_context
import terrareg.provider_search
from test.integration.terrareg.fixtures import test_github_provider_source
from terrareg.auth.admin_session_auth_method import AdminSessionAuthMethod
import terrareg.auth.github_auth_method
import terrareg.repository_model
import terrareg.database
import terrareg.provider_version_model
import terrareg.provider_model
import terrareg.provider_source.factory
import terrareg.models


@pytest.fixture
def test_repository_create(test_request_context):
    """Create test repository"""
    @contextmanager
    def inner(provider_source):
        with test_request_context:
            repository = terrareg.repository_model.Repository.create(
                provider_source=provider_source,
                provider_id="125563113",
                name="terraform-provider-publish",
                description="Repo for unit test",
                owner="initial-providers",
                clone_url="https://github.example.com/initial-providers/terraform-provider-publish",
                logo_url="https://example.com/publish.png",
            )
        try:
            yield repository
        finally:
            with test_request_context:
                db = terrareg.database.Database.get()
                with db.get_connection() as conn:
                    conn.execute(db.provider.delete().where(db.provider.c.repository_id==repository.pk))
                    conn.execute(db.repository.delete().where(db.repository.c.provider_id=="125563113"))
    return inner


class TestApiGithubRepositoryPublishProvider(TerraregIntegrationTest):
    """Test GithubRepositoryPublishProvider endpoint"""

    def test_invalid_provider_source(self, client, test_github_provider_source, test_repository_create):
        """Test endpoint with invalid provider source"""
        with unittest.mock.patch('terrareg.provider_source.github.GithubProviderSource.get_github_app_installation_id', unittest.mock.MagicMock(return_value='unittestinstallationid')) as mock_get_github_app_installation_id, \
                unittest.mock.patch('terrareg.csrf.check_csrf_token', unittest.mock.MagicMock(return_value=True)) as mock_check_csrf, \
                test_repository_create(test_github_provider_source) as test_repository, \
                unittest.mock.patch('terrareg.provider_model.Provider.refresh_versions', unittest.mock.MagicMock(return_value=[terrareg.provider_version_model.ProviderVersion(provider=None, version="5.2.1")])) as mock_refresh_versions:

            res = client.post(f"/doesnotexist/repositories/125563113/publish-provider", data={"csrf_token": "test", "category_id": 523})
            assert res.status_code == 404
            assert res.json == {'errors': ['Not Found']}
            mock_check_csrf.assert_called_once_with('test')
            mock_get_github_app_installation_id.assert_not_called()
            mock_refresh_versions.assert_not_called()


    def test_unauthenticated(self, client, test_github_provider_source, test_repository_create):
        """Test Endpoint without authentication"""
        self._get_current_auth_method_mock.stop()
        with unittest.mock.patch('terrareg.provider_source.github.GithubProviderSource.get_github_app_installation_id', unittest.mock.MagicMock(return_value='unittestinstallationid')) as mock_get_github_app_installation_id, \
                unittest.mock.patch('terrareg.csrf.check_csrf_token', unittest.mock.MagicMock(return_value=True)) as mock_check_csrf, \
                test_repository_create(test_github_provider_source) as test_repository, \
                unittest.mock.patch('terrareg.provider_model.Provider.refresh_versions', unittest.mock.MagicMock(return_value=[terrareg.provider_version_model.ProviderVersion(provider=None, version="5.2.1")])) as mock_refresh_versions:

            res = client.post(f"/test-github-provider/repositories/125563113/publish-provider", data={"csrf_token": "test", "category_id": 523})
            assert res.status_code == 401
            mock_check_csrf.assert_not_called()
            mock_get_github_app_installation_id.assert_not_called()
            mock_refresh_versions.assert_not_called()

    def test_non_existent_repository(self, client, test_github_provider_source, test_repository_create):
        """Test Endpoint with non-existent repository ID"""
        self._get_current_auth_method_mock.stop()
        mock_get_current_auth_method = unittest.mock.MagicMock(return_value=AdminSessionAuthMethod())

        with unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', mock_get_current_auth_method), \
                unittest.mock.patch('terrareg.provider_source.github.GithubProviderSource.get_github_app_installation_id', unittest.mock.MagicMock(return_value='unittestinstallationid')) as mock_get_github_app_installation_id, \
                unittest.mock.patch('terrareg.csrf.check_csrf_token', unittest.mock.MagicMock(return_value=True)) as mock_check_csrf, \
                test_repository_create(test_github_provider_source) as test_repository, \
                unittest.mock.patch('terrareg.provider_model.Provider.refresh_versions', unittest.mock.MagicMock(return_value=[terrareg.provider_version_model.ProviderVersion(provider=None, version="5.2.1")])) as mock_refresh_versions:

            res = client.post(f"/test-github-provider/repositories/12345/publish-provider", data={"csrf_token": "test", "category_id": 523})
            assert res.status_code == 404
            assert res.json == {'message': 'Repository does not exist', 'status': 'Error'}
            mock_check_csrf.assert_called_once_with('test')
            # Assert not called when using non-github authentication
            mock_get_github_app_installation_id.assert_not_called()
            mock_refresh_versions.assert_not_called()

    def test_invalid_category_id(self, client, test_github_provider_source, test_repository_create, test_request_context):
        """Test endpoint with invalid category ID"""
        self._get_current_auth_method_mock.stop()
        mock_get_current_auth_method = unittest.mock.MagicMock(return_value=AdminSessionAuthMethod())

        with unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', mock_get_current_auth_method), \
                unittest.mock.patch('terrareg.provider_source.github.GithubProviderSource.get_github_app_installation_id', unittest.mock.MagicMock(return_value='unittestinstallationid')) as mock_get_github_app_installation_id, \
                unittest.mock.patch('terrareg.csrf.check_csrf_token', unittest.mock.MagicMock(return_value=True)) as mock_check_csrf, \
                unittest.mock.patch('terrareg.provider_model.Provider.create', unittest.mock.MagicMock(side_effect=terrareg.provider_model.Provider.create)) as mock_provider_create, \
                test_repository_create(test_github_provider_source) as test_repository, \
                unittest.mock.patch('terrareg.provider_model.Provider.refresh_versions', unittest.mock.MagicMock(return_value=[terrareg.provider_version_model.ProviderVersion(provider=None, version="5.2.1")])) as mock_refresh_versions:

            res = client.post(f"/test-github-provider/repositories/125563113/publish-provider", data={"csrf_token": "test", "category_id": 135135135})
            assert res.status_code == 400
            assert res.json == {'message': 'Provider Category does not exist', 'status': 'Error'}

            mock_check_csrf.assert_called_once_with('test')
            mock_provider_create.assert_not_called()
            mock_get_github_app_installation_id.assert_not_called()
            mock_refresh_versions.assert_not_called()


    def test_non_user_selectable_category(self, client, test_github_provider_source, test_repository_create, test_request_context):
        """Test endpoint with non-user-selectable category ID"""
        self._get_current_auth_method_mock.stop()
        mock_get_current_auth_method = unittest.mock.MagicMock(return_value=AdminSessionAuthMethod())

        with unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', mock_get_current_auth_method), \
                unittest.mock.patch('terrareg.provider_source.github.GithubProviderSource.get_github_app_installation_id', unittest.mock.MagicMock(return_value='unittestinstallationid')) as mock_get_github_app_installation_id, \
                unittest.mock.patch('terrareg.csrf.check_csrf_token', unittest.mock.MagicMock(return_value=True)) as mock_check_csrf, \
                unittest.mock.patch('terrareg.provider_model.Provider.create', unittest.mock.MagicMock(side_effect=terrareg.provider_model.Provider.create)) as mock_provider_create, \
                test_repository_create(test_github_provider_source) as test_repository, \
                unittest.mock.patch('terrareg.provider_model.Provider.refresh_versions', unittest.mock.MagicMock(return_value=[terrareg.provider_version_model.ProviderVersion(provider=None, version="5.2.1")])) as mock_refresh_versions:

            res = client.post(f"/test-github-provider/repositories/125563113/publish-provider", data={"csrf_token": "test", "category_id": 99})
            assert res.status_code == 400
            assert res.json == {'message': 'Provider Category does not exist', 'status': 'Error'}

            mock_check_csrf.assert_called_once_with('test')
            mock_provider_create.assert_not_called()
            mock_get_github_app_installation_id.assert_not_called()
            mock_refresh_versions.assert_not_called()

    def test_provider_already_published(self, client, test_github_provider_source, test_repository_create, test_request_context):
        """Test endpoint with provider that has already been published"""
        self._get_current_auth_method_mock.stop()
        mock_get_current_auth_method = unittest.mock.MagicMock(return_value=AdminSessionAuthMethod())

        with unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', mock_get_current_auth_method), \
                unittest.mock.patch('terrareg.provider_source.github.GithubProviderSource.get_github_app_installation_id', unittest.mock.MagicMock(return_value='unittestinstallationid')) as mock_get_github_app_installation_id, \
                unittest.mock.patch('terrareg.csrf.check_csrf_token', unittest.mock.MagicMock(return_value=True)) as mock_check_csrf, \
                unittest.mock.patch('terrareg.provider_model.Provider.create', unittest.mock.MagicMock(side_effect=terrareg.provider_model.Provider.create)) as mock_provider_create, \
                test_repository_create(test_github_provider_source) as test_repository, \
                unittest.mock.patch('terrareg.provider_model.Provider.refresh_versions', unittest.mock.MagicMock(return_value=[terrareg.provider_version_model.ProviderVersion(provider=None, version="5.2.1")])) as mock_refresh_versions:

            with test_request_context:
                terrareg.provider_model.Provider.create(repository=test_repository, provider_category=ProviderCategory.get_by_pk(523), use_default_provider_source_auth=False, tier=ProviderTier.COMMUNITY)
            mock_provider_create.reset_mock()
            mock_get_github_app_installation_id.reset_mock()
            mock_refresh_versions.reset_mock()

            res = client.post(f"/test-github-provider/repositories/125563113/publish-provider", data={"csrf_token": "test", "category_id": 523})
            assert res.status_code == 500
            assert res.json == {'message': 'A duplicate provider exists with the same name in the namespace', 'status': 'Error'}

            mock_check_csrf.assert_called_once_with('test')
            mock_provider_create.assert_called_once_with(repository=test_repository, provider_category=ProviderCategory.get_by_pk(pk=523), use_default_provider_source_auth=True, tier=ProviderTier.COMMUNITY)
            mock_get_github_app_installation_id.assert_not_called()
            mock_refresh_versions.assert_not_called()

    def test_authenticated_with_admin(self, client, test_github_provider_source, test_repository_create):
        """Test Endpoint whilst authenticated with admin session"""
        self._get_current_auth_method_mock.stop()
        mock_get_current_auth_method = unittest.mock.MagicMock(return_value=AdminSessionAuthMethod())

        with unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', mock_get_current_auth_method), \
                unittest.mock.patch('terrareg.provider_source.github.GithubProviderSource.get_github_app_installation_id', unittest.mock.MagicMock(return_value='unittestinstallationid')) as mock_get_github_app_installation_id, \
                unittest.mock.patch('terrareg.csrf.check_csrf_token', unittest.mock.MagicMock(return_value=True)) as mock_check_csrf, \
                test_repository_create(test_github_provider_source) as test_repository, \
                unittest.mock.patch('terrareg.provider_model.Provider.refresh_versions', unittest.mock.MagicMock(return_value=[terrareg.provider_version_model.ProviderVersion(provider=None, version="5.2.1")])) as mock_refresh_versions:

            res = client.post(f"/test-github-provider/repositories/125563113/publish-provider", data={"csrf_token": "test", "category_id": 523})
            assert res.status_code == 200
            assert res.json == {'name': 'publish', 'namespace': 'initial-providers'}
            mock_check_csrf.assert_called_once_with('test')
            # Assert not called when using non-github authentication
            mock_get_github_app_installation_id.assert_not_called()
            mock_refresh_versions.assert_called_once_with(limit=1)

    def test_github_authenticated(self, client, test_github_provider_source, test_repository_create):
        """Test Endpoint whilst authenticated with github"""
        self._get_current_auth_method_mock.stop()
        mock_get_current_auth_method = unittest.mock.MagicMock(return_value=terrareg.auth.github_auth_method.GithubAuthMethod())

        with unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', mock_get_current_auth_method), \
                unittest.mock.patch('terrareg.provider_source.github.GithubProviderSource.get_github_app_installation_id', unittest.mock.MagicMock(return_value='unittestinstallationid')) as mock_get_github_app_installation_id, \
                unittest.mock.patch('terrareg.provider_source.github.GithubProviderSource.auto_generate_github_organisation_namespaces', True), \
                unittest.mock.patch('terrareg.csrf.check_csrf_token', unittest.mock.MagicMock(return_value=True)) as mock_check_csrf, \
                unittest.mock.patch('terrareg.provider_model.Provider.create', unittest.mock.MagicMock(side_effect=terrareg.provider_model.Provider.create)) as mock_provider_create, \
                test_repository_create(test_github_provider_source) as test_repository, \
                unittest.mock.patch('terrareg.provider_model.Provider.refresh_versions', unittest.mock.MagicMock(return_value=[terrareg.provider_version_model.ProviderVersion(provider=None, version="5.2.1")])) as mock_refresh_versions, \
                unittest.mock.patch('terrareg.auth.github_auth_method.GithubAuthMethod.check_namespace_access', unittest.mock.MagicMock(return_value=True)) as mock_check_namespace_access, \
                unittest.mock.patch('terrareg.auth.github_auth_method.GithubAuthMethod.get_current_instance',
                                    unittest.mock.MagicMock(return_value=terrareg.auth.github_auth_method.GithubAuthMethod())):

            res = client.post(f"/test-github-provider/repositories/125563113/publish-provider", data={"csrf_token": "test", "category_id": 523})
            assert res.status_code == 200
            assert res.json == {'name': 'publish', 'namespace': 'initial-providers'}

            mock_check_csrf.assert_called_once_with('test')
            mock_provider_create.assert_called_once_with(repository=test_repository, provider_category=ProviderCategory.get_by_pk(pk=523), use_default_provider_source_auth=False, tier=ProviderTier.COMMUNITY)
            mock_get_github_app_installation_id.assert_called_once_with(namespace=terrareg.models.Namespace.get("initial-providers"))
            mock_check_namespace_access.assert_called_once_with(UserGroupNamespacePermissionType.FULL, namespace='initial-providers')
            mock_refresh_versions.assert_called_once_with(limit=1)

    def test_github_without_permissions(self, client, test_github_provider_source, test_repository_create):
        """Test endpoint whilst authenticated with github without namespace access"""
        self._get_current_auth_method_mock.stop()
        mock_get_current_auth_method = unittest.mock.MagicMock(return_value=terrareg.auth.github_auth_method.GithubAuthMethod())

        with unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', mock_get_current_auth_method), \
                unittest.mock.patch('terrareg.provider_source.github.GithubProviderSource.get_github_app_installation_id', unittest.mock.MagicMock(return_value='unittestinstallationid')) as mock_get_github_app_installation_id, \
                unittest.mock.patch('terrareg.csrf.check_csrf_token', unittest.mock.MagicMock(return_value=True)) as mock_check_csrf, \
                unittest.mock.patch('terrareg.provider_model.Provider.create', unittest.mock.MagicMock(side_effect=terrareg.provider_model.Provider.create)) as mock_provider_create, \
                test_repository_create(test_github_provider_source) as test_repository, \
                unittest.mock.patch('terrareg.provider_model.Provider.refresh_versions', unittest.mock.MagicMock(return_value=[terrareg.provider_version_model.ProviderVersion(provider=None, version="5.2.1")])) as mock_refresh_versions, \
                unittest.mock.patch('terrareg.auth.github_auth_method.GithubAuthMethod.get_username', unittest.mock.MagicMock(return_value='unittestusername')), \
                unittest.mock.patch('terrareg.auth.github_auth_method.GithubAuthMethod.get_github_organisations', unittest.mock.MagicMock(return_value={
                    "moduleextraction": terrareg.namespace_type.NamespaceType.GITHUB_USER,
                    "does-not-exist": terrareg.namespace_type.NamespaceType.GITHUB_ORGANISATION
                })), \
                unittest.mock.patch('terrareg.auth.github_auth_method.GithubAuthMethod.get_current_instance',
                                    unittest.mock.MagicMock(return_value=terrareg.auth.github_auth_method.GithubAuthMethod())):

            res = client.post(f"/test-github-provider/repositories/125563113/publish-provider", data={"csrf_token": "test", "category_id": 523})
            assert res.status_code == 403
            assert res.json == {'message': "You don't have the permission to access the requested resource. It is either read-protected or not readable by the server."}

            mock_check_csrf.assert_not_called()
            mock_provider_create.assert_not_called()
            mock_get_github_app_installation_id.assert_not_called()
            mock_refresh_versions.assert_not_called()
