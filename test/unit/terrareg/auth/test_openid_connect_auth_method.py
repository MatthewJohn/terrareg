

from unittest import mock
import pytest

from terrareg.auth import OpenidConnectAuthMethod, UserGroupNamespacePermissionType
from test import BaseTest
from test.unit.terrareg import MockNamespace, MockUserGroup, MockUserGroupNamespacePermission, setup_test_data
from test.unit.terrareg.auth.base_session_auth_method_tests import BaseSessionAuthMethodTests
from test.unit.terrareg.auth.base_sso_auth_method_tests import BaseSsoAuthMethodTests, test_data, user_group_data

# Required as this is sued by BaseOpenidConnectAuthMethod
from test import test_request_context


class TestOpenidConnectAuthMethod(BaseSsoAuthMethodTests, BaseSessionAuthMethodTests):
    """Test methods of OpenidConnectAuthMethod auth method"""

    CLS = OpenidConnectAuthMethod

    def test_is_built_in_admin(self):
        """Test is_built_in_admin method"""
        obj = OpenidConnectAuthMethod()
        assert obj.is_built_in_admin() is False

    @pytest.mark.parametrize('sso_groups,expected_result', [
        ([], False),
        (['validgroup', False]),
        (['validgroup', 'invalidgroup'], False),
        # Passing
        (['siteadmingroup'], True),
        (['invalidgroup', 'validgroup', 'siteadmingroup'], True)
    ])
    @setup_test_data(None, user_group_data=user_group_data)
    def test_is_admin(self, sso_groups, expected_result):
        """Test is_admin method"""
        mock_get_group_memberships = mock.MagicMock(return_value=sso_groups)

        with mock.patch('terrareg.models.UserGroup', MockUserGroup), \
                mock.patch('terrareg.models.UserGroupNamespacePermission',
                           MockUserGroupNamespacePermission), \
                mock.patch('terrareg.models.Namespace', MockNamespace), \
                mock.patch(f'terrareg.auth.{self.CLS.__name__}.get_group_memberships', mock_get_group_memberships):
            obj = OpenidConnectAuthMethod()
            assert obj.is_admin() is expected_result

    def test_is_authenticated(self):
        """Test is_authenticated method"""
        obj = OpenidConnectAuthMethod()
        # @TODO Should this be True?
        assert obj.is_authenticated() is True

    @pytest.mark.parametrize('admin_authentication_token,expected_result', [
        (None, False),
        ('', False),
        ('ISAPASSWORD', True)
    ])
    def test_is_enabled(self, admin_authentication_token, expected_result):
        """Test is_enabled method"""
        with mock.patch('terrareg.config.Config.ADMIN_AUTHENTICATION_TOKEN', admin_authentication_token):
            obj = OpenidConnectAuthMethod()
            assert obj.is_enabled() is expected_result

    def test_requires_csrf_tokens(self):
        """Test requires_csrf_token method"""
        obj = OpenidConnectAuthMethod()
        assert obj.requires_csrf_tokens is True

    @pytest.mark.parametrize('public_api_keys', [
        (None),
        ([]),
        (['publishapikey'])
    ])
    def test_can_publish_module_version(self, public_api_keys):
        """Test can_publish_module_version method"""
        with mock.patch('terrareg.config.Config.PUBLISH_API_KEYS', public_api_keys):
            obj = OpenidConnectAuthMethod()
            assert obj.can_publish_module_version(namespace='testnamespace') is True

    @pytest.mark.parametrize('upload_api_keys', [
        None,
        [],
        ['uploadapikey']
    ])
    def test_can_upload_module_version(self, upload_api_keys):
        """Test can_upload_module_version method"""
        with mock.patch('terrareg.config.Config.UPLOAD_API_KEYS', upload_api_keys):
            obj = OpenidConnectAuthMethod()
            assert obj.can_upload_module_version(namespace='testnamespace') is True

    def test_check_session(self):
        """Test check_session method"""
        obj = OpenidConnectAuthMethod()
        assert obj.check_session() is True

    @pytest.mark.parametrize('session_groups,expected_result', [
        (None, []),
        ([], []),
        (['onegroup'], ['onegroup']),
        (['first-group', 'second group'], ['first-group', 'second group'])
    ])
    def test_get_group_memberships(self, session_groups, expected_result, test_request_context):
        """Test get_group_memberships method"""
        with test_request_context:
            test_request_context.session['openid_groups'] = session_groups
            test_request_context.session.modified = True
        
            obj = OpenidConnectAuthMethod()
            assert obj.get_group_memberships() == expected_result
