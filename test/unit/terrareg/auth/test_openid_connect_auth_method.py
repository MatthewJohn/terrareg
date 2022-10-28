

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

    @pytest.mark.parametrize('openid_connect_is_enabled,expected_result', [
        (False, False),
        (True, True)
    ])
    def test_is_enabled(self, openid_connect_is_enabled, expected_result):
        """Test is_enabled method"""
        with mock.patch('terrareg.openid_connect.OpenidConnect.is_enabled', mock.MagicMock(return_value=openid_connect_is_enabled)):
            obj = OpenidConnectAuthMethod()
            assert obj.is_enabled() is expected_result

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
