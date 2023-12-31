
import pytest

import terrareg.provider_documentation_type


class TestProviderSourceType:
    """Test ProviderDocumentationType"""

    @pytest.mark.parametrize('value, expected_enum', [
        ('overview', terrareg.provider_documentation_type.ProviderDocumentationType.OVERVIEW),
        ('provider', terrareg.provider_documentation_type.ProviderDocumentationType.PROVIDER),
        ('resources', terrareg.provider_documentation_type.ProviderDocumentationType.RESOURCE),
        ('data-sources', terrareg.provider_documentation_type.ProviderDocumentationType.DATA_SOURCE),
        ('guides', terrareg.provider_documentation_type.ProviderDocumentationType.GUIDE),
    ])
    def test_by_value(self, value, expected_enum):
        """Test getting by value"""
        assert terrareg.provider_documentation_type.ProviderDocumentationType(value) == expected_enum

    @pytest.mark.parametrize('value, expected_enum', [
        ('OVERVIEW', terrareg.provider_documentation_type.ProviderDocumentationType.OVERVIEW),
        ('PROVIDER', terrareg.provider_documentation_type.ProviderDocumentationType.PROVIDER),
        ('RESOURCE', terrareg.provider_documentation_type.ProviderDocumentationType.RESOURCE),
        ('DATA_SOURCE', terrareg.provider_documentation_type.ProviderDocumentationType.DATA_SOURCE),
        ('GUIDE', terrareg.provider_documentation_type.ProviderDocumentationType.GUIDE),
    ])
    def test_by_key(self, value, expected_enum):
        """Test getting by value"""
        assert terrareg.provider_documentation_type.ProviderDocumentationType[value] == expected_enum
