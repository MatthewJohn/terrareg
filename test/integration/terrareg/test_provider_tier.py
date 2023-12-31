
import pytest

from test.integration.terrareg import TerraregIntegrationTest
import terrareg.provider_tier


class TestProviderTier(TerraregIntegrationTest):
    """Test Registry resource type"""

    @pytest.mark.parametrize('value, expected_enum', [
        ('community', terrareg.provider_tier.ProviderTier.COMMUNITY),
        ('official', terrareg.provider_tier.ProviderTier.OFFICIAL),
    ])
    def test_by_value(self, value, expected_enum):
        """Test getting by value"""
        assert terrareg.provider_tier.ProviderTier(value) == expected_enum

    @pytest.mark.parametrize('value, expected_enum', [
        ('COMMUNITY', terrareg.provider_tier.ProviderTier.COMMUNITY),
        ('OFFICIAL', terrareg.provider_tier.ProviderTier.OFFICIAL),
    ])
    def test_by_key(self, value, expected_enum):
        """Test getting by value"""
        assert terrareg.provider_tier.ProviderTier[value] == expected_enum
