
import pytest

import terrareg.utils
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

        # Test with relative path within sub-directory
        ('/root_dir', ['/subdirectory', '../actualdirectory'], '/root_dir/actualdirectory'),

    ])
    def test_valid_paths(self, base_dir, sub_paths, expected_output):
        """Test valid path using safe_join_paths method."""
        assert terrareg.utils.safe_join_paths(base_dir, *sub_paths) == expected_output

    @pytest.mark.parametrize('base_dir,sub_paths', [
        # Basic test
        ('/root_dir', ['../subdirectory']),

        # Multiple sub-directories with exit root directory on first subdir
        ('/root_dir', ['../../subdirectory', 'subdir2']),

        # Leading exit root directory in second sub directory 
        ('/root_dir', ['./tosubdir', '../../outofrootdir']),

        # Leading slash in subdirectory
        ('/root_dir', ['/../../leadingdotslash']),
    ])
    def test_invalid_paths(self, base_dir, sub_paths):
        """Test valid path using safe_join_paths method."""
        with pytest.raises(terrareg.utils.PathIsNotWithinBaseDirectoryError):
            assert terrareg.utils.safe_join_paths(base_dir, *sub_paths)

