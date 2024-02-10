import unittest.mock

import pytest

from test.unit.terrareg import TerraregUnitTest
import terrareg.file_storage


class TestFileStorageFactory(TerraregUnitTest):

    @pytest.mark.parametrize('data_directory_path, expected_class', [
        ('/tmp/some/directory', terrareg.file_storage.LocalFileStorage),
        ('./some/relative/path', terrareg.file_storage.LocalFileStorage),
        ('s3://test-bucket', terrareg.file_storage.S3FileStorage),
        ('s3://test-bucket-trailing/', terrareg.file_storage.S3FileStorage),
        ('s3://test-bucket-trailing/with-path', terrareg.file_storage.S3FileStorage)
    ])
    def test_get_file_storage(self, data_directory_path, expected_class):
        """Test get_file_Storage method"""
        factory = terrareg.file_storage.FileStorageFactory()

        with unittest.mock.patch('terrareg.config.Config.DATA_DIRECTORY', data_directory_path):
            storage_instance = factory.get_file_storage()

            assert isinstance(storage_instance, expected_class)
            if expected_class is terrareg.file_storage.LocalFileStorage:
                assert storage_instance._base_directory == data_directory_path
            elif expected_class is terrareg.file_storage.S3FileStorage:
                assert storage_instance._s3_url == data_directory_path
            else:
                raise Exception('Unhandled storage type')
