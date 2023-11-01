
import json
import os
from typing import Dict, Union
import unittest.mock
import tempfile

import pytest

from test.test_gpg_key import public_ascii_armor
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


class MockProviderSource(terrareg.provider_source.BaseProviderSource):
    TYPE = "github"
    HAS_INSTALLATION_ID = True

    @classmethod
    def generate_db_config_from_source_config(cls, config: Dict[str, str]) -> Dict[str, Union[str, bool]]:
        """Mocked generate_db_config_from_source_config method"""
        return {}

    def get_github_app_installation_id(self, namespace):
        """"""
        return "12345-installation-id" if MockProviderSource.HAS_INSTALLATION_ID else None

    def get_public_source_url(self, repository):
        """Return mock public source URL"""
        return f"https://git.example.com/get_public_source_url/{repository.owner}/{repository.name}"


@pytest.fixture
def mock_provider_source():

    with unittest.mock.patch(
            'terrareg.provider_source.factory.ProviderSourceFactory._CLASS_MAPPING',
            {terrareg.provider_source_type.ProviderSourceType.GITHUB: MockProviderSource}):

        with unittest.mock.patch('terrareg.config.Config.PROVIDER_SOURCES', json.dumps(
            [{"name": "unittest-provider-source", "type": "github",
            "login_button_text": "Unit test login",
            "auto_generate_github_organisation_namespaces": False}]
        )):
            terrareg.provider_source.factory.ProviderSourceFactory.get().initialise_from_config()
        provider_source = terrareg.provider_source.factory.ProviderSourceFactory().get_provider_source_by_name("unittest-provider-source")
        provider_source_name = provider_source.name

        yield provider_source

    # Delete provider source
    db = terrareg.database.Database.get()
    with terrareg.database.Database.get_connection() as conn:
        conn.execute(db.provider_source.delete(db.provider_source.c.name==provider_source_name))


@pytest.fixture
def test_namespace():
    """Create test repository"""
    namespace = terrareg.models.Namespace.create("some-organisation", None, type_=None)
    yield namespace
    namespace.delete()


@pytest.fixture
def test_gpg_key(test_namespace):
    """Create test GPG key"""
    gpg_key = terrareg.models.GpgKey.create(namespace=test_namespace, ascii_armor=public_ascii_armor)
    yield gpg_key
    gpg_key.delete()


@pytest.fixture
def test_repository(test_namespace, mock_provider_source):
    """Create test repository"""
    repository = terrareg.repository_model.Repository.create(
        provider_source=mock_provider_source,
        provider_id="unittest-pid-123456",
        name="terraform-provider-unittest-create",
        description="Unit test repo for Terraform Provider",
        owner="some-organisation",
        clone_url="https://github.localhost/some-organisation/terraform-provider-unittest-create.git",
        logo_url="https://github.localhost/logos/some-organisation.png"
    )
    repository_pk = repository.pk
    yield repository
    db = terrareg.database.Database.get()
    with db.get_connection() as conn:
        conn.execute(db.repository.delete(db.repository.c.id==repository_pk))


@pytest.fixture
def test_provider_category():
    """Create test provider category"""
    with unittest.mock.patch('terrareg.config.Config.PROVIDER_SOURCES', json.dumps(
        [{"id": "1", "name": "Example Category", "slug": "example-category", "user-selectable": True}]
    )):
        terrareg.provider_category_model.ProviderCategoryFactory().initialise_from_config()

    provider_category = terrareg.provider_category_model.ProviderCategoryFactory().get_provider_category_by_slug("example-category")
    provider_category_pk = provider_category.pk
    yield provider_category
    db = terrareg.database.Database.get()
    with terrareg.database.Database.get_connection() as conn:
        conn.execute(db.provider_category.delete(db.provider_category.c.id==provider_category_pk))


@pytest.fixture
def test_provider(test_repository, test_namespace, test_provider_category):
    """Create test provider"""
    provider = terrareg.provider_model.Provider.create(
        repository=test_repository,
        provider_category=test_provider_category,
        use_default_provider_source_auth=True,
        tier=terrareg.provider_tier.ProviderTier.COMMUNITY,
    )
    provider_id = provider.pk

    # Set provider name/description to unique names
    # that do not match the repository to ensure
    # defails are being obtained from the provider row,
    # where applicable
    db = terrareg.database.Database.get()
    with db.get_connection() as conn:
        conn.execute(db.provider.update(db.provider.c.id==provider_id).values(
            name="unittest-create-provider-name",
            description=db.encode_blob("Unittest provider description")
        ))
    provider._name = "unittest-create-provider-name"
    provider._cache_db_row = None

    yield provider

    with db.get_connection() as conn:
        conn.execute(db.provider.delete(db.provider.c.id==provider_id))


class TestProvider(TerraregIntegrationTest):
    """Test provider model"""

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
                'description': b'Unit test repo for Terraform Provider',
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

    def test_create_without_github_installation(cls, test_namespace, mock_provider_source, test_repository, test_provider_category):
        """Test provider creation without valid github installation and without using default authentication"""

        try:
            MockProviderSource.HAS_INSTALLATION_ID = False
            with pytest.raises(terrareg.errors.NoGithubAppInstallationError):
                terrareg.provider_model.Provider.create(
                    repository=test_repository,
                    provider_category=test_provider_category,
                    use_default_provider_source_auth=False,
                    tier=terrareg.provider_tier.ProviderTier.COMMUNITY,
                )

        finally:
            MockProviderSource.HAS_INSTALLATION_ID = True
            db = terrareg.database.Database.get()
            with db.get_connection() as conn:
                conn.execute(db.provider.delete(db.provider.c.repository_id==test_repository.pk))

    def test_create_without_github_installation(cls, test_namespace, mock_provider_source, test_repository, test_provider_category):
        """Test provider creation without valid github installation and without using default authentication"""

        try:
            MockProviderSource.HAS_INSTALLATION_ID = False
            with pytest.raises(terrareg.errors.NoGithubAppInstallationError):
                terrareg.provider_model.Provider.create(
                    repository=test_repository,
                    provider_category=test_provider_category,
                    use_default_provider_source_auth=False,
                    tier=terrareg.provider_tier.ProviderTier.COMMUNITY,
                )

        finally:
            MockProviderSource.HAS_INSTALLATION_ID = True
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

    # def calculate_latest_version(self):
    #     """Obtain all versions of provider and sort by semantic version numbers to obtain latest version."""
    #     db = terrareg.database.Database.get()
    #     select = sqlalchemy.select(
    #         db.provider_version.c.version
    #     ).join(
    #         db.provider,
    #         db.provider_version.c.provider_id==db.provider.c.id
    #     ).where(
    #         db.provider.c.id==self.pk,
    #         db.provider_version.c.beta==False
    #     )
    #     with db.get_connection() as conn:
    #         res = conn.execute(select)

    #         # Convert to list
    #         rows = [r for r in res]

    #     # Sort rows by semantic versioning
    #     rows.sort(key=lambda x: LooseVersion(x['version']), reverse=True)

    #     # Ensure at least one row
    #     if not rows:
    #         return None

    #     # Obtain latest row
    #     return terrareg.provider_version_model.ProviderVersion(provider=self, version=rows[0]['version'])

    # def refresh_versions(self, limit: Union[int, None]=None) -> List['terrareg.provider_version_model.ProviderVersion']:
    #     """
    #     Refresh versions from provider source and create new provider versions

    #     Optional limit to determine the maximum number of releases to attempt to index
    #     """
    #     repository = self.repository

    #     releases_metadata = repository.get_new_releases(provider=self)

    #     provider_versions = []

    #     for release_metadata in releases_metadata:
    #         provider_version = terrareg.provider_version_model.ProviderVersion(provider=self, version=release_metadata.version)

    #         try:
    #             gpg_key = terrareg.provider_extractor.ProviderExtractor.obtain_gpg_key(
    #                 provider=self,
    #                 release_metadata=release_metadata,
    #                 namespace=self.namespace
    #             )
    #         except MissingSignureArtifactError:
    #             # Handle missing signature, and ignore release
    #             continue

    #         # However, if a signature was found, but the GPG key
    #         # could not be found, raise an exception, as the key is probably missing
    #         if not gpg_key:
    #             raise CouldNotFindGpgKeyForProviderVersionError(f"Could not find a valid GPG key to verify the signature of the release: {release_metadata.name}")

    #         current_transaction = terrareg.database.Database.get_current_transaction()
    #         nested_transaction = None
    #         if current_transaction:
    #             nested_transaction = current_transaction.begin_nested()
    #         try:
    #             with provider_version.create_extraction_wrapper(git_tag=release_metadata.tag, gpg_key=gpg_key):
    #                 provider_extractor = terrareg.provider_extractor.ProviderExtractor(
    #                     provider_version=provider_version,
    #                     release_metadata=release_metadata
    #                 )
    #                 provider_extractor.process_version()
    #         except TerraregError:
    #             # If an error occurs with the version, rollback nested transaction,
    #             # and try next version
    #             if nested_transaction:
    #                 nested_transaction.rollback()

    #         provider_versions.append(provider_version)
    #         if limit and len(provider_versions) >= limit:
    #             break

    #     return provider_versions

    # def update_attributes(self, **kwargs: dict) -> None:
    #     """Update DB row."""
    #     db = terrareg.database.Database.get()
    #     update = sqlalchemy.update(db.provider).where(
    #         db.provider.c.namespace_id==self.namespace.pk,
    #         db.provider.c.name==self.name
    #     ).values(**kwargs)
    #     with db.get_connection() as conn:
    #         conn.execute(update)

    #     # Remove cached DB row
    #     self._cache_db_row = None

    # def get_versions_api_details(self) -> dict:
    #     """Return API details for versions endpoint"""
    #     return {
    #         "id": self.id,
    #         "versions": [
    #             version.get_api_binaries_outline()
    #             for version in self.get_all_versions()
    #         ],
    #         "warnings": None
    #     }
