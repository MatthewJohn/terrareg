
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

        # Test with lots of slashes
        ('/root_dir', ['.//lots///of//slashes/'], '/root_dir/lots/of/slashes'),

        # Test starting from rooot
        ('/', ['test_subdir'], '/test_subdir')

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

    @pytest.mark.parametrize('is_allowed', [True, False])
    @pytest.mark.parametrize('base_dir,sub_paths', [
        ('/test', ['testsub/..']),
        ('/testdir', ['./testsubdirectory/../']),
        ('/testdir', ['./testsubdir', '../']),
        ('/testdir', ['./testsubdir', '..']),
        ('/testdir', ['/']),
        ('/testdir', ['./']),
        ('/testdir', ['/./']),
        ('/', ['/./']),
    ])
    def test_matching_directories(self, is_allowed, base_dir, sub_paths):
        """Test valid path using safe_join_paths method."""
        if is_allowed:
            assert terrareg.utils.safe_join_paths(base_dir, *sub_paths, allow_same_directory=True) == base_dir
        else:
            with pytest.raises(terrareg.utils.PathIsNotWithinBaseDirectoryError):
                assert terrareg.utils.safe_join_paths(base_dir, *sub_paths)

