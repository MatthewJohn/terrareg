
import os
import tempfile
import contextlib
from typing import ContextManager
import unittest.mock
import pytest

import terrareg.provider_extractor
import terrareg.provider_documentation_type
import terrareg.provider_version_documentation_model
import terrareg.provider_version_model
import terrareg.provider_model
import terrareg.database
import terrareg.provider_source.repository_release_metadata

from test.integration.terrareg import TerraregIntegrationTest


class TestProviderExtractorDocumentation(TerraregIntegrationTest):
    """Test provider extractor documentation extraction with real provider structure"""

    def _create_test_provider_version(self, version='2.0.0'):
        """Helper to create a test provider version"""
        # Get initial namespace and provider
        namespace = terrareg.models.Namespace.get("initial-providers")
        provider = terrareg.provider_model.Provider.get(namespace=namespace, name="multiple-versions")
        provider_version = terrareg.provider_version_model.ProviderVersion(
            provider=provider,
            version=version
        )
        return provider_version

    # Test: Existing docs only (no Go code, no tfplugindocs)
    def test_extract_documentation_with_existing_docs_only(self):
        """Test extraction with pre-generated docs - no tfplugindocs call"""
        provider_version = self._create_test_provider_version('2.0.0')

        # Create mock release metadata
        release_metadata = terrareg.provider_source.repository_release_metadata.RepositoryReleaseMetadata(
            name="Release v1.0.0",
            tag="v1.0.0",
            archive_url="https://example.com/test-provider-existing-docs/1.0.0.tar.gz",
            commit_hash="abc123",
            provider_id="test-release-id",
            release_artifacts=[]
        )

        # Create provider extractor instance
        extractor = terrareg.provider_extractor.ProviderExtractor(
            provider_version=provider_version,
            release_metadata=release_metadata
        )

        # Mock _obtain_source_code to return test provider directory
        with tempfile.TemporaryDirectory() as temp_dir:
            provider_dir = os.path.join(temp_dir, 'test-provider-existing-docs')
            os.makedirs(provider_dir)
            os.makedirs(os.path.join(provider_dir, 'docs'))

            # Copy docs files from fixtures
            fixtures_docs_dir = os.path.join(
                os.path.dirname(__file__),
                'fixtures',
                'providers',
                'test-provider-existing-docs',
                'docs'
            )

            # Copy index.md
            with open(os.path.join(fixtures_docs_dir, 'index.md'), 'r') as f:
                content = f.read()
            with open(os.path.join(provider_dir, 'docs', 'index.md'), 'w') as f:
                f.write(content)

            # Copy resources/test_example.md
            os.makedirs(os.path.join(provider_dir, 'docs', 'resources'))
            with open(os.path.join(fixtures_docs_dir, 'resources', 'test_example.md'), 'r') as f:
                content = f.read()
            with open(os.path.join(provider_dir, 'docs', 'resources', 'test_example.md'), 'w') as f:
                f.write(content)

            # Copy data-sources/test_example.md
            os.makedirs(os.path.join(provider_dir, 'docs', 'data-sources'))
            with open(os.path.join(fixtures_docs_dir, 'data-sources', 'test_example.md'), 'r') as f:
                content = f.read()
            with open(os.path.join(provider_dir, 'docs', 'data-sources', 'test_example.md'), 'w') as f:
                f.write(content)

            # Copy guides/example_guide.md
            os.makedirs(os.path.join(provider_dir, 'docs', 'guides'))
            with open(os.path.join(fixtures_docs_dir, 'guides', 'example_guide.md'), 'r') as f:
                content = f.read()
            with open(os.path.join(provider_dir, 'docs', 'guides', 'example_guide.md'), 'w') as f:
                f.write(content)

            # Mock _obtain_source_code context manager
            with unittest.mock.patch('terrareg.provider_extractor.ProviderExtractor._obtain_source_code') as mock_obtain_source:
                mock_cm = unittest.mock.MagicMock()
                mock_cm.__enter__ = unittest.mock.MagicMock(return_value=provider_dir)
                mock_cm.__exit__ = unittest.mock.MagicMock(return_value=False)
                mock_obtain_source.return_value = mock_cm

                # Run extraction - subprocess.call will be used normally if docs don't exist
                extractor.extract_documentation()

            # Verify documentation was collected
            docs = terrareg.provider_version_documentation_model.ProviderVersionDocumentation.get_by_provider_version(
                provider_version=provider_version
            )

            assert len(docs) >= 4, f"Expected at least 4 docs, got {len(docs)}"

            # Verify overview doc
            overview = terrareg.provider_version_documentation_model.ProviderVersionDocumentation.search(
                provider_version=provider_version,
                category=terrareg.provider_documentation_type.ProviderDocumentationType.OVERVIEW,
                slug='index',
                language='hcl'
            )
            assert overview is not None
            if isinstance(overview, list):
                overview = overview[0]
            assert overview.title == "Test Provider Overview"
            assert overview.description == "A test provider overview for documentation extraction testing"

            # Verify resource doc
            resource = terrareg.provider_version_documentation_model.ProviderVersionDocumentation.search(
                provider_version=provider_version,
                category=terrareg.provider_documentation_type.ProviderDocumentationType.RESOURCE,
                slug='test_example',
                language='hcl'
            )
            assert resource is not None
            if isinstance(resource, list):
                resource = resource[0]
            assert resource.title == "Test Provider: test_example"
            assert resource.description == "Creates a test example resource for documentation extraction"

            # Verify data source doc
            datasource = terrareg.provider_version_documentation_model.ProviderVersionDocumentation.search(
                provider_version=provider_version,
                category=terrareg.provider_documentation_type.ProviderDocumentationType.DATA_SOURCE,
                slug='test_example',
                language='hcl'
            )
            assert datasource is not None
            if isinstance(datasource, list):
                datasource = datasource[0]
            assert datasource.title == "Test Provider: test_example"
            assert datasource.description == "Creates a test example data source for documentation extraction"

            # Verify guide doc
            guide = terrareg.provider_version_documentation_model.ProviderVersionDocumentation.search(
                provider_version=provider_version,
                category=terrareg.provider_documentation_type.ProviderDocumentationType.GUIDE,
                slug='example_guide',
                language='hcl'
            )
            assert guide is not None
            if isinstance(guide, list):
                guide = guide[0]
            assert guide.title == "Test Provider Guide"
            assert guide.description == "A guide example for documentation extraction"

    def test_extract_markdown_metadata(self):
        """Test YAML frontmatter is correctly parsed (title, subcategory, description)"""
        # Test with frontmatter
        markdown_with_frontmatter = """---
subcategory: "testprovider/v1"
page_title: "Test Provider"
description: "Test description"
---
# Content here
"""
        title, subcategory, description, content = terrareg.provider_extractor.ProviderExtractor._extract_markdown_metadata(markdown_with_frontmatter)

        assert title == "Test Provider"
        assert subcategory == "testprovider/v1"
        assert description == "Test description"
        assert "---" not in content
        assert "# Content here" in content

        # Test without frontmatter
        markdown_without_frontmatter = "# Just markdown content\n\nNo frontmatter here."
        title, subcategory, description, content = terrareg.provider_extractor.ProviderExtractor._extract_markdown_metadata(markdown_without_frontmatter)

        assert title is None
        assert subcategory is None
        assert description is None
        assert content == markdown_without_frontmatter

    def test_collect_all_documentation_types(self):
        """Test all documentation types are collected and stored"""
        # This test is covered by test_extract_documentation_with_existing_docs_only
        # which verifies all documentation types are collected
        pass  # Implementation in test_extract_documentation_with_existing_docs_only covers this
