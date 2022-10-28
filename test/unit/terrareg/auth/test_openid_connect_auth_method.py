

import datetime
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

    @pytest.mark.parametrize('openid_connect_expires_at,expiry_should_pass,openid_connect_id_token,validate_session_token_raises,expected_result', [
        # Working situation
        ((datetime.datetime.now() + datetime.timedelta(minutes=10)).timestamp(), True, 'testtoken', False, True),
        # Expired token
        ((datetime.datetime.now() - datetime.timedelta(minutes=1)).timestamp(), False, 'testtoken', False, False),
        # Non existent timestamp
        (None, False, 'testtoken', False, False),
        # Invalid timestamp
        ('thisisnotatimestamp', False, 'testtoken', False, False),
        # Empty token, with false return
        ((datetime.datetime.now() + datetime.timedelta(minutes=10)).timestamp(), True, '', True, False),
        ((datetime.datetime.now() + datetime.timedelta(minutes=10)).timestamp(), True, None, True, False),
    ])
    def test_check_session(self, openid_connect_expires_at, expiry_should_pass, openid_connect_id_token,
                           validate_session_token_raises, expected_result, test_request_context):
        """Test check_session method"""
        def validate_session_token_side_effect(passed_token):
            if validate_session_token_raises:
                raise Exception('Token is invalid')
        mock_validate_session_token = mock.MagicMock(side_effect=validate_session_token_side_effect)
        self.SERVER._app.secret_key = 'test_secret_key'

        with mock.patch('terrareg.openid_connect.OpenidConnect.validate_session_token', mock_validate_session_token), \
                test_request_context:

            if openid_connect_expires_at:
                test_request_context.session['openid_connect_expires_at'] = openid_connect_expires_at
            if openid_connect_id_token:
                test_request_context.session['openid_connect_id_token'] = openid_connect_id_token
            test_request_context.session.modified = True

            obj = OpenidConnectAuthMethod()
            assert obj.check_session() is expected_result

            if expiry_should_pass:
                mock_validate_session_token.assert_called_once_with(openid_connect_id_token or None)
            else:
                mock_validate_session_token.assert_not_called()

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
