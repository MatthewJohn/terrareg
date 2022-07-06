
from datetime import datetime
import unittest.mock
import pytest
import sqlalchemy
from terrareg.analytics import AnalyticsEngine
from terrareg.database import Database

from terrareg.models import Example, ExampleFile, Module, Namespace, ModuleProvider, ModuleVersion
import terrareg.errors
from test.integration.terrareg import TerraregIntegrationTest

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

    def test_create_db_row(self):
        """Test creating DB row"""
        namespace = Namespace(name='testcreation')
        module = Module(namespace=namespace, name='test-module')
        module_provider = ModuleProvider.get(module=module, name='testprovider', create=True)
        module_provider_row = module_provider._get_db_row()

        module_version = ModuleVersion(module_provider=module_provider, version='1.0.0')

        # Ensure that no DB row is returned for new module version
        assert module_version._get_db_row() == None

        # Insert module version into database
        module_version._create_db_row()

        # Ensure that a DB row is now returned
        new_db_row = module_version._get_db_row()
        assert new_db_row['module_provider_id'] == module_provider_row['id']
        assert type(new_db_row['id']) == int

        assert new_db_row['published'] == False
        assert new_db_row['version'] == '1.0.0'

        assert new_db_row['beta'] == False

        for attr in ['description', 'module_details_id', 'owner',
                     'published_at', 'repo_base_url_template',
                     'repo_browse_url_template', 'repo_clone_url_template',
                     'variable_template']:
            assert new_db_row[attr] == None

    def test_create_beta_version(self):
        """Test creating DB row for beta version"""
        namespace = Namespace(name='testcreation')
        module = Module(namespace=namespace, name='test-module')
        module_provider = ModuleProvider.get(module=module, name='testprovider', create=True)
        module_provider_row = module_provider._get_db_row()

        module_version = ModuleVersion(module_provider=module_provider, version='1.0.0-beta')

        # Ensure that no DB row is returned for new module version
        assert module_version._get_db_row() == None

        # Insert module version into database
        module_version._create_db_row()

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

    def test_create_db_row_replace_existing(self):
        """Test creating DB row with pre-existing module version"""

        db = Database.get()

        with db.get_engine().connect() as conn:
            conn.execute(db.module_provider.insert().values(
                id=10000,
                namespace='testcreation',
                module='test-module',
                provider='testprovider'
            ))

            conn.execute(db.module_version.insert().values(
                id=10001,
                module_provider_id=10000,
                version='1.1.0',
                published=True,
                beta=False,
                internal=False
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

        namespace = Namespace(name='testcreation')
        module = Module(namespace=namespace, name='test-module')
        module_provider = ModuleProvider.get(module=module, name='testprovider')
        module_provider_row = module_provider._get_db_row()

        module_version = ModuleVersion(module_provider=module_provider, version='1.1.0')

        # Ensure that pre-existing row is returned
        pre_existing_row = module_version._get_db_row()
        assert pre_existing_row is not None
        assert pre_existing_row['id'] == 10001

        # Insert module version into database
        module_version._create_db_row()

        # Ensure that a DB row is now returned
        new_db_row = module_version._get_db_row()
        assert new_db_row['module_provider_id'] == module_provider_row['id']
        assert type(new_db_row['id']) == int

        assert new_db_row['published'] == False
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


    @pytest.mark.parametrize('template,version,expected_string', [
        ('= {major}.{minor}.{patch}', '1.5.0', '= 1.5.0'),
        ('<= {major_plus_one}.{minor_plus_one}.{patch_plus_one}', '1.5.0', '<= 2.6.1'),
        ('>= {major_minus_one}.{minor_minus_one}.{patch_minus_one}', '4.3.2', '>= 3.2.1'),
        ('>= {major_minus_one}.{minor_minus_one}.{patch_minus_one}', '0.0.0', '>= 0.0.0'),
        ('< {major_plus_one}.0.0', '10584.321.564', '< 10585.0.0'),
        # Test that beta version returns the version and
        # ignores the version template
        ('>= {major_minus_one}.{minor_minus_one}.{patch_minus_one}', '5.6.23-beta', '5.6.23-beta')
    ])
    def test_get_terraform_example_version_string(self, template, version, expected_string):
        """Test get_terraform_example_version_string method"""
        with unittest.mock.patch('terrareg.config.Config.TERRAFORM_EXAMPLE_VERSION_TEMPLATE', template):
            namespace = Namespace(name='test')
            module = Module(namespace=namespace, name='test')
            module_provider = ModuleProvider.get(module=module, name='test', create=True)
            module_version = ModuleVersion(module_provider=module_provider, version=version)
            module_version.prepare_module()
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
        module_version.prepare_module()
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
                    'type': 'text'
                }
            ]

        # Ensure when autogenerated usage builder is enabled, the missing required variables
        # are populated in the variable template
        with unittest.mock.patch('terrareg.config.Config.AUTOGENERATE_USAGE_BUILDER_VARIABLES', True):
            assert module_version.variable_template == [
                {
                    'additional_help': 'Provide the name of the application',
                    'name': 'name_of_application',
                    'quote_value': True,
                    'type': 'text'
                },
                {
                    'additional_help': 'Override the default string',
                    'name': 'undocumented_required_variable',
                    'quote_value': True,
                    'type': 'text'
                }
            ]
