
import datetime
from unittest import mock
import pytest

from terrareg.user_group_namespace_permission_type import UserGroupNamespacePermissionType
from terrareg.auth import AuthenticationType, GithubAuthMethod
from test import BaseTest
from test.unit.terrareg import setup_test_data, mock_models
from test.unit.terrareg.auth.base_session_auth_method_tests import BaseSessionAuthMethodTests
from test.unit.terrareg.auth.base_sso_auth_method_tests import BaseSsoAuthMethodTests, test_data, user_group_data
import terrareg.models

# Required as this is used by BaseOpenidConnectAuthMethod
from test import test_request_context


class TestGithubAuthMethod(BaseSsoAuthMethodTests, BaseSessionAuthMethodTests):
    """Test methods of Github auth method"""

    CLS = GithubAuthMethod

    @pytest.mark.parametrize('enabled', [
        True,
        False
    ])
    def test_is_enabled(self, enabled):
        """test is_enabled method"""
        with mock.patch('terrareg.github.Github.is_enabled', mock.MagicMock(return_value=enabled)):
            assert GithubAuthMethod().is_enabled() is enabled

    @pytest.mark.parametrize('organisations_session_value, expected_response', [
        (None, []),
        ('', []),
        (['test-github-org'], ['test-github-org']),
        (['test-github-org', 'user1'], ['test-github-org', 'user1']),
    ])
    def test__get_organisation_memeberships(self, organisations_session_value, expected_response, test_request_context):
        """Test _get_organisation_memeberships"""
        self.SERVER._app.secret_key = "asecretkey"
        with mock.patch('terrareg.config.Config.SECRET_KEY', "asecretkey"), \
                test_request_context:
            if organisations_session_value is not None:
                test_request_context.session['organisations'] = organisations_session_value

            assert GithubAuthMethod()._get_organisation_memeberships() == expected_response

    def test_get_group_memberships(self):
        """Test get_group_memberships"""
        with mock.patch('terrareg.auth.github_auth_method.GithubAuthMethod._get_organisation_memeberships',
                        mock.MagicMock(return_value=['unittest-group-1', 'unittest-group-2'])):
            assert GithubAuthMethod().get_group_memberships() == ['unittest-group-1', 'unittest-group-2']

    @pytest.mark.parametrize('username_session_value, expected_value', [
        (None, None),
        ('', None),
        ('testusername', 'testusername')
    ])
    def test_get_username(self, username_session_value, expected_value, test_request_context):
        """test get_username."""
        self.SERVER._app.secret_key = "asecretkey"
        with mock.patch('terrareg.config.Config.SECRET_KEY', "asecretkey"), \
                test_request_context:
            if username_session_value is not None:
                test_request_context.session['github_username'] = username_session_value

            assert GithubAuthMethod().get_username() == expected_value

    @pytest.mark.parametrize('auto_generate_github_organisation_namespaces, organisations, super_response, expect_super_called, check_namespace, check_permission_type, expected_response', [
        # Auto generate enabled and valid namespace from list of github organisation
        (True, ['organisation1', 'organisation2'], False, False, 'organisation2', UserGroupNamespacePermissionType.MODIFY, True),
        (True, ['organisation1', 'organisation2'], False, False, 'organisation2', UserGroupNamespacePermissionType.FULL, True),
        # Auto generate enabled, not valid namespace from github organisations, with varying super response
        (True, ['organisation1', 'organisation2'], True, True, 'organisation3', UserGroupNamespacePermissionType.MODIFY, True),
        (True, ['organisation1', 'organisation2'], True, True, 'organisation3', UserGroupNamespacePermissionType.FULL, True),
        (True, ['organisation1', 'organisation2'], False, True, 'organisation3', UserGroupNamespacePermissionType.MODIFY, False),
        (True, ['organisation1', 'organisation2'], False, True, 'organisation3', UserGroupNamespacePermissionType.FULL, False),

        # No organisations defined
        (True, None, False, True, 'organisation2', UserGroupNamespacePermissionType.FULL, False),
        (True, None, True, True, 'organisation2', UserGroupNamespacePermissionType.FULL, True),

        # Test with auto generated disabled
        (False, ['organisation1', 'organisation2'], True, True, 'organisation2', UserGroupNamespacePermissionType.MODIFY, True),
        (False, ['organisation1', 'organisation2'], True, True, 'organisation2', UserGroupNamespacePermissionType.FULL, True),
        (False, ['organisation1', 'organisation2'], False, True, 'organisation2', UserGroupNamespacePermissionType.MODIFY, False),
        (False, ['organisation1', 'organisation2'], False, True, 'organisation2', UserGroupNamespacePermissionType.FULL, False),

    ])
    def test_check_namespace_access(self, auto_generate_github_organisation_namespaces,
                                    organisations, super_response, expect_super_called, check_namespace,
                                    check_permission_type, expected_response, test_request_context):
        """Test check_namespace_access"""
        with mock.patch('terrareg.config.Config.AUTO_GENERATE_GITHUB_ORGANISATION_NAMESPACES', auto_generate_github_organisation_namespaces), \
                mock.patch('terrareg.auth.base_sso_auth_method.BaseSsoAuthMethod.check_namespace_access', mock.MagicMock(return_value=super_response)) as mock_check_namespace_access, \
                test_request_context:
            if organisations is not None:
                test_request_context.session['organisations'] = organisations

            obj = GithubAuthMethod()
            assert obj.check_namespace_access(permission_type=check_permission_type, namespace=check_namespace) is expected_response

            if expect_super_called:
                mock_check_namespace_access.assert_called_once_with(permission_type=check_permission_type, namespace=check_namespace)
            else:
                mock_check_namespace_access.assert_not_called()

    @pytest.mark.parametrize('auto_generate_github_organisation_namespaces, github_organisations, super_permissions, expected_permissions', [
        (True, None, {}, {}),
        (True, ['test-org1'], {}, {'test-org1': UserGroupNamespacePermissionType.FULL}),
        (True, ['test-org1'], {
            'testnamespace': UserGroupNamespacePermissionType.FULL,
            'moduledetails': UserGroupNamespacePermissionType.MODIFY
         }, {
            'test-org1': UserGroupNamespacePermissionType.FULL,
            'testnamespace': UserGroupNamespacePermissionType.FULL,
            'moduledetails': UserGroupNamespacePermissionType.MODIFY
        }),

        (False, None, {}, {}),
        (False, ['test-org1'], {}, {}),
        (False, ['test-org1'], {
            'testnamespace': UserGroupNamespacePermissionType.FULL,
            'moduledetails': UserGroupNamespacePermissionType.MODIFY
         }, {
            'testnamespace': UserGroupNamespacePermissionType.FULL,
            'moduledetails': UserGroupNamespacePermissionType.MODIFY
        }),

    ])
    @setup_test_data()
    def test_get_all_namespace_permissions(self, auto_generate_github_organisation_namespaces,
                                           github_organisations, super_permissions, expected_permissions,
                                           test_request_context, mock_models):
        """test test_get_all_namespace_permissions"""
        get_all_namespace_permissions_response = {
                terrareg.models.Namespace.get(namespace_name): permission
                for namespace_name, permission in super_permissions.items()
            }
        with mock.patch('terrareg.config.Config.AUTO_GENERATE_GITHUB_ORGANISATION_NAMESPACES', auto_generate_github_organisation_namespaces), \
                mock.patch('terrareg.auth.base_sso_auth_method.BaseSsoAuthMethod.get_all_namespace_permissions',
                           mock.MagicMock(return_value=get_all_namespace_permissions_response)) as mock_get_all_namespace_permissions, \
                test_request_context:
            if github_organisations is not None:
                test_request_context.session['organisations'] = github_organisations

            obj = GithubAuthMethod()
            assert obj.get_all_namespace_permissions() == {
                terrareg.models.Namespace.get(namespace_name): permission
                for namespace_name, permission in expected_permissions.items()
            }

            mock_get_all_namespace_permissions.assert_called_once_with()

            # Ensure namespace were created, if auto_generate_github_organisation_namespaces
            # is enabled, otherwise, ensure they're not
            for org_name in (github_organisations or []):
                if auto_generate_github_organisation_namespaces:
                    assert terrareg.models.Namespace.get(org_name) is not None
                else:
                    assert terrareg.models.Namespace.get(org_name) is None

    @pytest.mark.parametrize('auth_type,expected_result', [
        (None, False),
        (AuthenticationType.NOT_AUTHENTICATED, False),
        (AuthenticationType.NOT_CHECKED, False),
        (AuthenticationType.AUTHENTICATION_TOKEN, False),
        (AuthenticationType.SESSION_PASSWORD, False),
        (AuthenticationType.SESSION_OPENID_CONNECT, False),
        (AuthenticationType.SESSION_SAML, False),
        (AuthenticationType.SESSION_GITHUB, True),
    ])
    def test_check_session_auth_type(self, auth_type, expected_result, test_request_context):
        """Test check_session_auth_type"""
        self.SERVER._app.secret_key = "asecretkey"
        with mock.patch('terrareg.config.Config.SECRET_KEY', "asecretkey"), \
                test_request_context:
            if auth_type:
                test_request_context.session['authentication_type'] = auth_type.value
                test_request_context.session.modified = True

            obj = GithubAuthMethod()
            assert obj.check_session_auth_type() == expected_result

    @pytest.mark.parametrize('username, expected_result', [
        (None, False),
        ('', False),
        ('test-user', True)
    ])
    def test_check_session(self, username, expected_result, test_request_context):
        """Test check_session method"""
        self.SERVER._app.secret_key = "asecretkey"
        with mock.patch('terrareg.config.Config.SECRET_KEY', "asecretkey"), \
                test_request_context:
            if username:
                test_request_context.session['github_username'] = username
                test_request_context.session.modified = True

            obj = GithubAuthMethod()
            assert obj.check_session() == expected_result