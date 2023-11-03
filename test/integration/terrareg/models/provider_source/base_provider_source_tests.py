
from .test_base_provider_source import TestBaseProviderSource
import terrareg.database
import terrareg.provider_source


class BaseProviderSourceTests(TestBaseProviderSource):
    """Base tests for child classes of BaseProviderSource"""

    ADDITIONAL_CONFIG = None

    def test_generate_db_config_from_source_config(self):
        """Test generate_db_config_from_source_config"""
        raise NotImplementedError

    def test_login_button_text(self):
        """Test login_button_text property"""
        raise NotImplementedError

    def test_get_user_access_token(self):
        """Test get_user_access_token"""
        raise NotImplementedError

    def test_update_repositories(self):
        """Test update_repositories"""
        raise NotImplementedError

    def test_refresh_namespace_repositories(self):
        """Test refresh_namespace_repositories"""
        raise NotImplementedError

    def test_get_new_releases(self):
        """Test get_new_releases"""
        raise NotImplementedError

    def test_get_release_artifact(self):
        """Test get_release_artifact"""
        raise NotImplementedError

    def test_get_release_archive(self):
        """Test get_release_archive"""
        raise NotImplementedError

    def test_get_public_source_url(self):
        """Test get_public_source_url"""
        raise NotImplementedError

    def test_get_public_artifact_download_url(self):
        """Test get_public_artifact_download_url"""
        raise NotImplementedError
