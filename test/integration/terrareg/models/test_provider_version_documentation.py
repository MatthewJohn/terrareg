
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
        ('some_resource', terrareg.provider_documentation_type.ProviderDocumentationType.RESOURCE, "hcl", 6345),
        # Different lanugage
        ('some_resource', terrareg.provider_documentation_type.ProviderDocumentationType.RESOURCE, "python", 6346),
        # Wrong language
        ('some_resource', terrareg.provider_documentation_type.ProviderDocumentationType.RESOURCE, "other", None),
        # Wrong documentation type
        ('some_resource', terrareg.provider_documentation_type.ProviderDocumentationType.OVERVIEW, "hcl", None),
        # Documentation for older version
        ('some_old_resource', terrareg.provider_documentation_type.ProviderDocumentationType.RESOURCE, "hcl", None),
        # Non-existent documentation
        ('does_not_exist', terrareg.provider_documentation_type.ProviderDocumentationType.RESOURCE, "hcl", None),

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

    def test_exists(self):
        """Test exists property"""
        valid = terrareg.provider_version_documentation_model.ProviderVersionDocumentation(pk=6344)
        assert valid.exists is True

        non_existent = terrareg.provider_version_documentation_model.ProviderVersionDocumentation(pk=999987)
        assert non_existent.exists is False
        non_existent._cache_db_row = {"test": "row"}
        assert non_existent.exists is True

    @pytest.mark.parametrize('test_data, expected_title', [
        ({"name": "unittest-name", "title": "unittest-title"}, 'unittest-title'),
        ({"name": "unittest-name", "title": None}, 'unittest-name'),
        ({"name": "unittest-name-with-md.md", "title": None}, 'unittest-name-with-md'),
    ])
    def test_title(self, test_data, expected_title):
        """Test title property"""
        inst = terrareg.provider_version_documentation_model.ProviderVersionDocumentation.get_by_pk(pk=6344)
        assert inst.title == "Overview"

        inst._cache_db_row = test_data
        assert inst.title == expected_title

    def test_pk(self):
        """Test pk property"""
        inst = terrareg.provider_version_documentation_model.ProviderVersionDocumentation(pk=6344)
        assert inst.pk == 6344
        inst._pk = 1234
        assert inst.pk == 1234

    @pytest.mark.parametrize('pk, expected_type', [
        (6344, terrareg.provider_documentation_type.ProviderDocumentationType.OVERVIEW),
        (6345, terrareg.provider_documentation_type.ProviderDocumentationType.RESOURCE),
    ])
    def test_category(self, pk, expected_type):
        """Test category property"""
        inst = terrareg.provider_version_documentation_model.ProviderVersionDocumentation.get_by_pk(pk=pk)
        assert inst.category is expected_type

    @pytest.mark.parametrize('pk, expected_language', [
        (6345, "hcl"),
        (6346, "python"),
    ])
    def test_language(self, pk, expected_language):
        """Test language property"""
        inst = terrareg.provider_version_documentation_model.ProviderVersionDocumentation.get_by_pk(pk=pk)
        assert inst.language == expected_language

    @pytest.mark.parametrize('pk, expected_filename', [
        (6345, "data-sources/thing.md"),
        (6347, "resources/new-thing.md"),
    ])
    def test_filename(self, pk, expected_filename):
        """Test filename property"""
        inst = terrareg.provider_version_documentation_model.ProviderVersionDocumentation.get_by_pk(pk=pk)
        assert inst.filename == expected_filename

    @pytest.mark.parametrize('pk, expected_slug', [
        (6345, "some_resource"),
        (6346, "some_resource"),
        (6347, "some_new_resource"),
    ])
    def test_slug(self, pk, expected_slug):
        """Test slug property"""
        inst = terrareg.provider_version_documentation_model.ProviderVersionDocumentation.get_by_pk(pk=pk)
        assert inst.slug == expected_slug

    @pytest.mark.parametrize('pk, expected_subcategory', [
        (6344, None),
        (6345, "some-subcategory"),
        (6346, "some-subcategory"),
        (6347, "some-second-subcategory"),
    ])
    def test_subcategory(self, pk, expected_subcategory):
        """Test subcategory property"""
        inst = terrareg.provider_version_documentation_model.ProviderVersionDocumentation.get_by_pk(pk=pk)
        assert inst.subcategory == expected_subcategory

    def test___init__(self):
        """Test __init__ method"""
        inst = terrareg.provider_version_documentation_model.ProviderVersionDocumentation(pk=52312)
        assert inst._pk == 52312
        assert inst._cache_db_row is None

    def test__get_db_row(self):
        """Test _get_db_row method"""
        inst = terrareg.provider_version_documentation_model.ProviderVersionDocumentation.get_by_pk(pk=6345)
        inst._cache_db_row = None

        assert dict(inst._get_db_row()) == {
            'content': b'Documentation for generating a thing!',
            'description': b'Inital thing for multiple versions provider',
            'documentation_type': terrareg.provider_documentation_type.ProviderDocumentationType.RESOURCE,
            'filename': 'data-sources/thing.md',
            'id': 6345,
            'language': 'hcl',
            'name': 'some_resource',
            'provider_version_id': 7,
            'slug': 'some_resource',
            'subcategory': 'some-subcategory',
            'title': 'multiple_versions_thing',
        }
        assert dict(inst._cache_db_row) == {
            'content': b'Documentation for generating a thing!',
            'description': b'Inital thing for multiple versions provider',
            'documentation_type': terrareg.provider_documentation_type.ProviderDocumentationType.RESOURCE,
            'filename': 'data-sources/thing.md',
            'id': 6345,
            'language': 'hcl',
            'name': 'some_resource',
            'provider_version_id': 7,
            'slug': 'some_resource',
            'subcategory': 'some-subcategory',
            'title': 'multiple_versions_thing',
        }

        inst._cache_db_row = {"test": "row"}
        assert inst._get_db_row() == {"test": "row"}

        # Test non-existent
        inst = terrareg.provider_version_documentation_model.ProviderVersionDocumentation(pk=2313548)
        assert inst._get_db_row() is None
        assert inst._cache_db_row is None

    def test_get_api_outline(self):
        """Test get_api_outline method"""
        inst = terrareg.provider_version_documentation_model.ProviderVersionDocumentation.get_by_pk(pk=6345)
        assert inst.get_api_outline() == {
            'category': 'resources',
            'id': '6345',
            'language': 'hcl',
            'path': 'data-sources/thing.md',
            'slug': 'some_resource',
            'subcategory': 'some-subcategory',
            'title': 'multiple_versions_thing',
        }

    def test_get_v2_api_outline(self):
        """Test get_v2_api_outline"""
        inst = terrareg.provider_version_documentation_model.ProviderVersionDocumentation.get_by_pk(pk=6345)
        assert inst.get_v2_api_outline() == {
            'id': '6345',
            'type': 'provider-docs',
            'attributes': {
                'category': 'resources',
                'language': 'hcl',
                'path': 'data-sources/thing.md',
                'slug': 'some_resource',
                'subcategory': 'some-subcategory',
                'title': 'multiple_versions_thing',
                'truncated': False
            },
            'links': {'self': '/v2/provider-docs/6345'},
        }

    def test_get_v2_api_details(self, html: bool=False):
        """Test get_v2_api_details"""
        inst = terrareg.provider_version_documentation_model.ProviderVersionDocumentation.get_by_pk(pk=6347)
        assert inst.get_v2_api_details(html=False) == {
            'type': 'provider-docs',
            'id': '6347',
            'attributes': {
                'category': 'resources',
                'content': """
# Some Title!

## Second title

This module:

 * Creates something
 * Does something else

and it _really_ *does* work!
""",
                'language': 'hcl',
                'path': 'resources/new-thing.md',
                'slug': 'some_new_resource',
                'subcategory': 'some-second-subcategory',
                'title': 'multiple_versions_thing_new',
                'truncated': False
            },
          'links': {'self': '/v2/provider-docs/6347'},
        }
        assert inst.get_v2_api_details(html=True) == {
            'type': 'provider-docs',
            'id': '6347',
            'attributes': {
                'category': 'resources',
                'content': """
<h1 id="terrareg-anchor-resourcesnew-thingmd-some-title">Some Title!</h1>
<h2 id="terrareg-anchor-resourcesnew-thingmd-second-title">Second title</h2>
<p>This module:</p>
<ul>
<li>Creates something</li>
<li>Does something else</li>
</ul>
<p>and it <em>really</em> <em>does</em> work!</p>
""".strip(),
                'language': 'hcl',
                'path': 'resources/new-thing.md',
                'slug': 'some_new_resource',
                'subcategory': 'some-second-subcategory',
                'title': 'multiple_versions_thing_new',
                'truncated': False
            },
          'links': {'self': '/v2/provider-docs/6347'},
        }

    @pytest.mark.parametrize('html, expected_result', [
        (False, """
# Some Title!

## Second title

This module:

 * Creates something
 * Does something else

and it _really_ *does* work!
""",),
        (True, """
<h1 id="terrareg-anchor-resourcesnew-thingmd-some-title">Some Title!</h1>
<h2 id="terrareg-anchor-resourcesnew-thingmd-second-title">Second title</h2>
<p>This module:</p>
<ul>
<li>Creates something</li>
<li>Does something else</li>
</ul>
<p>and it <em>really</em> <em>does</em> work!</p>
""".strip())
    ])
    def test_get_content(self, html, expected_result):
        """Test get_content method"""
        inst = terrareg.provider_version_documentation_model.ProviderVersionDocumentation.get_by_pk(pk=6347)
        assert inst.get_content(html=html) == expected_result

    @pytest.mark.parametrize('html, content, expected_result', [
        (False, '# Hi!', '# Hi!'),
        (True, '# Hi!', '<h1 id="terrareg-anchor-testmd-hi">Hi!</h1>'),
        (False, None, ''),
        (True, None, ''),
        (False, '', ''),
        (True, '', ''),
    ])
    def test_get_content_conversion(self, html, content, expected_result):
        """Test get_content"""
        inst = terrareg.provider_version_documentation_model.ProviderVersionDocumentation.get_by_pk(pk=6347)
        inst._cache_db_row = {
            "filename": "test.md",
            "content": terrareg.database.Database.encode_blob(content)
        }
        assert inst.get_content(html=html) == expected_result
