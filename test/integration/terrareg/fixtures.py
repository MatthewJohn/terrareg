
from typing import Dict, Union
import unittest.mock
import json

import pytest

import terrareg.provider_source.repository_release_metadata
import terrareg.provider_source_type
import terrareg.provider_source.factory
import terrareg.database
import terrareg.models
import terrareg.provider_tier
import terrareg.repository_model
from test.test_gpg_key import public_ascii_armor


@pytest.fixture
def mock_provider_source_class():

    class MockProviderSource(terrareg.provider_source.BaseProviderSource):
        TYPE = "github"
        HAS_INSTALLATION_ID = True
        NEW_RELEASES = [
            terrareg.provider_source.repository_release_metadata.RepositoryReleaseMetadata(
                name="v1.0.0",
                tag="v1.0.0",
                archive_url=f"https://git.example.com/some-organisation/terraform-provider-unittest-create/1.0.0-source.tgz",
                commit_hash="abcefg123100",
                provider_id="provider-123-id",
                release_artifacts=[]
            ),
            terrareg.provider_source.repository_release_metadata.RepositoryReleaseMetadata(
                name="v1.5.0",
                tag="v1.5.0",
                archive_url=f"https://git.example.com/some-organisation/terraform-provider-unittest-create/1.5.0-source.tgz",
                commit_hash="abcefg123150",
                provider_id="provider-456-id",
                release_artifacts=[]
            )
        ]

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

        def get_new_releases(self, provider):
            """Return mocked method to obtain new releases"""
            return MockProviderSource.NEW_RELEASES

    yield MockProviderSource
    del MockProviderSource


@pytest.fixture
def mock_provider_source(mock_provider_source_class):
    with unittest.mock.patch(
            'terrareg.provider_source.factory.ProviderSourceFactory._CLASS_MAPPING',
            {terrareg.provider_source_type.ProviderSourceType.GITHUB: mock_provider_source_class}):

        with unittest.mock.patch('terrareg.config.Config.PROVIDER_SOURCES', json.dumps(
            [{"name": "unittest-provider-source",
              "type": "github",
              "login_button_text": "Unit test login",
              "auto_generate_github_organisation_namespaces": False,
              "base_url": "https://github.example.com",
              "api_url": "https://api.github.example.com",
              "client_id": "unittest-client-id",
              "client_secret": "unittest-client-secret",
              "private_key_path": "./path/to/key.pem",
              "app_id": "1234appid",
              "default_access_token": "pa-test-personal-access-token",
              "default_installation_id": "ut-default-installation-id-here",
            }]
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
def test_github_provider_source():
    """Return test github provider source instance"""
    db = terrareg.database.Database.get()
    name = "Test Github Provider"
    with terrareg.database.Database.get_connection() as conn:
        conn.execute(db.provider_source.insert().values(
            name=name,
            api_name="test-github-provider",
            provider_source_type=terrareg.provider_source_type.ProviderSourceType.GITHUB,
            config=db.encode_blob(json.dumps({
                "base_url": "https://github.example.com",
                "api_url": "https://api.github.example.com",
                "client_id": "unittest-client-id",
                "client_secret": "unittest-client-secret",
                "login_button_text": "Login via Github using this unit test",
                "private_key_path": "./path/to/key.pem",
                "app_id": "1234appid",
                "default_access_token": "pa-test-personal-access-token",
                "default_installation_id": "ut-default-installation-id-here",
                "auto_generate_github_organisation_namespaces": False
            }))
        ))

    yield terrareg.provider_source.factory.ProviderSourceFactory.get().get_provider_source_by_name(name)

    # Delete provider source
    db = terrareg.database.Database.get()
    with terrareg.database.Database.get_connection() as conn:
        conn.execute(db.provider_source.delete(db.provider_source.c.name==name))


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

