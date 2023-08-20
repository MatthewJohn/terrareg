
from unittest import mock
from terrareg.errors import IncorrectCSRFTokenError

from terrareg.user_group_namespace_permission_type import UserGroupNamespacePermissionType
import terrareg.audit_action

from test import client
from test.unit.terrareg import (
    TerraregUnitTest,
    setup_test_data, mock_models
)


class TestApiTerraregNamespaceDetails(TerraregUnitTest):
    """Test ApiTerraregNamespaceDetails resource."""

    def _get_mock_namespace_access(self):
        """Return mock object for get_current_auth_method for namespace access"""
        mock_auth_method = mock.MagicMock()
        mock_auth_method.check_namespace_access = mock.MagicMock(return_value=True)
        mock_auth_method.get_username.return_value = 'moduleversion-publish-username'
        mock_get_current_auth_method = mock.MagicMock(return_value=mock_auth_method)
        return mock_get_current_auth_method

    @setup_test_data()
    def test_with_non_existent_namespace(self, client, mock_models):
        """Test namespace details with non-existent namespace."""
        res = client.get('/v1/terrareg/namespaces/doesnotexist')

        assert res.status_code == 404
        assert res.json == {'errors': ['Not Found']}

    @setup_test_data()
    def test_with_existing_namespace(self, client, mock_models):
        """Test namespace details with existing namespace."""
        res = client.get('/v1/terrareg/namespaces/testnamespace')

        assert res.status_code == 200
        assert res.json == {'is_auto_verified': False, 'trusted': False, 'display_name': None}

    @setup_test_data()
    def test_with_trusted_namespace(self, client, mock_models):
        """Test namespace details with trusted namespace."""
        with mock.patch('terrareg.config.Config.TRUSTED_NAMESPACES', ['testnamespace']):
            res = client.get('/v1/terrareg/namespaces/testnamespace')

            assert res.status_code == 200
            assert res.json == {'is_auto_verified': False, 'trusted': True, 'display_name': None}

    @setup_test_data()
    def test_with_auto_verified_namespace(self, client, mock_models):
        """Test namespace details with auto-verified namespace."""
        with mock.patch('terrareg.config.Config.VERIFIED_MODULE_NAMESPACES', ['testnamespace']):
            res = client.get('/v1/terrareg/namespaces/testnamespace')

            assert res.status_code == 200
            assert res.json == {'is_auto_verified': True, 'trusted': False, 'display_name': None}

    @setup_test_data()
    def test_with_display_name(self, client, mock_models):
        """Test namespace details with auto-verified namespace."""
        with mock.patch('terrareg.models.Namespace.display_name', 'Unit test display Name'):
            res = client.get('/v1/terrareg/namespaces/testnamespace')

            assert res.status_code == 200
            assert res.json == {'is_auto_verified': False, 'trusted': False, 'display_name': 'Unit test display Name'}

    @setup_test_data()
    def test_update_no_changes(self, client, mock_models):
        """Test sending request to update attributes without any modifications and check CSRF token"""
        mock_get_current_auth_method = self._get_mock_namespace_access()

        with mock.patch('terrareg.models.Namespace.update_name') as mock_update_name, \
                mock.patch('terrareg.models.Namespace.update_display_name') as mock_update_display_name, \
                mock.patch('terrareg.csrf.check_csrf_token', return_value=True) as mock_check_csrf, \
                mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', mock_get_current_auth_method):

            res = client.post('/v1/terrareg/namespaces/testnamespace', json={'csrf_token': 'test'})

            assert res.status_code == 200
            assert res.json == {'name': 'testnamespace', 'view_href': '/modules/testnamespace', 'display_name': None}

            mock_update_display_name.assert_not_called()
            mock_update_name.assert_not_called()
            mock_check_csrf.assert_called_once_with('test')

    @setup_test_data()
    def test_update_display_name(self, client, mock_models):
        """Test sending request to update display_name attribute"""
        mock_get_current_auth_method = self._get_mock_namespace_access()

        with mock.patch('terrareg.models.Namespace.update_name') as mock_update_name, \
                mock.patch('terrareg.csrf.check_csrf_token', return_value=True), \
                mock.patch('terrareg.audit.AuditEvent.create_audit_event') as mock_create_audit_event, \
                mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', mock_get_current_auth_method):

            res = client.post('/v1/terrareg/namespaces/testnamespace', json={'display_name': 'New display Name'})

            assert res.status_code == 200
            assert res.json == {'name': 'testnamespace', 'view_href': '/modules/testnamespace', 'display_name': 'New display Name'}

            mock_update_name.assert_not_called()
            mock_create_audit_event.assert_called_once_with(
                action=terrareg.audit_action.AuditAction.NAMESPACE_MODIFY_DISPLAY_NAME,
                object_type='Namespace', object_id='testnamespace',
                old_value=None, new_value='New display Name'
            )

    @setup_test_data()
    def test_update_name(self, client, mock_models):
        """Test sending request to update attributes changing namespace name"""
        mock_get_current_auth_method = self._get_mock_namespace_access()

        with mock.patch('terrareg.models.Namespace.update_display_name') as mock_update_display_name, \
                mock.patch('terrareg.csrf.check_csrf_token', return_value=True), \
                mock.patch('terrareg.audit.AuditEvent.create_audit_event') as mock_create_audit_event, \
                mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', mock_get_current_auth_method):

            res = client.post('/v1/terrareg/namespaces/testnamespace', json={'name': 'newname'})

            assert res.status_code == 200
            assert res.json == {'name': 'newname', 'view_href': '/modules/newname', 'display_name': None}

            mock_update_display_name.assert_not_called()
            mock_create_audit_event.assert_called_once_with(
                action=terrareg.audit_action.AuditAction.NAMESPACE_MODIFY_NAME,
                object_type='Namespace', object_id='testnamespace',
                old_value='testnamespace', new_value='newname'
            )

    @setup_test_data()
    def test_update_without_access(self, client, mock_models):
        """Test sending request to update attributes with required namespace permissions"""
        mock_auth_method = mock.MagicMock()
        mock_auth_method.check_namespace_access = mock.MagicMock(return_value=False)
        mock_auth_method.get_username.return_value = 'moduleversion-publish-username'
        mock_get_current_auth_method = mock.MagicMock(return_value=mock_auth_method)

        with mock.patch('terrareg.models.Namespace.update_name') as mock_update_name, \
                mock.patch('terrareg.models.Namespace.update_display_name') as mock_update_display_name, \
                mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', mock_get_current_auth_method):

            res = client.post('/v1/terrareg/namespaces/testnamespace', json={'name': 'unauthname', 'display_name': 'Unauthorised Display Name'})
            assert res.status_code == 403

            mock_auth_method.check_namespace_access.assert_called_once_with(UserGroupNamespacePermissionType.FULL, namespace='testnamespace')
            mock_update_name.assert_not_called()
            mock_update_display_name.assert_not_called()

    @setup_test_data()
    def test_delete(self, client, mock_models):
        """Test sending request to delete namespace"""
        mock_auth_method = mock.MagicMock()
        mock_auth_method.check_namespace_access = mock.MagicMock(return_value=True)
        mock_auth_method.get_username.return_value = 'moduleversion-publish-username'
        mock_get_current_auth_method = mock.MagicMock(return_value=mock_auth_method)

        with mock.patch('terrareg.models.Namespace.delete') as mock_delete, \
                mock.patch('terrareg.csrf.check_csrf_token', return_value=True), \
                mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', mock_get_current_auth_method):

            res = client.delete('/v1/terrareg/namespaces/testnamespace', json={})
            assert res.json == {}
            assert res.status_code == 200

            mock_auth_method.check_namespace_access.assert_called_once_with(UserGroupNamespacePermissionType.FULL, namespace='testnamespace')
            mock_delete.assert_called_once_with()

    @setup_test_data()
    def test_delete_non_existent(self, client, mock_models):
        """Test sending request to delete non-existen namespace"""
        mock_auth_method = mock.MagicMock()
        mock_auth_method.check_namespace_access = mock.MagicMock(return_value=True)
        mock_auth_method.get_username.return_value = 'moduleversion-publish-username'
        mock_get_current_auth_method = mock.MagicMock(return_value=mock_auth_method)

        with mock.patch('terrareg.models.Namespace.delete') as mock_delete, \
                mock.patch('terrareg.csrf.check_csrf_token', return_value=True) as mock_check_csrf, \
                mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', mock_get_current_auth_method):

            res = client.delete('/v1/terrareg/namespaces/doesnotexist', json={})
            assert res.status_code == 404
            assert res.json == {'errors': ['Not Found']}

            mock_auth_method.check_namespace_access.assert_called_once_with(UserGroupNamespacePermissionType.FULL, namespace='doesnotexist')
            mock_delete.assert_not_called()

    @setup_test_data()
    def test_delete_invalid_csrf(self, client, mock_models):
        """Test sending request to delete non-existen namespace"""
        mock_auth_method = mock.MagicMock()
        mock_auth_method.check_namespace_access = mock.MagicMock(return_value=True)
        mock_auth_method.get_username.return_value = 'moduleversion-publish-username'
        mock_get_current_auth_method = mock.MagicMock(return_value=mock_auth_method)

        def check_csrf_token(token):
            raise IncorrectCSRFTokenError("Invalid token")

        with mock.patch('terrareg.models.Namespace.delete') as mock_delete, \
                mock.patch('terrareg.csrf.check_csrf_token', side_effect=check_csrf_token) as mock_check_csrf, \
                mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', mock_get_current_auth_method):

            res = client.delete('/v1/terrareg/namespaces/testnamespace', json={'csrf_token': 'test-delete-csrf'})
            assert res.status_code == 500
            assert res.json == {'message': 'Invalid token', 'status': 'Error'}

            mock_auth_method.check_namespace_access.assert_called_once_with(UserGroupNamespacePermissionType.FULL, namespace='testnamespace')
            mock_check_csrf.assert_called_once_with('test-delete-csrf')
            mock_delete.assert_not_called()

    @setup_test_data()
    def test_update_without_access(self, client, mock_models):
        """Test request to delete namespace without access"""
        mock_auth_method = mock.MagicMock()
        mock_auth_method.check_namespace_access = mock.MagicMock(return_value=False)
        mock_auth_method.get_username.return_value = 'moduleversion-publish-username'
        mock_get_current_auth_method = mock.MagicMock(return_value=mock_auth_method)

        with mock.patch('terrareg.models.Namespace.delete') as mock_delete, \
                mock.patch('terrareg.csrf.check_csrf_token', return_value=True), \
                mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', mock_get_current_auth_method):

            res = client.post('/v1/terrareg/namespaces/testnamespace', json={})
            assert res.status_code == 403

            mock_auth_method.check_namespace_access.assert_called_once_with(UserGroupNamespacePermissionType.FULL, namespace='testnamespace')
            mock_delete.assert_not_called()

