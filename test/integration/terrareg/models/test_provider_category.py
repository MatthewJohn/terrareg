
from contextlib import contextmanager
import pytest

from test.integration.terrareg import TerraregIntegrationTest
import terrareg.database
import terrareg.provider_category_model


@pytest.fixture
def generate_test_provider_category():
    @contextmanager
    def generate(user_selectable=True):
        db = terrareg.database.Database.get()
        with db.get_connection() as conn:
            conn.execute(db.provider_category.insert().values(
                id=564341,
                name='Unittest Provider Category',
                slug='unittest-provider-category',
                user_selectable=user_selectable
            ))
        try:
            yield terrareg.provider_category_model.ProviderCategory(pk=564341)
        finally:
            with db.get_connection() as conn:
                conn.execute(db.provider_category.delete().where(
                    db.provider_category.c.id==564341
                ))

    return generate


class TestProviderCategory(TerraregIntegrationTest):
    """Test ProviderCategory class"""

    @pytest.mark.parametrize('exists, user_selectable', [
        (True, True),
        (True, False),
        (False, False)
    ])
    def test_get_by_pk(self, exists, user_selectable):
        """Test get_by_pk method"""
        # Create test ProviderCategory
        db = terrareg.database.Database.get()
        if exists:
            with db.get_connection() as conn:
                conn.execute(db.provider_category.insert().values(
                    id=564341,
                    name='Unittest Provider Category',
                    slug='unittest-provider-category',
                    user_selectable=user_selectable
                ))

        try:

            # Attempt to obtain instance using get_by_pk
            res = terrareg.provider_category_model.ProviderCategory.get_by_pk(pk=564341)
            if exists:
                assert isinstance(res, terrareg.provider_category_model.ProviderCategory)
                assert res._pk == 564341
            else:
                assert res is None

        finally:
            with db.get_connection() as conn:
                res = conn.execute(db.provider_category.delete().where(
                    db.provider_category.c.id==564341
                ))

    def test_name(self, generate_test_provider_category):
        """Test name property"""
        with generate_test_provider_category() as test_provider_category:
            assert test_provider_category.name == "Unittest Provider Category"

    def test_slug(self, generate_test_provider_category):
        """Test slug property"""
        with generate_test_provider_category() as test_provider_category:
            assert test_provider_category.slug == "unittest-provider-category"

    def test_pk(self, generate_test_provider_category):
        """Test PK property"""
        with generate_test_provider_category() as test_provider_category:
            assert test_provider_category.pk == 564341

    @pytest.mark.parametrize('pk, exists', [
        (564341, True),
        (958888, False)
    ])
    def test_exists(self, pk, exists, generate_test_provider_category):
        """Test exists property"""
        with generate_test_provider_category():
            instance = terrareg.provider_category_model.ProviderCategory(pk=pk)
            assert instance.exists is exists

    @pytest.mark.parametrize('user_selectable', [
        True,
        False
    ])
    def test_user_selectable(self, user_selectable, generate_test_provider_category):
        """Test user_selectable property"""
        with generate_test_provider_category(user_selectable=user_selectable) as test_provider_category:
            assert test_provider_category.user_selectable is user_selectable

    def test___init__(self):
        """Test __init__ method"""
        instance = terrareg.provider_category_model.ProviderCategory(pk=123)
        assert instance._pk == 123
        assert instance._cache_db_row is None

    @pytest.mark.parametrize('exists', [
        True,
        False
    ])
    def test__get_db_row(self, exists, generate_test_provider_category):
        """Test _get_db_row method."""
        with generate_test_provider_category():
            instance = terrareg.provider_category_model.ProviderCategory(pk=564341 if exists else 59999912)
            assert instance._cache_db_row is None

            row = instance._get_db_row()
            if exists:
                assert row is not None
                assert dict(row) == {
                    'id': 564341,
                    'name': 'Unittest Provider Category',
                    'slug': 'unittest-provider-category',
                    'user_selectable': True
                }
            else:
                assert row is None

            assert instance._cache_db_row is row

            instance._cache_db_row = {"a": "b"}
            assert instance._get_db_row() == {"a": "b"}

    def test_get_v2_include(self, generate_test_provider_category):
        """Test get_v2_include method"""
        with generate_test_provider_category() as test_provider_category:
            assert test_provider_category.get_v2_include() == {
                "type": "categories",
                "id": "564341",
                "attributes": {
                    "name": "Unittest Provider Category",
                    "slug": "unittest-provider-category",
                    "user-selectable": True,
                },
                "links": {
                    "self": f"/v2/categories/564341"
                }
            }
