
from enum import Enum
import re

import semantic_version


class VersionCompatibilityType(Enum):
    """Version compatibility status types"""

    # There is no constraint on the module version,
    # so compatibility cannot be determined
    NO_CONSTRAINT = "no_constraint"
    # Compatible
    COMPATIBLE = "compatible"
    # The version is compatible but due to a missing
    # upper or lower bounds
    IMPLICIT_COMPATIBLE = "implicit_compatible"
    # Not compatible
    INCOMPATIBLE = "incompatible"
    # Error whilst processing constraint
    ERROR = "error"


class VersionConstraint:

    # Match:
    # - group 1 - major version
    # - group 2 - minor version (optional)
    # - group 3 - patch version (optional)
    RE_VERSION_MATCH_PART = re.compile(r'^\s*([0-9]+)(?:\.([0-9]+)(?:\.([0-9]+))?)?\s*$')

    @classmethod
    def version_string_to_sem_version(cls, version_string):
        """Convert version string to semantic version. Returns semantic version, whether minor matched and whether patch matched"""
        version_match = cls.RE_VERSION_MATCH_PART.match(version_string)
        if version_match is None:
            return None, False, False

        try:
            sem_version = semantic_version.Version(
                "{major}.{minor}.{patch}".format(
                    major=version_match.group(1),
                    minor=version_match.group(2) if version_match.group(2) is not None else '0',
                    patch=version_match.group(3) if version_match.group(3) is not None else '0'
                )
            )
        # Catch value error caused by invalid version string
        except ValueError:
            return None, False, False

        return (
            sem_version,
            version_match.group(2) is not None,
            version_match.group(3) is not None
        )

    @classmethod
    def is_compatible(cls, constraint, target_version):
        """Determine if a version constraint is compatible with a target version"""
        try:
            target_version_sem = semantic_version.Version(target_version)
        except ValueError:
            # Catch value error due to invalid target version string
            return VersionCompatibilityType.ERROR

        has_lower_bounds = False
        has_upper_bounds = False

        if not constraint:
            return VersionCompatibilityType.NO_CONSTRAINT

        for constraint_part in constraint.split(','):
            constraint_part = constraint_part.strip()

            if not constraint_part:
                continue

            if constraint_part.startswith('>='):
                constraint_sem, _, _ = cls.version_string_to_sem_version(constraint_part[2:])
                if not constraint_sem:
                    return VersionCompatibilityType.ERROR
                
                # If constraint is higher than the target version, return incompatible
                if constraint_sem > target_version_sem:
                    return VersionCompatibilityType.INCOMPATIBLE
                
                # Mark as having found a lower bound
                has_lower_bounds = True

            elif constraint_part.startswith('>'):
                constraint_sem, _, _ = cls.version_string_to_sem_version(constraint_part[1:])
                if not constraint_sem:
                    return VersionCompatibilityType.ERROR

                # If constraint is higher or equal than the target version, return incompatible
                if constraint_sem >= target_version_sem:
                    return VersionCompatibilityType.INCOMPATIBLE
                
                has_lower_bounds = True
                
            elif constraint_part.startswith('<='):
                constraint_sem, _, _ = cls.version_string_to_sem_version(constraint_part[2:])
                if not constraint_sem:
                    return VersionCompatibilityType.ERROR
                
                # If constraint is higher than the target version, return incompatible
                if constraint_sem < target_version_sem:
                    return VersionCompatibilityType.INCOMPATIBLE
                
                # Mark as having found a lower bound
                has_upper_bounds = True

            elif constraint_part.startswith('<'):
                constraint_sem, _, _ = cls.version_string_to_sem_version(constraint_part[1:])
                if not constraint_sem:
                    return VersionCompatibilityType.ERROR
                
                # If constraint is higher or equal than the target version, return incompatible
                if constraint_sem <= target_version_sem:
                    return VersionCompatibilityType.INCOMPATIBLE
                
                has_upper_bounds = True

            elif constraint_part.startswith('~>'):

                version_match_sem, minor_matched, patch_matched = cls.version_string_to_sem_version(constraint_part[2:])
                if not version_match_sem:
                    return VersionCompatibilityType.ERROR

                # Ignore the last matching group and compare the remaining version
                if patch_matched:
                    # Match major and minor
                    if version_match_sem.major != target_version_sem.major or version_match_sem.minor != target_version_sem.minor:
                        return VersionCompatibilityType.INCOMPATIBLE

                elif minor_matched:
                    # Match just major
                    if version_match_sem.major != target_version_sem.major:
                        return VersionCompatibilityType.INCOMPATIBLE

                has_upper_bounds = True
                has_lower_bounds = True

            elif constraint_part.startswith('='):
                version_match_sem, _, _ = cls.version_string_to_sem_version(constraint_part[1:])
                if not version_match_sem:
                    return VersionCompatibilityType.ERROR

                if version_match_sem != target_version_sem:
                    return VersionCompatibilityType.INCOMPATIBLE

                has_upper_bounds = True
                has_lower_bounds = True

            elif constraint_part.startswith('!='):
                version_match_sem, _, _ = cls.version_string_to_sem_version(constraint_part[2:])
                if not version_match_sem:
                    return VersionCompatibilityType.ERROR

                if version_match_sem == target_version_sem:
                    return VersionCompatibilityType.INCOMPATIBLE
            else:
                # Attempt to match just the string
                version_match_sem, _, _ = cls.version_string_to_sem_version(constraint_part)
                if not version_match_sem:
                    return VersionCompatibilityType.ERROR

                if version_match_sem != target_version_sem:
                    return VersionCompatibilityType.INCOMPATIBLE

                has_upper_bounds = True
                has_lower_bounds = True

        # If either upper or lower bound was not found,
        # return implicitly compatible, otherwise,
        # return compatible
        return (
            VersionCompatibilityType.COMPATIBLE
            if has_upper_bounds and has_lower_bounds else
            VersionCompatibilityType.IMPLICIT_COMPATIBLE
        )
