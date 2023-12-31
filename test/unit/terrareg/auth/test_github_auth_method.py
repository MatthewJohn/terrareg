
import datetime
from unittest import mock
import pytest

from terrareg.user_group_namespace_permission_type import UserGroupNamespacePermissionType
from terrareg.auth import AuthenticationType, GithubAuthMethod
from test import BaseTest
from test.unit.terrareg import setup_test_data, mock_models
from test.unit.terrareg.auth.base_session_auth_method_tests import BaseSessionAuthMethodTests
from test.unit.terrareg.auth.base_sso_auth_method_tests import BaseSsoAuthMethodTests, test_data, user_group_data
import terrareg.namespace_type
import terrareg.models
from test.integration.terrareg.fixtures import (
    test_github_provider_source
)

# Required as this is used by BaseOpenidConnectAuthMethod
from test import test_request_context


class TestGithubAuthMethod(BaseSsoAuthMethodTests, BaseSessionAuthMethodTests):
    """Test methods of Github auth method"""

    CLS = GithubAuthMethod

    def test_is_enabled(self):
        """test is_enabled method"""
        assert GithubAuthMethod().is_enabled() is True

    @pytest.mark.parametrize('organisations_session_value, expected_response', [
        (None, {}),
        ('', {}),
        ({'test-github-org': terrareg.namespace_type.NamespaceType.GITHUB_ORGANISATION.value}, {'test-github-org': terrareg.namespace_type.NamespaceType.GITHUB_ORGANISATION}),
        ({'test-github-org': terrareg.namespace_type.NamespaceType.GITHUB_ORGANISATION.value, 'user1': terrareg.namespace_type.NamespaceType.GITHUB_USER.value},
         {'test-github-org': terrareg.namespace_type.NamespaceType.GITHUB_ORGANISATION, 'user1': terrareg.namespace_type.NamespaceType.GITHUB_USER}),
    ])
    def test__get_organisation_memeberships(self, organisations_session_value, expected_response, test_github_provider_source, test_request_context):
        """Test _get_organisation_memeberships"""
        self.SERVER._app.secret_key = "asecretkey"
        with mock.patch('terrareg.config.Config.SECRET_KEY', "asecretkey"), \
                test_request_context:

            test_request_context.session['provider_source'] = test_github_provider_source.name

            if organisations_session_value is not None:
                test_request_context.session['organisations'] = organisations_session_value

            assert GithubAuthMethod()._get_organisation_memeberships() == expected_response

    def test_get_group_memberships(self, test_github_provider_source, test_request_context):
        """Test get_group_memberships"""
        self.SERVER._app.secret_key = "asecretkey"
        with mock.patch('terrareg.config.Config.SECRET_KEY', "asecretkey"), \
                test_request_context:
            
            test_request_context.session['provider_source'] = test_github_provider_source.name
            test_request_context.session['organisations'] = {'unittest-group-1': terrareg.namespace_type.NamespaceType.GITHUB_ORGANISATION.value,
                                                             'unittest-group-2': terrareg.namespace_type.NamespaceType.GITHUB_ORGANISATION.value}
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

    @pytest.mark.parametrize('valid_provider_source, auto_generate_github_organisation_namespaces, organisations, super_response, expect_super_called, check_namespace, check_permission_type, expected_response', [
        # Auto generate enabled and valid namespace from list of github organisation
        (True, True, {'organisation1': terrareg.namespace_type.NamespaceType.GITHUB_ORGANISATION,
                      'organisation2': terrareg.namespace_type.NamespaceType.GITHUB_ORGANISATION},
         False, False, 'organisation2', UserGroupNamespacePermissionType.MODIFY, True),
        (True, True, {'organisation1': terrareg.namespace_type.NamespaceType.GITHUB_ORGANISATION,
                      'organisation2': terrareg.namespace_type.NamespaceType.GITHUB_ORGANISATION},
         False, False, 'organisation2', UserGroupNamespacePermissionType.FULL, True),
        # Auto generate enabled, not valid namespace from github organisations, with varying super response
        (True, True, {'organisation1': terrareg.namespace_type.NamespaceType.GITHUB_ORGANISATION,
                      'organisation2': terrareg.namespace_type.NamespaceType.GITHUB_ORGANISATION},
         True, True, 'organisation3', UserGroupNamespacePermissionType.MODIFY, True),
        (True, True, {'organisation1': terrareg.namespace_type.NamespaceType.GITHUB_ORGANISATION,
                      'organisation2': terrareg.namespace_type.NamespaceType.GITHUB_ORGANISATION},
         True, True, 'organisation3', UserGroupNamespacePermissionType.FULL, True),
        (True, True, {'organisation1': terrareg.namespace_type.NamespaceType.GITHUB_ORGANISATION,
                      'organisation2': terrareg.namespace_type.NamespaceType.GITHUB_ORGANISATION},
         False, True, 'organisation3', UserGroupNamespacePermissionType.MODIFY, False),
        (True, True, {'organisation1': terrareg.namespace_type.NamespaceType.GITHUB_ORGANISATION,
                      'organisation2': terrareg.namespace_type.NamespaceType.GITHUB_ORGANISATION},
         False, True, 'organisation3', UserGroupNamespacePermissionType.FULL, False),

        # No organisations defined
        (True, True, None, False, True, 'organisation2', UserGroupNamespacePermissionType.FULL, False),
        (True, True, None, True, True, 'organisation2', UserGroupNamespacePermissionType.FULL, True),

        # Test with auto generated disabled
        (True, False, {'organisation1': terrareg.namespace_type.NamespaceType.GITHUB_ORGANISATION,
                       'organisation2': terrareg.namespace_type.NamespaceType.GITHUB_ORGANISATION},
         True, True, 'organisation2', UserGroupNamespacePermissionType.MODIFY, True),
        (True, False, {'organisation1': terrareg.namespace_type.NamespaceType.GITHUB_ORGANISATION,
                       'organisation2': terrareg.namespace_type.NamespaceType.GITHUB_ORGANISATION},
         True, True, 'organisation2', UserGroupNamespacePermissionType.FULL, True),
        (True, False, {'organisation1': terrareg.namespace_type.NamespaceType.GITHUB_ORGANISATION,
                       'organisation2': terrareg.namespace_type.NamespaceType.GITHUB_ORGANISATION},
         False, True, 'organisation2', UserGroupNamespacePermissionType.MODIFY, False),
        (True, False, {'organisation1': terrareg.namespace_type.NamespaceType.GITHUB_ORGANISATION,
                       'organisation2': terrareg.namespace_type.NamespaceType.GITHUB_ORGANISATION},
         False, True, 'organisation2', UserGroupNamespacePermissionType.FULL, False),

        # Invalid provider source
        (False, True, {'organisation1': terrareg.namespace_type.NamespaceType.GITHUB_ORGANISATION,
                       'organisation2': terrareg.namespace_type.NamespaceType.GITHUB_ORGANISATION},
         True, False, 'organisation2', UserGroupNamespacePermissionType.MODIFY, False),
    ])
    @setup_test_data()
    def test_check_namespace_access(self, valid_provider_source, auto_generate_github_organisation_namespaces,
                                    organisations, super_response, expect_super_called, check_namespace,
                                    check_permission_type, expected_response, test_request_context,
                                    test_github_provider_source):
        """Test check_namespace_access"""
        self.SERVER._app.secret_key = "a-secret-key"

        with mock.patch('terrareg.provider_source.github.GithubProviderSource.auto_generate_github_organisation_namespaces', auto_generate_github_organisation_namespaces), \
                mock.patch('terrareg.auth.base_sso_auth_method.BaseSsoAuthMethod.check_namespace_access', mock.MagicMock(return_value=super_response)) as mock_check_namespace_access, \
                mock.patch('terrareg.config.Config.SECRET_KEY', "a-secret-key"), \
                test_request_context:

            if organisations is not None:
                test_request_context.session['organisations'] = organisations

            if valid_provider_source:
                test_request_context.session['provider_source'] = test_github_provider_source.name
            else:
                test_request_context.session['provider_source'] = None

            obj = GithubAuthMethod()
            assert obj.check_namespace_access(permission_type=check_permission_type, namespace=check_namespace) is expected_response

            if expect_super_called:
                mock_check_namespace_access.assert_called_once_with(permission_type=check_permission_type, namespace=check_namespace)
            else:
                mock_check_namespace_access.assert_not_called()

    @pytest.mark.parametrize('valid_provider_source, auto_generate_github_organisation_namespaces, github_organisations, super_permissions, expected_permissions', [
        (True, True, None, {}, {}),
        (True, True, {'test-org1': terrareg.namespace_type.NamespaceType.GITHUB_ORGANISATION}, {}, {'test-org1': UserGroupNamespacePermissionType.FULL}),
        (True, True, {'test-org1': terrareg.namespace_type.NamespaceType.GITHUB_ORGANISATION}, {
            'testnamespace': UserGroupNamespacePermissionType.FULL,
            'moduledetails': UserGroupNamespacePermissionType.MODIFY
         }, {
            'test-org1': UserGroupNamespacePermissionType.FULL,
            'testnamespace': UserGroupNamespacePermissionType.FULL,
            'moduledetails': UserGroupNamespacePermissionType.MODIFY
        }),

        (True, False, None, {}, {}),
        (True, False, {'test-org1': terrareg.namespace_type.NamespaceType.GITHUB_ORGANISATION}, {}, {}),
        (True, False, {'test-org1': terrareg.namespace_type.NamespaceType.GITHUB_ORGANISATION}, {
            'testnamespace': UserGroupNamespacePermissionType.FULL,
            'moduledetails': UserGroupNamespacePermissionType.MODIFY
         }, {
            'testnamespace': UserGroupNamespacePermissionType.FULL,
            'moduledetails': UserGroupNamespacePermissionType.MODIFY
        }),

        # Invalid provider source
        (False, True, {'test-org1': terrareg.namespace_type.NamespaceType.GITHUB_ORGANISATION}, {
            'testnamespace': UserGroupNamespacePermissionType.FULL,
            'moduledetails': UserGroupNamespacePermissionType.MODIFY
         }, {}),
    ])
    @setup_test_data()
    def test_get_all_namespace_permissions(self, valid_provider_source, auto_generate_github_organisation_namespaces,
                                           github_organisations, super_permissions, expected_permissions,
                                           test_request_context, test_github_provider_source, mock_models):
        """test test_get_all_namespace_permissions"""
        self.SERVER._app.secret_key = "a-secret-key"
        get_all_namespace_permissions_response = {
                terrareg.models.Namespace.get(namespace_name): permission
                for namespace_name, permission in super_permissions.items()
            }
        with mock.patch('terrareg.provider_source.github.GithubProviderSource.auto_generate_github_organisation_namespaces', auto_generate_github_organisation_namespaces), \
                mock.patch('terrareg.auth.base_sso_auth_method.BaseSsoAuthMethod.get_all_namespace_permissions',
                           mock.MagicMock(return_value=get_all_namespace_permissions_response)) as mock_get_all_namespace_permissions, \
                mock.patch('terrareg.config.Config.SECRET_KEY', "a-secret-key"), \
                test_request_context:
            if github_organisations is not None:
                test_request_context.session['organisations'] = github_organisations
            if valid_provider_source:
                test_request_context.session['provider_source'] = test_github_provider_source.name

            obj = GithubAuthMethod()
            assert obj.get_all_namespace_permissions() == {
                terrareg.models.Namespace.get(namespace_name): permission
                for namespace_name, permission in expected_permissions.items()
            }

            if valid_provider_source:
                mock_get_all_namespace_permissions.assert_called_once_with()
            else:
                mock_get_all_namespace_permissions.assert_not_called()

            # Ensure namespace were created, if auto_generate_github_organisation_namespaces
            # is enabled, otherwise, ensure they're not
            for org_name in (github_organisations or {}):
                if auto_generate_github_organisation_namespaces and valid_provider_source:
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

    @pytest.mark.parametrize('username, provider_source, expected_result', [
        (None, "Test Github Provider", False),
        ('', "Test Github Provider", False),
        ('test-user', "Test Github Provider", True),

        ('test-user', None, False),
        ('test-user', "", False),
        ('test-user', "invalid", False),
    ])
    def test_check_session(self, username, provider_source, expected_result, test_request_context, test_github_provider_source):
        """Test check_session method"""
        self.SERVER._app.secret_key = "asecretkey"
        with mock.patch('terrareg.config.Config.SECRET_KEY', "asecretkey"), \
                test_request_context:
            if username is not None:
                test_request_context.session['github_username'] = username
            if provider_source is not None:
                test_request_context.session['provider_source'] = provider_source
            test_request_context.session.modified = True

            obj = GithubAuthMethod()
            assert obj.check_session() == expected_result