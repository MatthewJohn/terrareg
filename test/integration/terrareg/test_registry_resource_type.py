
import pytest

from test.integration.terrareg import TerraregIntegrationTest
import terrareg.registry_resource_type


class TestRegistryResourceType(TerraregIntegrationTest):
    """Test Registry resource type"""

    @pytest.mark.parametrize('value, expected_enum', [
        ('module', terrareg.registry_resource_type.RegistryResourceType.MODULE),
        ('provider', terrareg.registry_resource_type.RegistryResourceType.PROVIDER),
    ])
    def test_by_value(self, value, expected_enum):
        """Test getting by value"""
        assert terrareg.registry_resource_type.RegistryResourceType(value) == expected_enum

    @pytest.mark.parametrize('value, expected_enum', [
        ('MODULE', terrareg.registry_resource_type.RegistryResourceType.MODULE),
        ('PROVIDER', terrareg.registry_resource_type.RegistryResourceType.PROVIDER),
    ])
    def test_by_key(self, value, expected_enum):
        """Test getting by value"""
        assert terrareg.registry_resource_type.RegistryResourceType[value] == expected_enum
