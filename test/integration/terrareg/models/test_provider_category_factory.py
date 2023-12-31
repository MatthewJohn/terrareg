
from unicodedata import category
import unittest.mock

import pytest

from test.integration.terrareg import TerraregIntegrationTest
from test.integration.terrareg.fixtures import test_provider_category
import terrareg.provider_category_model
import terrareg.database
import terrareg.errors


class TestProviderCategoryFactory(TerraregIntegrationTest):

    _TEST_DATA = {}

    def test_get(self):
        """Test get method"""
        terrareg.provider_category_model.ProviderCategoryFactory._INSTANCE = None
        instance = terrareg.provider_category_model.ProviderCategoryFactory.get()
        assert isinstance(instance, terrareg.provider_category_model.ProviderCategoryFactory)
        assert terrareg.provider_category_model.ProviderCategoryFactory._INSTANCE is instance
        assert terrareg.provider_category_model.ProviderCategoryFactory.get() is instance

    @pytest.mark.parametrize('slug, exists', [
        ("example-category", True),
        ("another-example-category", False),
    ])
    def test_get_provider_category_by_slug(self, slug, exists, test_provider_category):
        """Test get_provider_category_by_slug"""
        res = terrareg.provider_category_model.ProviderCategoryFactory.get().get_provider_category_by_slug(slug=slug)

        if exists:
            assert isinstance(res, terrareg.provider_category_model.ProviderCategory)
            assert res._pk == 1
        else:
            assert res is None

    @pytest.mark.parametrize('pk, exists', [
        (54, True),
        (99998, False),
    ])
    def test_get_provider_category_by_pk(self, pk, exists):
        """Test get_provider_category_by_pk"""
        res = terrareg.provider_category_model.ProviderCategoryFactory.get().get_provider_category_by_pk(pk=pk)

        if exists:
            assert isinstance(res, terrareg.provider_category_model.ProviderCategory)
            assert res._pk == pk
        else:
            assert res is None

    def test_get_all_provider_categories(self):
        """Test get_all_provider_categories"""
        all = terrareg.provider_category_model.ProviderCategoryFactory.get().get_all_provider_categories()
        assert sorted([category._pk for category in all]) == [
            54, 55, 99, 100, 101, 523
        ]

    @pytest.mark.parametrize('name, expected_slug', [
        ("test", "test"),
        ("test name", "test-name"),
        (" test  name ", "test-name"),
        ("test name/or something", "test-name-or-something"),
        ("test@a-to_123/that", "test-a-to-123-that"),
    ])
    def test_name_to_slug(self, name, expected_slug):
        """Test name_to_slug"""
        assert terrareg.provider_category_model.ProviderCategoryFactory.get().name_to_slug(name) == expected_slug

    @pytest.mark.parametrize('config, throws_error, expected_provider_categories', [
        # No provider categories defined
        ('[]', False, []),

        # Single category
        ('[{"id": 1, "name": "Unittest Category", "slug": "unittest-category", "user-selectable": false}]', False, [(1, "Unittest Category", "unittest-category", False)]),

        # Mulitple categories
        ('[{"id": 1, "name": "Unittest Category", "slug": "unittest-category", "user-selectable": false}, {"id": 2, "name": "Unittest Category2", "slug": "unittest-category2", "user-selectable": true}]',
         False, [(1, "Unittest Category", "unittest-category", False), (2, "Unittest Category2", "unittest-category2", True)]),

        # Without slug
        ('[{"id": 1, "name": "Test Category-without_slug", "user-selectable": true}]', False, [(1, "Test Category-without_slug", "test-category-without-slug", True)]),
        # Without user-selectable
        ('[{"id": 1, "name": "Unittest Category3", "slug": "unittest-category3"}]', False, [(1, "Unittest Category3", "unittest-category3", True)]),

        # With non-matching slug
        ('[{"id": 1, "name": "Category Name Here", "slug": "this-is-a-slug", "user-selectable": true}]', False, [(1, "Category Name Here", "this-is-a-slug", True)]),

        # Duplicate ID
        ('[{"id": 1, "name": "Unittest Category", "slug": "unittest-category", "user-selectable": false}, {"id": 1, "name": "Unittest Category2", "slug": "unittest-category2", "user-selectable": true}]', False,
         [(1, "Unittest Category2", "unittest-category2", True)]),

        # Invalid JSON
        ('[', True, ()),
        # Missing ID
        ('[{"name": "Unittest Category", "slug": "unittest-category", "user-selectable": false}]', True, ()),
        # Invalid ID
        ('[{"id": "blah", "name": "Unittest Category", "slug": "unittest-category", "user-selectable": false}]', True, ()),
        # Missing name
        ('[{"id": 1, "slug": "unittest-category", "user-selectable": false}]', True, ()),
        # Invalid Name
        ('[{"id": 1, "name": {"test": 2}, "slug": "unittest-category", "user-selectable": false}]', True, ()),
    ])
    def test_initialise_from_config(self, config, throws_error, expected_provider_categories):
        """Test initialise_from_config method."""
        db = terrareg.database.Database.get()
        with db.get_connection() as conn:
            conn.execute(db.provider_category.delete())

        with unittest.mock.patch('terrareg.config.Config.PROVIDER_CATEGORIES', config):
            factory = terrareg.provider_category_model.ProviderCategoryFactory.get()
            if throws_error:
                with pytest.raises(terrareg.errors.InvalidProviderCategoryConfigError):
                    factory.initialise_from_config()
            else:
                factory.initialise_from_config()

                all_categories = [
                    (category.pk, category.name, category.slug, category.user_selectable)
                    for category in factory.get_all_provider_categories()
                ]

                for category_ in all_categories:
                    assert category_ in expected_provider_categories
                    expected_provider_categories.remove(category_)
                
                assert len(expected_provider_categories) == 0
