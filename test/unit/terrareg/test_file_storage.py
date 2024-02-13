import contextlib
import tempfile
import unittest.mock
import os

import pytest
import boto3
import botocore.exceptions

from test.unit.terrareg import TerraregUnitTest
import terrareg.file_storage
from terrareg.errors import InvalidDataDirectoryError


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

class TestLocalFileStorage(TerraregUnitTest):
    """Handle local file storage."""

    def test___init__(self):
        """Test __init__ method"""
        instance = terrareg.file_storage.LocalFileStorage("/test/directory")
        assert instance._base_directory == "/test/directory"

    @pytest.mark.parametrize('base_directory, paths, expected_response', [
        # Paths defined with slashes and replaced by test
        ('/root/dir', ['base_dir/second_dir'], "/root/dir/base_dir/second_dir"),

        ('/root/dir', ['base_dir', 'combined'], "/root/dir/base_dir/combined"),
        ('/root/dir', ['multiple/base_dir', 'multiple/combined'], "/root/dir/multiple/base_dir/multiple/combined"),
        # Leading slash
        ('/root/dir', ['/multiple/base_dir', 'multiple/combined'], "/root/dir/multiple/base_dir/multiple/combined"),
        ('/root/dir', ['/multiple/base_dir', '/multiple/combined'], "/root/dir/multiple/base_dir/multiple/combined"),
        # Trailing slash
        ('/root/dir', ['multiple/base_dir/', 'multiple/combined'], "/root/dir/multiple/base_dir/multiple/combined"),
        ('/root/dir', ['multiple/base_dir/', 'multiple/combined/'], "/root/dir/multiple/base_dir/multiple/combined"),
        # Multiple slashes
        ('/root/dir', ['//multiple//base_dir/', '//multiple//combined//'], "/root/dir/multiple/base_dir/multiple/combined"),

        # Root path with missing leading and trailing slashes
        ('root/dir', ['multiple/base_dir', 'multiple/combined'], "root/dir/multiple/base_dir/multiple/combined"),
        ('root/dir/', ['multiple/base_dir', 'multiple/combined'], "root/dir/multiple/base_dir/multiple/combined"),
        ('root/dir//', ['/multiple/base_dir', 'multiple/combined'], "root/dir/multiple/base_dir/multiple/combined"),
    ])
    def test__generate_path(self, base_directory, paths, expected_response):
        """Test _generate_path method"""
        paths = [
            path.replace('/', os.path.sep)
            for path in paths
        ]
        expected_response = expected_response.replace('/', os.path.sep)

        instance = terrareg.file_storage.LocalFileStorage(base_directory)
        assert instance._generate_path(*paths) == expected_response

    @pytest.mark.parametrize('directory_to_create', [
        # Single directory
        ('test_dir_create'),
        # Directory with parents
        ('some/dir/path/test_dir_create'),
    ])
    def test_make_directory(self, directory_to_create):
        """Test make_directory method"""
        directory_to_create = directory_to_create.replace('/', os.path.sep)

        with tempfile.TemporaryDirectory() as temp_dir:
            full_test_path = os.path.join(temp_dir, directory_to_create)
            assert os.path.isdir(full_test_path) is False

            instance = terrareg.file_storage.LocalFileStorage(temp_dir)
            instance.make_directory(directory=directory_to_create)

            assert os.path.isdir(full_test_path) is True

    def test_make_directory_preexisting(self):
        """Test make_directory method with directory that already exists"""

        with tempfile.TemporaryDirectory() as temp_dir:
            test_dir_name = "test_dir"
            full_test_path = os.path.join(temp_dir, test_dir_name)
            os.mkdir(full_test_path)

            instance = terrareg.file_storage.LocalFileStorage(temp_dir)
            instance.make_directory(directory=test_dir_name)

            assert os.path.isdir(full_test_path) is True

    def test_upload_file(self):
        """Test upload_file"""
        with tempfile.TemporaryDirectory() as temp_dir:
            source_file = os.path.join(temp_dir, "source_file_name")
            with open(source_file, "w") as fh:
                fh.write("Test source file content")

            instance = terrareg.file_storage.LocalFileStorage(temp_dir)
            instance.upload_file(
                source_path=source_file,
                dest_directory="dest_directory",
                dest_filename="dest_file"
            )

            assert os.path.isdir(os.path.join(temp_dir, "dest_directory"))
            assert os.path.isfile(os.path.join(temp_dir, "dest_directory", "dest_file"))
            with open(os.path.join(temp_dir, "dest_directory", "dest_file"), "r") as fh:
                assert fh.read() == "Test source file content"

    def test_upload_file_already_exists(self):
        """Test upload_file with file that already exists in destination"""
        with tempfile.TemporaryDirectory() as temp_dir:
            source_file = os.path.join(temp_dir, "source_file_name")
            with open(source_file, "w") as fh:
                fh.write("Test source file content")

            with open(os.path.join(temp_dir, "dest_file_name"), "w") as fh:
                fh.write("Original destination content")

            instance = terrareg.file_storage.LocalFileStorage(temp_dir)
            instance.upload_file(
                source_path=source_file,
                dest_directory="",
                dest_filename="dest_file_name"
            )

            assert os.path.isfile(os.path.join(temp_dir, "dest_file_name"))
            with open(os.path.join(temp_dir, "dest_file_name"), "r") as fh:
                assert fh.read() == "Test source file content"

    @pytest.mark.parametrize('is_dir, is_file, raises', [
        (False, False, False),
        (False, True, False),
        (True, False, True),
    ])
    @pytest.mark.parametrize('paths', [
        ('test_path'),
        (['tested', 'test', 'path'])
    ])
    def test__check_not_directory(self, paths, is_dir, is_file, raises):
        """Test _check_not_directory method"""
        test_path = os.path.join(*paths)

        with tempfile.TemporaryDirectory() as temp_dir:
            instance = terrareg.file_storage.LocalFileStorage(temp_dir)

            full_path = os.path.join(temp_dir, test_path)
            # If test object is a directory or file, create directory
            # and parents. If it's a file, it will then be removed and replaced with
            # a file (allowing for parents to be created in the process)
            if is_dir or is_file:
                os.makedirs(full_path)
            if is_file:
                os.rmdir(full_path)
                with open(full_path, "w"):
                    pass

            if raises:
                with pytest.raises(terrareg.errors.FileUploadError):
                    instance._check_not_directory(*paths)
            else:
                instance._check_not_directory(*paths)

    @pytest.mark.parametrize('is_dir, is_file, expected_value', [
        (False, False, False),
        (False, True, True),
        (True, False, False),
    ])
    @pytest.mark.parametrize('test_path', [
        ('test_path'),
        ('tested/test/path')
    ])
    def test_file_exists(self, test_path, is_dir, is_file, expected_value):
        """Test file exists"""
        test_path = test_path.replace('/', os.path.sep)

        with tempfile.TemporaryDirectory() as temp_dir:
            instance = terrareg.file_storage.LocalFileStorage(temp_dir)

            full_path = os.path.join(temp_dir, test_path)
            # If test object is a directory or file, create directory
            # and parents. If it's a file, it will then be removed and replaced with
            # a file (allowing for parents to be created in the process)
            if is_dir or is_file:
                os.makedirs(full_path)
            if is_file:
                os.rmdir(full_path)
                with open(full_path, "w"):
                    pass

            assert instance.file_exists(test_path) is expected_value

    @pytest.mark.parametrize('is_dir, is_file, expected_value', [
        (False, False, False),
        (False, True, False),
        (True, False, True),
    ])
    @pytest.mark.parametrize('test_path', [
        ('test_path'),
        ('tested/test/path')
    ])
    def test_file_exists(self, test_path, is_dir, is_file, expected_value):
        """Test file exists"""
        test_path = test_path.replace('/', os.path.sep)

        with tempfile.TemporaryDirectory() as temp_dir:
            instance = terrareg.file_storage.LocalFileStorage(temp_dir)

            full_path = os.path.join(temp_dir, test_path)
            # If test object is a directory or file, create directory
            # and parents. If it's a file, it will then be removed and replaced with
            # a file (allowing for parents to be created in the process)
            if is_dir or is_file:
                os.makedirs(full_path)
            if is_file:
                os.rmdir(full_path)
                with open(full_path, "w"):
                    pass

            assert instance.directory_exists(test_path) is expected_value

    def test_delete_file(self):
        """Test delete_file"""
        with tempfile.TemporaryDirectory() as temp_dir, \
                unittest.mock.patch('os.unlink') as mock_os_unlink:
            instance = terrareg.file_storage.LocalFileStorage(temp_dir)

            test_directory = os.path.join('some', 'test', 'file')
            instance.delete_file(test_directory)

            mock_os_unlink.assert_called_once_with(os.path.join(temp_dir, test_directory))

    def test_delete_directory(self):
        """Test delete_directory"""
        with tempfile.TemporaryDirectory() as temp_dir, \
                unittest.mock.patch('os.rmdir') as mock_os_rmdir:
            instance = terrareg.file_storage.LocalFileStorage(temp_dir)

            test_directory = os.path.join('some', 'test', 'directory')
            instance.delete_directory(test_directory)

            mock_os_rmdir.assert_called_once_with(os.path.join(temp_dir, test_directory))

    @pytest.mark.parametrize('bytes_mode, expected_mode', [
        (False, 'r'),
        (True, 'rb')
    ])
    def test_read_file(self, bytes_mode, expected_mode):
        """Return file handler for file"""
        with tempfile.TemporaryDirectory() as temp_dir:
            instance = terrareg.file_storage.LocalFileStorage(temp_dir)

            file = 'test_file'
            with open(os.path.join(temp_dir, file), "w") as fh:
                fh.write("Test content")

            with instance.read_file(file, bytes_mode=bytes_mode) as test_fh:
                assert test_fh.mode == expected_mode

                expected_content = "Test content"
                if bytes_mode:
                    expected_content = expected_content.encode('utf-8')
                assert test_fh.read() == expected_content

    @pytest.mark.parametrize('binary', [
        (False),
        (True)
    ])
    def test_write_file(self, binary):
        """Test write_file method"""
        with tempfile.TemporaryDirectory() as temp_dir:
            instance = terrareg.file_storage.LocalFileStorage(temp_dir)

            file = 'test_file'

            test_content = "Test write content"
            if binary:
                test_content = test_content.encode('utf-8')

            instance.write_file(path=file, content=test_content, binary=binary)

            with open(os.path.join(temp_dir, file), "r") as fh:
                assert fh.read() == "Test write content"


@contextlib.contextmanager
def create_s3_file_storage_with_bucket(bucket_name, bucket_path):
    """Create instance of S3FileStorage and S3 bucket"""
    instance = terrareg.file_storage.S3FileStorage(s3_url=f's3://{bucket_name}{bucket_path}')
    try:
        instance._s3_client.create_bucket(
            ACL='private',
            Bucket=bucket_name
        )
    except Exception as exc:
        # If bucket already exists from a previous test, ignore it
        # @TODO What to do with content
        if "BucketAlreadyOwnedByYou" not in str(exc):
            raise

    try:
        yield instance
    finally:
        # Destroy test bucket
        response = instance._s3_client.list_objects_v2(Bucket=bucket_name)

        for object in response.get('Contents', []):
            instance._s3_client.delete_object(Bucket=bucket_name, Key=object['Key'])

        instance._s3_client.delete_bucket(Bucket=bucket_name)


class TestS3FileStorage(TerraregUnitTest):
    """Handle file storage in s3"""


    def test___init__(self):
        """Test __init__ method"""
        instance = terrareg.file_storage.S3FileStorage(s3_url="s3://unittest-bucket-name/some_path/for-ut/")
        assert instance._s3_url == "s3://unittest-bucket-name/some_path/for-ut/"
        assert instance._bucket_name == "unittest-bucket-name"
        assert instance._base_s3_path == "/some_path/for-ut"
        assert isinstance(instance._session, boto3.session.Session)

    @pytest.mark.parametrize('s3_url, expected_bucket, expected_path', [
        ('s3://test-bucket', 'test-bucket', ''),
        ('s3://test-bucket/', 'test-bucket', ''),
        ('s3://test-bucket/some-path', 'test-bucket', '/some-path'),
        ('s3://test-bucket/some-path/', 'test-bucket', '/some-path'),
        ('s3://test-bucket/some-path//', 'test-bucket', '/some-path'),
        ('s3://test-bucket/some/extra//path//', 'test-bucket', '/some/extra/path'),
    ])
    def test__get_path_details(self, s3_url, expected_bucket, expected_path):
        """Obtain bucket name and base path from s3 path"""
        instance = terrareg.file_storage.S3FileStorage(s3_url="s3://unittest-bucket-name/some_path/for-ut/")
        assert instance._get_path_details(s3_url=s3_url) == (expected_bucket, expected_path)

    @pytest.mark.parametrize('s3_url', [
        (''),
        ('s3://'),
        ('s3:///adg'),
        ('adg'),
    ])
    def test__get_path_details_invalid(self, s3_url):
        """Obtain bucket name and base path from s3 path"""
        instance = terrareg.file_storage.S3FileStorage(s3_url="s3://unittest-bucket-name/some_path/for-ut/")
        with pytest.raises(InvalidDataDirectoryError):
            instance._get_path_details(s3_url=s3_url)

    def test__get_bucket(self):
        """Get bucket object"""
        instance = terrareg.file_storage.S3FileStorage(s3_url="s3://unittest-bucket-name/some_path/for-ut")
        bucket = instance._get_bucket()
        assert bucket
        assert bucket.name == "unittest-bucket-name"
        assert bucket.__class__.__name__ == "s3.Bucket"

    @pytest.mark.parametrize('s3_url, paths, expected_key', [
        ('s3://test-bucket', ['path1'], '/path1'),
        ('s3://test-bucket', ['path1', 'path2'], '/path1/path2'),
        ('s3://test-bucket/', ['path1', 'path2'], '/path1/path2'),
        ('s3://test-bucket/', ['/path1', 'path2'], '/path1/path2'),
        ('s3://test-bucket/', ['/path1/', '/path2/'], '/path1/path2'),
        ('s3://test-bucket/', ['/path1//', '/path2//'], '/path1/path2'),
        ('s3://test-bucket/some-path', ['path1', 'path2'], '/some-path/path1/path2'),
        ('s3://test-bucket/some-path/', ['/path1', 'path2'], '/some-path/path1/path2'),
    ])
    def test__generate_key(self, s3_url, paths, expected_key):
        """Generate s3 key"""
        instance = terrareg.file_storage.S3FileStorage(s3_url=s3_url)
        assert instance._generate_key(*paths) == expected_key

    @pytest.mark.parametrize('bucket_path, dest_directory, dest_filename, test_key', [
        ("", "/test/directory/in/s3", "test_dest_file", "/test/directory/in/s3/test_dest_file"),
        ("/", "/test/directory/in/s3", "test_dest_file", "/test/directory/in/s3/test_dest_file"),
        ("/test-path", "/test/directory/in/s3", "test_dest_file", "/test-path//test/directory/in/s3/test_dest_file"),
        ("/test-path/test-path2", "/test/directory/in/s3/", "test_dest_file", "/test-path/test-path2/test/directory/in/s3/test_dest_file"),
        ("/test-path/test-path3/", "/test/dir/", "/leadingslash", "/test-path/test-path3/test/dir/leadingslash"),
    ])
    @pytest.mark.skipif(not os.environ.get('AWS_ENDPOINT_URL'), reason="Skipping due to minio not configured")
    def test_upload_file(self, bucket_path, dest_directory, dest_filename, test_key):
        """Upload file to s3"""

        with create_s3_file_storage_with_bucket(bucket_name="test-bucket", bucket_path=bucket_path) as instance, \
                tempfile.TemporaryDirectory() as temp_dir:
            # Create test file
            source_file = os.path.join(temp_dir, "test_upload_file")
            with open(source_file, "w") as fh:
                fh.write("Test upload content")
            
            instance.upload_file(source_path=source_file, dest_directory=dest_directory, dest_filename=dest_filename)

            # Obtain file from s3 to ensure it was uploaded correctly
            res = instance._s3_client.get_object(
                Bucket="test-bucket",
                Key=test_key
            )
            assert res['Body'].read() == "Test upload content".encode('utf-8')


    @pytest.mark.parametrize('bucket_path, dest_path, test_key', [
        ("", "/test/directory/in/s3/test_dest_file", "/test/directory/in/s3/test_dest_file"),
        ("/", "/test/directory/in/s3/test_dest_file", "/test/directory/in/s3/test_dest_file"),
        ("/test-path", "/test/directory/in/s3/test_dest_file", "/test-path//test/directory/in/s3/test_dest_file"),
        ("/test-path/test-path2", "/test/directory/in/s3/test_dest_file", "/test-path/test-path2/test/directory/in/s3/test_dest_file"),
        ("/test-path/test-path3/", "/test/dir/leadingslash", "/test-path/test-path3/test/dir/leadingslash"),
    ])
    @pytest.mark.parametrize('binary', [
        True,
        False
    ])
    @pytest.mark.skipif(not os.environ.get('AWS_ENDPOINT_URL'), reason="Skipping due to minio not configured")
    def test_write_file(self, bucket_path, dest_path, test_key, binary):
        """Upload file to s3"""

        with create_s3_file_storage_with_bucket(bucket_name="test-bucket", bucket_path=bucket_path) as instance:
            # Create test file - binary parameter should be unused
            instance.write_file(path=dest_path, content="Test Write content", binary=binary)

            # Obtain file from s3 to ensure it was uploaded correctly
            res = instance._s3_client.get_object(
                Bucket="test-bucket",
                Key=test_key
            )
            assert res['Body'].read() == "Test Write content".encode('utf-8')

    def test_delete_directory(self):
        """Test delete_directory method"""
        instance = terrareg.file_storage.S3FileStorage(s3_url='s3://test-bucket')
        assert instance.make_directory(directory="/does/not/exist") is None

    @pytest.mark.skipif(not os.environ.get('AWS_ENDPOINT_URL'), reason="Skipping due to minio not configured")
    def test_delete_file(self):
        """Test delete_file method"""
        with create_s3_file_storage_with_bucket(bucket_name="test-bucket", bucket_path="/test-base-dir/") as instance:
            # Upload file to s3
            bucket = instance._s3_resource.Bucket("test-bucket")
            bucket.put_object(
                Key="/test-base-dir/some-test/file-to-delete",
                Body="Test Content"
            )

            # Ensure file exists
            instance._s3_client.head_object(Bucket="test-bucket", Key="/test-base-dir/some-test/file-to-delete")

            instance.delete_file(path="/some-test/file-to-delete")

            with pytest.raises(botocore.exceptions.ClientError):
                instance._s3_client.head_object(Bucket="test-bucket", Key="/test-base-dir/some-test/file-to-delete")

    @pytest.mark.skipif(not os.environ.get('AWS_ENDPOINT_URL'), reason="Skipping due to minio not configured")
    def test_read_file_bytes_mode(self):
        """Test read_file method"""
        with create_s3_file_storage_with_bucket(bucket_name="test-bucket", bucket_path="/another-base-dir/") as instance:
            # Upload file to s3
            bucket = instance._s3_resource.Bucket("test-bucket")
            bucket.put_object(
                Key="/another-base-dir/some-test/file-to-read",
                Body="Test Content To Read"
            )

            expected_content = "Test Content To Read".encode("utf-8")

            file_handler = instance.read_file(path="/some-test/file-to-read", bytes_mode=True)
            assert file_handler is not None
            assert file_handler.read() == expected_content

    @pytest.mark.skipif(not os.environ.get('AWS_ENDPOINT_URL'), reason="Skipping due to minio not configured")
    def test_read_file_text_mode(self):
        """Test read_file method"""
        with create_s3_file_storage_with_bucket(bucket_name="test-bucket", bucket_path="/another-base-dir/") as instance:
            # Upload file to s3
            bucket = instance._s3_resource.Bucket("test-bucket")
            bucket.put_object(
                Key="/another-base-dir/some-test/file-to-read",
                Body="Test Content To Read"
            )

            with pytest.raises(NotImplementedError):
                instance.read_file(path="/some-test/file-to-read", bytes_mode=False)

    @pytest.mark.skipif(not os.environ.get('AWS_ENDPOINT_URL'), reason="Skipping due to minio not configured")
    def test_read_file_non_existent(self):
        """Test read_file method with non-existent file"""
        with create_s3_file_storage_with_bucket(bucket_name="test-bucket", bucket_path="/another-base-dir/") as instance:
            assert instance.read_file(path="/does/not/exist", bytes_mode=True) is None

    @pytest.mark.parametrize('exists', [
        True,
        False
    ])
    @pytest.mark.skipif(not os.environ.get('AWS_ENDPOINT_URL'), reason="Skipping due to minio not configured")
    def test_file_exists(self, exists):
        """Test file_exists method"""
        with create_s3_file_storage_with_bucket(bucket_name="test-bucket", bucket_path="/exists-base-dir/") as instance:
            # Upload file to s3
            if exists:
                bucket = instance._s3_resource.Bucket("test-bucket")
                bucket.put_object(
                    Key="/exists-base-dir/some-test/file-to-exist",
                    Body="Test Content To Read"
                )

            assert instance.file_exists("/some-test/file-to-exist") is exists

    def test_directory_exists(self):
        """Test test_directory_exists"""
        instance = terrareg.file_storage.S3FileStorage(s3_url='s3://test-bucket')
        assert instance.directory_exists(path="/does/not/exist") is True

    def test_make_directory(self):
        """Test make_directory"""
        instance = terrareg.file_storage.S3FileStorage(s3_url='s3://test-bucket')
        instance.make_directory(directory="/test/dir")
