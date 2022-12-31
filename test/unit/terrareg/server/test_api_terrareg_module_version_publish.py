
import unittest.mock

import terrareg.audit_action
from test.unit.terrareg import (
    mock_models,
    setup_test_data, TerraregUnitTest
)
from test import client


class TestApiModuleVersionPublish(TerraregUnitTest):
    """Test module version publish resource."""

    def _get_mock_get_current_auth_method(self, allowed_to_create):
        """Return mock auth method"""
        mock_auth_method = unittest.mock.MagicMock()
        mock_auth_method.can_publish_module_version = unittest.mock.MagicMock(return_value=allowed_to_create)
        mock_auth_method.get_username.return_value = 'moduleversion-publish-username'
        mock_get_current_auth_method = unittest.mock.MagicMock(return_value=mock_auth_method)
        return mock_get_current_auth_method

    @setup_test_data()
    def test_publish_unpublished_module_version(self, client, mock_models):
        """Test publishing a module version."""
        with unittest.mock.patch(
                    'terrareg.models.ModuleVersion.update_attributes') as mocked_update_attributes, \
                unittest.mock.patch('terrareg.audit.AuditEvent.create_audit_event') as mock_create_audit_event, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._get_mock_get_current_auth_method(True)):

            res = client.post(
                '/v1/terrareg/modules/testnamespace/modulenotpublished/testprovider/10.2.1/publish')
            assert res.status_code == 200
            assert res.json == {'status': 'Success'}

            mocked_update_attributes.assert_called_once_with(published=True)
            mock_create_audit_event.assert_called_once_with(
                action=terrareg.audit_action.AuditAction.MODULE_VERSION_PUBLISH,
                object_type='ModuleVersion',
                object_id='testnamespace/modulenotpublished/testprovider/10.2.1',
                old_value=None, new_value=None
            )

    @setup_test_data()
    def test_publish_published_module_version(self, client, mock_models):
        """Test publishing a module version that is already in a published state."""
        with unittest.mock.patch(
                    'terrareg.models.ModuleVersion.update_attributes') as mocked_update_attributes, \
                unittest.mock.patch('terrareg.audit.AuditEvent.create_audit_event') as mock_create_audit_event, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._get_mock_get_current_auth_method(True)):

            res = client.post(
                '/v1/terrareg/modules/testnamespace/testmodulename/testprovider/2.4.1/publish')
            assert res.status_code == 200
            assert res.json == {'status': 'Success'}

            mock_create_audit_event.assert_called_once_with(
                action=terrareg.audit_action.AuditAction.MODULE_VERSION_PUBLISH,
                object_type='ModuleVersion',
                object_id='testnamespace/testmodulename/testprovider/2.4.1',
                old_value=None, new_value=None
            )
            mocked_update_attributes.assert_called_once_with(published=True)

    @setup_test_data()
    def test_publish_non_existent_module_provider(self, client, mock_models):
        """Test publishing endpoint against non-existent module provider."""
        with unittest.mock.patch(
                    'terrareg.models.ModuleVersion.update_attributes') as mocked_update_attributes, \
                unittest.mock.patch('terrareg.audit.AuditEvent.create_audit_event') as mock_create_audit_event, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._get_mock_get_current_auth_method(True)):

            res = client.post(
                '/v1/terrareg/modules/testnamespace/modulenotpublished/doesnotexistprovider/5.5.4/publish')
            assert res.status_code == 400
            assert res.json == {'message': 'Module provider does not exist'}

            mock_create_audit_event.assert_not_called()
            mocked_update_attributes.assert_not_called()

    @setup_test_data()
    def test_publish_non_existent_module(self, client, mock_models):
        """Test publishing endpoint against non-existent module."""
        with unittest.mock.patch(
                    'terrareg.models.ModuleVersion.update_attributes') as mocked_update_attributes, \
                unittest.mock.patch('terrareg.audit.AuditEvent.create_audit_event') as mock_create_audit_event, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._get_mock_get_current_auth_method(True)):

            res = client.post(
                '/v1/terrareg/modules/testnamespace/providerdoesnotexist/doesnotexistprovider/5.5.4/publish')
            assert res.status_code == 400
            assert res.json == {'message': 'Module provider does not exist'}

            mock_create_audit_event.assert_not_called()
            mocked_update_attributes.assert_not_called()

    @setup_test_data()
    def test_publish_non_existent_namespace(self, client, mock_models):
        """Test publishing endpoint against non-existent namespace."""
        with unittest.mock.patch(
                    'terrareg.models.ModuleVersion.update_attributes') as mocked_update_attributes, \
                unittest.mock.patch('terrareg.audit.AuditEvent.create_audit_event') as mock_create_audit_event, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._get_mock_get_current_auth_method(True)):

            res = client.post(
                '/v1/terrareg/modules/namespacedoesnotexist/providerdoesnotexist/doesnotexistprovider/5.5.4/publish')
            assert res.status_code == 400
            assert res.json == {'message': 'Namespace does not exist'}

            mock_create_audit_event.assert_not_called()
            mocked_update_attributes.assert_not_called()

    @setup_test_data()
    def test_publish_non_existent_version(self, client, mock_models):
        """Test publishing endpoint against non-existent version."""
        with unittest.mock.patch(
                    'terrareg.models.ModuleVersion.update_attributes') as mocked_update_attributes, \
                unittest.mock.patch('terrareg.audit.AuditEvent.create_audit_event') as mock_create_audit_event, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._get_mock_get_current_auth_method(True)):

            res = client.post(
                '/v1/terrareg/modules/testnamespace/modulenotpublished/testprovider/99.21.22/publish')
            assert res.status_code == 400
            assert res.json == {'message': 'Module version does not exist'}

            mock_create_audit_event.assert_not_called()
            mocked_update_attributes.assert_not_called()

    @setup_test_data()
    def test_publish_invalid_api_key(self, client, mock_models):
        """Test publishing endpoint with invalid API keys."""
        with unittest.mock.patch(
                    'terrareg.models.ModuleVersion.update_attributes') as mocked_update_attributes, \
                unittest.mock.patch('terrareg.audit.AuditEvent.create_audit_event') as mock_create_audit_event, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._get_mock_get_current_auth_method(False)):

            res = client.post(
                '/v1/terrareg/modules/testnamespace/modulenotpublished/testprovider/10.2.1/publish')
            assert res.status_code == 403
            assert res.json == {'message': "You don't have the permission to access the requested resource. "
                                           "It is either read-protected or not readable by the server."}

            mock_create_audit_event.assert_not_called()
            mocked_update_attributes.assert_not_called()