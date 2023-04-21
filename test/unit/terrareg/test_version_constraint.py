
import pytest


import terrareg.version_constraint
from test.unit.terrareg import TerraregUnitTest

class TestVersionConstraint(TerraregUnitTest):

    @pytest.mark.parametrize('constraint, target_version', [
        # Invalid constraints
        ('abcde', '1.0.0'),
        ('=abcdefg', '1.0.0'),
        ('>abcdefg', '1.0.0'),
        ('=1.2.3abcd', '1.0.0'),

        # Test invalid target version
        ('=1.0.0', None),
        ('=1.0.0', ''),
        ('=1.0.0', 'blah'),
        ('=1.0.0', '...')
    ])
    def test_failure_conditions(self, constraint, target_version):
        """Test failure conditions"""
        assert terrareg.version_constraint.VersionConstraint.is_compatible(
            constraint=constraint,
            target_version=target_version) == terrareg.version_constraint.VersionCompatibilityType.ERROR

    @pytest.mark.parametrize('constraint', [
        None,
        '',
    ])
    def test_no_constraint(self, constraint):
        """Test constraint compatiblity with no constraint"""
        assert terrareg.version_constraint.VersionConstraint.is_compatible(
            constraint=constraint,
            target_version="1.0.0") == terrareg.version_constraint.VersionCompatibilityType.NO_CONSTRAINT


    @pytest.mark.parametrize('constraint, target_version, expected_result', [
        ('=1.2.3', '1.2.3', terrareg.version_constraint.VersionCompatibilityType.IMPLICIT_COMPATIBLE),
        ('=1.2.3', '2.2.3', terrareg.version_constraint.VersionCompatibilityType.INCOMPATIBLE),
        ('=1.2.3', '1.3.3', terrareg.version_constraint.VersionCompatibilityType.INCOMPATIBLE),
        ('=1.2.3', '1.2.4', terrareg.version_constraint.VersionCompatibilityType.INCOMPATIBLE),
        ('=1.2.3', '0.2.3', terrareg.version_constraint.VersionCompatibilityType.INCOMPATIBLE),
        ('=1.2.3', '1.1.3', terrareg.version_constraint.VersionCompatibilityType.INCOMPATIBLE),
        ('=1.2.3', '1.2.2', terrareg.version_constraint.VersionCompatibilityType.INCOMPATIBLE),

        ('!=1.2.3', '1.2.3', terrareg.version_constraint.VersionCompatibilityType.INCOMPATIBLE),
        ('!=1.2.3', '2.2.3', terrareg.version_constraint.VersionCompatibilityType.IMPLICIT_COMPATIBLE),
        ('!=1.2.3', '1.3.3', terrareg.version_constraint.VersionCompatibilityType.IMPLICIT_COMPATIBLE),
        ('!=1.2.3', '1.2.4', terrareg.version_constraint.VersionCompatibilityType.IMPLICIT_COMPATIBLE),
        ('!=1.2.3', '0.2.3', terrareg.version_constraint.VersionCompatibilityType.IMPLICIT_COMPATIBLE),
        ('!=1.2.3', '1.1.3', terrareg.version_constraint.VersionCompatibilityType.IMPLICIT_COMPATIBLE),
        ('!=1.2.3', '1.2.2', terrareg.version_constraint.VersionCompatibilityType.IMPLICIT_COMPATIBLE),

        ('>3.2.1', '3.2.1', terrareg.version_constraint.VersionCompatibilityType.INCOMPATIBLE),
        ('>3.2.1', '1.5.5', terrareg.version_constraint.VersionCompatibilityType.INCOMPATIBLE),
        ('>3.2.1', '3.2.2', terrareg.version_constraint.VersionCompatibilityType.IMPLICIT_COMPATIBLE),
        ('>3.2.1', '3.3.1', terrareg.version_constraint.VersionCompatibilityType.IMPLICIT_COMPATIBLE),
        ('>3.2.1', '4.1.1', terrareg.version_constraint.VersionCompatibilityType.IMPLICIT_COMPATIBLE),

        ('>=3.2.1', '3.2.1', terrareg.version_constraint.VersionCompatibilityType.IMPLICIT_COMPATIBLE),
        ('>=3.2.1', '3.2.0', terrareg.version_constraint.VersionCompatibilityType.INCOMPATIBLE),
        ('>=3.2.1', '3.1.1', terrareg.version_constraint.VersionCompatibilityType.INCOMPATIBLE),
        ('>=3.2.1', '2.2.1', terrareg.version_constraint.VersionCompatibilityType.INCOMPATIBLE),
        ('>=3.2.1', '1.5.5', terrareg.version_constraint.VersionCompatibilityType.INCOMPATIBLE),
        ('>=3.2.1', '3.2.2', terrareg.version_constraint.VersionCompatibilityType.IMPLICIT_COMPATIBLE),
        ('>=3.2.1', '3.3.1', terrareg.version_constraint.VersionCompatibilityType.IMPLICIT_COMPATIBLE),
        ('>=3.2.1', '4.2.1', terrareg.version_constraint.VersionCompatibilityType.IMPLICIT_COMPATIBLE),

        ('<3.2.1', '3.2.1', terrareg.version_constraint.VersionCompatibilityType.INCOMPATIBLE),
        ('<3.2.1', '4.5.5', terrareg.version_constraint.VersionCompatibilityType.INCOMPATIBLE),
        ('<3.2.1', '3.2.0', terrareg.version_constraint.VersionCompatibilityType.IMPLICIT_COMPATIBLE),
        ('<3.2.1', '3.1.1', terrareg.version_constraint.VersionCompatibilityType.IMPLICIT_COMPATIBLE),
        ('<3.2.1', '2.2.1', terrareg.version_constraint.VersionCompatibilityType.IMPLICIT_COMPATIBLE),

        ('<=3.2.1', '3.2.1', terrareg.version_constraint.VersionCompatibilityType.IMPLICIT_COMPATIBLE),
        ('<=3.2.1', '3.2.2', terrareg.version_constraint.VersionCompatibilityType.INCOMPATIBLE),
        ('<=3.2.1', '3.3.1', terrareg.version_constraint.VersionCompatibilityType.INCOMPATIBLE),
        ('<=3.2.1', '4.2.1', terrareg.version_constraint.VersionCompatibilityType.INCOMPATIBLE),
        ('<=3.2.1', '4.1.1', terrareg.version_constraint.VersionCompatibilityType.INCOMPATIBLE),
        ('<=3.2.1', '3.2.0', terrareg.version_constraint.VersionCompatibilityType.IMPLICIT_COMPATIBLE),
        ('<=3.2.1', '3.1.1', terrareg.version_constraint.VersionCompatibilityType.IMPLICIT_COMPATIBLE),
        ('<=3.2.1', '1.2.1', terrareg.version_constraint.VersionCompatibilityType.IMPLICIT_COMPATIBLE),

        ('~>3.2.1', '3.2.1', terrareg.version_constraint.VersionCompatibilityType.COMPATIBLE),
        ('~>3.2.1', '3.2.2', terrareg.version_constraint.VersionCompatibilityType.COMPATIBLE),
        ('~>3.2.1', '3.2.0', terrareg.version_constraint.VersionCompatibilityType.COMPATIBLE),
        ('~>3.2.1', '3.1.1', terrareg.version_constraint.VersionCompatibilityType.INCOMPATIBLE),
        ('~>3.2.1', '3.3.1', terrareg.version_constraint.VersionCompatibilityType.INCOMPATIBLE),
        ('~>3.2.1', '2.2.1', terrareg.version_constraint.VersionCompatibilityType.INCOMPATIBLE),
        ('~>3.2.1', '4.2.1', terrareg.version_constraint.VersionCompatibilityType.INCOMPATIBLE),

        ('~>3.2', '3.2.1', terrareg.version_constraint.VersionCompatibilityType.COMPATIBLE),
        ('~>3.2', '3.2.2', terrareg.version_constraint.VersionCompatibilityType.COMPATIBLE),
        ('~>3.2', '3.2.0', terrareg.version_constraint.VersionCompatibilityType.COMPATIBLE),
        ('~>3.2', '3.1.1', terrareg.version_constraint.VersionCompatibilityType.COMPATIBLE),
        ('~>3.2', '3.3.1', terrareg.version_constraint.VersionCompatibilityType.COMPATIBLE),
        ('~>3.2', '2.2.1', terrareg.version_constraint.VersionCompatibilityType.INCOMPATIBLE),
        ('~>3.2', '4.2.1', terrareg.version_constraint.VersionCompatibilityType.INCOMPATIBLE),

        ('~>3', '3.2.1', terrareg.version_constraint.VersionCompatibilityType.COMPATIBLE),
        ('~>3', '3.2.2', terrareg.version_constraint.VersionCompatibilityType.COMPATIBLE),
        ('~>3', '3.2.0', terrareg.version_constraint.VersionCompatibilityType.COMPATIBLE),
        ('~>3', '3.1.1', terrareg.version_constraint.VersionCompatibilityType.COMPATIBLE),
        ('~>3', '3.3.1', terrareg.version_constraint.VersionCompatibilityType.COMPATIBLE),
        ('~>3', '2.2.1', terrareg.version_constraint.VersionCompatibilityType.COMPATIBLE),
        ('~>3', '4.2.1', terrareg.version_constraint.VersionCompatibilityType.COMPATIBLE),

        # Test multiple
        ('>2.0.0, <3.0.0', '2.0.0', terrareg.version_constraint.VersionCompatibilityType.INCOMPATIBLE),
        ('>2.0.0, <3.0.0', '3.0.0', terrareg.version_constraint.VersionCompatibilityType.INCOMPATIBLE),
        ('>2.0.0, <3.0.0', '2.0.1', terrareg.version_constraint.VersionCompatibilityType.COMPATIBLE),
        ('>2.0.0, <3.0.1', '3.0.0', terrareg.version_constraint.VersionCompatibilityType.COMPATIBLE),

        ('>=2.0.1, <=3.0.0', '2.0.0', terrareg.version_constraint.VersionCompatibilityType.INCOMPATIBLE),
        ('>=2.0.0, <=3.0.0', '3.0.1', terrareg.version_constraint.VersionCompatibilityType.INCOMPATIBLE),
        ('>=2.0.0, <=3.0.0', '2.0.0', terrareg.version_constraint.VersionCompatibilityType.COMPATIBLE),
        ('>=2.0.0, <=3.0.0', '3.0.0', terrareg.version_constraint.VersionCompatibilityType.COMPATIBLE),
        ('>=2.0.0, <=3.0.0', '2.0.1', terrareg.version_constraint.VersionCompatibilityType.COMPATIBLE),
        ('>=2.0.0, <=3.0.1', '3.0.0', terrareg.version_constraint.VersionCompatibilityType.COMPATIBLE),

        ('>=2.0.0, <=3.0.1, != 2.5.4', '2.5.4', terrareg.version_constraint.VersionCompatibilityType.INCOMPATIBLE),
        ('>=2.0.0, <=3.0.1, != 2.5.4', '2.5.3', terrareg.version_constraint.VersionCompatibilityType.COMPATIBLE),

        # Test with spacing
        (' >= 2.0.0 , <= 3.0.0 ', '3.0.1', terrareg.version_constraint.VersionCompatibilityType.INCOMPATIBLE),
        (' >= 2.0.0 , <= 3.0.0 ', '2.0.0', terrareg.version_constraint.VersionCompatibilityType.COMPATIBLE),
    ])
    def test_is_compatible(self, constraint, target_version, expected_result):
        """Test constraint compatiblity"""
        assert terrareg.version_constraint.VersionConstraint.is_compatible(
            constraint=constraint,
            target_version=target_version
        ) == expected_result

