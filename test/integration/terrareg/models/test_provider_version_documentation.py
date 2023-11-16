
import pytest

from test.integration.terrareg import TerraregIntegrationTest
import terrareg.provider_version_documentation_model
import terrareg.provider_version_model
import terrareg.provider_model
import terrareg.models
import terrareg.provider_documentation_type
import terrareg.database
import terrareg.utils
from test.integration.terrareg.fixtures import (
    test_provider_version, test_provider, test_repository,
    test_gpg_key, test_namespace, mock_provider_source,
    mock_provider_source_class, test_provider_category
)


class TestProviderVersionDocumentation(TerraregIntegrationTest):
    """Test ProviderVersionDocumentation"""

    @pytest.mark.parametrize('name, expected_slug', [
        ('test_name', 'test_name'),
        ('test_name.md', 'test_name'),
        ('test_name.html', 'test_name'),
        ('test_name.markdown', 'test_name'),
        ('test_name.html.md', 'test_name'),
        ('test_name.html.markdown', 'test_name'),
    ])
    def test_generate_slug_from_name(self, name, expected_slug):
        """Test generate_slug_from_name"""
        assert terrareg.provider_version_documentation_model.ProviderVersionDocumentation.generate_slug_from_name(
            name=name
        ) == expected_slug

    @pytest.mark.parametrize('type_', [
        terrareg.provider_documentation_type.ProviderDocumentationType.DATA_SOURCE,
        terrareg.provider_documentation_type.ProviderDocumentationType.GUIDE,
        terrareg.provider_documentation_type.ProviderDocumentationType.OVERVIEW,
        terrareg.provider_documentation_type.ProviderDocumentationType.PROVIDER,
        terrareg.provider_documentation_type.ProviderDocumentationType.RESOURCE
    ])
    @pytest.mark.parametrize('subcategory', [
        'Some Subcategory',
        None
    ])
    @pytest.mark.parametrize('description', [
        'Some description',
        None
    ])
    @pytest.mark.parametrize('title', [
        'Unittest Title',
        None
    ])
    def test_create(self, type_, subcategory, description, title, test_provider_version):
        """Test create method"""
        documentation = terrareg.provider_version_documentation_model.ProviderVersionDocumentation.create(
            provider_version=test_provider_version,
            documentation_type=type_,
            name="test-provider-documentation.md",
            title=title,
            description=description,
            filename="docs/resources/test-provider-documentation.md",
            language="hcl",
            subcategory=subcategory,
            content="Some test documentation\nContent!!!"
        )

        try:
            assert isinstance(documentation, terrareg.provider_version_documentation_model.ProviderVersionDocumentation)
            assert isinstance(documentation._pk, int)
            db = terrareg.database.Database.get()
            with db.get_connection() as conn:
                res = conn.execute(db.provider_version_documentation.select().where(
                    db.provider_version_documentation.c.id==documentation._pk
                )).all()
                assert len(res) == 1
                assert dict(res[0]) == {
                    "id": documentation._pk,
                    "provider_version_id": test_provider_version.pk,
                    "name": "test-provider-documentation.md",
                    "title": title,
                    "description": (description or '').encode('utf-8'),
                    "filename": "docs/resources/test-provider-documentation.md",
                    "language": "hcl",
                    "subcategory": subcategory,
                    "content": b"Some test documentation\nContent!!!",
                    "documentation_type": type_,
                    "slug": "test-provider-documentation"
                }
        finally:
            with db.get_connection() as conn:
                conn.execute(db.provider_version_documentation.delete().where(
                    db.provider_version_documentation.c.id==documentation._pk
                ))

    def test__insert_db_row(self, test_provider_version):
        """Test _insert_db_row"""
        pk = terrareg.provider_version_documentation_model.ProviderVersionDocumentation._insert_db_row(
            provider_version=test_provider_version,
            documentation_type=terrareg.provider_documentation_type.ProviderDocumentationType.DATA_SOURCE,
            name="unit-test-provider-documentation.md",
            title="Unit test title",
            description="Test insert description",
            filename="docs/resources/test-insert-provider-documentation.md",
            language="hcl",
            subcategory="Test inserting subcategory",
            content="Some test insert documentation\nContent!!!",
            slug="some-unittest-slug"
        )
        assert isinstance(pk, int)
        db = terrareg.database.Database.get()
        with db.get_connection() as conn:
            res = conn.execute(db.provider_version_documentation.select().where(
                db.provider_version_documentation.c.id==pk
            )).all()
            assert len(res) == 1
            assert dict(res[0]) == {
                "id": pk,
                "provider_version_id": test_provider_version.pk,
                "name": "unit-test-provider-documentation.md",
                "title": "Unit test title",
                "description": "Test insert description".encode('utf-8'),
                "filename": "docs/resources/test-insert-provider-documentation.md",
                "language": "hcl",
                "subcategory": "Test inserting subcategory",
                "content": b"Some test insert documentation\nContent!!!",
                "documentation_type": terrareg.provider_documentation_type.ProviderDocumentationType.DATA_SOURCE,
                "slug": "some-unittest-slug"
            }

    def test_get_by_pk(cls, test_provider_version):
        """Test get_by_pk method"""
        db = terrareg.database.Database.get()
        with db.get_connection() as conn:
            db_res = conn.execute(db.provider_version_documentation.insert().values(
                id=9999923,
                provider_version_id=test_provider_version.pk,
                name="unittest-docs",
                content=b"test",
                slug="some-unittest-slug",
                language="hcl",
                filename="some-testfile.md",
                documentation_type=terrareg.provider_documentation_type.ProviderDocumentationType.GUIDE
            ))
            pk = db_res.lastrowid
        
        res = terrareg.provider_version_documentation_model.ProviderVersionDocumentation.get_by_pk(pk=pk)
        assert isinstance(res, terrareg.provider_version_documentation_model.ProviderVersionDocumentation)
        assert res._pk == pk

        # Test non-existent
        res = terrareg.provider_version_documentation_model.ProviderVersionDocumentation.get_by_pk(pk=83158315)
        assert res is None

    @pytest.mark.parametrize('slug, type_, language, expected_pk', [
        ('some-resource', terrareg.provider_documentation_type.ProviderDocumentationType.RESOURCE, "hcl", 6345),
        # Different lanugage
        ('some-resource', terrareg.provider_documentation_type.ProviderDocumentationType.RESOURCE, "python", 6346),
        # Wrong language
        ('some-resource', terrareg.provider_documentation_type.ProviderDocumentationType.RESOURCE, "other", None),
        # Wrong documentation type
        ('some-resource', terrareg.provider_documentation_type.ProviderDocumentationType.OVERVIEW, "hcl", None),
        # Documentation for older version
        ('some-old-resource', terrareg.provider_documentation_type.ProviderDocumentationType.RESOURCE, "hcl", None),
        # Non-existent documentation
        ('does-not-exist', terrareg.provider_documentation_type.ProviderDocumentationType.RESOURCE, "hcl", None),

    ])
    def test_get(self, slug, type_, language, expected_pk):
        """Test get method"""
        namespace = terrareg.models.Namespace.get("initial-providers")
        provider = terrareg.provider_model.Provider.get(namespace=namespace, name="multiple-versions")
        provider_version = terrareg.provider_version_model.ProviderVersion.get(provider=provider, version="1.5.0")

        # Ensure can successfully obtain documentation
        doc = terrareg.provider_version_documentation_model.ProviderVersionDocumentation.get(
            provider_version=provider_version,
            documentation_type=type_,
            language=language,
            slug=slug
        )
        if expected_pk is not None:
            assert isinstance(doc, terrareg.provider_version_documentation_model.ProviderVersionDocumentation)
            assert doc._pk == expected_pk
        else:
            assert doc is None

    def test_get_by_provider_version(self):
        """Test get_by_provider_version"""
        namespace = terrareg.models.Namespace.get("initial-providers")
        provider = terrareg.provider_model.Provider.get(namespace=namespace, name="multiple-versions")
        provider_version = terrareg.provider_version_model.ProviderVersion.get(provider=provider, version="1.5.0")

        docs = terrareg.provider_version_documentation_model.ProviderVersionDocumentation.get_by_provider_version(
            provider_version=provider_version
        )

        assert isinstance(docs, list)
        assert len(docs) == 4
        for doc_itx in docs:
            assert isinstance(doc_itx, terrareg.provider_version_documentation_model.ProviderVersionDocumentation)
        assert sorted([doc._pk for doc in docs]) == [
            6344, 6345, 6346, 6347
        ]

    @pytest.mark.parametrize('category, slug, language, expected_ids', [
        (terrareg.provider_documentation_type.ProviderDocumentationType.OVERVIEW, 'overview', 'hcl', [6344]),
        # Wrong type
        (terrareg.provider_documentation_type.ProviderDocumentationType.DATA_SOURCE, 'overview', 'hcl', []),
        # Non-existent title
        (terrareg.provider_documentation_type.ProviderDocumentationType.RESOURCE, 'some-non-existent', 'hcl', []),
        # Wrong language
        (terrareg.provider_documentation_type.ProviderDocumentationType.RESOURCE, 'overview', 'python', []),
    ])
    def test_search(self, category, slug, language, expected_ids):
        """Search for provider version documentation using query filters"""
        namespace = terrareg.models.Namespace.get("initial-providers")
        provider = terrareg.provider_model.Provider.get(namespace=namespace, name="multiple-versions")
        provider_version = terrareg.provider_version_model.ProviderVersion.get(provider=provider, version="1.5.0")
        
        res = terrareg.provider_version_documentation_model.ProviderVersionDocumentation.search(
            provider_version=provider_version,
            category=category,
            slug=slug,
            language=language
        )
        assert isinstance(res, list)
        for doc_itx in res:
            assert isinstance(doc_itx, terrareg.provider_version_documentation_model.ProviderVersionDocumentation)
        assert [doc._pk for doc in res] == expected_ids
