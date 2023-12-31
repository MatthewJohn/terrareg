
import pytest

from test.integration.terrareg import TerraregIntegrationTest
import terrareg.repository_kind


class TestRepositoryKind(TerraregIntegrationTest):
    """Test Registry resource type"""

    @pytest.mark.parametrize('value, expected_enum', [
        ('module', terrareg.repository_kind.RepositoryKind.MODULE),
        ('provider', terrareg.repository_kind.RepositoryKind.PROVIDER),
    ])
    def test_by_value(self, value, expected_enum):
        """Test getting by value"""
        assert terrareg.repository_kind.RepositoryKind(value) == expected_enum

    @pytest.mark.parametrize('value, expected_enum', [
        ('MODULE', terrareg.repository_kind.RepositoryKind.MODULE),
        ('PROVIDER', terrareg.repository_kind.RepositoryKind.PROVIDER),
    ])
    def test_by_key(self, value, expected_enum):
        """Test getting by value"""
        assert terrareg.repository_kind.RepositoryKind[value] == expected_enum
