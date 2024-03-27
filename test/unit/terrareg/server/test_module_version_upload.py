
from contextlib import contextmanager
import io
import unittest.mock
import zipfile

import pytest

from terrareg.analytics import AnalyticsEngine
from test.unit.terrareg import (
    mock_models,
    setup_test_data, TerraregUnitTest
)
import terrareg.models
import terrareg.config
from test import client, mock_create_audit_event
from . import mock_record_module_version_download


class TestApiModuleVersionUpload(TerraregUnitTest):
    """Test ApiModuleVersionUpload resource."""

    @staticmethod
    def _generate_zip_file():
        zip_io = io.BytesIO()
        with zipfile.ZipFile(zip_io, "w") as zip:
            zip.writestr("main.tf", b'variable "test" { }')
        zip_io.seek(0)
        return zip_io

    # EMPTY_ZIP_FILE = io.BytesIO(b'PK\x05\x06\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00')

    @setup_test_data()
    def test_non_existent_module(self, client, mock_models):
        """Test endpoint with non-existent module"""
        data = {
            'file': (self._generate_zip_file(), "test_file.zip")
        }
        with unittest.mock.patch('terrareg.config.Config.AUTO_CREATE_MODULE_PROVIDER', False), \
                unittest.mock.patch('terrareg.config.Config.AUTO_CREATE_NAMESPACE', False), \
                unittest.mock.patch('terrareg.module_extractor.ApiUploadModuleExtractor', unittest.mock.MagicMock()) as mock_extractor_patch:
            res = client.post(
                "/v1/terrareg/modules/testnamespace/modulename/doesnotexist/5.76.4/upload",
                data=data,
                headers={"content-type": "multipart/form-data"})

        assert res.json == {'message': 'Module provider does not exist'}
        assert res.status_code == 400
        mock_extractor_patch.assert_not_called()

    @setup_test_data()
    def test_non_existent_namespace(self, client, mock_models):
        """Test endpoint with non-existent module"""
        data = {
            'file': (self._generate_zip_file(), "test_file.zip")
        }
        with unittest.mock.patch('terrareg.config.Config.AUTO_CREATE_MODULE_PROVIDER', False), \
                unittest.mock.patch('terrareg.config.Config.AUTO_CREATE_NAMESPACE', False), \
                unittest.mock.patch('terrareg.module_extractor.ApiUploadModuleExtractor', unittest.mock.MagicMock()) as mock_extractor_patch:
            res = client.post(
                "/v1/terrareg/modules/doesnotexist/modulename/doesnotexist/5.76.4/upload",
                data=data,
                headers={"content-type": "multipart/form-data"})

        assert res.json == {'message': 'Namespace does not exist'}
        assert res.status_code == 400
        mock_extractor_patch.assert_not_called()

    @setup_test_data()
    def test_no_file(self, client, mock_models):
        """Test endpoint with no files"""
        data = {
        }
        namespace = terrareg.models.Namespace.get("testnamespace")
        module = terrareg.models.Module(namespace, "test-upload")
        provider = terrareg.models.ModuleProvider.create(module=module, name="test")
        with unittest.mock.patch('terrareg.config.Config.AUTO_CREATE_MODULE_PROVIDER', False), \
                unittest.mock.patch('terrareg.config.Config.AUTO_CREATE_NAMESPACE', False), \
                unittest.mock.patch('terrareg.module_extractor.ApiUploadModuleExtractor', unittest.mock.MagicMock()) as mock_extractor_patch:
            res = client.post(
                "/v1/terrareg/modules/testnamespace/test-upload/test/5.76.67/upload",
                data=data,
                headers={"content-type": "multipart/form-data"})

        assert res.json == {'status': 'Error', 'message': 'One file can be uploaded'}
        assert res.status_code == 500
        mock_extractor_patch.assert_not_called()

    @setup_test_data()
    def test_multiple_files(self, client, mock_models):
        """Test endpoint with multiple files in request"""
        zip_file = self._generate_zip_file()
        data = {
            'file': (zip_file, "test_file.zip"),
            'second_file': (zip_file, "test_file2.zip")
        }
        namespace = terrareg.models.Namespace.get("testnamespace")
        module = terrareg.models.Module(namespace, "test-upload")
        provider = terrareg.models.ModuleProvider.create(module=module, name="test")
        with unittest.mock.patch('terrareg.config.Config.AUTO_CREATE_MODULE_PROVIDER', False), \
                unittest.mock.patch('terrareg.config.Config.AUTO_CREATE_NAMESPACE', False), \
                unittest.mock.patch('terrareg.module_extractor.ApiUploadModuleExtractor', unittest.mock.MagicMock()) as mock_extractor_patch:
            res = client.post(
                "/v1/terrareg/modules/testnamespace/test-upload/test/5.76.67/upload",
                data=data,
                headers={"content-type": "multipart/form-data"})

        assert res.json == {'status': 'Error', 'message': 'One file can be uploaded'}
        assert res.status_code == 500
        mock_extractor_patch.assert_not_called()

    @setup_test_data()
    def test_non_zip_file(self, client, mock_models):
        """Test endpoint with non-zip file"""
        zip_file = self._generate_zip_file()
        data = {
            'file': (zip_file, "test_file.tar.gz")
        }
        namespace = terrareg.models.Namespace.get("testnamespace")
        module = terrareg.models.Module(namespace, "test-upload")
        provider = terrareg.models.ModuleProvider.create(module=module, name="test")
        with unittest.mock.patch('terrareg.config.Config.AUTO_CREATE_MODULE_PROVIDER', False), \
                unittest.mock.patch('terrareg.config.Config.AUTO_CREATE_NAMESPACE', False), \
                unittest.mock.patch('terrareg.module_extractor.ApiUploadModuleExtractor', unittest.mock.MagicMock()) as mock_extractor_patch:
            res = client.post(
                "/v1/terrareg/modules/testnamespace/test-upload/test/5.76.67/upload",
                data=data,
                headers={"content-type": "multipart/form-data"})

        assert res.json == {'status': 'Error', 'message': 'Error occurred - unknown file extension'}
        assert res.status_code == 500
        mock_extractor_patch.assert_not_called()

    @setup_test_data()
    def test_upload(self, client, mock_models):
        """Test upload"""
        zip_file = self._generate_zip_file()
        zip_file.seek(0)
        zip_data = zip_file.read()
        zip_file.seek(0)
        data = {
            'file': (zip_file, "test_file.zip")
        }
        namespace = terrareg.models.Namespace.get("testnamespace")
        module = terrareg.models.Module(namespace, "test-upload")
        provider = terrareg.models.ModuleProvider.create(module=module, name="test")
        expected_version = terrareg.models.ModuleVersion(module_provider=provider, version='5.76.4')

        class MockExtractor:
            process_upload = unittest.mock.MagicMock()
        mock_extractor = MockExtractor()

        def mock_extractor_constructor(upload_file, module_version):
            assert upload_file.read() == zip_data
            assert module_version.version == expected_version.version
            assert module_version.module_provider == provider

            @contextmanager
            def return_context():
                yield mock_extractor
            return return_context()
        
        with unittest.mock.patch('terrareg.config.Config.AUTO_CREATE_MODULE_PROVIDER', False), \
                unittest.mock.patch('terrareg.config.Config.AUTO_CREATE_NAMESPACE', False), \
                unittest.mock.patch('terrareg.module_extractor.ApiUploadModuleExtractor', unittest.mock.MagicMock(side_effect=mock_extractor_constructor)) as mock_extractor_patch:
            res = client.post(
                "/v1/terrareg/modules/testnamespace/test-upload/test/5.76.4/upload",
                data=data,
                headers={"content-type": "multipart/form-data"})

        assert res.json == {'status': 'Success'}
        assert res.status_code == 200

        mock_extractor_patch.assert_called_once()
        mock_extractor.process_upload.assert_called_once_with()

    @pytest.mark.parametrize('allow_module_hosting, allowed', [
        (terrareg.config.ModuleHostingMode.ALLOW, True),
        (terrareg.config.ModuleHostingMode.ENFORCE, True),
        (terrareg.config.ModuleHostingMode.DISALLOW, False),
    ])
    @setup_test_data()
    def test_non_zip_file(self, allow_module_hosting, allowed, client, mock_models):
        """Test endpoint with non-zip file"""
        zip_file = self._generate_zip_file()
        data = {
            'file': (zip_file, "test_file.zip")
        }
        namespace = terrareg.models.Namespace.get("testnamespace")
        module = terrareg.models.Module(namespace, "test-upload")
        terrareg.models.ModuleProvider.create(module=module, name="test")
        with unittest.mock.patch('terrareg.config.Config.AUTO_CREATE_MODULE_PROVIDER', False), \
                unittest.mock.patch('terrareg.config.Config.AUTO_CREATE_NAMESPACE', False), \
                unittest.mock.patch('terrareg.module_extractor.ApiUploadModuleExtractor', unittest.mock.MagicMock()) as mock_extractor_patch, \
                unittest.mock.patch('terrareg.config.Config.ALLOW_MODULE_HOSTING', allow_module_hosting):
            res = client.post(
                "/v1/terrareg/modules/testnamespace/test-upload/test/5.76.67/upload",
                data=data,
                headers={"content-type": "multipart/form-data"})

        if allowed:
            assert res.json == {'status': 'Success'}
            assert res.status_code == 200
            mock_extractor_patch.assert_called()
        else:
            assert res.json == {'message': 'Module upload is disabled.'}
            assert res.status_code == 400
            mock_extractor_patch.assert_not_called()
