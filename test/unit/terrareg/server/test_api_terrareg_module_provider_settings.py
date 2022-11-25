
import unittest.mock

import pytest

import terrareg.audit_action
from test.unit.terrareg import (
    MockModule,
    MockModuleProvider,
    MockNamespace,
    mocked_server_namespace_fixture,
    setup_test_data, TerraregUnitTest
)
from terrareg.auth import UserGroupNamespacePermissionType
from test import client, app_context, test_request_context



class TestApiTerraregModuleProviderSettings(TerraregUnitTest):
    """Test module provider settings endpoint"""

    def _get_mock_get_current_auth_method(self, allowed_to_create):
        """Return mock auth method"""
        mock_auth_method = unittest.mock.MagicMock()
        mock_auth_method.check_namespace_access = unittest.mock.MagicMock(return_value=allowed_to_create)
        mock_auth_method.get_username.return_value = 'mps-user'
        mock_get_current_auth_method = unittest.mock.MagicMock(return_value=mock_auth_method)
        return mock_get_current_auth_method, mock_auth_method

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
        mock_get_auth_method, mock_auth_method = self._get_mock_get_current_auth_method(True)
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', mock_get_auth_method), \
                unittest.mock.patch('terrareg.audit.AuditEvent.create_audit_event') as mock_create_audit_event, \
                unittest.mock.patch('terrareg.server.check_csrf_token', return_value=True) as mock_check_csrf, \
                unittest.mock.patch('test.unit.terrareg.MockModuleProvider.update_attributes') as mock_update_attributes:

            res = client.post(
                '/v1/terrareg/modules/testnamespace/testmodulename/testprovider/settings',
                json={
                    'repo_clone_url_template': repository_url,
                    'csrf_token': 'unittestcsrf'
                }
            )
            assert res.json == {}
            assert res.status_code == 200

            # Ensure required checks are called
            mock_check_csrf.assert_called_once_with('unittestcsrf')
            mock_auth_method.check_namespace_access.assert_called_once_with(UserGroupNamespacePermissionType.MODIFY, namespace='testnamespace')
            mock_update_attributes.assert_called_once_with(repo_clone_url_template=repository_url)
            mock_create_audit_event.assert_called_once_with(
                action=terrareg.audit_action.AuditAction.MODULE_PROVIDER_UPDATE_GIT_CUSTOM_CLONE_URL,
                object_type='MockModuleProvider',
                object_id='testnamespace/testmodulename/testprovider',
                old_value=None, new_value=repository_url
            )

    @setup_test_data()
    def test_update_repository_url_invalid_protocol(self, app_context, test_request_context, mocked_server_namespace_fixture, client):
        """Test update of repository URL with invalid protocol."""
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._get_mock_get_current_auth_method(True)[0]), \
                unittest.mock.patch('terrareg.audit.AuditEvent.create_audit_event') as mock_create_audit_event, \
                unittest.mock.patch('terrareg.server.check_csrf_token', return_value=True) as mock_check_csrf, \
                unittest.mock.patch('test.unit.terrareg.MockModuleProvider.update_attributes') as mock_update_attributes:

            res = client.post(
                '/v1/terrareg/modules/testnamespace/testmodulename/testprovider/settings',
                json={
                    'repo_clone_url_template': 'nope://unittest.com/module.git',
                    'csrf_token': 'unittestcsrf'
                }
            )
            assert res.json == {'message': 'Repo clone URL: Repository URL contains an unknown scheme (e.g. https/ssh/http)'}
            assert res.status_code == 400

            # Ensure required checks are called
            mock_check_csrf.assert_called_once_with('unittestcsrf')
            mock_update_attributes.assert_not_called()
            mock_create_audit_event.assert_not_called()

    @setup_test_data()
    def test_update_repository_url_invalid_domain(self, app_context, test_request_context, mocked_server_namespace_fixture, client):
        """Test update of repository URL with an invalid domain."""
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._get_mock_get_current_auth_method(True)[0]), \
                unittest.mock.patch('terrareg.audit.AuditEvent.create_audit_event') as mock_create_audit_event, \
                unittest.mock.patch('terrareg.server.check_csrf_token', return_value=True) as mock_check_csrf, \
                unittest.mock.patch('test.unit.terrareg.MockModuleProvider.update_attributes') as mock_update_attributes:

            res = client.post(
                '/v1/terrareg/modules/testnamespace/testmodulename/testprovider/settings',
                json={
                    'repo_clone_url_template': 'https:///module.git',
                    'csrf_token': 'unittestcsrf'
                }
            )
            assert res.json == {'message': 'Repo clone URL: Repository URL does not contain a host/domain'}
            assert res.status_code == 400

            # Ensure required checks are called
            mock_check_csrf.assert_called_once_with('unittestcsrf')
            mock_update_attributes.assert_not_called()
            mock_create_audit_event.assert_not_called()

    @setup_test_data()
    def test_update_repository_url_without_path(self, app_context, test_request_context, mocked_server_namespace_fixture, client):
        """Test update of repository URL without a path."""
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._get_mock_get_current_auth_method(True)[0]), \
                unittest.mock.patch('terrareg.audit.AuditEvent.create_audit_event') as mock_create_audit_event, \
                unittest.mock.patch('terrareg.server.check_csrf_token', return_value=True) as mock_check_csrf, \
                unittest.mock.patch('test.unit.terrareg.MockModuleProvider.update_attributes') as mock_update_attributes:

            res = client.post(
                '/v1/terrareg/modules/testnamespace/testmodulename/testprovider/settings',
                json={
                    'repo_clone_url_template': 'https://example.com',
                    'csrf_token': 'unittestcsrf'
                }
            )
            assert res.json == {'message': 'Repo clone URL: Repository URL does not contain a path'}
            assert res.status_code == 400

            # Ensure required checks are called
            mock_check_csrf.assert_called_once_with('unittestcsrf')
            mock_update_attributes.assert_not_called()
            mock_create_audit_event.assert_not_called()

    @setup_test_data()
    def test_update_repository_without_csrf(self, app_context, test_request_context, mocked_server_namespace_fixture, client):
        """Test update of repository URL without a CSRF token."""
        mock_get_auth_method, mock_auth_method = self._get_mock_get_current_auth_method(True)
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', mock_get_auth_method), \
                unittest.mock.patch('terrareg.audit.AuditEvent.create_audit_event') as mock_create_audit_event, \
                unittest.mock.patch('test.unit.terrareg.MockModuleProvider.update_attributes') as mock_update_attributes:

            res = client.post(
                '/v1/terrareg/modules/testnamespace/testmodulename/testprovider/settings',
                json={
                    'repo_clone_url_template': 'https://example.com/test.git'
                }
            )
            assert res.json == {'message': 'No session is presesnt to check CSRF token', 'status': 'Error'}
            assert res.status_code == 500

            # Ensure required checks are called
            mock_update_attributes.assert_not_called()
            mock_create_audit_event.assert_not_called()

    @setup_test_data()
    def test_update_repository_invalid_auth(self, app_context, test_request_context, mocked_server_namespace_fixture, client):
        """Test update of repository URL without a CSRF token."""
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._get_mock_get_current_auth_method(False)[0]), \
                unittest.mock.patch('terrareg.audit.AuditEvent.create_audit_event') as mock_create_audit_event, \
                unittest.mock.patch('terrareg.server.check_csrf_token', return_value=True) as mock_check_csrf, \
                unittest.mock.patch('test.unit.terrareg.MockModuleProvider.update_attributes') as mock_update_attributes:

            res = client.post(
                '/v1/terrareg/modules/testnamespace/testmodulename/testprovider/settings',
                json={
                    'repo_clone_url_template': 'https://example.com/test.git'
                }
            )
            assert res.json == {'message': "You don't have the permission to access the requested resource. "
                                           "It is either read-protected or not readable by the server."}
            assert res.status_code == 403

            # Ensure required checks are called
            mock_update_attributes.assert_not_called()
            mock_create_audit_event.assert_not_called()

    @setup_test_data()
    def test_update_git_tag_format(
            self, app_context,
            test_request_context, mocked_server_namespace_fixture,
            client
        ):
        """Test update of git tag format."""
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._get_mock_get_current_auth_method(True)[0]), \
                unittest.mock.patch('terrareg.audit.AuditEvent.create_audit_event') as mock_create_audit_event, \
                unittest.mock.patch('terrareg.server.check_csrf_token', return_value=True) as mock_check_csrf, \
                unittest.mock.patch('test.unit.terrareg.MockModuleProvider.update_attributes') as mock_update_attributes:

            res = client.post(
                '/v1/terrareg/modules/testnamespace/testmodulename/testprovider/settings',
                json={
                    'git_tag_format': 'unittest{version}gittagformat',
                    'csrf_token': 'unittestcsrf'
                }
            )
            assert res.json == {}
            assert res.status_code == 200

            # Ensure required checks are called
            mock_check_csrf.assert_called_once_with('unittestcsrf')
            mock_update_attributes.assert_called_once_with(git_tag_format='unittest{version}gittagformat')
            mock_create_audit_event.assert_called_once_with(
                action=terrareg.audit_action.AuditAction.MODULE_PROVIDER_UPDATE_GIT_TAG_FORMAT,
                object_type='MockModuleProvider',
                object_id='testnamespace/testmodulename/testprovider',
                old_value='{version}',
                new_value='unittest{version}gittagformat'
            )

    @setup_test_data()
    def test_update_empty_git_tag_format(
            self, app_context,
            test_request_context, mocked_server_namespace_fixture,
            client
        ):
        """Test update of git tag format with empty value."""
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._get_mock_get_current_auth_method(True)[0]), \
                unittest.mock.patch('terrareg.audit.AuditEvent.create_audit_event') as mock_create_audit_event, \
                unittest.mock.patch('terrareg.server.check_csrf_token', return_value=True) as mock_check_csrf, \
                unittest.mock.patch('test.unit.terrareg.MockModuleProvider.update_attributes') as mock_update_attributes:

            res = client.post(
                '/v1/terrareg/modules/moduleextraction/gitextraction/complexgittagformat/settings',
                json={
                    'git_tag_format': '',
                    'csrf_token': 'unittestcsrf'
                }
            )
            assert res.json == {}
            assert res.status_code == 200

            # Ensure required checks are called
            mock_check_csrf.assert_called_once_with('unittestcsrf')
            mock_update_attributes.assert_called_once_with(git_tag_format=None)
            mock_create_audit_event.assert_called_once_with(
                action=terrareg.audit_action.AuditAction.MODULE_PROVIDER_UPDATE_GIT_TAG_FORMAT,
                object_type='MockModuleProvider',
                object_id='moduleextraction/gitextraction/complexgittagformat',
                old_value='unittest{version}value',
                new_value=None
            )

    @pytest.mark.parametrize('verified_state', [True, False])
    @setup_test_data()
    def test_update_verified_flag(
            self, verified_state, app_context,
            test_request_context, mocked_server_namespace_fixture,
            client
        ):
        """Test update of verified flag."""
        ns = MockNamespace.get('testnamespace')
        module = MockModule(ns, 'testmodulename')
        provider = MockModuleProvider.get(module, 'testprovider')
        provider.update_attributes(verified=(not verified_state))

        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._get_mock_get_current_auth_method(True)[0]), \
                unittest.mock.patch('terrareg.audit.AuditEvent.create_audit_event') as mock_create_audit_event, \
                unittest.mock.patch('terrareg.server.check_csrf_token', return_value=True) as mock_check_csrf, \
                unittest.mock.patch('test.unit.terrareg.MockModuleProvider.update_attributes') as mock_update_attributes:

            res = client.post(
                '/v1/terrareg/modules/testnamespace/testmodulename/testprovider/settings',
                json={
                    'verified': verified_state,
                    'csrf_token': 'unittestcsrf'
                }
            )
            assert res.json == {}
            assert res.status_code == 200

            # Ensure required checks are called
            mock_check_csrf.assert_called_once_with('unittestcsrf')
            mock_update_attributes.assert_called_with(verified=verified_state)
            mock_create_audit_event.assert_called_once_with(
                action=terrareg.audit_action.AuditAction.MODULE_PROVIDER_UPDATE_VERIFIED,
                object_type='MockModuleProvider',
                object_id='testnamespace/testmodulename/testprovider',
                old_value=(not verified_state), new_value=verified_state
            )

    @pytest.mark.parametrize('verified_state', ['', 'isastring'])
    @setup_test_data()
    def test_update_verified_flag_invalid_value(
            self, verified_state, app_context,
            test_request_context, mocked_server_namespace_fixture,
            client
        ):
        """Test update of verified flag with invalid value."""
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._get_mock_get_current_auth_method(True)[0]), \
                unittest.mock.patch('terrareg.audit.AuditEvent.create_audit_event') as mock_create_audit_event, \
                unittest.mock.patch('terrareg.server.check_csrf_token', return_value=True) as mock_check_csrf, \
                unittest.mock.patch('test.unit.terrareg.MockModuleProvider.update_attributes') as mock_update_attributes:

            res = client.post(
                '/v1/terrareg/modules/testnamespace/testmodulename/testprovider/settings',
                json={
                    'verified': verified_state,
                    'csrf_token': 'unittestcsrf'
                }
            )
            assert res.json == {'message': {'verified': 'Whether module provider is marked as verified.'}}
            assert res.status_code == 400

            # Ensure required checks are called
            mock_check_csrf.assert_not_called()
            mock_update_attributes.assert_not_called()
            mock_create_audit_event.assert_not_called()
