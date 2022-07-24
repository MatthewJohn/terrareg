
import pytest

from terrareg.utils import safe_join_paths
from test.unit.terrareg import TerraregUnitTest


class TestSafeJoinPaths(TerraregUnitTest):

    @pytest.mark.parametrize('base_dir,sub_paths,expected_output', [
        # Basic test
        ('/root_dir', ['subdirectory'], '/root_dir/subdirectory'),

        # Multiple sub-directories
        ('/root_dir', ['subdirectory', 'subdir2'], '/root_dir/subdirectory/subdir2'),

        # Leading dot-slash in subdirectory
        ('/root_dir', ['./leadingslash'], '/root_dir/leadingslash'),

        # Leading slash in subdirectory
        ('/root_dir', ['/leadingdotslash'], '/root_dir/leadingdotslash'),
    ])
    def test_valid_paths(self, base_dir, sub_paths, expected_output):
        """Test valid path using safe_join_paths method."""
        assert safe_join_paths(base_dir, *sub_paths) == expected_output
