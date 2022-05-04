
import pytest
import sqlalchemy
from terrareg.database import Database

from terrareg.models import Module, Namespace, ModuleProvider, ModuleVersion
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
        '1.2.3-a',
        '1.2.2-1'
    ])
    def test_invalid_module_versions(self, version):
        """Test invalid module versions"""
        namespace = Namespace(name='test')
        module = Module(namespace=namespace, name='test')
        module_provider = ModuleProvider(module=module, name='test')
        with pytest.raises(terrareg.errors.InvalidVersionError):
            ModuleVersion(module_provider=module_provider, version=version)

    @pytest.mark.parametrize('version', [
        '1.1.1',
        '13.14.16',
        '1.10.10',
        '01.01.01'  # @TODO Should this be allowed?
    ])
    def test_valid_module_versions(self, version):
        """Test valid module versions"""
        namespace = Namespace(name='test')
        module = Module(namespace=namespace, name='test')
        module_provider = ModuleProvider(module=module, name='test')
        ModuleVersion(module_provider=module_provider, version=version)

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

        for attr in ['artifact_location', 'description', 'module_details', 'owner',
                     'published_at', 'readme_content', 'repo_base_url_template',
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
                published=True
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

        for attr in ['artifact_location', 'description', 'module_details', 'owner',
                     'published_at', 'readme_content', 'repo_base_url_template',
                     'repo_browse_url_template', 'repo_clone_url_template',
                     'variable_template']:
            assert new_db_row[attr] == None

        # Ensure that all moduleversion and submodules have been removed
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
