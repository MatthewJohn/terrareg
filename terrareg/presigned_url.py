
import datetime
import jwt

import terrareg.config
from terrareg.errors import InvalidPresignedUrlKeyError, PresignedUrlsNotConfiguredError
from terrareg.utils import get_datetime_now


class TerraformSourcePresignedUrl:

    @classmethod
    def is_enabled(cls):
        """Whether pre-signed URLs are available"""
        return bool(cls.get_secret())

    @classmethod
    def get_secret(cls):
        """Get secret for JWT signature encryption"""
        return terrareg.config.Config().TERRAFORM_PRESIGNED_URL_SECRET

    @classmethod
    def get_expiry(cls):
        """Get expiry"""
        expiry = terrareg.config.Config().TERRAFORM_PRESIGNED_URL_EXPIRY_SECONDS
        return (get_datetime_now() + datetime.timedelta(seconds=expiry)).isoformat()

    @classmethod
    def get_algorithm(cls):
        """Return JWT algorithms supported"""
        return "HS256"

    @classmethod
    def expiry_is_valid(cls, expiry):
        """Check expiry value from payload"""
        if not expiry or type(expiry) is not str:
            return False

        try:
            expiry_dt = datetime.datetime.fromisoformat(expiry)
        except ValueError:
            # Handle invalid format
            return False
        
        # If expiry is in the past, do not allow
        if expiry_dt < get_datetime_now():
            return False

        # If all checks have passed, return False
        return True

    @classmethod
    def generate_presigned_key(cls, url):
        """Generate presigned key for URL"""
        if not cls.is_enabled():
            raise PresignedUrlsNotConfiguredError("Presigned URL configurations are not present. Please see documentation")

        payload = {
            "expiry": cls.get_expiry(),
            "url": url
        }
        return jwt.encode(payload=payload, key=cls.get_secret(), algorithm=cls.get_algorithm())

    @classmethod
    def validate_presigned_key(cls, url, payload):
        """Ensure provided pre-signed key is valid"""
        if not cls.is_enabled():
            raise PresignedUrlsNotConfiguredError("Presigned URL configurations are not present. Please see documentation")

        # Generate exception that is identicle to end user,
        # but can be raised multiple in multiple places to allow
        # identification of specific problem to system maintainers 
        generic_exception = InvalidPresignedUrlKeyError("Invalid pre-signed URL key")

        try:
            decrypted_payload = jwt.decode(jwt=payload, key=cls.get_secret(), algorithms=[cls.get_algorithm()])

        except (jwt.exceptions.DecodeError, jwt.exceptions.InvalidSignatureError):
            raise generic_exception

        if type(decrypted_payload) is not dict:
            raise generic_exception

        # Ensure expiry is still valid
        if not cls.expiry_is_valid(decrypted_payload.get("expiry")):
            raise generic_exception
        
        # Ensure URL in presigned token matches the actual URL
        if url != decrypted_payload.get("url"):
            raise generic_exception
