
import unittest.mock

import pytest

from terrareg.analytics import AnalyticsEngine
import terrareg.errors
from test.unit.terrareg import (
    mock_models,
    setup_test_data, TerraregUnitTest
)
import terrareg.models
import terrareg.config
from test import client, mock_create_audit_event
from . import mock_record_module_version_download


class TestApiModuleVersionSourceDownload(TerraregUnitTest):
    """Test ApiModuleVersionDownload resource."""

            # '/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/source.zip',
            # '/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/<string:presign>/source.zip'

    @setup_test_data()
    def test_disable_module_hosting(self, client, mock_models, mock_record_module_version_download):
        """Test module version download with invalid analytics token"""
        with unittest.mock.patch('terrareg.config.Config.ALLOW_MODULE_HOSTING', terrareg.config.ModuleHostingMode.DISALLOW):
            res = client.get(f"/v1/terrareg/modules/testnamespace/testmodulename/testprovider/2.4.1/source.zip")
        assert res.status_code == 500
        assert res.json == {'message': 'Module hosting is disbaled'}

    @setup_test_data()
    @pytest.mark.parametrize('namespace, module, provider, version, error', [
        # Non-existent version
        ('testnamespace', 'testmodulename', 'testprovider', '2.4.2', 'Module version does not exist'),
        ('testnamespace', 'testmodulename', 'nonexistent', '2.4.2', 'Module provider does not exist'),
        ('testnamespace', 'nonexistent', 'nonexistent', '2.4.2', 'Module provider does not exist'),
        ('nonexistent', 'nonexistent', 'nonexistent', '2.4.2', 'Namespace does not exist'),
    ])
    def test_non_existent(self, namespace, module, provider, version, error, client, mock_models):
        """Test module version download with invalid analytics token"""
        with unittest.mock.patch('terrareg.config.Config.ALLOW_MODULE_HOSTING', terrareg.config.ModuleHostingMode.ENFORCE):
            res = client.get(f"/v1/terrareg/modules/{namespace}/{module}/{provider}/{version}/source.zip")
        assert res.status_code == 400
        assert res.json == {'message': error}

    @setup_test_data()
    @pytest.mark.parametrize('url, expected_presign_key', [
        ('/v1/terrareg/modules/testnamespace/testmodulename/testprovider/2.4.1/source.zip', None),

        # Pre-sign key in URL
        ('/v1/terrareg/modules/testnamespace/testmodulename/testprovider/2.4.1/unittest-presign-key/source.zip', 'unittest-presign-key'),

        # Presign key in params
        ('/v1/terrareg/modules/testnamespace/testmodulename/testprovider/2.4.1/source.zip?presign=unittest-presign-key', 'unittest-presign-key'),
    ])
    def test_invalid_presign_key(self, url, expected_presign_key, client, mock_models):
        """Ensure invalid pre-sign key throws an error"""
        def raise_exception(*args, **kwargs):
            raise terrareg.errors.InvalidPresignedUrlKeyError('Invalid pre-sign key')

        mock_validate_presigned_key = unittest.mock.MagicMock(side_effect=raise_exception)
        with unittest.mock.patch('terrareg.config.Config.ALLOW_MODULE_HOSTING', terrareg.config.ModuleHostingMode.ALLOW), \
                unittest.mock.patch('terrareg.config.Config.ALLOW_UNAUTHENTICATED_ACCESS', False), \
                unittest.mock.patch('terrareg.presigned_url.TerraformSourcePresignedUrl.validate_presigned_key', mock_validate_presigned_key):
            res = client.get(url)
        assert res.status_code == 403
        assert res.json == {'message': 'Invalid pre-sign key'}

        mock_validate_presigned_key.assert_called_once_with(url='/v1/terrareg/modules/testnamespace/testmodulename/testprovider/2.4.1', payload=expected_presign_key)

    @setup_test_data()
    @pytest.mark.parametrize('url, expected_presign_key', [
        ('/v1/terrareg/modules/testnamespace/testmodulename/testprovider/2.4.1/source.zip', None),

        # Pre-sign key in URL
        ('/v1/terrareg/modules/testnamespace/testmodulename/testprovider/2.4.1/unittest-presign-key/source.zip', 'unittest-presign-key'),

        # Presign key in params
        ('/v1/terrareg/modules/testnamespace/testmodulename/testprovider/2.4.1/source.zip?presign=unittest-presign-key', 'unittest-presign-key'),
    ])
    def test_presign_validation_disabled(self, url, expected_presign_key, client, mock_models):
        """Ensure pre-sign validation is not performed when unauthenticated access is allowed"""
        def raise_exception(*args, **kwargs):
            raise terrareg.errors.InvalidPresignedUrlKeyError('Invalid pre-sign key')

        mock_get_file_storage = unittest.mock.MagicMock()
        mock_send_file = unittest.mock.MagicMock(return_value="UNIT TEST BINARY OUTPUT")

        mock_validate_presigned_key = unittest.mock.MagicMock(side_effect=raise_exception)
        with unittest.mock.patch('terrareg.config.Config.ALLOW_MODULE_HOSTING', terrareg.config.ModuleHostingMode.ALLOW), \
                unittest.mock.patch('terrareg.config.Config.ALLOW_UNAUTHENTICATED_ACCESS', True), \
                unittest.mock.patch('terrareg.presigned_url.TerraformSourcePresignedUrl.validate_presigned_key', mock_validate_presigned_key), \
                unittest.mock.patch('flask.send_file', mock_send_file), \
                unittest.mock.patch('terrareg.file_storage.FileStorageFactory.get_file_storage', mock_get_file_storage):
            res = client.get(url)

        assert res.status_code == 200
        assert res.json == "UNIT TEST BINARY OUTPUT"

        mock_validate_presigned_key.assert_not_called()

    @setup_test_data()
    @pytest.mark.parametrize('url, allow_unauthenticated_access', [
        ('/v1/terrareg/modules/testnamespace/testmodulename/testprovider/2.4.1/source.zip', True),

        # Pre-sign key in URL
        ('/v1/terrareg/modules/testnamespace/testmodulename/testprovider/2.4.1/unittest-presign-key/source.zip', False),
        ('/v1/terrareg/modules/testnamespace/testmodulename/testprovider/2.4.1/unittest-presign-key/source.zip', True),

        # Presign key in params
        ('/v1/terrareg/modules/testnamespace/testmodulename/testprovider/2.4.1/source.zip?presign=unittest-presign-key', False),
        ('/v1/terrareg/modules/testnamespace/testmodulename/testprovider/2.4.1/source.zip?presign=unittest-presign-key', True),
    ])
    def test_send_file(self, url, allow_unauthenticated_access, client, mock_models):
        """Ensure send file and file storage is handled correctly"""
        mock_file_storage = unittest.mock.MagicMock()
        mock_get_file_storage = unittest.mock.MagicMock(return_value=mock_file_storage)
        mock_read_file_response = unittest.mock.MagicMock()
        mock_read_file = unittest.mock.MagicMock(return_value=mock_read_file_response)
        mock_file_storage.read_file = mock_read_file

        mock_validate_presigned_key = unittest.mock.MagicMock()

        mock_send_file = unittest.mock.MagicMock(return_value="UNIT TEST BINARY OUTPUT")

        with unittest.mock.patch('terrareg.config.Config.ALLOW_MODULE_HOSTING', terrareg.config.ModuleHostingMode.ALLOW), \
                unittest.mock.patch('terrareg.config.Config.ALLOW_UNAUTHENTICATED_ACCESS', allow_unauthenticated_access), \
                unittest.mock.patch('flask.send_file', mock_send_file), \
                unittest.mock.patch('terrareg.presigned_url.TerraformSourcePresignedUrl.validate_presigned_key', mock_validate_presigned_key), \
                unittest.mock.patch('terrareg.file_storage.FileStorageFactory.get_file_storage', mock_get_file_storage):
            res = client.get(url)

        assert res.status_code == 200
        assert res.json == "UNIT TEST BINARY OUTPUT"

        mock_get_file_storage.assert_called_once()
        mock_read_file.assert_called_once_with('/modules/testnamespace/testmodulename/testprovider/2.4.1/source.zip', bytes_mode=True)
        mock_send_file.assert_called_once_with(mock_read_file_response, download_name='source.zip', mimetype='application/zip')

        if not allow_unauthenticated_access:
            mock_validate_presigned_key.assert_called_once_with(url='/v1/terrareg/modules/testnamespace/testmodulename/testprovider/2.4.1', payload='unittest-presign-key')
        else:
            mock_validate_presigned_key.assert_not_called()
