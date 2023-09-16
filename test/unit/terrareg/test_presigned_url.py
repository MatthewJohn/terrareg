
import datetime
import unittest.mock

import pytest
import jwt

from terrareg.errors import InvalidPresignedUrlKeyError, PresignedUrlsNotConfiguredError
from terrareg.presigned_url import TerraformSourcePresignedUrl


class TestTerraformSourcePresignedUrl:

    @pytest.mark.parametrize('now, expiry_config, expected_value', [
        # 10 seconds from now
        (datetime.datetime(2023, 9, 16, 6, 50, 24, 968701), 10, "2023-09-16T06:50:34.968701"),
        # 60 seconds from now
        (datetime.datetime(2023, 9, 16, 6, 50, 24, 968701), 60, "2023-09-16T06:51:24.968701"),
    ])
    def test_get_expiry(self, now, expiry_config, expected_value):
        """Test get_expiry method"""
        with unittest.mock.patch('terrareg.presigned_url.get_datetime_now', unittest.mock.MagicMock(return_value=now)), \
                unittest.mock.patch('terrareg.config.Config.TERRAFORM_PRESIGNED_URL_EXPIRY_SECONDS', expiry_config):

            assert TerraformSourcePresignedUrl.get_expiry() == expected_value

    def test_get_secret(self):
        """Test get secret method"""
        with unittest.mock.patch('terrareg.config.Config.TERRAFORM_PRESIGNED_URL_SECRET', "unittest-secret"):
            assert TerraformSourcePresignedUrl.get_secret() == "unittest-secret"

    def test_get_algorithm(self):
        """Test get_algorithm method"""
        assert TerraformSourcePresignedUrl.get_algorithm() == "HS256"

    def test_generate_presigned_key_no_secret(self):
        """Test generate_presigned_key_no_secret with no secret value"""
        with unittest.mock.patch('terrareg.config.Config.TERRAFORM_PRESIGNED_URL_SECRET', None):
            with pytest.raises(PresignedUrlsNotConfiguredError):
                TerraformSourcePresignedUrl.generate_presigned_key('/test-url')

    def test_generate_presigned_key(self):
        now = datetime.datetime(2023, 9, 15, 4, 32, 20, 123456)

        with unittest.mock.patch('terrareg.presigned_url.get_datetime_now', unittest.mock.MagicMock(return_value=now)), \
                unittest.mock.patch('terrareg.config.Config.TERRAFORM_PRESIGNED_URL_EXPIRY_SECONDS', 10), \
                unittest.mock.patch('terrareg.config.Config.TERRAFORM_PRESIGNED_URL_SECRET', "unittest-secret-key"):

            presign_key = TerraformSourcePresignedUrl.generate_presigned_key('/some-test/path')

            assert type(presign_key) == str

            # Decode JWT
            val = jwt.decode(presign_key, key="unittest-secret-key", algorithms=["HS256"])
            assert val["expiry"] == "2023-09-15T04:32:30.123456"
            assert val["url"] == "/some-test/path"

    def test_validate_presigned_key(self):
        now = datetime.datetime(2023, 9, 15, 4, 32, 1, 123456)

        with unittest.mock.patch('terrareg.presigned_url.get_datetime_now', unittest.mock.MagicMock(return_value=now)), \
                unittest.mock.patch('terrareg.config.Config.TERRAFORM_PRESIGNED_URL_EXPIRY_SECONDS', 10), \
                unittest.mock.patch('terrareg.config.Config.TERRAFORM_PRESIGNED_URL_SECRET', "unittest-secret-key"):

            # Encode JWT
            val = jwt.encode(
                {"expiry": "2023-09-15T04:32:30.123456", "url": "/some-test/path"},
                key="unittest-secret-key",
                algorithm="HS256"
            )

            TerraformSourcePresignedUrl.validate_presigned_key(url='/some-test/path', payload=val)

    @pytest.mark.parametrize('payload_expiry, payload_path, payload_key', [
        # Incorrect path
        ("2023-09-15T04:35:05.123456", "/some-test/another", "unittest-secret-key"),
        ("2023-09-15T04:35:05.123456", "", "unittest-secret-key"),
        ("2023-09-15T04:35:05.123456", None, "unittest-secret-key"),

        # Expired
        ("2023-09-15T04:31:00.123456", "/some-test/path", "unittest-secret-key"),

        # Invalid date field values
        ("2023-09-15", "/some-test/path", "unittest-secret-key"),
        ("", "/some-test/path", "unittest-secret-key"),
        (None, "/some-test/path", "unittest-secret-key"),
        (["a" "list"], "/some-test/path", "unittest-secret-key"),
        ({"a": "map"}, "/some-test/path", "unittest-secret-key"),

        # Signed with different key
        ("2023-09-15T04:35:05.123456", "/some-test/path", "another-key"),
    ])
    def test_validate_presigned_key_invalid(self, payload_expiry, payload_path, payload_key):
        now = datetime.datetime(2023, 9, 15, 4, 32, 1, 531521)

        with unittest.mock.patch('terrareg.presigned_url.get_datetime_now', unittest.mock.MagicMock(return_value=now)), \
                unittest.mock.patch('terrareg.config.Config.TERRAFORM_PRESIGNED_URL_SECRET', "unittest-secret-key"):

            # Encode JWT
            val = jwt.encode(
                {"expiry": payload_expiry, "url": payload_path},
                key=payload_key,
                algorithm="HS256"
            )
            print({"expiry": payload_expiry, "url": payload_path})

            with pytest.raises(InvalidPresignedUrlKeyError):
                TerraformSourcePresignedUrl.validate_presigned_key(url='/some-test/path', payload=val)
