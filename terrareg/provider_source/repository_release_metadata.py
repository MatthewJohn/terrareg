
from typing import Union, List

import re


class RepositoryReleaseMetadata:
    """Structure for storing repository release metadata"""

    _TAG_REGEX = re.compile(r'^v([0-9]+\.[0-9]+\.[0-9]+(:?-[a-z0-9]+)?)$')

    def __init__(self, name: str, tag: str, archive_url: str, commit_hash: str,
                 provider_id: Union[str, int], release_artifacts: List['ReleaseArtifactMetadata']):
        """Store member variables"""
        self.name = name
        self.tag = tag
        self.provider_id = provider_id
        self.release_artifacts = release_artifacts
        self.archive_url = archive_url
        self.commit_hash = commit_hash

    def __eq__(self, __o):
        """Check if two repository releases are the same"""
        if isinstance(__o, self.__class__):
            return (
                self.name == __o.name and
                self.tag == __o.tag and
                self.provider_id == __o.provider_id and
                self.release_artifacts == __o.release_artifacts and
                self.archive_url == __o.archive_url and
                self.commit_hash == __o.commit_hash
            )
        return super(RepositoryReleaseMetadata, self).__eq__(__o)

    @classmethod
    def tag_to_version(cls, tag: str) -> Union[None, str]:
        """Convert tag to version"""
        # Since we currently support the original tagging strategy 'v{version}', simply strip off the v
        if (match := cls._TAG_REGEX.match(tag)):
            return match.group(1)
        return None

    @property
    def version(self) -> Union[None, str]:
        """Convert tag to version"""
        return self.tag_to_version(self.tag)


class ReleaseArtifactMetadata:
    """Structure for storing release artifact metadata"""

    def __init__(self, name: str, provider_id: Union[str, int]):
        """Store member variables"""
        self.name = name
        self.provider_id = provider_id

    def __eq__(self, __o):
        """Check if two release artifacts are the same"""
        if isinstance(__o, self.__class__):
            return self.name == __o.name and self.provider_id == __o.provider_id
        return super(ReleaseArtifactMetadata, self).__eq__(__o)
