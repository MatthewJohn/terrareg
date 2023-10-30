
from test.integration.terrareg import TerraregIntegrationTest
import terrareg.registry_resource_type


class TestRegistryResourceType(TerraregIntegrationTest):
    """Test Registry resource type"""

    def test_module(self):
        """Test module"""
        assert terrareg.registry_resource_type.RegistryResourceType.MODULE.value == "module"

    def test_provider(self):
        """Test provider"""
        assert terrareg.registry_resource_type.RegistryResourceType.PROVIDER.value == "provider"