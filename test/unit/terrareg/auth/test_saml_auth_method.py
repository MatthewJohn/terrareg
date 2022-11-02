

import datetime
from unittest import mock
import pytest

from terrareg.auth import SamlAuthMethod, UserGroupNamespacePermissionType
from test import BaseTest
from test.unit.terrareg import MockNamespace, MockUserGroup, MockUserGroupNamespacePermission, setup_test_data
from test.unit.terrareg.auth.base_session_auth_method_tests import BaseSessionAuthMethodTests
from test.unit.terrareg.auth.base_sso_auth_method_tests import BaseSsoAuthMethodTests, test_data, user_group_data

# Required as this is sued by BaseSamlAuthMethod
from test import test_request_context


class TestSamlAuthMethod(BaseSsoAuthMethodTests, BaseSessionAuthMethodTests):
    """Test methods of SamlAuthMethod auth method"""

    CLS = SamlAuthMethod

    @pytest.mark.parametrize('saml_is_enabled,expected_result', [
        (False, False),
        (True, True)
    ])
    def test_is_enabled(self, saml_is_enabled, expected_result):
        """Test is_enabled method"""
        with mock.patch('terrareg.saml.Saml2.is_enabled', mock.MagicMock(return_value=saml_is_enabled)):
            obj = SamlAuthMethod()
            assert obj.is_enabled() is expected_result

    @pytest.mark.parametrize('session_userdata,expected_result', [
        (None, False),
        ('', False),
        ({'valid': 'user_data'}, True),
    ])
    def test_check_session(self, session_userdata, expected_result, test_request_context):
        """Test check_session method"""
        with test_request_context:
            if session_userdata:
                test_request_context.session['samlUserdata'] = session_userdata
            test_request_context.session.modified = True
        
            obj = SamlAuthMethod()
            assert obj.check_session() == expected_result

    @pytest.mark.parametrize('userdata_group_session_value,expected_result', [
        (None, []),
        ({}, []),
        ('invalidstring', []),
        ({'groups': None}, []),
        ({'groups': 'invalidstring'}, []),
        ({'groups': []}, []),
        ({'groups': ['onegroup']}, ['onegroup']),
        ({'groups': ['first-group', 'second group']}, ['first-group', 'second group'])
    ])
    def test_get_group_memberships(self, userdata_group_session_value, expected_result, test_request_context):
        """Test get_group_memberships method"""
        with test_request_context:
            test_request_context.session['samlUserdata'] = userdata_group_session_value
            test_request_context.session.modified = True
        
            obj = SamlAuthMethod()
            assert obj.get_group_memberships() == expected_result
