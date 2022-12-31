

import datetime
from unittest import mock
import pytest

from terrareg.auth import AuthenticationType, SamlAuthMethod, UserGroupNamespacePermissionType
from test import BaseTest
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
        self.SERVER._app.secret_key = "asecretkey"
        with mock.patch('terrareg.config.Config.SECRET_KEY', "asecretkey"), \
                test_request_context:
            if session_userdata:
                test_request_context.session['samlUserdata'] = session_userdata
            test_request_context.session.modified = True
        
            obj = SamlAuthMethod()
            assert obj.check_session() == expected_result

    @pytest.mark.parametrize('userdata_group_session_value,saml2_group_attribute,expected_result', [
        (None, 'groups', []),
        ({}, 'groups', []),
        ('invalidstring', 'groups', []),
        ({'groups': None}, 'groups', []),
        ({'groups': 'invalidstring'}, 'groups', []),
        ({'groups': []}, 'groups', []),
        ({'groups': ['onegroup']}, 'groups', ['onegroup']),
        ({'groups': ['first-group', 'second group']}, 'groups', ['first-group', 'second group']),
        ({'groups': ['first-group', 'second group'],
          'alternativeattr': ['third-group', 'forth-group']}, 'alternativeattr',
          ['third-group', 'forth-group'])
    ])
    def test_get_group_memberships(self, userdata_group_session_value, saml2_group_attribute, expected_result, test_request_context):
        """Test get_group_memberships method"""
        self.SERVER._app.secret_key = "asecretkey"
        with mock.patch('terrareg.config.Config.SECRET_KEY', "asecretkey"), \
                mock.patch('terrareg.config.Config.SAML2_GROUP_ATTRIBUTE', saml2_group_attribute), \
                test_request_context:
            test_request_context.session['samlUserdata'] = userdata_group_session_value
            test_request_context.session.modified = True
        
            obj = SamlAuthMethod()
            assert obj.get_group_memberships() == expected_result

    @pytest.mark.parametrize('auth_type,expected_result', [
        (None, False),
        (AuthenticationType.NOT_AUTHENTICATED, False),
        (AuthenticationType.NOT_CHECKED, False),
        (AuthenticationType.AUTHENTICATION_TOKEN, False),
        (AuthenticationType.SESSION_PASSWORD, False),
        (AuthenticationType.SESSION_OPENID_CONNECT, False),
        (AuthenticationType.SESSION_SAML, True)
    ])
    def test_check_session_auth_type(self, auth_type, expected_result, test_request_context):
        """Test check_session_auth_type"""
        self.SERVER._app.secret_key = "asecretkey"
        with mock.patch('terrareg.config.Config.SECRET_KEY', "asecretkey"), \
                test_request_context:
            if auth_type:
                test_request_context.session['authentication_type'] = auth_type.value
                test_request_context.session.modified = True
            
            obj = SamlAuthMethod()
            assert obj.check_session_auth_type() == expected_result

    @pytest.mark.parametrize('saml_name_id', [
        None,
        '',
        'ausername'
    ])
    def test_get_username(self, saml_name_id, test_request_context):
        """Test get_username method"""
        with test_request_context:

            test_request_context.session['samlNameId'] = saml_name_id
            test_request_context.session.modified = True

            obj = SamlAuthMethod()
            assert obj.get_username() == saml_name_id
