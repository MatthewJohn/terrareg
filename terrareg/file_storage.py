
import re
from typing import TextIO, Tuple
import abc
from io import BytesIO, TextIOWrapper
import os
import shutil

import boto3
import botocore.exceptions

import terrareg.config
from terrareg.errors import FileUploadError


class BaseFileStorage(abc.ABC):

    @abc.abstractmethod
    def upload_file(self, source_path: str, dest_directory: str, dest_filename: str) -> None:
        """Upload file to storage"""
        ...

    @abc.abstractmethod
    def file_exists(self, path: str) -> bool:
        """Check if file exists"""
        ...

    @abc.abstractmethod
    def directory_exists(self, path: str) -> bool:
        """Return if a directory exists"""
        ...

    @abc.abstractmethod
    def delete_file(self, path: str) -> None:
        """Delete file from storage"""
        ...

    @abc.abstractmethod
    def delete_directory(self, path: str) -> None:
        """Delete directory recursively from storage"""
        ...

    @abc.abstractmethod
    def read_file(self, path: str, bytes_mode: bool=False) -> TextIOWrapper:
        """Obtain file handle of file from storage"""
        ...

    @abc.abstractmethod
    def make_directory(self, directory: str) -> None:
        """Recursively create directory"""
        ...

    @abc.abstractmethod
    def write_file(self, path: str, content: any, binary: bool):
        """Write file to file storage from content"""
        ...


class LocalFileStorage(BaseFileStorage):
    """Handle local file storage."""

    def __init__(self, base_directory):
        """Store base directory"""
        self._base_directory = base_directory
        super().__init__()

    def _generate_path(self, *paths: str) -> str:
        """Generate real path of file, prepending base directory"""
        # Remove leading slash, as it does not allow the base
        # directory to be prepended
        paths = [
            path.lstrip(os.path.sep)
            for path in paths
        ]
        path = os.path.join(self._base_directory, *paths)

        # Add leading slash to path, if it does not contain one
        # before passing to abspath, as abspath will
        # prefix with CWD
        contains_leading_slash = path.startswith(os.path.sep)
        if not contains_leading_slash:
            path = f"{os.path.sep}{path}"

        path = os.path.abspath(path)
        if not contains_leading_slash:
            path = path[1:]

        return path

    def make_directory(self, directory: str):
        """Recursively create directory"""
        directory = self._generate_path(directory)
        os.makedirs(directory, exist_ok=True)

    def upload_file(self, source_path: str, dest_directory: str, dest_filename: str):
        """Upload file"""
        self._check_not_directory(dest_directory, dest_filename)

        # Create directory to store file
        self.make_directory(dest_directory)

        dest_directory = self._generate_path(dest_directory)
        dest_full_path = os.path.join(dest_directory, dest_filename)

        # Copy source file to destination
        shutil.copyfile(source_path, dest_full_path)

    def _check_not_directory(self, *paths):
        """Ensure path is not a directory"""
        # If destination already exists, but isn't a file, raise error.
        path = self._generate_path(*paths)
        if os.path.exists(path) and not os.path.isfile(path):
            raise FileUploadError("Destination already exists, but is not a file")

    def file_exists(self, path: str) -> bool:
        """Return if a file exists"""
        path = self._generate_path(path)
        return os.path.isfile(path)

    def directory_exists(self, path: str) -> bool:
        """Return if a directory exists"""
        path = self._generate_path(path)
        return os.path.isdir(path)

    def delete_file(self, path: str) -> None:
        """Delete path"""
        path = self._generate_path(path)
        os.unlink(path)

    def delete_directory(self, path: str) -> None:
        """Delete path"""
        path = self._generate_path(path)
        os.rmdir(path)

    def read_file(self, path: str, bytes_mode: bool=False) -> TextIOWrapper:
        """Return file handler for file"""
        path = self._generate_path(path)
        mode = "r"
        if bytes_mode:
            mode += "b"
        return open(path, mode)

    def write_file(self, path: str, content: any, binary: bool):
        """Write file to file storage from content"""
        # Ensure destination is not a directory
        self._check_not_directory(path)

        # Create directory to store file
        dest_directory = os.path.abspath(os.path.join(path, '..'))
        self.make_directory(dest_directory)

        path = self._generate_path(path)
        mode = "w"
        if binary:
            mode += "b"
        with open(path, mode) as fh:
            fh.write(content)


class S3FileStorage(BaseFileStorage):
    """Handle file storage in s3"""

    def __init__(self, s3_url) -> None:
        """Store member variables"""
        self._s3_url = s3_url
        self._bucket_name, self._base_s3_path = self._get_path_details(s3_url)

        self._session = boto3.session.Session()
        self._s3_client = self._session.client('s3')
        self._s3_resource = self._session.resource('s3')
        super().__init__()

    def _get_path_details(self, s3_url) -> Tuple[str, str]:
        """Obtain bucket name and base path from s3 path"""
        match = re.match(r"s3://([^/]+)((:?/.*)?)$", s3_url)
        if not match:
            raise Exception("Invalid s3 path for DATA_DIRECTORY. Must be in the form: s3://BUCKETNAME/")

        bucket = match.group(1)
        path = match.group(2)
        return bucket, path

    def _get_bucket(self):
        """Get bucket object"""
        return self._s3_resource.Bucket(self._bucket_name)

    def _generate_key(self, *paths):
        """Generate s3 key"""
        path = "/".join([self._base_s3_path, *paths])
        # Replace any double slashes
        path = path.replace('//', '/')
        if not path.startswith('/'):
            path = f'/{path}'
        return path


    def upload_file(self, source_path: str, dest_directory: str, dest_filename: str) -> None:
        """Upload file to s3"""
        with open(source_path, 'rb') as fh:
            content = fh.read()
        self.write_file(path=f'{dest_directory}/{dest_filename}', content=content, binary=True)

    def write_file(self, path: str, content: any, binary: bool):
        """Write file to file storage from content"""
        key = self._generate_key(path)

        self._get_bucket().put_object(
            Key=key,
            Body=content
        )

    def delete_directory(self, path: str) -> None:
        """Delete directory from s3"""
        # There is no method required to delete a directory
        pass

    def delete_file(self, path: str) -> None:
        """Delete key from s3"""
        path = self._generate_key(path)
        self._s3_client.delete_object(Bucket=self._bucket_name, Key=path)

    def read_file(self, path: str, bytes_mode: bool = False) -> TextIOWrapper:
        """Obtain FH containing contents of file from s3"""
        if bytes_mode:
            content = BytesIO()
        else:
            content = TextIO()

        key = self._generate_key(path)

        try:
            self._get_bucket().download_fileobj(Key=key, Fileobj=content)
        except botocore.exceptions.ClientError:
            return None

        content.seek(0)
        return content

    def file_exists(self, path: str) -> bool:
        """Check if object exists in s3"""
        path = self._generate_key(path)
        try:
            self._s3_client.head_object(Bucket=self._bucket_name, Key=path)
            return True
        except botocore.exceptions.ClientError as exc:
            if "Not Found" in str(exc):
                return False
            raise

    def directory_exists(self, path: str) -> bool:
        """Check if directory exists"""
        # Always return True.
        # This will lead to logic where, directories are always removed,
        # but also means files can be uploaded, since directories
        # do not need creating.
        return True

    def make_directory(self, directory: str) -> None:
        """Create directory"""
        # Directories do not need to be created in s3
        pass


class FileStorageFactory:

    def get_file_storage(self) -> 'BaseFileStorage':
        """Generate file storage instance"""
        config = terrareg.config.Config()
        if config.DATA_DIRECTORY.startswith("s3://"):
            return S3FileStorage(config.DATA_DIRECTORY)
        else:
            return LocalFileStorage(config.DATA_DIRECTORY)
