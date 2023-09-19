
from test.unit.terrareg import TerraregUnitTest


class BaseAuthMethodTest(TerraregUnitTest):
    """Test common smaller methods of auth methods"""

    def test_is_built_in_admin(self):
        """Test is_built_in_admin method"""
        raise NotImplementedError

    def test_is_admin(self):
        """Test is_admin method"""
        raise NotImplementedError

    def test_is_authenticated(self):
        """Test is_authenticated method"""
        raise NotImplementedError

    def test_is_enabled(self):
        """Test is_enabled method"""
        raise NotImplementedError

    def test_requires_csrf_tokens(self):
        """Test requires_csrf_token method"""
        raise NotImplementedError

    def test_can_publish_module_version(self):
        """Test can_publish_module_version method"""
        raise NotImplementedError

    def test_can_upload_module_version(self):
        """Test can_upload_module_version method"""
        raise NotImplementedError

    def test_check_auth_state(self):
        """Test check_auth_state method"""
        raise NotImplementedError

    def test_check_namespace_access(self):
        """Test check_namespace_access method"""
        raise NotImplementedError

    def test_get_username(self):
        """Test check_username method"""
        raise NotImplementedError

    def test_can_access_read_api(self):
        """Test can_access_read_api method"""
        raise NotImplementedError

    def test_can_access_terraform_api(self):
        """Test can_access_terraform_api method"""
        raise NotImplementedError
