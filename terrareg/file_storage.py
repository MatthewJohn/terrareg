

from abc import ABC
from io import TextIOWrapper
import os
import shutil

import terrareg.config


class BaseFileStorage(ABC):

    def upload_file(self, source_path: str, dest_directory: str, dest_filename: str):
        ...

    def file_exists(self, path: str) -> bool:
        ...

    def delete_file(self, path: str):
        ...

    def read_file(self, path: str, bytes_mode: bool=False) -> TextIOWrapper:
        ...

    def make_directory(self, directory: str) -> None:
        """Recursively create directory"""
        ...


class LocalFileStorage(BaseFileStorage):

    def __init__(self, base_directory):
        """Store base directory"""
        self._base_directory = base_directory

    def _generate_path(self, *paths: str) -> str:
        return os.path.join(self._base_directory, *paths)

    def make_directory(self, directory: str):

        directory = self._generate_path(directory)
        # os.mkdir(directory)
        os.makedirs(directory, exist_ok=True)

    def upload_file(self, source_path: str, dest_directory: str, dest_filename: str):
        """Upload file"""
        dest_directory = self._generate_path(dest_directory)
        # Create all parent directories
        os.makedirs(dest_directory, exist_ok=True)
        dest_full_path = os.path.join(dest_directory, dest_filename)
        # If destination already exists, but isnt a file, raise error.
        if os.path.exists(dest_full_path) and not os.path.isfile(dest_full_path):
            raise Exception("Destination already exists, but is not a file")

        # Copy source file to destination
        shutil.copyfile(source_path, dest_full_path)

    def file_exists(self, path: str) -> bool:
        """Return if a file exists"""
        path = self._generate_path(path)
        return os.path.isfile(path)

    def delete_file(self, path: str):
        """Delete path"""
        path = self._generate_path(path)
        if self.file_exists(path):
            os.unlink(path)

    def delete_directory(self, path: str):
        """Delete path"""
        path = self._generate_path(path)
        raise Exception(f"Going to delete direcxtory: {path}")
        if os.path.exists(path):
            shutil.rmtree(path)

    def read_file(self, path: str, bytes_mode: bool=False):
        """Return filehandler for file"""
        path = self._generate_path(path)
        mode = "r"
        if bytes_mode:
            mode += "b"
        return open(path, mode)


class S3FileStorage(BaseFileStorage):
    pass


class FileStorageFactory:

    def get_file_storage(self) -> 'BaseFileStorage':
        """Generate file storage instance"""
        config = terrareg.config.Config()
        if config.DATA_DIRECTORY.startswith("s3://"):
            return S3FileStorage(config.DATA_DIRECTORY)
        else:
            return LocalFileStorage(config.DATA_DIRECTORY)
