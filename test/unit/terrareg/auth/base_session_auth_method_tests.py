

from unittest import mock
import pytest
from terrareg.auth import AuthenticationType

from test.unit.terrareg.auth.base_auth_method_test import BaseAuthMethodTest
from test import test_request_context


class BaseSessionAuthMethodTests(BaseAuthMethodTest):

    CLS = None

    @pytest.mark.parametrize('secret_key,session_id,session_check_session_result,is_admin_authenticated,check_session_auth_type_result,cls_check_session_result,expected_result', [
        # Test working
        ('TEST_KEY', 'test_session_id', True, True, True, True, True),
        # Test without secret_key
        ('', 'test_session_id', True, True, True, True, False),
        (None, 'test_session_id', True, True, True, True, False),
        # Test without session ID
        ('TEST_KEY', '', False, True, True, True, False),
        # Test with bad is_admin_authenticated headers
        ('TEST_KEY', 'test_session_id', True, False, True, True, False),
        ('TEST_KEY', 'test_session_id', True, None, True, True, False),
        # Test with check_session_auth_type failing
        ('TEST_KEY', 'test_session_id', True, True, False, True, False),
        # Test with cls.check_session failing
        ('TEST_KEY', 'test_session_id', True, True, True, False, False),
    ])
    def test_check_auth_state(self, secret_key, session_id, session_check_session_result,
                              is_admin_authenticated, check_session_auth_type_result,
                              cls_check_session_result, expected_result, test_request_context):
        """Test check_auth_state method"""

        mock_check_session = mock.MagicMock(return_value=session_check_session_result)
        mock_check_session_auth_type = mock.MagicMock(return_value=check_session_auth_type_result)
        mock_cls_check_session = mock.MagicMock(return_value=cls_check_session_result)

        self.SERVER._app.secret_key = secret_key
        with mock.patch('terrareg.config.Config.SECRET_KEY', secret_key), \
                mock.patch('terrareg.models.Session.check_session', mock_check_session), \
                mock.patch('terrareg.auth.base_session_auth_method.BaseSessionAuthMethod.check_session_auth_type', mock_check_session_auth_type), \
                mock.patch(f'terrareg.auth.{self.CLS.__name__}.check_session', mock_cls_check_session), \
                test_request_context:

            if secret_key:
                if session_id is not None:
                    test_request_context.session['session_id'] = session_id
                if is_admin_authenticated is not None:
                    test_request_context.session['is_admin_authenticated'] = is_admin_authenticated
                test_request_context.session.modified = True

            obj = self.CLS()
            assert obj.check_auth_state() == expected_result

            if secret_key:
                mock_check_session.assert_called_once_with(session_id)

    def test_check_session_auth_type(self):
        """Test check_session_auth_type"""
        raise NotImplementedError

    def test_check_session(self):
        """test check_session method"""
        raise NotImplementedError

    def test_can_access_read_api(self):
        """Test can_access_read_api method"""
        obj = self.CLS()
        assert obj.can_access_read_api() == True

    def test_can_access_terraform_api(self):
        """Test can_access_terraform_api method"""
        obj = self.CLS()
        assert obj.can_access_terraform_api() == True

    def test_should_record_terraform_analytics(self):
        """Test should_record_terraform_analytics method"""
        obj = self.CLS()
        assert obj.should_record_terraform_analytics() is True

    def test_get_terraform_auth_token(self):
        """Test get_terraform_auth_token method"""
        obj = self.CLS()
        assert obj.get_terraform_auth_token() is None
