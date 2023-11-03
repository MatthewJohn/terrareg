
import json

import pytest

import terrareg.database
import terrareg.provider_source_type


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
