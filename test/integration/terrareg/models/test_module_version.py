
from datetime import datetime
import unittest.mock
import pytest
import sqlalchemy
from terrareg.analytics import AnalyticsEngine
from terrareg.config import ModuleVersionReindexMode
from terrareg.database import Database

from terrareg.models import Example, ExampleFile, Module, Namespace, ModuleProvider, ModuleVersion
import terrareg.errors
from test.integration.terrareg import TerraregIntegrationTest
from terrareg.module_version_create import module_version_create


class TestModuleVersion(TerraregIntegrationTest):

    @pytest.mark.parametrize('version', [
        'astring',
        '',
        '1',
        '1.1',
        '.23.1',
        '1.1.1.1',
        '1.1.1.',
        '1.2.3-dottedsuffix1.2',
        '1.2.3-invalid-suffix',
        '1.0.9-'
    ])
    def test_invalid_module_versions(self, version):
        """Test invalid module versions"""
        namespace = Namespace(name='test')
        module = Module(namespace=namespace, name='test')
        module_provider = ModuleProvider(module=module, name='test')
        with pytest.raises(terrareg.errors.InvalidVersionError):
            ModuleVersion(module_provider=module_provider, version=version)

    @pytest.mark.parametrize('version,beta', [
        ('1.1.1', False),
        ('13.14.16', False),
        ('1.10.10', False),
        ('01.01.01', False),  # @TODO Should this be allowed?
        ('1.2.3-alpha', True),
        ('1.2.3-beta', True),
        ('1.2.3-anothersuffix1', True),
        ('1.2.2-123', True)
    ])
    def test_valid_module_versions(self, version, beta):
        """Test valid module versions"""
        namespace = Namespace(name='test')
        module = Module(namespace=namespace, name='test')
        module_provider = ModuleProvider(module=module, name='test')
        module_version = ModuleVersion(module_provider=module_provider, version=version)
        assert module_version._extracted_beta_flag == beta

    def test_module_version_create(self):
        """Test creating DB row"""
        namespace = Namespace.get(name='testcreation', create=True)
        module = Module(namespace=namespace, name='test-module')
        module_provider = ModuleProvider.get(module=module, name='testprovider', create=True)
        module_provider_row = module_provider._get_db_row()

        module_version = ModuleVersion(module_provider=module_provider, version='1.0.0')

        # Ensure that no DB row is returned for new module version
        assert module_version._get_db_row() == None

        # Insert module version into database
        with module_version_create(module_version):
            pass

        # Ensure that a DB row is now returned
        new_db_row = module_version._get_db_row()
        assert new_db_row['module_provider_id'] == module_provider_row['id']
        assert new_db_row['extraction_complete'] is True
        assert type(new_db_row['id']) == int

        assert new_db_row['published'] == False
        assert new_db_row['version'] == '1.0.0'

        assert new_db_row['beta'] == False

        for attr in ['description', 'module_details_id', 'owner',
                     'published_at', 'repo_base_url_template',
                     'repo_browse_url_template', 'repo_clone_url_template',
                     'variable_template']:
            assert new_db_row[attr] == None

    def test_module_version_create_beta_version(self):
        """Test creating DB row for beta version"""
        namespace = Namespace.get(name='testcreation', create=True)
        module = Module(namespace=namespace, name='test-module')
        module_provider = ModuleProvider.get(module=module, name='testprovider', create=True)
        module_provider_row = module_provider._get_db_row()

        module_version = ModuleVersion(module_provider=module_provider, version='1.0.0-beta')

        # Ensure that no DB row is returned for new module version
        assert module_version._get_db_row() == None

        # Insert module version into database
        with module_version_create(module_version):
            pass

        # Ensure that a DB row is now returned
        new_db_row = module_version._get_db_row()
        assert new_db_row['module_provider_id'] == module_provider_row['id']
        assert type(new_db_row['id']) == int

        assert new_db_row['published'] == False
        assert new_db_row['version'] == '1.0.0-beta'

        assert new_db_row['beta'] == True

        for attr in ['description', 'module_details_id', 'owner',
                     'published_at', 'repo_base_url_template',
                     'repo_browse_url_template', 'repo_clone_url_template',
                     'variable_template']:
            assert new_db_row[attr] == None

    @pytest.mark.parametrize('module_version_reindex_mode,previous_publish_state,expected_published_state,should_raise_error', [
        # Legacy mode should allow the re-index and ignore pre-existing version for setting published
        (ModuleVersionReindexMode.LEGACY, False, False, False),
        ## With previous version published
        (ModuleVersionReindexMode.LEGACY, True, False, False),

        # Auto-publish mode should return the previously indexed module's published state
        (ModuleVersionReindexMode.AUTO_PUBLISH, False, False, False),
        (ModuleVersionReindexMode.AUTO_PUBLISH, True, True, False),

        # Prohibit mode should raise an error
        (ModuleVersionReindexMode.PROHIBIT, False, False, True)
    ])
    def test_module_version_create_replace_existing(self, module_version_reindex_mode,
                                            previous_publish_state, expected_published_state,
                                            should_raise_error):
        """Test module_version_create with pre-existing module version"""

        db = Database.get()

        try:
            with db.get_engine().connect() as conn:
                conn.execute(db.namespace.insert().values(
                    id=9999,
                    namespace='testcreationunique'
                ))

                conn.execute(db.module_provider.insert().values(
                    id=10000,
                    namespace_id=9999,
                    module='test-module',
                    provider='testprovider'
                ))

                conn.execute(db.module_version.insert().values(
                    id=10001,
                    module_provider_id=10000,
                    version='1.1.0',
                    published=previous_publish_state,
                    beta=False,
                    internal=False,
                    extraction_complete=True
                ))

                # Create submodules
                conn.execute(db.sub_module.insert().values(
                    id=10002,
                    parent_module_version=10001,
                    type='example',
                    path='example/test-modal-db-row-create-here'
                ))
                conn.execute(db.sub_module.insert().values(
                    id=10003,
                    parent_module_version=10001,
                    type='submodule',
                    path='modules/test-modal-db-row-create-there'
                ))

                # Create example file
                conn.execute(db.example_file.insert().values(
                    id=10004,
                    submodule_id=10002,
                    path='testfile.tf',
                    content=None
                ))

                # Create download analytics
                conn.execute(db.analytics.insert().values(
                    id=10005,
                    parent_module_version=10001,
                    timestamp=datetime.now(),
                    terraform_version='1.0.0',
                    analytics_token='unittest-download',
                    auth_token='abcefg',
                    environment='test'
                ))

            namespace = Namespace.get(name='testcreationunique')
            assert namespace is not None
            module = Module(namespace=namespace, name='test-module')
            module_provider = ModuleProvider.get(module=module, name='testprovider')
            assert module_provider is not None
            module_provider_row = module_provider._get_db_row()

            module_version = ModuleVersion(module_provider=module_provider, version='1.1.0')

            # Ensure that pre-existing row is returned
            pre_existing_row = module_version._get_db_row()
            assert pre_existing_row is not None
            assert pre_existing_row['id'] == 10001

            with unittest.mock.patch('terrareg.config.Config.MODULE_VERSION_REINDEX_MODE', module_version_reindex_mode):
                # If confiugred to raise an error, check that it is
                if should_raise_error:
                    with pytest.raises(terrareg.errors.ReindexingExistingModuleVersionsIsProhibitedError):
                        with module_version_create(module_version):
                            pass

                    # Do not run any further tests as the exception will
                    # rollback any changes
                    return
                else:
                    # Otherwise check the return value
                    with module_version_create(module_version):
                        pass

            # Ensure that a DB row is now returned
            new_db_row = module_version._get_db_row()
            assert new_db_row['module_provider_id'] == module_provider_row['id']
            assert type(new_db_row['id']) == int
            assert new_db_row['id'] != 10001

            assert new_db_row['published'] == expected_published_state
            assert new_db_row['version'] == '1.1.0'

            for attr in ['description', 'module_details_id', 'owner',
                        'published_at', 'repo_base_url_template',
                        'repo_browse_url_template', 'repo_clone_url_template',
                        'variable_template']:
                assert new_db_row[attr] == None

            # Ensure that all moduleversion, submodules and example files have been removed
            with db.get_engine().connect() as conn:
                mv_res = conn.execute(db.module_version.select(
                    db.module_version.c.id == 10001
                ))
                assert [r for r in mv_res] == []

                # Check for any submodules with the original IDs
                # or with the previous module ID or with the example
                # paths
                sub_module_res = conn.execute(db.sub_module.select().where(
                    sqlalchemy.or_(
                        db.sub_module.c.id.in_((10002, 10003)),
                        db.sub_module.c.parent_module_version == 10001,
                        db.sub_module.c.path.in_(('example/test-modal-db-row-create-here',
                                                'modules/test-modal-db-row-create-there'))
                    )
                ))
                assert [r for r in sub_module_res] == []

                # Ensure example files have been removed
                example_file_res = conn.execute(db.example_file.select().where(
                    db.example_file.c.id == 10004
                ))
                assert [r for r in example_file_res] == []

                # Ensure analytics are retained
                analytics_res = conn.execute(db.analytics.select().where(
                    db.analytics.c.id==10005
                ))
                analytics_res = list(analytics_res)
                assert len(analytics_res) == 1
                assert analytics_res[0]['id'] == 10005
                # Assert that analytics row has been updated to new module version ID
                assert analytics_res[0]['parent_module_version'] == new_db_row['id']
                assert analytics_res[0]['environment'] == 'test'

                # Ensure namespace still exists
                namespace = Namespace.get('testcreationunique')
                assert namespace is not None
                assert namespace.pk == 9999
        finally:
            # Clear down test data
            ns = Namespace.get('testcreationunique')
            if ns:
                module = Module(ns, 'test-module')
                module_provider = ModuleProvider.get(module, 'testprovider')
                if module_provider:
                    module_provider.delete()
                with db.get_engine().connect() as conn:
                    conn.execute(db.namespace.delete().where(
                        db.namespace.c.id==9999
                    ))

    def test_module_version_create_error(self):
        """Test module_version_create with extraction error"""

        db = Database.get()

        try:
            with db.get_engine().connect() as conn:
                conn.execute(db.namespace.insert().values(
                    id=9999,
                    namespace='testcreationunique'
                ))

                conn.execute(db.module_provider.insert().values(
                    id=10000,
                    namespace_id=9999,
                    module='test-module',
                    provider='testprovider'
                ))

                conn.execute(db.module_version.insert().values(
                    id=10001,
                    module_provider_id=10000,
                    version='1.1.0',
                    published=True,
                    beta=False,
                    internal=False,
                    extraction_complete=True
                ))

                # Create submodules
                conn.execute(db.sub_module.insert().values(
                    id=10002,
                    parent_module_version=10001,
                    type='example',
                    path='example/test-modal-db-row-create-here'
                ))
                conn.execute(db.sub_module.insert().values(
                    id=10003,
                    parent_module_version=10001,
                    type='submodule',
                    path='modules/test-modal-db-row-create-there'
                ))

                # Create example file
                conn.execute(db.example_file.insert().values(
                    id=10004,
                    submodule_id=10002,
                    path='testfile.tf',
                    content=None
                ))

                # Create download analytics
                conn.execute(db.analytics.insert().values(
                    id=10005,
                    parent_module_version=10001,
                    timestamp=datetime.now(),
                    terraform_version='1.0.0',
                    analytics_token='unittest-download',
                    auth_token='abcefg',
                    environment='test'
                ))

            namespace = Namespace.get(name='testcreationunique')
            assert namespace is not None
            module = Module(namespace=namespace, name='test-module')
            module_provider = ModuleProvider.get(module=module, name='testprovider')
            assert module_provider is not None

            old_module_version = ModuleVersion(module_provider=module_provider, version='1.1.0')

            # Ensure that pre-existing row is returned
            pre_existing_row = old_module_version._get_db_row()
            assert pre_existing_row is not None
            assert pre_existing_row['id'] == 10001

            # Create new instance for creation
            create_module_version = ModuleVersion(module_provider=module_provider, version='1.1.0')

            class UnittestException(Exception):
                pass

            with unittest.mock.patch('terrareg.config.Config.MODULE_VERSION_REINDEX_MODE', ModuleVersionReindexMode.AUTO_PUBLISH):
                # If confiugred to raise an error, check that it is
                with pytest.raises(UnittestException):
                    with module_version_create(create_module_version):
                        raise UnittestException('Unit test extraction exception')

            # Check member variable for PK pinning of old version
            new_module_pk = create_module_version._pk
            assert type(new_module_pk) is int

            # Ensure previous module version is in it's previous state
            # and the new module has been rememoved from the database
            check_module_version = ModuleVersion(module_provider=module_provider, version='1.1.0')
            assert check_module_version._get_db_row() is not None
            assert check_module_version._get_db_row() == pre_existing_row

            # Check that new module version no longer exists
            new_module_version = ModuleVersion(module_provider=module_provider, version='1.1.0')
            ## Pin to PK of new module version
            new_module_version._pk = new_module_pk
            assert new_module_version._get_db_row() is None

            # @TODO Check that all submodules and indexed files are still present

        finally:
            # Clear down test data
            ns = Namespace.get('testcreationunique')
            if ns:
                module = Module(ns, 'test-module')
                module_provider = ModuleProvider.get(module, 'testprovider')
                if module_provider:
                    module_provider.delete()
                with db.get_engine().connect() as conn:
                    conn.execute(db.namespace.delete().where(
                        db.namespace.c.id==9999
                    ))

    @pytest.mark.parametrize('template,version,published,expected_string', [
        ('>= {major_minus_one}.{minor_minus_one}.{patch_minus_one}', '0.0.0', True, '>= 0.0.0'),
        ('= {major}.{minor}.{patch}', '1.5.0', True, '= 1.5.0'),
        ('<= {major_plus_one}.{minor_plus_one}.{patch_plus_one}', '1.5.0', True, '<= 2.6.1'),
        ('>= {major_minus_one}.{minor_minus_one}.{patch_minus_one}', '4.3.2', True, '>= 3.2.1'),
        ('< {major_plus_one}.0.0', '10584.321.564', True, '< 10585.0.0'),
        # Test older version to ensure it is shown as specific version
        ('>= {major_minus_one}.{minor_minus_one}.{patch_minus_one}', '0.0.1', True, '0.0.1'),
        # Test that beta version returns the version and
        # ignores the version template
        ('>= {major_minus_one}.{minor_minus_one}.{patch_minus_one}', '5.6.23-beta', False, '5.6.23-beta'),
        ('>= {major_minus_one}.{minor_minus_one}.{patch_minus_one}', '5.6.24-beta', True, '5.6.24-beta'),
        # Non-published version
        ('>= {major_minus_one}.{minor_minus_one}.{patch_minus_one}', '5.6.25', False, '5.6.25'),
    ])
    def test_get_terraform_example_version_string(self, template, version, published, expected_string):
        """Test get_terraform_example_version_string method"""
        with unittest.mock.patch('terrareg.config.Config.TERRAFORM_EXAMPLE_VERSION_TEMPLATE', template):
            namespace = Namespace.get(name='test', create=True)
            module = Module(namespace=namespace, name='test')
            module_provider = ModuleProvider.get(module=module, name='test', create=True)
            module_version = ModuleVersion(module_provider=module_provider, version=version)
            with module_version_create(module_version):
                pass
            if published:
                module_version.publish()
            assert module_version.get_terraform_example_version_string() == expected_string

    def test_delete(self):
        """Test deletion of module version."""
        namespace = Namespace(name='testnamespace')
        module = Module(namespace=namespace, name='wrongversionorder')
        module_provider = ModuleProvider.get(module=module, name='testprovider')

        existing_module_versions = [v.version for v in module_provider.get_versions()]
        assert '2.5.5' not in existing_module_versions

        # Create test module version
        module_version = ModuleVersion(module_provider=module_provider, version='2.5.5')
        with module_version_create(module_version):
            pass
        module_version.publish()
        module_version_pk = module_version.pk

        new_versions = [v.version for v in module_provider.get_versions()]
        assert '2.5.5' in new_versions
        assert len(new_versions) == (len(existing_module_versions) + 1)

        # Create example and files
        example = Example.create(module_version=module_version, module_path='examples/test_example')
        example_pk = example.pk
        example_file = ExampleFile.create(example=example, path='main.tf')
        example_file_pk = example_file.pk

        # Create analytics
        AnalyticsEngine.record_module_version_download(
            module_version=module_version,
            terraform_version='1.0.0',
            analytics_token='unittest',
            user_agent='',
            auth_token=None
        )

        # Ensure that all of the rows can be fetched
        db = Database.get()
        with db.get_connection() as conn:

            res = conn.execute(db.module_version.select().where(db.module_version.c.id==module_version_pk))
            assert res.fetchone() is not None

            res = conn.execute(db.sub_module.select().where(db.sub_module.c.id==example_pk))
            assert res.fetchone() is not None

            res = conn.execute(db.example_file.select().where(db.example_file.c.id==example_file_pk))
            assert res.fetchone() is not None

            analytics_row_id = conn.execute(
                db.analytics.select().where(
                    db.analytics.c.parent_module_version==module_version_pk
                )
            ).fetchone()['id']

        # Delete module version
        module_version.delete()

        # Check module_version, example and example file have been removed
        with db.get_connection() as conn:

            res = conn.execute(db.module_version.select().where(db.module_version.c.id==module_version_pk))
            assert res.fetchone() is None

            res = conn.execute(db.sub_module.select().where(db.sub_module.c.id==example_pk))
            assert res.fetchone() is None

            res = conn.execute(db.example_file.select().where(db.example_file.c.id==example_file_pk))
            assert res.fetchone() is None

            # Ensure that the analytics has been removed
            analytics_res = conn.execute(
                db.analytics.select().where(
                    db.analytics.c.id==analytics_row_id
                )
            )
            assert analytics_res.fetchone() is None

    @pytest.mark.parametrize('delete_module_extracted', [
        True,
        False
    ])
    def test_delete_with_duplicate_module_versions(self, delete_module_extracted):
        """Test deletion of module version."""
        namespace = Namespace.get(name='test_delete_with_duplicate_module_versions', create=True)
        module = Module(namespace=namespace, name='test-module')
        module_provider = ModuleProvider.get(module=module, name='testprovider', create=True)

        try:
            existing_module_versions = [v.version for v in module_provider.get_versions()]
            assert '33.0.0' not in existing_module_versions

            # Create test module version to keep
            module_version_to_keep = ModuleVersion(module_provider=module_provider, version='33.0.0')
            module_version_to_keep.prepare_module()
            module_version_to_keep.update_attributes(extraction_complete=not delete_module_extracted)
            module_version_to_keep_pk = module_version_to_keep.pk
            module_version_to_keep_row = module_version_to_keep._get_db_row()

            module_version_to_delete = ModuleVersion(module_provider=module_provider, version='33.0.0')
            module_version_to_delete.prepare_module()
            module_version_to_delete.update_attributes(extraction_complete=delete_module_extracted)
            module_version_to_delete_pk = module_version_to_delete.pk

            module_version_to_delete.delete()

            test_module_version = ModuleVersion(module_provider=module_provider, version='33.0.0', pk=module_version_to_keep_pk)
            assert test_module_version._get_db_row() == module_version_to_keep_row

            test_module_version = ModuleVersion(module_provider=module_provider, version='33.0.0', pk=module_version_to_delete_pk)
            assert test_module_version._get_db_row() == None
        finally:
            module_provider.delete()
            namespace.delete()

    def test_get(self):
        """Test deletion of module version."""
        namespace = Namespace.get(name='test_delete_with_duplicate_module_versions', create=True)
        module = Module(namespace=namespace, name='test-module')
        module_provider = ModuleProvider.get(module=module, name='testprovider', create=True)

        try:
            existing_module_versions = [v.version for v in module_provider.get_versions()]
            assert '33.0.0' not in existing_module_versions

            # Create test module version to keep
            previous_version = ModuleVersion(module_provider=module_provider, version='33.0.0')
            previous_version.prepare_module()
            previous_version.update_attributes(extraction_complete=False)
            previous_pk = previous_version.pk

            test_version = ModuleVersion(module_provider=module_provider, version='33.0.0')
            test_version.prepare_module()

            test_version.update_attributes(extraction_complete=True)
            expected_id = test_version.pk
            expected_row = test_version._get_db_row()

            # Ensure previous and new version does not have the same pk
            assert previous_pk != expected_id

            post_version = ModuleVersion(module_provider=module_provider, version='33.0.0')
            post_version.prepare_module()
            post_version.update_attributes(extraction_complete=False)

            test_module_version = ModuleVersion.get(module_provider=module_provider, version='33.0.0')
            assert test_module_version.pk == expected_id
            assert test_module_version._get_db_row() == expected_row
        finally:
            module_provider.delete()
            namespace.delete()

    def test_variable_template(self):
        """Test variable template of module version."""

        module_provider = ModuleProvider.get(Module(Namespace('moduledetails'), 'withterraformdocs'), 'testprovider')
        module_version = ModuleVersion.get(module_provider, '1.5.0')

        # Ensure when the autogenerated usage builder is disabled, only the pre-defined
        # variables in the module version are returned
        with unittest.mock.patch('terrareg.config.Config.AUTOGENERATE_USAGE_BUILDER_VARIABLES', False):
            assert module_version.variable_template == [
                {
                    'additional_help': 'Provide the name of the application',
                    'name': 'name_of_application',
                    'quote_value': True,
                    'type': 'text',
                    'default_value': None,
                    'required': True
                }
            ]

        # Ensure when autogenerated usage builder is enabled, the missing required variables
        # are populated in the variable template
        with unittest.mock.patch('terrareg.config.Config.AUTOGENERATE_USAGE_BUILDER_VARIABLES', True):
            assert module_version.variable_template ==  [
                {
                    'additional_help': 'Provide the name of the application',
                    'name': 'name_of_application',
                    'quote_value': True,
                    'type': 'text',
                    'default_value': None,
                    'required': True
                },
                {
                    'additional_help': 'Override the default string',
                    'default_value': 'this is the default',
                    'name': 'string_with_default_value',
                    'quote_value': True,
                    'required': False,
                    'type': 'text'
                },
                {
                    'additional_help': 'Override the default string',
                    'default_value': None,
                    'name': 'undocumented_required_variable',
                    'quote_value': True,
                    'required': True,
                    'type': 'text'
                },
                {
                    'additional_help': 'required boolean variable',
                    'default_value': None,
                    'name': 'example_boolean_input',
                    'quote_value': False,
                    'required': True,
                    'type': 'boolean'
                },
                {
                    'additional_help': 'A required list',
                    'default_value': None,
                    'name': 'required_list_variable',
                    'quote_value': True,
                    'required': True,
                    'type': 'list'
                },
                {
                    'additional_help': 'Override the stringy list',
                    'default_value': ['value 1',
                                    'value 2'],
                    'name': 'example_list_input',
                    'quote_value': True,
                    'required': False,
                    'type': 'text'
                }
            ]

    @pytest.mark.parametrize('readme_content,expected_output', [
        # Test README with basic formatting
        (
            """
# Test terraform module

This is a terraform module to create a README example.

It performs the following:

 * Creates a README
 * Tests the README
 * Passes tests
""",
"""
<h1>Test terraform module</h1>
<p>This is a terraform module to create a README example.</p>
<p>It performs the following:</p>
<ul>
<li>Creates a README</li>
<li>Tests the README</li>
<li>Passes tests</li>
</ul>
"""
        ),
        # Test README with external module call
        (
            """
# Test external module

```
module "test-usage" {
  source  = "an-external-module/test"
  version = "1.0.0"

  some_variable = true
  another       = "value"
}
```
""",
"""
<h1>Test external module</h1>
<pre><code>module &quot;test-usage&quot; {
  source  = &quot;an-external-module/test&quot;
  version = &quot;1.0.0&quot;

  some_variable = true
  another       = &quot;value&quot;
}
</code></pre>
"""
        ),
        # Test README with call to root module
        (
            """
# Test external module

```
module "test-usage" {
  source  = "./"

  some_variable = true
  another       = "value"
}
```
""",
"""
<h1>Test external module</h1>
<pre><code>module &quot;test-usage&quot; {
  source  = &quot;example.com/moduledetails/readme-tests/provider&quot;
  version = &quot;&gt;= 1.0.0, &lt; 2.0.0&quot;

  some_variable = true
  another       = &quot;value&quot;
}
</code></pre>
"""
        ),

        # Test README with call outside of module root
        (
            """
# Test external module

```
module "test-usage" {
  source  = "../../"

  some_variable = true
  another       = "value"
}
```
""",
"""
<h1>Test external module</h1>
<pre><code>module &quot;test-usage&quot; {
  source  = &quot;example.com/moduledetails/readme-tests/provider&quot;
  version = &quot;&gt;= 1.0.0, &lt; 2.0.0&quot;

  some_variable = true
  another       = &quot;value&quot;
}
</code></pre>
"""
        ),

        # Test with call to submodule
        (
            """
# Test external module

```
module "test-usage" {
  source  = "./modules/testsubmodule"

  some_variable = true
  another       = "value"
}
```
""",
"""
<h1>Test external module</h1>
<pre><code>module &quot;test-usage&quot; {
  source  = &quot;example.com/moduledetails/readme-tests/provider//modules/testsubmodule&quot;
  version = &quot;&gt;= 1.0.0, &lt; 2.0.0&quot;

  some_variable = true
  another       = &quot;value&quot;
}
</code></pre>
"""
        ),

        # Test with call to submodule, with content from an example file in examples directory
        (
            """
# Test external module

```
module "test-usage" {
  source = "../../modules/testsubmodule"

  some_variable = true
  another       = "value"
}
```
""",
"""
<h1>Test external module</h1>
<pre><code>module &quot;test-usage&quot; {
  source  = &quot;example.com/moduledetails/readme-tests/provider//modules/testsubmodule&quot;
  version = &quot;&gt;= 1.0.0, &lt; 2.0.0&quot;

  some_variable = true
  another       = &quot;value&quot;
}
</code></pre>
"""
        ),

        # Test README with multiple modules
        (
            """
# Test external module

```
module "test-usage1" {
  source = "./"

  some_variable = true
  another       = "value"
}
module "test-usage2" {
  source = "./modules/testsubmodule"

  some_variable = true
  another       = "value"
}
module "test-external-call" {
  source  = "external-module"
  version = "1.0.3"
}
```
""",
"""
<h1>Test external module</h1>
<pre><code>module &quot;test-usage1&quot; {
  source  = &quot;example.com/moduledetails/readme-tests/provider&quot;
  version = &quot;&gt;= 1.0.0, &lt; 2.0.0&quot;

  some_variable = true
  another       = &quot;value&quot;
}
module &quot;test-usage2&quot; {
  source  = &quot;example.com/moduledetails/readme-tests/provider//modules/testsubmodule&quot;
  version = &quot;&gt;= 1.0.0, &lt; 2.0.0&quot;

  some_variable = true
  another       = &quot;value&quot;
}
module &quot;test-external-call&quot; {
  source  = &quot;external-module&quot;
  version = &quot;1.0.3&quot;
}
</code></pre>
"""
        ),

        # Test module call with different indentation
        (
            """
# Test external module

```
module "test-usage1" {
  source        = "./"
  some_variable = true
  another       = "value"
}
module "test-usage2" {
    source = "./modules/testsubmodule"
}
module "test-usage3" {
          source =         "./modules/testsubmodule"
}
```
""",
"""
<h1>Test external module</h1>
<pre><code>module &quot;test-usage1&quot; {
  source        = &quot;example.com/moduledetails/readme-tests/provider&quot;
  version       = &quot;&gt;= 1.0.0, &lt; 2.0.0&quot;
  some_variable = true
  another       = &quot;value&quot;
}
module &quot;test-usage2&quot; {
    source  = &quot;example.com/moduledetails/readme-tests/provider//modules/testsubmodule&quot;
    version = &quot;&gt;= 1.0.0, &lt; 2.0.0&quot;
}
module &quot;test-usage3&quot; {
          source  = &quot;example.com/moduledetails/readme-tests/provider//modules/testsubmodule&quot;
          version = &quot;&gt;= 1.0.0, &lt; 2.0.0&quot;
}
</code></pre>
"""
        ),
    ])
    def test_get_readme_html(self, readme_content, expected_output):
        """Test get_readme_html method, ensuring it replaces example source and converts from markdown to HTML."""

        with unittest.mock.patch('terrareg.config.Config.TERRAFORM_EXAMPLE_VERSION_TEMPLATE', '>= {major}.{minor}.{patch}, < {major_plus_one}.0.0'):
            module_version = ModuleVersion(ModuleProvider(Module(Namespace('moduledetails'), 'readme-tests'), 'provider'), '1.0.0')
            # Set README in module version
            module_version.module_details.update_attributes(readme_content=readme_content)

            assert module_version.get_readme_html(server_hostname='example.com').strip() == expected_output.strip()

    def test_git_path(self):
        """Test git_path property"""
        # Ensure the git_path from the module provider is returned
        with unittest.mock.patch('terrareg.models.ModuleProvider.git_path', 'unittest-git-path'):
            module_provider = ModuleProvider.get(Module(Namespace('moduledetails'), 'git-path'), 'provider')
            module_version = ModuleVersion.get(module_provider, '1.0.0')
            assert module_version.git_path == 'unittest-git-path'


    @pytest.mark.parametrize('module_name,module_version,git_path,path,expected_browse_url', [
        # Test no browse URL in any configuration
        ('no-git-provider', '1.0.0', None, None, None),
        ('no-git-provider', '1.0.0', None, 'unittestpath', None),
        # Test browse URL only configured in module version
        ('no-git-provider', '1.4.0', None, None, 'https://mv-browse-url.com/repo_url_tests/no-git-provider-test/browse/1.4.0/suffix'),
        ('no-git-provider', '1.4.0', None, 'unittestpath', 'https://mv-browse-url.com/repo_url_tests/no-git-provider-test/browse/1.4.0/unittestpathsuffix'),

        # Test URI encoded tags in browse URL template
        # - from module provider
        ('no-git-provider-uri-encoded', '1.4.0', None, 'testpath', 'https://mv-browse-url.com/repo_url_tests/no-git-provider-uri-encoded-test/browse/testpath?at=release%40test%2F1.4.0%2F'),
        # - from git provider
        ('git-provider-uri-encoded', '1.4.0', None, 'testpath', 'https://browse-url.com/repo_url_tests/git-provider-uri-encoded-test/browse/testpath?at=release%40test%2F1.4.0%2F'),

        # Test with git provider configured in module provider
        ('git-provider-urls', '1.1.0', None, None, 'https://browse-url.com/repo_url_tests/git-provider-urls-test/browse/1.1.0/suffix'),
        ('git-provider-urls', '1.1.0', None, 'unittestpath', 'https://browse-url.com/repo_url_tests/git-provider-urls-test/browse/1.1.0/unittestpathsuffix'),
        # Test with git provider configured in module provider and override in module version
        ('git-provider-urls', '1.4.0', None, None, 'https://mv-browse-url.com/repo_url_tests/git-provider-urls-test/browse/1.4.0/suffix'),
        ('git-provider-urls', '1.4.0', None, 'unittestpath', 'https://mv-browse-url.com/repo_url_tests/git-provider-urls-test/browse/1.4.0/unittestpathsuffix'),

        # Test with git provider configured in module provider
        ('module-provider-urls', '1.2.0', None, None, 'https://mp-browse-url.com/repo_url_tests/module-provider-urls-test/browse/1.2.0/suffix'),
        ('module-provider-urls', '1.2.0', None, 'unittestpath', 'https://mp-browse-url.com/repo_url_tests/module-provider-urls-test/browse/1.2.0/unittestpathsuffix'),
        # Test with URls configured in module provider and override in module version
        ('module-provider-urls', '1.4.0', None, None, 'https://mv-browse-url.com/repo_url_tests/module-provider-urls-test/browse/1.4.0/suffix'),
        ('module-provider-urls', '1.4.0', None, 'unittestpath', 'https://mv-browse-url.com/repo_url_tests/module-provider-urls-test/browse/1.4.0/unittestpathsuffix'),

        # Test with git provider configured in module provider and URLs configured in git provider
        ('module-provider-override-git-provider', '1.3.0', None, None, 'https://mp-browse-url.com/repo_url_tests/module-provider-override-git-provider-test/browse/1.3.0/suffix'),
        ('module-provider-override-git-provider', '1.3.0', None, 'unittestpath', 'https://mp-browse-url.com/repo_url_tests/module-provider-override-git-provider-test/browse/1.3.0/unittestpathsuffix'),
        # Test with URls configured in module provider and override in module version
        ('module-provider-override-git-provider', '1.4.0', None, None, 'https://mv-browse-url.com/repo_url_tests/module-provider-override-git-provider-test/browse/1.4.0/suffix'),
        ('module-provider-override-git-provider', '1.4.0', None, 'unittestpath', 'https://mv-browse-url.com/repo_url_tests/module-provider-override-git-provider-test/browse/1.4.0/unittestpathsuffix'),

        ## Test with git_url set in module provider
        # Test no browse URL in any configuration
        ('no-git-provider', '1.0.0', 'module/sub/directory', None, None),
        ('no-git-provider', '1.0.0', 'module/sub/directory', 'unittestpath', None),
        # Test browse URL only configured in module version
        ('no-git-provider', '1.4.0', 'module/sub/directory', None, 'https://mv-browse-url.com/repo_url_tests/no-git-provider-test/browse/1.4.0/module/sub/directorysuffix'),
        ('no-git-provider', '1.4.0', 'module/sub/directory', 'unittestpath', 'https://mv-browse-url.com/repo_url_tests/no-git-provider-test/browse/1.4.0/module/sub/directory/unittestpathsuffix'),

        # Test URI encoded tags in browse URL template
        # - from module provider
        ('no-git-provider-uri-encoded', '1.4.0', 'module/sub/directory', 'testpath', 'https://mv-browse-url.com/repo_url_tests/no-git-provider-uri-encoded-test/browse/module/sub/directory/testpath?at=release%40test%2F1.4.0%2F'),
        # - from git provider
        ('git-provider-uri-encoded', '1.4.0', 'module/sub/directory', 'testpath', 'https://browse-url.com/repo_url_tests/git-provider-uri-encoded-test/browse/module/sub/directory/testpath?at=release%40test%2F1.4.0%2F'),

        # Test with git provider configured in module provider
        ('git-provider-urls', '1.1.0', 'module/sub/directory', None, 'https://browse-url.com/repo_url_tests/git-provider-urls-test/browse/1.1.0/module/sub/directorysuffix'),
        ('git-provider-urls', '1.1.0', 'module/sub/directory', 'unittestpath', 'https://browse-url.com/repo_url_tests/git-provider-urls-test/browse/1.1.0/module/sub/directory/unittestpathsuffix'),
        # Test with git provider configured in module provider and override in module version
        ('git-provider-urls', '1.4.0', 'module/sub/directory', None, 'https://mv-browse-url.com/repo_url_tests/git-provider-urls-test/browse/1.4.0/module/sub/directorysuffix'),
        ('git-provider-urls', '1.4.0', 'module/sub/directory', 'unittestpath', 'https://mv-browse-url.com/repo_url_tests/git-provider-urls-test/browse/1.4.0/module/sub/directory/unittestpathsuffix'),

        # Test with git provider configured in module provider
        ('module-provider-urls', '1.2.0', 'module/sub/directory', None, 'https://mp-browse-url.com/repo_url_tests/module-provider-urls-test/browse/1.2.0/module/sub/directorysuffix'),
        ('module-provider-urls', '1.2.0', 'module/sub/directory', 'unittestpath', 'https://mp-browse-url.com/repo_url_tests/module-provider-urls-test/browse/1.2.0/module/sub/directory/unittestpathsuffix'),
        # Test with URls configured in module provider and override in module version
        ('module-provider-urls', '1.4.0', 'module/sub/directory', None, 'https://mv-browse-url.com/repo_url_tests/module-provider-urls-test/browse/1.4.0/module/sub/directorysuffix'),
        ('module-provider-urls', '1.4.0', 'module/sub/directory', 'unittestpath', 'https://mv-browse-url.com/repo_url_tests/module-provider-urls-test/browse/1.4.0/module/sub/directory/unittestpathsuffix'),

        # Test with git provider configured in module provider and URLs configured in git provider
        ('module-provider-override-git-provider', '1.3.0', 'module/sub/directory', None, 'https://mp-browse-url.com/repo_url_tests/module-provider-override-git-provider-test/browse/1.3.0/module/sub/directorysuffix'),
        ('module-provider-override-git-provider', '1.3.0', 'module/sub/directory', 'unittestpath', 'https://mp-browse-url.com/repo_url_tests/module-provider-override-git-provider-test/browse/1.3.0/module/sub/directory/unittestpathsuffix'),
        # Test with URls configured in module provider and override in module version
        ('module-provider-override-git-provider', '1.4.0', 'module/sub/directory', None, 'https://mv-browse-url.com/repo_url_tests/module-provider-override-git-provider-test/browse/1.4.0/module/sub/directorysuffix'),
        ('module-provider-override-git-provider', '1.4.0', 'module/sub/directory', 'unittestpath', 'https://mv-browse-url.com/repo_url_tests/module-provider-override-git-provider-test/browse/1.4.0/module/sub/directory/unittestpathsuffix')
    ])
    def test_get_source_browse_url(self, module_name, module_version, git_path, path, expected_browse_url):
        """Ensure browse URL matches the expected values."""
        namespace = Namespace(name='repo_url_tests')
        module = Module(namespace=namespace, name=module_name)
        module_provider = ModuleProvider.get(module=module, name='test')
        assert module_provider is not None

        module_provider.update_git_path(git_path)

        try:
            module_version = ModuleVersion.get(module_provider=module_provider, version=module_version)
            assert module_version is not None

            kwargs = {'path': path} if path else {}
            assert module_version.get_source_browse_url(**kwargs) == expected_browse_url

        finally:
            module_provider.update_git_path(None)

    @pytest.mark.parametrize('module_name,module_version,path,expected_browse_url', [
        # Test no browse URL in any configuration
        ('no-git-provider', '1.0.0', None, None),
        ('no-git-provider', '1.0.0', 'unittestpath', None),
        # Test browse URL only configured in module version
        ('no-git-provider', '1.4.0', None, None),
        ('no-git-provider', '1.4.0', 'unittestpath', None),

        # Test with git provider configured in module provider
        ('git-provider-urls', '1.1.0', None, 'https://browse-url.com/repo_url_tests/git-provider-urls-test/browse/1.1.0/suffix'),
        ('git-provider-urls', '1.1.0', 'unittestpath', 'https://browse-url.com/repo_url_tests/git-provider-urls-test/browse/1.1.0/unittestpathsuffix'),
        # Test with git provider configured in module provider and override in module version
        ('git-provider-urls', '1.4.0', None, 'https://browse-url.com/repo_url_tests/git-provider-urls-test/browse/1.4.0/suffix'),
        ('git-provider-urls', '1.4.0', 'unittestpath', 'https://browse-url.com/repo_url_tests/git-provider-urls-test/browse/1.4.0/unittestpathsuffix'),

        # Test with repo URLs configured in module provider
        ('module-provider-urls', '1.2.0', None, 'https://mp-browse-url.com/repo_url_tests/module-provider-urls-test/browse/1.2.0/suffix'),
        ('module-provider-urls', '1.2.0', 'unittestpath', 'https://mp-browse-url.com/repo_url_tests/module-provider-urls-test/browse/1.2.0/unittestpathsuffix'),
        # Test with repo URls configured in module provider and override in module version
        ('module-provider-urls', '1.4.0', None, 'https://mp-browse-url.com/repo_url_tests/module-provider-urls-test/browse/1.4.0/suffix'),
        ('module-provider-urls', '1.4.0', 'unittestpath', 'https://mp-browse-url.com/repo_url_tests/module-provider-urls-test/browse/1.4.0/unittestpathsuffix'),

        # Test with git provider configured in module provider and URLs configured in git provider
        ('module-provider-override-git-provider', '1.3.0', None, 'https://mp-browse-url.com/repo_url_tests/module-provider-override-git-provider-test/browse/1.3.0/suffix'),
        ('module-provider-override-git-provider', '1.3.0', 'unittestpath', 'https://mp-browse-url.com/repo_url_tests/module-provider-override-git-provider-test/browse/1.3.0/unittestpathsuffix'),
        # Test with URls configured in module provider and override in module version
        ('module-provider-override-git-provider', '1.4.0', None, 'https://mp-browse-url.com/repo_url_tests/module-provider-override-git-provider-test/browse/1.4.0/suffix'),
        ('module-provider-override-git-provider', '1.4.0', 'unittestpath', 'https://mp-browse-url.com/repo_url_tests/module-provider-override-git-provider-test/browse/1.4.0/unittestpathsuffix')
    ])
    def test_get_source_browse_url_with_custom_module_version_urls_disabled(self, module_name, module_version, path, expected_browse_url):
        """Ensure browse URL matches the expected values when module version repo urls are disabled."""
        namespace = Namespace(name='repo_url_tests')
        module = Module(namespace=namespace, name=module_name)
        module_provider = ModuleProvider.get(module=module, name='test')
        assert module_provider is not None
        module_version = ModuleVersion.get(module_provider=module_provider, version=module_version)
        assert module_version is not None

        kwargs = {'path': path} if path else {}
        with unittest.mock.patch('terrareg.config.Config.ALLOW_CUSTOM_GIT_URL_MODULE_VERSION', False):
            assert module_version.get_source_browse_url(**kwargs) == expected_browse_url

    @pytest.mark.parametrize('module_name,module_version,path,expected_browse_url', [
        # Test no browse URL in any configuration
        ('no-git-provider', '1.0.0', None, None),
        ('no-git-provider', '1.0.0', 'unittestpath', None),
        # Test browse URL only configured in module version
        ('no-git-provider', '1.4.0', None, 'https://mv-browse-url.com/repo_url_tests/no-git-provider-test/browse/1.4.0/suffix'),
        ('no-git-provider', '1.4.0', 'unittestpath', 'https://mv-browse-url.com/repo_url_tests/no-git-provider-test/browse/1.4.0/unittestpathsuffix'),

        # Test with git provider configured in module provider
        ('git-provider-urls', '1.1.0', None, 'https://browse-url.com/repo_url_tests/git-provider-urls-test/browse/1.1.0/suffix'),
        ('git-provider-urls', '1.1.0', 'unittestpath', 'https://browse-url.com/repo_url_tests/git-provider-urls-test/browse/1.1.0/unittestpathsuffix'),
        # Test with git provider configured in module provider and override in module version
        ('git-provider-urls', '1.4.0', None, 'https://mv-browse-url.com/repo_url_tests/git-provider-urls-test/browse/1.4.0/suffix'),
        ('git-provider-urls', '1.4.0', 'unittestpath', 'https://mv-browse-url.com/repo_url_tests/git-provider-urls-test/browse/1.4.0/unittestpathsuffix'),

        # Test with repo URLs configured in module provider
        ('module-provider-urls', '1.2.0', None, None),
        ('module-provider-urls', '1.2.0', 'unittestpath', None),
        # Test with repo URls configured in module provider and override in module version
        ('module-provider-urls', '1.4.0', None, 'https://mv-browse-url.com/repo_url_tests/module-provider-urls-test/browse/1.4.0/suffix'),
        ('module-provider-urls', '1.4.0', 'unittestpath', 'https://mv-browse-url.com/repo_url_tests/module-provider-urls-test/browse/1.4.0/unittestpathsuffix'),

        # Test with git provider configured in module provider and URLs configured in git provider
        ('module-provider-override-git-provider', '1.3.0', None, 'https://browse-url.com/repo_url_tests/module-provider-override-git-provider-test/browse/1.3.0/suffix'),
        ('module-provider-override-git-provider', '1.3.0', 'unittestpath', 'https://browse-url.com/repo_url_tests/module-provider-override-git-provider-test/browse/1.3.0/unittestpathsuffix'),
        # Test with URls configured in module provider and override in module version
        ('module-provider-override-git-provider', '1.4.0', None, 'https://mv-browse-url.com/repo_url_tests/module-provider-override-git-provider-test/browse/1.4.0/suffix'),
        ('module-provider-override-git-provider', '1.4.0', 'unittestpath', 'https://mv-browse-url.com/repo_url_tests/module-provider-override-git-provider-test/browse/1.4.0/unittestpathsuffix')
    ])
    def test_get_source_browse_url_with_custom_module_provider_urls_disabled(self, module_name, module_version, path, expected_browse_url):
        """Ensure browse URL matches the expected values when module provider repo urls are disabled."""
        namespace = Namespace(name='repo_url_tests')
        module = Module(namespace=namespace, name=module_name)
        module_provider = ModuleProvider.get(module=module, name='test')
        assert module_provider is not None
        module_version = ModuleVersion.get(module_provider=module_provider, version=module_version)
        assert module_version is not None

        kwargs = {'path': path} if path else {}
        with unittest.mock.patch('terrareg.config.Config.ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER', False):
            assert module_version.get_source_browse_url(**kwargs) == expected_browse_url

    @pytest.mark.parametrize('module_name,module_version,path,expected_browse_url', [
        # Test no browse URL in any configuration
        ('no-git-provider', '1.0.0', None, None),
        ('no-git-provider', '1.0.0', 'unittestpath', None),
        # Test browse URL only configured in module version
        ('no-git-provider', '1.4.0', None, None),
        ('no-git-provider', '1.4.0', 'unittestpath', None),

        # Test with git provider configured in module provider
        ('git-provider-urls', '1.1.0', None, 'https://browse-url.com/repo_url_tests/git-provider-urls-test/browse/1.1.0/suffix'),
        ('git-provider-urls', '1.1.0', 'unittestpath', 'https://browse-url.com/repo_url_tests/git-provider-urls-test/browse/1.1.0/unittestpathsuffix'),
        # Test with git provider configured in module provider and override in module version
        ('git-provider-urls', '1.4.0', None, 'https://browse-url.com/repo_url_tests/git-provider-urls-test/browse/1.4.0/suffix'),
        ('git-provider-urls', '1.4.0', 'unittestpath', 'https://browse-url.com/repo_url_tests/git-provider-urls-test/browse/1.4.0/unittestpathsuffix'),

        # Test with repo URLs configured in module provider
        ('module-provider-urls', '1.2.0', None, None),
        ('module-provider-urls', '1.2.0', 'unittestpath', None),
        # Test with repo URls configured in module provider and override in module version
        ('module-provider-urls', '1.4.0', None, None),
        ('module-provider-urls', '1.4.0', 'unittestpath', None),

        # Test with git provider configured in module provider and URLs configured in git provider
        ('module-provider-override-git-provider', '1.3.0', None, 'https://browse-url.com/repo_url_tests/module-provider-override-git-provider-test/browse/1.3.0/suffix'),
        ('module-provider-override-git-provider', '1.3.0', 'unittestpath', 'https://browse-url.com/repo_url_tests/module-provider-override-git-provider-test/browse/1.3.0/unittestpathsuffix'),
        # Test with URls configured in module provider and override in module version
        ('module-provider-override-git-provider', '1.4.0', None, 'https://browse-url.com/repo_url_tests/module-provider-override-git-provider-test/browse/1.4.0/suffix'),
        ('module-provider-override-git-provider', '1.4.0', 'unittestpath', 'https://browse-url.com/repo_url_tests/module-provider-override-git-provider-test/browse/1.4.0/unittestpathsuffix')
    ])
    def test_get_source_browse_url_with_custom_module_provider_and_module_version_urls_disabled(self, module_name, module_version, path, expected_browse_url):
        """Ensure browse URL matches the expected values when module provider and module version repo urls are disabled."""
        namespace = Namespace(name='repo_url_tests')
        module = Module(namespace=namespace, name=module_name)
        module_provider = ModuleProvider.get(module=module, name='test')
        assert module_provider is not None
        module_version = ModuleVersion.get(module_provider=module_provider, version=module_version)
        assert module_version is not None

        kwargs = {'path': path} if path else {}
        with unittest.mock.patch('terrareg.config.Config.ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER', False):
            with unittest.mock.patch('terrareg.config.Config.ALLOW_CUSTOM_GIT_URL_MODULE_VERSION', False):
                assert module_version.get_source_browse_url(**kwargs) == expected_browse_url

    @pytest.mark.parametrize('module_name,module_version,expected_clone_url', [
        # Test no clone URL in any configuration
        ('no-git-provider', '1.0.0', None),
        # Test clone URL only configured in module version
        ('no-git-provider', '1.4.0', 'ssh://mv-clone-url.com/repo_url_tests/no-git-provider-test'),

        # Test with git provider configured in module provider
        ('git-provider-urls', '1.1.0', 'ssh://clone-url.com/repo_url_tests/git-provider-urls-test'),
        # Test with git provider configured in module provider and override in module version
        ('git-provider-urls', '1.4.0', 'ssh://mv-clone-url.com/repo_url_tests/git-provider-urls-test'),

        # Test with git provider configured in module provider
        ('module-provider-urls', '1.2.0', 'ssh://mp-clone-url.com/repo_url_tests/module-provider-urls-test'),
        # Test with URls configured in module provider and override in module version
        ('module-provider-urls', '1.4.0', 'ssh://mv-clone-url.com/repo_url_tests/module-provider-urls-test'),

        # Test with git provider configured in module provider and URLs configured in git provider
        ('module-provider-override-git-provider', '1.3.0', 'ssh://mp-clone-url.com/repo_url_tests/module-provider-override-git-provider-test'),
        # Test with URls configured in module provider and override in module version
        ('module-provider-override-git-provider', '1.4.0', 'ssh://mv-clone-url.com/repo_url_tests/module-provider-override-git-provider-test'),
    ])
    def test_get_source_clone_url(self, module_name, module_version, expected_clone_url):
        """Ensure clone URL matches the expected values."""
        namespace = Namespace(name='repo_url_tests')
        module = Module(namespace=namespace, name=module_name)
        module_provider = ModuleProvider.get(module=module, name='test')
        assert module_provider is not None
        module_version = ModuleVersion.get(module_provider=module_provider, version=module_version)
        assert module_version is not None

        assert module_version.get_git_clone_url() == expected_clone_url

    @pytest.mark.parametrize('module_name,module_version,expected_clone_url', [
        # Test no clone URL in any configuration
        ('no-git-provider', '1.0.0', None),
        # Test clone URL only configured in module version
        ('no-git-provider', '1.4.0', None),

        # Test with git provider configured in module provider
        ('git-provider-urls', '1.1.0', 'ssh://clone-url.com/repo_url_tests/git-provider-urls-test'),
        # Test with git provider configured in module provider and override in module version
        ('git-provider-urls', '1.4.0', 'ssh://clone-url.com/repo_url_tests/git-provider-urls-test'),

        # Test with repo URLs configured in module provider
        ('module-provider-urls', '1.2.0', 'ssh://mp-clone-url.com/repo_url_tests/module-provider-urls-test'),
        # Test with repo URls configured in module provider and override in module version
        ('module-provider-urls', '1.4.0', 'ssh://mp-clone-url.com/repo_url_tests/module-provider-urls-test'),

        # Test with git provider configured in module provider and URLs configured in git provider
        ('module-provider-override-git-provider', '1.3.0', 'ssh://mp-clone-url.com/repo_url_tests/module-provider-override-git-provider-test'),
        # Test with URls configured in module provider and override in module version
        ('module-provider-override-git-provider', '1.4.0', 'ssh://mp-clone-url.com/repo_url_tests/module-provider-override-git-provider-test'),
    ])
    def test_get_source_clone_url_with_custom_module_version_urls_disabled(self, module_name, module_version, expected_clone_url):
        """Ensure clone URL matches the expected values when module version repo urls are disabled."""
        namespace = Namespace(name='repo_url_tests')
        module = Module(namespace=namespace, name=module_name)
        module_provider = ModuleProvider.get(module=module, name='test')
        assert module_provider is not None
        module_version = ModuleVersion.get(module_provider=module_provider, version=module_version)
        assert module_version is not None

        with unittest.mock.patch('terrareg.config.Config.ALLOW_CUSTOM_GIT_URL_MODULE_VERSION', False):
            assert module_version.get_git_clone_url() == expected_clone_url

    @pytest.mark.parametrize('module_name,module_version,expected_clone_url', [
        # Test no clone URL in any configuration
        ('no-git-provider', '1.0.0', None),
        # Test clone URL only configured in module version
        ('no-git-provider', '1.4.0', 'ssh://mv-clone-url.com/repo_url_tests/no-git-provider-test'),

        # Test with git provider configured in module provider
        ('git-provider-urls', '1.1.0', 'ssh://clone-url.com/repo_url_tests/git-provider-urls-test'),
        # Test with git provider configured in module provider and override in module version
        ('git-provider-urls', '1.4.0', 'ssh://mv-clone-url.com/repo_url_tests/git-provider-urls-test'),

        # Test with repo URLs configured in module provider
        ('module-provider-urls', '1.2.0', None),
        # Test with repo URls configured in module provider and override in module version
        ('module-provider-urls', '1.4.0', 'ssh://mv-clone-url.com/repo_url_tests/module-provider-urls-test'),

        # Test with git provider configured in module provider and URLs configured in git provider
        ('module-provider-override-git-provider', '1.3.0', 'ssh://clone-url.com/repo_url_tests/module-provider-override-git-provider-test'),
        # Test with URls configured in module provider and override in module version
        ('module-provider-override-git-provider', '1.4.0', 'ssh://mv-clone-url.com/repo_url_tests/module-provider-override-git-provider-test'),
    ])
    def test_get_source_clone_url_with_custom_module_provider_urls_disabled(self, module_name, module_version, expected_clone_url):
        """Ensure clone URL matches the expected values when module provider repo urls are disabled."""
        namespace = Namespace(name='repo_url_tests')
        module = Module(namespace=namespace, name=module_name)
        module_provider = ModuleProvider.get(module=module, name='test')
        assert module_provider is not None
        module_version = ModuleVersion.get(module_provider=module_provider, version=module_version)
        assert module_version is not None

        with unittest.mock.patch('terrareg.config.Config.ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER', False):
            assert module_version.get_git_clone_url() == expected_clone_url

    @pytest.mark.parametrize('module_name,module_version,expected_clone_url', [
        # Test no clone URL in any configuration
        ('no-git-provider', '1.0.0', None),
        # Test clone URL only configured in module version
        ('no-git-provider', '1.4.0', None),

        # Test with git provider configured in module provider
        ('git-provider-urls', '1.1.0', 'ssh://clone-url.com/repo_url_tests/git-provider-urls-test'),
        # Test with git provider configured in module provider and override in module version
        ('git-provider-urls', '1.4.0', 'ssh://clone-url.com/repo_url_tests/git-provider-urls-test'),

        # Test with repo URLs configured in module provider
        ('module-provider-urls', '1.2.0', None),
        # Test with repo URls configured in module provider and override in module version
        ('module-provider-urls', '1.4.0', None),

        # Test with git provider configured in module provider and URLs configured in git provider
        ('module-provider-override-git-provider', '1.3.0', 'ssh://clone-url.com/repo_url_tests/module-provider-override-git-provider-test'),
        # Test with URls configured in module provider and override in module version
        ('module-provider-override-git-provider', '1.4.0', 'ssh://clone-url.com/repo_url_tests/module-provider-override-git-provider-test'),
    ])
    def test_get_source_clone_url_with_custom_module_provider_and_module_version_urls_disabled(self, module_name, module_version, expected_clone_url):
        """Ensure clone URL matches the expected values when module provider and module version repo urls are disabled."""
        namespace = Namespace(name='repo_url_tests')
        module = Module(namespace=namespace, name=module_name)
        module_provider = ModuleProvider.get(module=module, name='test')
        assert module_provider is not None
        module_version = ModuleVersion.get(module_provider=module_provider, version=module_version)
        assert module_version is not None

        with unittest.mock.patch('terrareg.config.Config.ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER', False):
            with unittest.mock.patch('terrareg.config.Config.ALLOW_CUSTOM_GIT_URL_MODULE_VERSION', False):
                assert module_version.get_git_clone_url() == expected_clone_url

    @pytest.mark.parametrize('module_name,module_version,expected_base_url', [
        # Test no base URL in any configuration
        ('no-git-provider', '1.0.0', None),
        # Test base URL only configured in module version
        ('no-git-provider', '1.4.0', 'https://mv-base-url.com/repo_url_tests/no-git-provider-test'),

        # Test with git provider configured in module provider
        ('git-provider-urls', '1.1.0', 'https://base-url.com/repo_url_tests/git-provider-urls-test'),
        # Test with git provider configured in module provider and override in module version
        ('git-provider-urls', '1.4.0', 'https://mv-base-url.com/repo_url_tests/git-provider-urls-test'),

        # Test with git provider configured in module provider
        ('module-provider-urls', '1.2.0', 'https://mp-base-url.com/repo_url_tests/module-provider-urls-test'),
        # Test with URls configured in module provider and override in module version
        ('module-provider-urls', '1.4.0', 'https://mv-base-url.com/repo_url_tests/module-provider-urls-test'),

        # Test with git provider configured in module provider and URLs configured in git provider
        ('module-provider-override-git-provider', '1.3.0', 'https://mp-base-url.com/repo_url_tests/module-provider-override-git-provider-test'),
        # Test with URls configured in module provider and override in module version
        ('module-provider-override-git-provider', '1.4.0', 'https://mv-base-url.com/repo_url_tests/module-provider-override-git-provider-test'),
    ])
    def test_get_source_base_url(self, module_name, module_version, expected_base_url):
        """Ensure base URL matches the expected values."""
        namespace = Namespace(name='repo_url_tests')
        module = Module(namespace=namespace, name=module_name)
        module_provider = ModuleProvider.get(module=module, name='test')
        assert module_provider is not None
        module_version = ModuleVersion.get(module_provider=module_provider, version=module_version)
        assert module_version is not None

        assert module_version.get_source_base_url() == expected_base_url

    @pytest.mark.parametrize('module_name,module_version,expected_base_url', [
        # Test no base URL in any configuration
        ('no-git-provider', '1.0.0', None),
        # Test base URL only configured in module version
        ('no-git-provider', '1.4.0', None),

        # Test with git provider configured in module provider
        ('git-provider-urls', '1.1.0', 'https://base-url.com/repo_url_tests/git-provider-urls-test'),
        # Test with git provider configured in module provider and override in module version
        ('git-provider-urls', '1.4.0', 'https://base-url.com/repo_url_tests/git-provider-urls-test'),

        # Test with repo URLs configured in module provider
        ('module-provider-urls', '1.2.0', 'https://mp-base-url.com/repo_url_tests/module-provider-urls-test'),
        # Test with repo URls configured in module provider and override in module version
        ('module-provider-urls', '1.4.0', 'https://mp-base-url.com/repo_url_tests/module-provider-urls-test'),

        # Test with git provider configured in module provider and URLs configured in git provider
        ('module-provider-override-git-provider', '1.3.0', 'https://mp-base-url.com/repo_url_tests/module-provider-override-git-provider-test'),
        # Test with URls configured in module provider and override in module version
        ('module-provider-override-git-provider', '1.4.0', 'https://mp-base-url.com/repo_url_tests/module-provider-override-git-provider-test'),
    ])
    def test_get_source_base_url_with_custom_module_version_urls_disabled(self, module_name, module_version, expected_base_url):
        """Ensure base URL matches the expected values when module version repo urls are disabled."""
        namespace = Namespace(name='repo_url_tests')
        module = Module(namespace=namespace, name=module_name)
        module_provider = ModuleProvider.get(module=module, name='test')
        assert module_provider is not None
        module_version = ModuleVersion.get(module_provider=module_provider, version=module_version)
        assert module_version is not None

        with unittest.mock.patch('terrareg.config.Config.ALLOW_CUSTOM_GIT_URL_MODULE_VERSION', False):
            assert module_version.get_source_base_url() == expected_base_url

    @pytest.mark.parametrize('module_name,module_version,expected_base_url', [
        # Test no base URL in any configuration
        ('no-git-provider', '1.0.0', None),
        # Test base URL only configured in module version
        ('no-git-provider', '1.4.0', 'https://mv-base-url.com/repo_url_tests/no-git-provider-test'),

        # Test with git provider configured in module provider
        ('git-provider-urls', '1.1.0', 'https://base-url.com/repo_url_tests/git-provider-urls-test'),
        # Test with git provider configured in module provider and override in module version
        ('git-provider-urls', '1.4.0', 'https://mv-base-url.com/repo_url_tests/git-provider-urls-test'),

        # Test with repo URLs configured in module provider
        ('module-provider-urls', '1.2.0', None),
        # Test with repo URls configured in module provider and override in module version
        ('module-provider-urls', '1.4.0', 'https://mv-base-url.com/repo_url_tests/module-provider-urls-test'),

        # Test with git provider configured in module provider and URLs configured in git provider
        ('module-provider-override-git-provider', '1.3.0', 'https://base-url.com/repo_url_tests/module-provider-override-git-provider-test'),
        # Test with URls configured in module provider and override in module version
        ('module-provider-override-git-provider', '1.4.0', 'https://mv-base-url.com/repo_url_tests/module-provider-override-git-provider-test'),
    ])
    def test_get_source_base_url_with_custom_module_provider_urls_disabled(self, module_name, module_version, expected_base_url):
        """Ensure base URL matches the expected values when module provider repo urls are disabled."""
        namespace = Namespace(name='repo_url_tests')
        module = Module(namespace=namespace, name=module_name)
        module_provider = ModuleProvider.get(module=module, name='test')
        assert module_provider is not None
        module_version = ModuleVersion.get(module_provider=module_provider, version=module_version)
        assert module_version is not None

        with unittest.mock.patch('terrareg.config.Config.ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER', False):
            assert module_version.get_source_base_url() == expected_base_url

    @pytest.mark.parametrize('module_name,module_version,expected_base_url', [
        # Test no base URL in any configuration
        ('no-git-provider', '1.0.0', None),
        # Test base URL only configured in module version
        ('no-git-provider', '1.4.0', None),

        # Test with git provider configured in module provider
        ('git-provider-urls', '1.1.0', 'https://base-url.com/repo_url_tests/git-provider-urls-test'),
        # Test with git provider configured in module provider and override in module version
        ('git-provider-urls', '1.4.0', 'https://base-url.com/repo_url_tests/git-provider-urls-test'),

        # Test with repo URLs configured in module provider
        ('module-provider-urls', '1.2.0', None),
        # Test with repo URls configured in module provider and override in module version
        ('module-provider-urls', '1.4.0', None),

        # Test with git provider configured in module provider and URLs configured in git provider
        ('module-provider-override-git-provider', '1.3.0', 'https://base-url.com/repo_url_tests/module-provider-override-git-provider-test'),
        # Test with URls configured in module provider and override in module version
        ('module-provider-override-git-provider', '1.4.0', 'https://base-url.com/repo_url_tests/module-provider-override-git-provider-test'),
    ])
    def test_get_source_base_url_with_custom_module_provider_and_module_version_urls_disabled(self, module_name, module_version, expected_base_url):
        """Ensure base URL matches the expected values when module provider and module version repo urls are disabled."""
        namespace = Namespace(name='repo_url_tests')
        module = Module(namespace=namespace, name=module_name)
        module_provider = ModuleProvider.get(module=module, name='test')
        assert module_provider is not None
        module_version = ModuleVersion.get(module_provider=module_provider, version=module_version)
        assert module_version is not None

        with unittest.mock.patch('terrareg.config.Config.ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER', False):
            with unittest.mock.patch('terrareg.config.Config.ALLOW_CUSTOM_GIT_URL_MODULE_VERSION', False):
                assert module_version.get_source_base_url() == expected_base_url

    @pytest.mark.parametrize('module_name,module_version,git_path,expected_source_download_url', [
        # Test no clone URL in any configuration, defaulting to source archive download
        ('no-git-provider', '1.0.0', None, '/v1/terrareg/modules/repo_url_tests/no-git-provider/test/1.0.0/source.zip'),
        # Test clone URL only configured in module version
        ('no-git-provider', '1.4.0', None, 'git::ssh://mv-clone-url.com/repo_url_tests/no-git-provider-test?ref=1.4.0'),

        # Test with git provider configured in module provider
        ('git-provider-urls', '1.1.0', None, 'git::ssh://clone-url.com/repo_url_tests/git-provider-urls-test?ref=1.1.0'),
        # Test with git provider configured in module provider and override in module version
        ('git-provider-urls', '1.4.0', None, 'git::ssh://mv-clone-url.com/repo_url_tests/git-provider-urls-test?ref=1.4.0'),

        # Test with git provider configured in module provider
        ('module-provider-urls', '1.2.0', None, 'git::ssh://mp-clone-url.com/repo_url_tests/module-provider-urls-test?ref=1.2.0'),
        # Test with URls configured in module provider and override in module version
        ('module-provider-urls', '1.4.0', None, 'git::ssh://mv-clone-url.com/repo_url_tests/module-provider-urls-test?ref=1.4.0'),

        # Test with git provider configured in module provider and URLs configured in git provider
        ('module-provider-override-git-provider', '1.3.0', None, 'git::ssh://mp-clone-url.com/repo_url_tests/module-provider-override-git-provider-test?ref=1.3.0'),
        # Test with URls configured in module provider and override in module version
        ('module-provider-override-git-provider', '1.4.0', None, 'git::ssh://mv-clone-url.com/repo_url_tests/module-provider-override-git-provider-test?ref=1.4.0'),

        ## Tests with git_path set for sub-directory of repo
        # Test no clone URL in any configuration, defaulting to source archive download
        ('no-git-provider', '1.0.0', 'sub/directory/of/repo', '/v1/terrareg/modules/repo_url_tests/no-git-provider/test/1.0.0/source.zip'),
        # Test clone URL only configured in module version
        ('no-git-provider', '1.4.0', 'sub/directory/of/repo', 'git::ssh://mv-clone-url.com/repo_url_tests/no-git-provider-test//sub/directory/of/repo?ref=1.4.0'),

        # Test with git provider configured in module provider
        ('git-provider-urls', '1.1.0', 'sub/directory/of/repo', 'git::ssh://clone-url.com/repo_url_tests/git-provider-urls-test//sub/directory/of/repo?ref=1.1.0'),
        # Test with git provider configured in module provider and override in module version
        ('git-provider-urls', '1.4.0', 'sub/directory/of/repo', 'git::ssh://mv-clone-url.com/repo_url_tests/git-provider-urls-test//sub/directory/of/repo?ref=1.4.0'),

        # Test with git provider configured in module provider
        ('module-provider-urls', '1.2.0', 'sub/directory/of/repo', 'git::ssh://mp-clone-url.com/repo_url_tests/module-provider-urls-test//sub/directory/of/repo?ref=1.2.0'),
        # Test with URls configured in module provider and override in module version
        ('module-provider-urls', '1.4.0', 'sub/directory/of/repo', 'git::ssh://mv-clone-url.com/repo_url_tests/module-provider-urls-test//sub/directory/of/repo?ref=1.4.0'),

        # Test with git provider configured in module provider and URLs configured in git provider
        ('module-provider-override-git-provider', '1.3.0', 'sub/directory/of/repo', 'git::ssh://mp-clone-url.com/repo_url_tests/module-provider-override-git-provider-test//sub/directory/of/repo?ref=1.3.0'),
        # Test with URls configured in module provider and override in module version
        ('module-provider-override-git-provider', '1.4.0', 'sub/directory/of/repo', 'git::ssh://mv-clone-url.com/repo_url_tests/module-provider-override-git-provider-test//sub/directory/of/repo?ref=1.4.0'),
    ])
    def test_get_source_download_url(self, module_name, module_version, git_path, expected_source_download_url):
        """Ensure clone URL matches the expected values."""
        namespace = Namespace(name='repo_url_tests')
        module = Module(namespace=namespace, name=module_name)
        module_provider = ModuleProvider.get(module=module, name='test')
        module_provider.update_git_path(git_path)

        try:
            assert module_provider is not None
            module_version = ModuleVersion.get(module_provider=module_provider, version=module_version)
            assert module_version is not None

            assert module_version.get_source_download_url() == expected_source_download_url

        finally:
            module_provider.update_git_path(None)

    @pytest.mark.parametrize('published,beta,is_latest_version,expected_value', [
        # Latest published non-beta
        (True, False, True, []),
        # Non-latest published non-beta
        (True, False, False, ['This version of the module is not the latest version.', 'To use this specific version, it must be pinned in Terraform']),

        # Non-latest un-published non-beta
        (False, False, False, ['This version of this module has not yet been published,', 'meaning that it cannot yet be used by Terraform']),

        # Un-published beta
        (False, True, False, ['This version of this module has not yet been published,', 'meaning that it cannot yet be used by Terraform']),
        # Published beta
        (True, True, False, ['This version of the module is a beta version.', 'To use this version, it must be pinned in Terraform']),
    ])
    def test_get_terraform_example_version_comment(self, published, beta, is_latest_version, expected_value):
        """Test get_terraform_example_version_comment"""
        with unittest.mock.patch("terrareg.models.ModuleVersion.beta", beta), \
                unittest.mock.patch("terrareg.models.ModuleVersion.published", published), \
                unittest.mock.patch("terrareg.models.ModuleVersion.is_latest_version", is_latest_version):
            module_provider = ModuleProvider.get(Module(Namespace('moduledetails'), 'withterraformdocs'), 'testprovider')
            module_version = ModuleVersion.get(module_provider, '1.5.0')

            assert module_version.get_terraform_example_version_comment() == expected_value
