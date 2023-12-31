
import pytest

import terrareg.provider_source_type


class TestProviderSourceType:
    """Test ProviderSourceType"""

    @pytest.mark.parametrize('value, expected_enum', [
        ('github', terrareg.provider_source_type.ProviderSourceType.GITHUB)
    ])
    def test_by_value(self, value, expected_enum):
        """Test getting by value"""
        assert terrareg.provider_source_type.ProviderSourceType(value) == expected_enum

    @pytest.mark.parametrize('value, expected_enum', [
        ('GITHUB', terrareg.provider_source_type.ProviderSourceType.GITHUB)
    ])
    def test_by_key(self, value, expected_enum):
        """Test getting by value"""
        assert terrareg.provider_source_type.ProviderSourceType[value] == expected_enum
