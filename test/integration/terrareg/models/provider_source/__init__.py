
import json
import unittest.mock

import pytest

import terrareg.database
import terrareg.provider_source_type
import terrareg.models
import terrareg.repository_model
from test.test_gpg_key import public_ascii_armor
import terrareg.provider_tier
import terrareg.provider_category_model
import terrareg.provider_model
from test.integration.terrareg.fixtures import (
    test_namespace,
    test_provider_category
)


@pytest.fixture()
def test_provider_source(request):
    """Create provider source object"""
    db = terrareg.database.Database.get()
    with db.get_connection() as conn:
        conn.execute(db.provider_source.insert().values(
            name="Test Provider Source",
            api_name="test-provider-source",
            # Use one of the request types, but since
            # this is only used by the factory to generate a class,
            # it won't be used
            provider_source_type=terrareg.provider_source_type.ProviderSourceType.GITHUB,
            config=db.encode_blob(json.dumps(request.cls.ADDITIONAL_CONFIG))
        ))

    yield request.cls._CLASS(name="Test Provider Source")

    # Delete provider source
    with db.get_connection() as conn:
        conn.execute(db.provider_source.delete(
            db.provider_source.c.api_name=="test-provider-source"
        ))


@pytest.fixture
def test_repository(test_namespace, test_provider_source):
    """Create test repository"""
    repository = terrareg.repository_model.Repository.create(
        provider_source=test_provider_source,
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

