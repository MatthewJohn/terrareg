
import pytest

import terrareg.models
from terrareg.analytics import AnalyticsEngine
from . import AnalyticsIntegrationTest


class TestRecordModuleVersionDownload(AnalyticsIntegrationTest):
    """Test record_module_version_download function."""

    _TEST_ANALYTICS_DATA = {}

    @pytest.mark.parametrize("user_agent, terraform_version", [
        # No user agent
        (None, "1.5.3"),

        # No version
        ("Terraform/1.5.2", None),

        # Invalid user-agent
        ("Go-http-client/1.1", "1.5.3"),
    ])
    def test_ignore_terraform_version(self, user_agent, terraform_version):
        """Test function with ignoring terraform version headers."""
        namespace = "testnamespace"
        module = "publishedmodule"
        provider = "testprovider"
        analytics_token = "test-invalid-version-headers"

        namespace_obj = terrareg.models.Namespace.get(namespace)
        module_obj = terrareg.models.Module(namespace_obj, module)
        provider_obj = terrareg.models.ModuleProvider.get(module_obj, provider)
        version_obj = terrareg.models.ModuleVersion.get(provider_obj, "1.4.0")

        # Clean up any analytics
        AnalyticsEngine.delete_analytics_for_module_version(version_obj)

        AnalyticsEngine.record_module_version_download(
            namespace_name=namespace, module_name=module, provider_name=provider,
            module_version=version_obj, terraform_version=terraform_version,
            analytics_token=analytics_token, user_agent=user_agent,
            auth_token=None
        )

        results = AnalyticsEngine.get_module_provider_token_versions(provider_obj)
        assert results == {
            'test-invalid-version-headers': {
                'environment': 'Default',
                'module_version': '1.4.0',
                'terraform_version': '0.0.0'
            }
        }

    @pytest.mark.parametrize("user_agent, terraform_version", [
        # Terraform
        ("Terraform/1.5.0", "1.5.3"),

        # OpenTofu
        ("OpenTofu/1.5.2", "1.5.3"),
    ])
    def test_valid_terraform_version(self, user_agent, terraform_version):
        """Test function with ignoring terraform version headers."""
        namespace = "testnamespace"
        module = "publishedmodule"
        provider = "testprovider"
        analytics_token = "test-with-version"

        namespace_obj = terrareg.models.Namespace.get(namespace)
        module_obj = terrareg.models.Module(namespace_obj, module)
        provider_obj = terrareg.models.ModuleProvider.get(module_obj, provider)
        version_obj = terrareg.models.ModuleVersion.get(provider_obj, "1.4.0")

        # Clean up any analytics
        AnalyticsEngine.delete_analytics_for_module_version(version_obj)

        AnalyticsEngine.record_module_version_download(
            namespace_name=namespace, module_name=module, provider_name=provider,
            module_version=version_obj, terraform_version=terraform_version,
            analytics_token=analytics_token, user_agent=user_agent,
            auth_token=None
        )

        results = AnalyticsEngine.get_module_provider_token_versions(provider_obj)
        assert results == {
            'test-with-version': {
                'environment': 'Default',
                'module_version': '1.4.0',
                'terraform_version': '1.5.3'
            }
        }
