

from unittest import mock
import pytest

from terrareg.auth import AdminSessionAuthMethod, UserGroupNamespacePermissionType
from test.unit.terrareg.auth.base_auth_method_test import BaseAuthMethodTest


class BaseAdminSessionAuthMethod(BaseAuthMethodTest):

    CLS = None

    @pytest.mark.parametrize('secret_key,session_id,session_check_session_result,is_admin_authenticated,check_session_auth_type_result,cls_check_session_result,expected_result', [
        ('TEST_KEY', 'test_session_id', True, True, True, True, True)
    ])
    def test_check_auth_state(self, secret_key, session_id, session_check_session_result,
                              is_admin_authenticated, check_session_auth_type_result,
                              cls_check_session_result, expected_result):
        """Test check_auth_state method"""

        def mock_session_get_side_effect(key):
            print('I WAS CALLED')
            if key == 'session_id':
                return session_id
            elif key == 'is_admin_authenticated':
                return is_admin_authenticated
            assert False
        mock_flask_session_get = mock.MagicMock(side_affect=mock_session_get_side_effect)
        mock_flask_session = mock.MagicMock()
        mock_flask_session.get = mock_flask_session_get
        mock_check_session = mock.MagicMock(return_value=session_check_session_result)
        mock_check_session_auth_type = mock.MagicMock(return_value=check_session_auth_type_result)
        mock_cls_check_session = mock.MagicMock(return_value=cls_check_session_result)

        obj = self.CLS()
        with mock.patch('terrareg.config.Config.SECRET_KEY', secret_key), \
                mock.patch('terrareg.models.Session.check_session', mock_check_session), \
                mock.patch('flask.session', mock_flask_session), \
                mock.patch(f'terrareg.auth.{self.CLS.__name__}.check_session_auth_type', mock_check_session_auth_type), \
                mock.patch(f'terrareg.auth.{self.CLS.__name__}.check_session', mock_cls_check_session):
            assert obj.check_auth_state() == expected_result

        # if (not terrareg.config.Config().SECRET_KEY or
        #         not Session.check_session(session.get('session_id', None)) or
        #         not session.get('is_admin_authenticated', False)):
        #     return False

        # # Ensure session type is set to the current session and session is valid
        # return cls.check_session_auth_type() and cls.check_session()

    def test_check_session_auth_type(self):
        """Test check_session_auth_type"""
        pass