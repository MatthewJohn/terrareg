
import json
import os
from typing import Dict, Union
import unittest.mock
import tempfile

import pytest

from test.integration.terrareg import TerraregIntegrationTest
import terrareg.provider_model
import terrareg.repository_model
import terrareg.provider_category_model
import terrareg.provider_source.factory
import terrareg.provider_source
import terrareg.provider_source_type
import terrareg.database
import terrareg.provider_tier
import terrareg.models
import terrareg.errors
import terrareg.audit_action
import terrareg.config
import terrareg.provider_version_model
import terrareg.provider_version_binary_model
import terrareg.provider_binary_types
import terrareg.provider_source.repository_release_metadata
from test.integration.terrareg.fixtures import (
    mock_provider_source_class,
    mock_provider_source,
    test_gpg_key,
    test_namespace,
    test_provider,
    test_provider_category,
    test_repository
)


class TestProvider(TerraregIntegrationTest):
    """Test provider model"""

    _PROVIDER_SOURCES = []
    _TEST_DATA = {}

    @pytest.mark.parametrize("repository_name, expected_result", [
        ('terraform-provider-jmon', 'jmon'),
        ('terraform-provider-some-service', 'some-service'),
        ('terraform-some-service', None),
        ('some-service', None),
        ('', None),
        (None, None),
    ])
    def test_repository_name_to_provider_name(cls, repository_name, expected_result):
        """Convert repository name to provider name"""
        assert terrareg.provider_model.Provider.repository_name_to_provider_name(repository_name=repository_name) == expected_result

    @pytest.mark.parametrize('use_default_provider_source_auth', [
        True,
        False,
    ])
    def test_create(cls, use_default_provider_source_auth, test_namespace, mock_provider_source, test_repository, test_provider_category):
        """Test create method."""

        try:
            provider = terrareg.provider_model.Provider.create(
                repository=test_repository,
                provider_category=test_provider_category,
                use_default_provider_source_auth=use_default_provider_source_auth,
                tier=terrareg.provider_tier.ProviderTier.COMMUNITY,
            )
            assert isinstance(provider, terrareg.provider_model.Provider)
            assert provider.pk

            db = terrareg.database.Database.get()
            with db.get_connection() as conn:
                res = conn.execute(db.provider.select(db.provider.c.repository_id==test_repository.pk)).all()

            assert len(res) == 1
            row = dict(res[0])
            # Since we don't care what the ID is, set the result ID to a known value
            row['id'] = 1
            assert row == {
                'default_provider_source_auth': use_default_provider_source_auth,
                'description': 'Unit test repo for Terraform Provider',
                'id': 1,
                'latest_version_id': None,
                'name': 'unittest-create',
                'namespace_id': test_namespace.pk,
                'provider_category_id': test_provider_category.pk,
                'repository_id': test_repository.pk,
                'tier': terrareg.provider_tier.ProviderTier.COMMUNITY,
            }
            
            # Ensure audit event was created correct
            with db.get_connection() as conn:
                res = conn.execute(db.audit_history.select().order_by(db.audit_history.c.timestamp.desc()).limit(1)).all()

            assert len(res) == 1
            audit_event = dict(res[0])
            assert audit_event['action'] == terrareg.audit_action.AuditAction.PROVIDER_CREATE
            assert audit_event['object_type'] == "Provider"
            assert audit_event['object_id'] == "some-organisation/unittest-create"
            assert audit_event['old_value'] == None
            assert audit_event['new_value'] == None

        finally:
            db = terrareg.database.Database.get()
            with db.get_connection() as conn:
                conn.execute(db.provider.delete(db.provider.c.repository_id==test_repository.pk))

    def test_create_duplicate(cls, test_namespace, mock_provider_source, test_repository, test_provider_category):
        """Test attempting to create with duplicate provider"""
        try:
            provider = terrareg.provider_model.Provider.create(
                repository=test_repository,
                provider_category=test_provider_category,
                use_default_provider_source_auth=True,
                tier=terrareg.provider_tier.ProviderTier.COMMUNITY,
            )
            assert isinstance(provider, terrareg.provider_model.Provider)

            with pytest.raises(terrareg.errors.DuplicateProviderError):
                terrareg.provider_model.Provider.create(
                    repository=test_repository,
                    provider_category=test_provider_category,
                    use_default_provider_source_auth=True,
                    tier=terrareg.provider_tier.ProviderTier.COMMUNITY,
                )

        finally:
            db = terrareg.database.Database.get()
            with db.get_connection() as conn:
                conn.execute(db.provider.delete(db.provider.c.repository_id==test_repository.pk))

    def test_create_without_namespace(cls, test_namespace, mock_provider_source, test_repository, test_provider_category):
        """Test attempting to create without a namespace present for repository"""
        try:
            test_namespace.delete()

            with pytest.raises(terrareg.errors.NonExistentNamespaceError):
                terrareg.provider_model.Provider.create(
                    repository=test_repository,
                    provider_category=test_provider_category,
                    use_default_provider_source_auth=True,
                    tier=terrareg.provider_tier.ProviderTier.COMMUNITY,
                )

        finally:
            db = terrareg.database.Database.get()
            with db.get_connection() as conn:
                conn.execute(db.provider.delete(db.provider.c.repository_id==test_repository.pk))

    def test_create_without_github_installation(cls, test_namespace, mock_provider_source_class, mock_provider_source, test_repository, test_provider_category):
        """Test provider creation without valid github installation and without using default authentication"""

        try:
            mock_provider_source_class.HAS_INSTALLATION_ID = False
            with pytest.raises(terrareg.errors.NoGithubAppInstallationError):
                terrareg.provider_model.Provider.create(
                    repository=test_repository,
                    provider_category=test_provider_category,
                    use_default_provider_source_auth=False,
                    tier=terrareg.provider_tier.ProviderTier.COMMUNITY,
                )

        finally:
            mock_provider_source_class.HAS_INSTALLATION_ID = True
            db = terrareg.database.Database.get()
            with db.get_connection() as conn:
                conn.execute(db.provider.delete(db.provider.c.repository_id==test_repository.pk))

    def test_create_without_github_installation(cls, test_namespace, mock_provider_source_class, mock_provider_source, test_repository, test_provider_category):
        """Test provider creation without valid github installation and without using default authentication"""

        try:
            mock_provider_source_class.HAS_INSTALLATION_ID = False
            with pytest.raises(terrareg.errors.NoGithubAppInstallationError):
                terrareg.provider_model.Provider.create(
                    repository=test_repository,
                    provider_category=test_provider_category,
                    use_default_provider_source_auth=False,
                    tier=terrareg.provider_tier.ProviderTier.COMMUNITY,
                )

        finally:
            mock_provider_source_class.HAS_INSTALLATION_ID = True
            db = terrareg.database.Database.get()
            with db.get_connection() as conn:
                conn.execute(db.provider.delete(db.provider.c.repository_id==test_repository.pk))

    @pytest.mark.parametrize('repository_name', [
        'invalidname',
        'terraform-invalidname',
        'terraform-module-invalidname',
    ])
    def test_create_with_invalid_repository(cls, repository_name, test_namespace, mock_provider_source, test_provider_category):
        """Test provider creation with an invalid repository name"""

        repository = terrareg.repository_model.Repository.create(
            provider_source=mock_provider_source,
            provider_id="unittest-pid-123456",
            name=repository_name,
            description="Unit test repo for Terraform Provider",
            owner="some-organisation",
            clone_url="https://github.localhost/some-organisation/terraform-provider-unittest-create.git",
            logo_url="https://github.localhost/logos/some-organisation.png"
        )
        repository_pk = repository.pk

        try:
            with pytest.raises(terrareg.errors.InvalidRepositoryNameError):
                terrareg.provider_model.Provider.create(
                    repository=repository,
                    provider_category=test_provider_category,
                    use_default_provider_source_auth=False,
                    tier=terrareg.provider_tier.ProviderTier.COMMUNITY,
                )

        finally:
            db = terrareg.database.Database.get()
            with db.get_connection() as conn:
                conn.execute(db.provider.delete(db.provider.c.repository_id==repository.pk))
                conn.execute(db.repository.delete(db.repository.c.id==repository_pk))

    def test_get_by_pk(cls, test_provider):
        """Test get_by_pk method of provider"""
        test_pk = test_provider.pk

        # Test with valid PK
        result = terrareg.provider_model.Provider.get_by_pk(pk=test_pk)
        assert result is not None
        assert isinstance(result, terrareg.provider_model.Provider)

        assert result.pk == test_pk
        assert result._get_db_row() == test_provider._get_db_row()

    def test_get_by_pk_non_existent(cls, test_provider):
        """Test get_by_pk method of provider with non-existent PK"""
        invalid_result = terrareg.provider_model.Provider.get_by_pk(pk=9999999)
        assert invalid_result is None

    def test_get_by_repository(cls, test_provider, test_repository):
        """Test get_by_repository with repository that is associated with a provider"""
        result = terrareg.provider_model.Provider.get_by_repository(repository=test_repository)
        assert result is not None
        assert isinstance(result, terrareg.provider_model.Provider)

        assert result.pk == test_provider.pk
        assert result._get_db_row() == test_provider._get_db_row()

    def test_get_by_repository_non_existent(cls, test_repository):
        """Test get_by_repository with non-existent provider"""
        result = terrareg.provider_model.Provider.get_by_repository(repository=test_repository)
        assert result is None

    def test_get(self, test_provider, test_namespace):
        """Test get method with valid namespace and name combination."""
        res = terrareg.provider_model.Provider.get(namespace=test_namespace, name='unittest-create-provider-name')

        assert res is not None
        assert isinstance(res, terrareg.provider_model.Provider)
        assert res.pk == test_provider.pk

    def test_get_non_existent(self, test_namespace):
        """Test get method with invalid namespace and name combination."""
        res = terrareg.provider_model.Provider.get(namespace=test_namespace, name='does-not-exist')
        assert res is None

    def test_id(self, test_provider):
        """Test ID property of provider"""
        assert test_provider.id == "some-organisation/unittest-create-provider-name"

    def test_namespace(self, test_provider, test_namespace):
        """Test namespace property of provider"""
        namespace = test_provider.namespace
        assert isinstance(namespace, terrareg.models.Namespace)
        assert namespace.pk == test_namespace.pk
        assert namespace._get_db_row() == test_namespace._get_db_row()

    def test_name(self, test_provider):
        """Test name property of provider"""
        assert test_provider.name == "unittest-create-provider-name"

    def test_full_name(self, test_provider):
        """Test full_name property of Provider"""
        assert test_provider.full_name == "terraform-provider-unittest-create-provider-name"

    def test_pk(self, test_provider):
        """Test pk property of Provider"""
        assert isinstance(test_provider.pk, int)
        test_provider._cache_db_row = {"id": "newid"}
        assert test_provider.pk == "newid"

    @pytest.mark.parametrize('provider_tier', [
        terrareg.provider_tier.ProviderTier.COMMUNITY,
        terrareg.provider_tier.ProviderTier.OFFICIAL
    ])
    def test_tier(self, provider_tier, test_repository, test_provider_category):
        """Test tier property of Provider"""
        provider = terrareg.provider_model.Provider.create(
            repository=test_repository,
            provider_category=test_provider_category,
            use_default_provider_source_auth=False,
            tier=provider_tier
        )
        try:
            assert isinstance(provider, terrareg.provider_model.Provider)
            assert provider.tier is provider_tier

            # Ensure tier is correctly loaded for new instance
            new_provider = terrareg.provider_model.Provider.get_by_pk(provider.pk)
            assert new_provider.tier is provider_tier
        finally:
            db = terrareg.database.Database.get()
            with db.get_connection() as conn:
                conn.execute(db.provider.delete(db.provider.c.id==provider.pk))

    def test_base_directory(self, test_provider):
        """Test base_directory property of Provider."""
        assert test_provider.base_directory == f"{os.getcwd()}/data/providers/some-organisation/unittest-create-provider-name"

    def test_repository(self, test_provider, test_repository):
        """Test repository property of provider"""
        res = test_provider.repository
        assert isinstance(res, terrareg.repository_model.Repository)
        assert res.pk == test_repository.pk

    def test_category(self, test_provider, test_provider_category):
        """Test category property of Provider"""
        res = test_provider.category
        assert isinstance(res, terrareg.provider_category_model.ProviderCategory)
        assert res.pk == test_provider_category.pk

    def test_source_url(self, test_provider):
        """Test source_url property"""
        assert test_provider.source_url == "https://git.example.com/get_public_source_url/some-organisation/terraform-provider-unittest-create"

    def test_description(self, test_provider):
        """Test description property"""
        assert test_provider.description == "Unit test repo for Terraform Provider"

    def test_alias(self, test_provider):
        """Test alias property"""
        assert test_provider.alias is None

    def test_featured(self, test_provider):
        """Test featured property"""
        assert test_provider.featured is False

    def test_logo_url(self, test_provider):
        """Test logo_url property"""
        assert test_provider.logo_url == "https://github.localhost/logos/some-organisation.png"

    def test_owner_name(self, test_provider):
        """Test owner_name property"""
        assert test_provider.owner_name == "some-organisation"
    
    def test_repository_id(self, test_provider, test_repository) -> int:
        """Test repository_id property"""
        assert test_provider.repository_id == test_repository.pk
    
    def test_robots_noindex(self, test_provider):
        """Test robots_noindex property"""
        assert test_provider.robots_noindex is False

    def unlisted(self, test_provider):
        """Test unlisted property"""
        assert test_provider.unlisted is False

    def test_warning(self, test_provider):
        """Test warning property"""
        assert test_provider.warning == ""

    @pytest.mark.parametrize('default_provider_source_auth', [
        True,
        False
    ])
    def test_use_default_provider_source_auth(self, default_provider_source_auth, test_provider):
        """Test use_default_provider_source_auth property"""
        test_provider._cache_db_row = {"default_provider_source_auth": default_provider_source_auth}
        assert test_provider.use_default_provider_source_auth is default_provider_source_auth

    def test_create_data_directory(self, test_provider):
        """"Test create_data_directory method"""
        with tempfile.TemporaryDirectory() as tempdir, \
                unittest.mock.patch("terrareg.config.Config.DATA_DIRECTORY", tempdir):
            os.mkdir(os.path.join(tempdir, "providers"))

            test_provider.create_data_directory()

            assert os.path.isdir(os.path.join(tempdir, "providers", "some-organisation", "unittest-create-provider-name"))

    @pytest.mark.parametrize('provider_versions, expected_latest_version', [
        ([], None),
        (['1.0.0'], '1.0.0'),
        (['1.0.0', '3.0.0', '1.5.2', '2.1.0'], '3.0.0'),
    ])
    def test_get_latest_version(self, test_provider, test_gpg_key, provider_versions, expected_latest_version):
        """Test get_latest_version method"""
        created_version_mapping = {}
        try:
            for version_ in provider_versions:
                provider_version = terrareg.provider_version_model.ProviderVersion(provider=test_provider, version=version_)
                provider_version._create_db_row(git_tag=f"v{version_}", gpg_key=test_gpg_key)
                provider_version.publish()
                created_version_mapping[version_] = provider_version.pk

            returned_version = test_provider.get_latest_version()
            # If no versions were created, expect None
            if expected_latest_version is None:
                assert returned_version is None
            else:
                # Otherwise, ensure the version and PK match
                assert isinstance(returned_version, terrareg.provider_version_model.ProviderVersion)
                assert returned_version.version == expected_latest_version
                assert returned_version.pk == created_version_mapping[expected_latest_version]

        finally:
            db = terrareg.database.Database.get()
            with db.get_connection() as conn:
                for provider_version_id in created_version_mapping.values():
                    conn.execute(db.provider_version.delete(db.provider_version.c.id==provider_version_id))

    @pytest.mark.parametrize('provider_versions, expected_return_order', [
        ([], []),
        (['1.0.0'], ['1.0.0']),
        (['1.0.0', '3.0.0', '1.5.2', '2.1.0', '2.0.5', '2.0.0'], ['3.0.0', '2.1.0', '2.0.5', '2.0.0', '1.5.2', '1.0.0']),
    ])
    def test_get_all_versions(self, test_provider, test_gpg_key, provider_versions, expected_return_order):
        """Test get_all_versions method"""
        created_version_mapping = {}
        try:
            for version_ in provider_versions:
                provider_version = terrareg.provider_version_model.ProviderVersion(provider=test_provider, version=version_)
                provider_version._create_db_row(git_tag=f"v{version_}", gpg_key=test_gpg_key)
                provider_version.publish()
                created_version_mapping[version_] = provider_version.pk

            versions_response = test_provider.get_all_versions()
            assert [
                version_.version
                for version_ in versions_response
            ] == expected_return_order

        finally:
            db = terrareg.database.Database.get()
            with db.get_connection() as conn:
                for provider_version_id in created_version_mapping.values():
                    conn.execute(db.provider_version.delete(db.provider_version.c.id==provider_version_id))

    @pytest.mark.parametrize('provider_versions, expected_latest_version', [
        ([], None),
        (['1.0.0'], '1.0.0'),
        (['1.0.0', '3.0.0', '1.5.2', '2.1.0'], '3.0.0'),
    ])
    def test_calculate_latest_version(self, test_provider, test_gpg_key, provider_versions, expected_latest_version):
        """Test calculate_latest_version method."""
        created_version_mapping = {}
        try:
            for version_ in provider_versions:
                provider_version = terrareg.provider_version_model.ProviderVersion(provider=test_provider, version=version_)
                provider_version._create_db_row(git_tag=f"v{version_}", gpg_key=test_gpg_key)
                created_version_mapping[version_] = provider_version.pk

            test_provider._cache_db_row = None
            assert test_provider._get_db_row()['latest_version_id'] is None

            returned_version = test_provider.calculate_latest_version()

            # If no versions were created, expect None
            if expected_latest_version is None:
                assert returned_version is None
            else:
                # Otherwise, ensure the version and PK match
                assert isinstance(returned_version, terrareg.provider_version_model.ProviderVersion)
                assert returned_version.version == expected_latest_version
                assert returned_version.pk == created_version_mapping[expected_latest_version]

        finally:
            db = terrareg.database.Database.get()
            with db.get_connection() as conn:
                for provider_version_id in created_version_mapping.values():
                    conn.execute(db.provider_version.delete(db.provider_version.c.id==provider_version_id))

    def test_refresh_versions(self, mock_provider_source_class, test_provider, test_gpg_key, test_namespace):
        """Test refresh_versions method"""

        try:
            with unittest.mock.patch(
                        'terrareg.provider_extractor.ProviderExtractor.obtain_gpg_key',
                        unittest.mock.MagicMock(return_value=test_gpg_key)) as mock_obtain_gpg_key, \
                    unittest.mock.patch('terrareg.provider_extractor.ProviderExtractor.process_version', unittest.mock.MagicMock()) as mock_process_version:

                created_versions = test_provider.refresh_versions()

                mock_obtain_gpg_key.assert_has_calls(calls=[
                        unittest.mock.call(provider=test_provider, release_metadata=mock_provider_source_class.NEW_RELEASES[0], namespace=test_namespace),
                        unittest.mock.call(provider=test_provider, release_metadata=mock_provider_source_class.NEW_RELEASES[1], namespace=test_namespace),
                    ]
                )
                mock_process_version.assert_has_calls(calls=[
                        unittest.mock.call(),
                        unittest.mock.call(),
                    ]
                )

                assert len(created_versions) == 2
                for created_version in created_versions:
                    assert isinstance(created_version, terrareg.provider_version_model.ProviderVersion)
        finally:
            db = terrareg.database.Database.get()
            with db.get_connection() as conn:
                conn.execute(db.provider_version.delete(db.provider_version.c.provider_id==test_provider.pk))

    def test_refresh_versions_no_gpg_key(self, mock_provider_source_class, test_provider, test_namespace):
        """Test refresh_versions method with no GPG key found for release"""

        try:
            with unittest.mock.patch(
                        'terrareg.provider_extractor.ProviderExtractor.obtain_gpg_key',
                        unittest.mock.MagicMock(return_value=None)) as mock_obtain_gpg_key, \
                    unittest.mock.patch('terrareg.provider_extractor.ProviderExtractor.process_version', unittest.mock.MagicMock()) as mock_process_version:

                with pytest.raises(terrareg.errors.CouldNotFindGpgKeyForProviderVersionError):
                    created_versions = test_provider.refresh_versions()

                mock_obtain_gpg_key.assert_has_calls(calls=[
                        unittest.mock.call(provider=test_provider, release_metadata=mock_provider_source_class.NEW_RELEASES[0], namespace=test_namespace),
                    ]
                )
                mock_process_version.assert_not_called()

        finally:
            db = terrareg.database.Database.get()
            with db.get_connection() as conn:
                conn.execute(db.provider_version.delete(db.provider_version.c.provider_id==test_provider.pk))

    def test_refresh_versions_get_gpg_key_exception(self, mock_provider_source_class, test_provider, test_namespace):
        """Test refresh_versions method with exception raised when attempting to obtain GPG key"""

        def raise_obtain_gpg_key_error(*args, **kwargs):
            """Raise exception when attepmting to obtain GPG key"""
            raise terrareg.errors.MissingSignureArtifactError("Unit test no GPG Key found")

        try:
            with unittest.mock.patch(
                        'terrareg.provider_extractor.ProviderExtractor.obtain_gpg_key',
                        unittest.mock.MagicMock(side_effect=raise_obtain_gpg_key_error)) as mock_obtain_gpg_key, \
                    unittest.mock.patch('terrareg.provider_extractor.ProviderExtractor.process_version', unittest.mock.MagicMock()) as mock_process_version:

                created_versions = test_provider.refresh_versions()
                assert len(created_versions) == 0

                mock_obtain_gpg_key.assert_has_calls(calls=[
                        unittest.mock.call(provider=test_provider, release_metadata=mock_provider_source_class.NEW_RELEASES[0], namespace=test_namespace),
                        unittest.mock.call(provider=test_provider, release_metadata=mock_provider_source_class.NEW_RELEASES[1], namespace=test_namespace),
                    ]
                )
                mock_process_version.assert_not_called()

                # Ensure no provider versions were created in database
                db = terrareg.database.Database.get()
                with db.get_connection() as conn:
                    rows = conn.execute(db.provider_version.select(db.provider_version.c.provider_id==test_provider.pk)).all()
                    assert len(rows) == 0

        finally:
            db = terrareg.database.Database.get()
            with db.get_connection() as conn:
                conn.execute(db.provider_version.delete(db.provider_version.c.provider_id==test_provider.pk))

    def test_refresh_versions_get_gpg_key_generic_exception(self, mock_provider_source_class, test_provider, test_namespace):
        """Test refresh_versions method with generic exception raised when attempting to obtain GPG key"""

        class UnittestGpgException(Exception):
            pass


        def raise_obtain_gpg_key_error(*args, **kwargs):
            """Raise exception when attepmting to obtain GPG key"""
            raise UnittestGpgException("Unit test generic exception")

        try:
            with unittest.mock.patch(
                        'terrareg.provider_extractor.ProviderExtractor.obtain_gpg_key',
                        unittest.mock.MagicMock(side_effect=raise_obtain_gpg_key_error)) as mock_obtain_gpg_key, \
                    unittest.mock.patch('terrareg.provider_extractor.ProviderExtractor.process_version', unittest.mock.MagicMock()) as mock_process_version:

                with pytest.raises(UnittestGpgException):
                    test_provider.refresh_versions()
    
                mock_obtain_gpg_key.assert_has_calls(calls=[
                        unittest.mock.call(provider=test_provider, release_metadata=mock_provider_source_class.NEW_RELEASES[0], namespace=test_namespace)
                    ]
                )
                mock_process_version.assert_not_called()

                # Ensure no provider versions were created in database
                db = terrareg.database.Database.get()
                with db.get_connection() as conn:
                    rows = conn.execute(db.provider_version.select(db.provider_version.c.provider_id==test_provider.pk)).all()
                    assert len(rows) == 0

        finally:
            db = terrareg.database.Database.get()
            with db.get_connection() as conn:
                conn.execute(db.provider_version.delete(db.provider_version.c.provider_id==test_provider.pk))

    def test_refresh_versions_extraction_terrareg_exception(self, mock_provider_source_class, test_provider, test_gpg_key, test_namespace):
        """Test refresh_versions method with Terrareg exception raised when extracting version"""

        class UnittestExtractionException(terrareg.errors.TerraregError):
            pass


        def raise_process_version_exception(*args, **kwargs):
            """Raise exception when attempting to extract provider"""
            raise UnittestExtractionException("Unit test generic exception")

        try:
            with unittest.mock.patch('terrareg.provider_extractor.ProviderExtractor.obtain_gpg_key',
                                     unittest.mock.MagicMock(return_value=test_gpg_key)) as mock_obtain_gpg_key, \
                    unittest.mock.patch('terrareg.provider_extractor.ProviderExtractor.process_version',
                                         unittest.mock.MagicMock(side_effect=raise_process_version_exception)) as mock_process_version:

                created_versions = test_provider.refresh_versions()

                assert len(created_versions) == 0
    
                mock_obtain_gpg_key.assert_has_calls(calls=[
                        unittest.mock.call(provider=test_provider, release_metadata=mock_provider_source_class.NEW_RELEASES[0], namespace=test_namespace),
                        unittest.mock.call(provider=test_provider, release_metadata=mock_provider_source_class.NEW_RELEASES[1], namespace=test_namespace)
                    ]
                )
                mock_process_version.assert_has_calls(calls=[
                    unittest.mock.call(),
                    unittest.mock.call(),
                ])

                # Ensure no provider versions were created in database
                db = terrareg.database.Database.get()
                with db.get_connection() as conn:
                    rows = conn.execute(db.provider_version.select(db.provider_version.c.provider_id==test_provider.pk)).all()
                    assert len(rows) == 0

        finally:
            db = terrareg.database.Database.get()
            with db.get_connection() as conn:
                conn.execute(db.provider_version.delete(db.provider_version.c.provider_id==test_provider.pk))

    def test_refresh_versions_extraction_generic_exception(self, mock_provider_source_class, test_provider, test_gpg_key, test_namespace):
        """Test refresh_versions method with generic exception raised when extracting version"""

        class UnittestExtractionException(Exception):
            pass


        def raise_process_version_exception(*args, **kwargs):
            """Raise exception when attempting to extract provider"""
            raise UnittestExtractionException("Unit test generic exception")

        try:
            with unittest.mock.patch('terrareg.provider_extractor.ProviderExtractor.obtain_gpg_key',
                                     unittest.mock.MagicMock(return_value=test_gpg_key)) as mock_obtain_gpg_key, \
                    unittest.mock.patch('terrareg.provider_extractor.ProviderExtractor.process_version',
                                         unittest.mock.MagicMock(side_effect=raise_process_version_exception)) as mock_process_version:

                with pytest.raises(UnittestExtractionException):
                    test_provider.refresh_versions()
    
                mock_obtain_gpg_key.assert_has_calls(calls=[
                        unittest.mock.call(provider=test_provider, release_metadata=mock_provider_source_class.NEW_RELEASES[0], namespace=test_namespace)
                    ]
                )
                mock_process_version.assert_called_once_with()

                # Ensure no provider versions were created in database
                db = terrareg.database.Database.get()
                with db.get_connection() as conn:
                    rows = conn.execute(db.provider_version.select(db.provider_version.c.provider_id==test_provider.pk)).all()
                    assert len(rows) == 0

        finally:
            db = terrareg.database.Database.get()
            with db.get_connection() as conn:
                conn.execute(db.provider_version.delete(db.provider_version.c.provider_id==test_provider.pk))

    def test_update_attributes(self, test_provider):
        """Test update_attributes method"""
        test_provider._cache_db_row = None
        db_row = test_provider._get_db_row()
        assert db_row["description"] == "Unittest provider description"
        assert db_row["tier"] is terrareg.provider_tier.ProviderTier.COMMUNITY
        assert db_row["default_provider_source_auth"] is True

        test_provider.update_attributes(
            description="New Description",
            tier=terrareg.provider_tier.ProviderTier.OFFICIAL,
            default_provider_source_auth=False
        )

        # Ensure cached DB row is flushed and get_db_row immediately returns new data
        assert test_provider._cache_db_row is None
        new_db_row = test_provider._get_db_row()
        assert new_db_row["description"] == "New Description"
        assert new_db_row["tier"] is terrareg.provider_tier.ProviderTier.OFFICIAL
        assert new_db_row["default_provider_source_auth"] is False

    def test_get_versions_api_details(self, test_provider, test_gpg_key):
        """Test get_versions_api_details method"""
        created_version_mapping = {}
        try:
            for version_ in ["1.5.0", "1.0.0"]:
                provider_version = terrareg.provider_version_model.ProviderVersion(provider=test_provider, version=version_)
                provider_version._create_db_row(git_tag=f"v{version_}", gpg_key=test_gpg_key)
                created_version_mapping[version_] = provider_version.pk

                for os_, platform_ in [
                        (terrareg.provider_binary_types.ProviderBinaryOperatingSystemType.LINUX, terrareg.provider_binary_types.ProviderBinaryArchitectureType.AMD64),
                        (terrareg.provider_binary_types.ProviderBinaryOperatingSystemType.LINUX, terrareg.provider_binary_types.ProviderBinaryArchitectureType.ARM64),
                        (terrareg.provider_binary_types.ProviderBinaryOperatingSystemType.WINDOWS, terrareg.provider_binary_types.ProviderBinaryArchitectureType.AMD64)]:
                    terrareg.provider_version_binary_model.ProviderVersionBinary.create(
                        provider_version=provider_version,
                        name=f"{test_provider.full_name}_{provider_version.version}_{os_.value}_{platform_.value}.zip",
                        checksum=f"abcefg{provider_version.version}{os_.value}{platform_.value}",
                        content=b"sometestcontent"
                    )

            assert test_provider.get_versions_api_details() == {
                'id': 'some-organisation/unittest-create-provider-name',
                'versions': [
                    {
                        'platforms': [
                            {'arch': 'amd64', 'os': 'linux'},
                            {'arch': 'arm64', 'os': 'linux'},
                            {'arch': 'amd64', 'os': 'windows'}
                        ],
                        'protocols': ['5.0'],
                        'version': '1.5.0'
                    },
                    {
                        'platforms': [
                            {'arch': 'amd64', 'os': 'linux'},
                            {'arch': 'arm64', 'os': 'linux'},
                            {'arch': 'amd64', 'os': 'windows'}
                        ],
                        'protocols': ['5.0'],
                        'version': '1.0.0'
                    }
                ],
                'warnings': None,
            }

        finally:
            db = terrareg.database.Database.get()
            with db.get_connection() as conn:
                for provider_version_id in created_version_mapping.values():
                    conn.execute(db.provider_version_binary.delete(db.provider_version_binary.c.provider_version_id==provider_version_id))
                    conn.execute(db.provider_version.delete(db.provider_version.c.id==provider_version_id))
