
import os
import shutil
import tempfile
from unittest import mock
import pytest
from terrareg.audit import AuditEvent
from terrareg.database import Database

from terrareg.models import GitProvider, Module, ModuleVersion, Namespace, ModuleProvider
import terrareg.errors
from test.integration.terrareg import TerraregIntegrationTest
import terrareg.audit_action


class TestModuleProvider(TerraregIntegrationTest):

    @pytest.mark.parametrize('module_provider_name', [
        'invalid@atsymbol',
        'invalid"doublequote',
        "invalid'singlequote",
        '-startwithdash',
        'endwithdash-',
        '_startwithunderscore',
        'endwithunscore_',
        'a:colon',
        'or;semicolon',
        'who?knows',
        'with-dash',
        'with_underscore',
        'withAcapital',
        'StartwithCaptital',
        'endwithcapitaL',
        ''
    ])
    def test_invalid_module_provider_names(self, module_provider_name):
        """Test invalid module names"""
        namespace = Namespace(name='test')
        module = Module(namespace=namespace, name='test')
        with pytest.raises(terrareg.errors.InvalidModuleProviderNameError):
            ModuleProvider(module=module, name=module_provider_name)

    @pytest.mark.parametrize('module_provider_name', [
        'normalname',
        'name2withnumber',
        '2startendiwthnumber2',
        'contains4number'
    ])
    def test_valid_module_provider_names(self, module_provider_name):
        """Test valid module names"""
        namespace = Namespace(name='test')
        module = Module(namespace=namespace, name='test')
        ModuleProvider(module=module, name=module_provider_name)


    def test_module_provider_name_in_allow_list(self):
        """Test module provider name that is not in allow list"""
        with mock.patch('terrareg.config.Config.ALLOWED_PROVIDERS', ['aws', 'azure', 'test']):
            namespace = Namespace(name='test')
            module = Module(namespace=namespace, name='test')
            ModuleProvider(module=module, name='aws')
            ModuleProvider(module=module, name='azure')
            ModuleProvider(module=module, name='test')


    def test_module_provider_name_not_in_allow_list(self):
        """Test module provider name that is not in allow list"""
        with mock.patch('terrareg.config.Config.ALLOWED_PROVIDERS', ['onlyallow']):
            namespace = Namespace(name='test')
            module = Module(namespace=namespace, name='test')
            with pytest.raises(terrareg.errors.ProviderNameNotPermittedError):
                ModuleProvider(module=module, name='notallowed')

    def test_module_provider_get_versions(self):
        """Test that a module provider with versions in the wrong order are still returned correctly."""
        namespace = Namespace(name='testnamespace')
        module = Module(namespace=namespace, name='wrongversionorder')
        module_provider = ModuleProvider.get(module=module, name='testprovider')

        assert [mv.version for mv in module_provider.get_versions()] == [
            '23.2.3-beta', '10.23.0', '2.1.0',
            '1.5.4', '0.1.10', '0.1.09', '0.1.8',
            '0.1.1', '0.0.9'
        ]

    def test_module_provider_get_versions_without_beta(self):
        """Test that a module provider with versions in the wrong order are still returned correctly."""
        namespace = Namespace(name='testnamespace')
        module = Module(namespace=namespace, name='wrongversionorder')
        module_provider = ModuleProvider.get(module=module, name='testprovider')

        assert [mv.version for mv in module_provider.get_versions(include_beta=False)] == [
            '10.23.0', '2.1.0', '1.5.4',
            '0.1.10', '0.1.09', '0.1.8',
            '0.1.1', '0.0.9'
        ]

    def test_module_provider_get_latest_version(self):
        """
        Test that a module provider with versions in the wrong order return correct
        latest version and ignores beta version.
        """
        namespace = Namespace(name='testnamespace')
        module = Module(namespace=namespace, name='wrongversionorder')
        module_provider = ModuleProvider.get(module=module, name='testprovider')
        module_version = module_provider.get_latest_version()

        assert module_version.version == '10.23.0'

    @pytest.mark.parametrize('module_name', [
        # Module with no versions at all
        'noversions',
        # Module with only unpublished version
        'onlyunpublished',
        # Module with only a published beta version
        'onlybeta'
    ])
    def test_module_provider_get_latest_version_with_no_version(self, module_name):
        """
        Test that a module provider without any versions does not return
        a latest version.
        """
        namespace = Namespace(name='testnamespace')
        module = Module(namespace=namespace, name=module_name)
        module_provider = ModuleProvider.get(module=module, name='testprovider')
        module_version = module_provider.get_latest_version()

        assert module_version is None

    def test_module_provider_calculate_latest_version(self):
        """
        Test that a module provider with versions in the wrong order return correct
        latest version and ignores beta version with calculate_latest_version.
        """
        namespace = Namespace(name='testnamespace')
        module = Module(namespace=namespace, name='wrongversionorder')
        module_provider = ModuleProvider.get(module=module, name='testprovider')
        module_version = module_provider.calculate_latest_version()

        assert module_version.version == '10.23.0'

    @pytest.mark.parametrize('module_name', [
        # Module with no versions at all
        'noversions',
        # Module with only unpublished version
        'onlyunpublished',
        # Module with only a published beta version
        'onlybeta'
    ])
    def test_module_provider_calculate_latest_version_with_no_version(self, module_name):
        """
        Test that a module provider without any versions does not return
        a latest version using calculate_latest_version.
        """
        namespace = Namespace(name='testnamespace')
        module = Module(namespace=namespace, name=module_name)
        module_provider = ModuleProvider.get(module=module, name='testprovider')
        module_version = module_provider.calculate_latest_version()

        assert module_version is None

    def test_get_total_count(self):
        """Test get_total_count method"""
        assert ModuleProvider.get_total_count() == 43

    def test_get_module_provider_existing(self):
        """Attempt to get existing module provider"""
        namespace = Namespace.get(name='genericmodules')
        module = Module(namespace=namespace, name='modulename')
        module_provider = ModuleProvider.get(module=module, name='providername')
        assert module_provider is not None
        row = module_provider._get_db_row()
        assert row['id'] == 48
        assert row['namespace_id'] == namespace.pk
        assert row['module'] == 'modulename'
        assert row['provider'] == 'providername'

    def test_get_module_provider_non_existent(self):
        """Attempt to get non-existent module provider"""
        namespace = Namespace.get(name='genericmodules')
        module = Module(namespace=namespace, name='modulename')
        module_provider = ModuleProvider.get(module=module, name='doesnotexist')
        assert module_provider is None

    def test_get_module_provider_with_create(self):
        """Attempt to get non-existent module provider with create"""
        namespace = Namespace.get(name='genericmodules')
        module = Module(namespace=namespace, name='modulename')
        with mock.patch('terrareg.config.Config.AUTO_CREATE_MODULE_PROVIDER', True):
            module_provider = ModuleProvider.get(module=module, name='doesnotexistgetcreate', create=True)
            assert module_provider is not None
            assert module_provider._get_db_row()['provider'] == 'doesnotexistgetcreate'

    def test_get_module_provider_with_create_auto_create_disabled(self):
        """Attempt to get non-existent module provider with auto-creation disabled"""
        namespace = Namespace.get(name='genericmodules')
        module = Module(namespace=namespace, name='modulename')
        with mock.patch('terrareg.config.Config.AUTO_CREATE_MODULE_PROVIDER', False):
            module_provider = ModuleProvider.get(module=module, name='doesnotexist', create=True)
            assert module_provider is None

    def test_get_module_provider_with_create_existing(self):
        """Attempt to get non-existent module provider with create"""
        namespace = Namespace.get(name='genericmodules')
        module = Module(namespace=namespace, name='modulename')
        with mock.patch('terrareg.config.Config.AUTO_CREATE_MODULE_PROVIDER', True):
            module_provider = ModuleProvider.get(module=module, name='providername', create=True)
            assert module_provider is not None
            assert module_provider._get_db_row()['id'] == 48

    def test_get_module_provider_with_create_auto_create_disabled_existing(self):
        """Attempt to get non-existent module provider with auto-creation disabled"""
        namespace = Namespace.get(name='genericmodules')
        module = Module(namespace=namespace, name='modulename')
        with mock.patch('terrareg.config.Config.AUTO_CREATE_MODULE_PROVIDER', False):
            module_provider = ModuleProvider.get(module=module, name='providername', create=True)
            assert module_provider is not None
            assert module_provider._get_db_row()['id'] == 48

    @pytest.mark.parametrize('git_path,expected_git_path', [
        (None, None),
        ('', None),
        ('./', None),
        ('/', None),
        ('subpath', 'subpath'),
        ('/subpath', 'subpath'),
        ('./subpath', 'subpath'),
        ('./subpath/', 'subpath'),
        ('./test/another/dir', 'test/another/dir'),
        ('./test/another/dir/', 'test/another/dir'),
        ('.//lots/of///slashes//', 'lots/of/slashes')
    ])
    def test_git_path(self, git_path, expected_git_path):
        """Test git_path property"""
        module_provider = ModuleProvider.get(Module(Namespace('moduledetails'), 'git-path'), 'provider')
        module_provider.update_git_path(git_path)
        assert module_provider.git_path == expected_git_path

    def test_delete(self):
        """Test deletion of module version."""
        existing_module_provider_count = ModuleProvider.get_total_count()
        namespace = Namespace.get(name='testnamespace')
        module = Module(namespace=namespace, name='to-delete')

        # Create test module provider
        module_provider = ModuleProvider.get(module=module, name='testprovider', create=True)
        module_provider_pk = module_provider.pk

        # Create test module versions
        module_version_pks = []
        for itx in ['1.0.0', '1.1.1', '1.2.3']:
            module_version = ModuleVersion(module_provider=module_provider, version=itx)
            module_version.prepare_module()
            module_version.publish()
            module_version_pks.append(module_version.pk)

        assert ModuleProvider.get_total_count() == (existing_module_provider_count + 1)

        # Ensure that all of the rows can be fetched
        db = Database.get()
        with db.get_connection() as conn:

            res = conn.execute(db.module_provider.select().where(db.module_provider.c.id==module_provider_pk))
            assert res.fetchone() is not None

            for mv_pk in module_version_pks:
                res = conn.execute(db.module_version.select().where(db.module_version.c.id==mv_pk))
                assert res.fetchone() is not None

        # Delete module provider
        module_provider.delete()

        assert ModuleProvider.get_total_count() == existing_module_provider_count

        # Check module_version, example and example file have been removed
        with db.get_connection() as conn:

            res = conn.execute(db.module_provider.select().where(db.module_provider.c.id==module_provider_pk))
            assert res.fetchone() is None

            for mv_pk in module_version_pks:
                res = conn.execute(db.module_version.select().where(db.module_version.c.id==mv_pk))
                assert res.fetchone() is None

    @pytest.mark.parametrize('module_directory_exists, module_provider_directory_exists, module_provider_directory_non_empty', [
        # Data directory does not exist for module
        (False, False, False),
        # Data directory for module provider does not exist
        (True, False, False),
        # Directories exist
        (True, True, False),
        # Pre-existing files are present in module provider directory
        (True, True, True),
    ])
    def test_delete_removes_data_files(self, module_directory_exists, module_provider_directory_exists, module_provider_directory_non_empty):
        """Ensure removal of module provider removes any data files for the module provider"""
        namespace = Namespace.get(name='testnamespace')
        module = Module(namespace=namespace, name='to-delete')

        # Create test module provider
        module_provider = ModuleProvider.get(module=module, name='datadelete', create=True)

        # Patch data directory to a temporary directory
        data_directory = tempfile.mkdtemp()
        try:
            with mock.patch('terrareg.config.Config.DATA_DIRECTORY', data_directory):
                # Create module provider data directory tree
                module_provider_directory = os.path.join(data_directory, module_provider.base_directory.lstrip(os.path.sep))
                module_directory = os.path.join(data_directory, module.base_directory.lstrip(os.path.sep))
                os.makedirs(module_provider_directory)

                # Create additional test file in module provider directory
                # to ensure it is not accidently removed
                test_module_provider_file = os.path.join(module_provider_directory, 'test_file')
                if module_provider_directory_non_empty:
                    with open(test_module_provider_file, 'w'):
                        pass

                # Remove module provider/module directories to match
                # test case
                if not module_provider_directory_exists:
                    os.rmdir(module_provider_directory)
                if not module_directory_exists:
                    os.rmdir(module_directory)

                # Remove module version
                module_provider.delete()

                # Ensure directory was removed, if there are not pre-existing files
                # in the module provider directory
                assert os.path.exists(module_provider_directory) is module_provider_directory_non_empty

                # Ensure module provider directory exists, if it
                # existed in test case
                if module_directory_exists:
                    assert os.path.isdir(module_directory)

        finally:
            shutil.rmtree(data_directory)
            # Delete module provider if it still exists
            if module_provider._get_db_row():
                module_provider.delete()

    @pytest.mark.parametrize('original_git_provider_id, new_git_provider_id', [
        (None, None),
        (None, 1),
        (1, None),
        (1, 2),
        (2, 2)
    ])
    def test_update_git_provider(self, original_git_provider_id, new_git_provider_id):
        """Test update_git_provider method"""

        original_git_provider = GitProvider(original_git_provider_id) if original_git_provider_id else None
        new_git_provider = GitProvider(new_git_provider_id) if new_git_provider_id else None

        module_provider = ModuleProvider.get(Module(Namespace.get('testnamespace'), 'noversions'), 'testprovider')
        module_provider.update_attributes(git_provider_id=original_git_provider_id)

        assert module_provider.get_git_provider() == original_git_provider

        if original_git_provider_id != new_git_provider_id:
            assert module_provider.get_git_provider() != new_git_provider
        else:
            assert module_provider.get_git_provider() == new_git_provider

        # Update git provider
        module_provider.update_git_provider(new_git_provider)

        # Re-obtain module provider object
        module_provider = ModuleProvider.get(Module(Namespace.get('testnamespace'), 'noversions'), 'testprovider')

        assert module_provider.get_git_provider() == new_git_provider

        if original_git_provider_id != new_git_provider_id:
            assert module_provider.get_git_provider() != original_git_provider
        else:
            assert module_provider.get_git_provider() == original_git_provider

    @pytest.mark.parametrize('url', [
        None,
        '',
        'https://github.com/example/blah.git',
        'http://github.com/example/blah.git',
        'ssh://github.com/example/blah.git',
        'ssh://github.com:7999/example/blah.git',
        'ssh://github.com:7999/{namespace}/{provider}-{module}.git',
    ])
    def test_update_repo_clone_url_template(self, url):
        """Ensure update_repo_clone_url_template successfully updates path"""
        module_provider = ModuleProvider.get(Module(Namespace.get('testnamespace'), 'noversions'), 'testprovider')
        module_provider.update_attributes(repo_clone_url_template=None)

        module_provider.update_repo_clone_url_template(url)

        # Create new module provider object
        module_provider = ModuleProvider.get(Module(Namespace.get('testnamespace'), 'noversions'), 'testprovider')
        assert module_provider._get_db_row()['repo_clone_url_template'] == url

    @pytest.mark.parametrize('url, expected_exception, expected_message', [
        ('://github.com/example/blah.git',
         terrareg.errors.RepositoryUrlDoesNotContainValidSchemeError,
         'URL does not contain a scheme (e.g. ssh://)'),
        ('ftp://github.com/example/blah.git',
         terrareg.errors.RepositoryUrlContainsInvalidSchemeError,
         'URL contains an unknown scheme (e.g. https/ssh/http)'),
        ('ssh://github.com:example/blah.git',
         terrareg.errors.RepositoryUrlContainsInvalidPortError,
         'URL contains a invalid port. Only use a colon to for specifying a port, otherwise a forward slash should be used.'),
        ('ssh://github.com',
         terrareg.errors.RepositoryUrlDoesNotContainPathError,
         'URL does not contain a path'),
        ('ssh:///example/blah.git',
         terrareg.errors.RepositoryUrlDoesNotContainHostError,
         'URL does not contain a host/domain'),
        ('ssh://{invalidvalue}/example',
         terrareg.errors.RepositoryUrlContainsInvalidTemplateError,
         'URL contains invalid template value. Only the following template values are allowed: {namespace}, {module}, {provider}'),
        ('ssh://{tag}/example',
         terrareg.errors.RepositoryUrlContainsInvalidTemplateError,
         'URL contains invalid template value. Only the following template values are allowed: {namespace}, {module}, {provider}'),
        ('ssh://{path}/example',
         terrareg.errors.RepositoryUrlContainsInvalidTemplateError,
         'URL contains invalid template value. Only the following template values are allowed: {namespace}, {module}, {provider}'),
    ])
    def test_update_repo_clone_url_template_invalid_url(self, url, expected_exception, expected_message):
        """Ensure update_repo_clone_url_template with invalid URLs"""
        module_provider = ModuleProvider.get(Module(Namespace.get('testnamespace'), 'noversions'), 'testprovider')
        module_provider.update_attributes(repo_clone_url_template='old-value')

        with pytest.raises(expected_exception) as exc:
            module_provider.update_repo_clone_url_template(url)
        assert str(exc.value) == expected_message

        # Create new module provider object
        module_provider = ModuleProvider.get(Module(Namespace.get('testnamespace'), 'noversions'), 'testprovider')

        # Ensure clone URL hasn't been modified
        assert module_provider._get_db_row()['repo_clone_url_template'] == 'old-value'

    @pytest.mark.parametrize('url', [
        None,
        '',
        'https://github.com/example/blah/{tag}/{path}',
        'http://github.com/example/blah/{tag}/{path}',
        'https://github.com:7999/{namespace}/{provider}-{module}/{tag}/{path}',
    ])
    def test_update_repo_browse_url_template(self, url):
        """Ensure update_repo_browse_url_template successfully updates path"""
        module_provider = ModuleProvider.get(Module(Namespace.get('testnamespace'), 'noversions'), 'testprovider')
        module_provider.update_attributes(repo_browse_url_template=None)

        module_provider.update_repo_browse_url_template(url)

        # Create new module provider object
        module_provider = ModuleProvider.get(Module(Namespace.get('testnamespace'), 'noversions'), 'testprovider')
        assert module_provider._get_db_row()['repo_browse_url_template'] == url

    @pytest.mark.parametrize('url, expected_exception, expected_message', [
        ('://github.com/example/blah/{tag}/{path}',
         terrareg.errors.RepositoryUrlDoesNotContainValidSchemeError,
         'URL does not contain a scheme (e.g. https://)'),
        ('ftp://github.com/example/blah/{tag}/{path}',
         terrareg.errors.RepositoryUrlContainsInvalidSchemeError,
         'URL contains an unknown scheme (e.g. https/http)'),
        ('https://github.com:example/blah/{tag}/{path}',
         terrareg.errors.RepositoryUrlContainsInvalidPortError,
         'URL contains a invalid port. Only use a colon to for specifying a port, otherwise a forward slash should be used.'),
        ('https://github.com-{tag}-{path}',
         terrareg.errors.RepositoryUrlDoesNotContainPathError,
         'URL does not contain a path'),
        ('https:///example/blah/{tag}/{path}',
         terrareg.errors.RepositoryUrlDoesNotContainHostError,
         'URL does not contain a host/domain'),
        ('https://example.com/example/blah',
         terrareg.errors.RepositoryUrlParseError,
         'tag or tag_uri_encoded placeholder not present in URL'),
        ('https://example.com/example/blah/{tag}',
         terrareg.errors.RepositoryUrlParseError,
         'Path placeholder not present in URL'),
        ('https://example.com/example/blah/{tag_uri_encoded}',
         terrareg.errors.RepositoryUrlParseError,
         'Path placeholder not present in URL'),
        ('https://{invalidvalue}/{tag}/{path}/example',
         terrareg.errors.RepositoryUrlContainsInvalidTemplateError,
         'URL contains invalid template value. Only the following template values are allowed: {namespace}, {module}, {provider}, {tag}, {path}'),
    ])
    def test_update_repo_browse_url_invalid_url(self, url, expected_exception, expected_message):
        """Ensure update_repo_browse_url with invalid URLs"""
        module_provider = ModuleProvider.get(Module(Namespace.get('testnamespace'), 'noversions'), 'testprovider')
        module_provider.update_attributes(repo_browse_url_template='old-value')

        with pytest.raises(expected_exception) as exc:
            module_provider.update_repo_browse_url_template(url)
        assert str(exc.value) == expected_message

        # Create new module provider object
        module_provider = ModuleProvider.get(Module(Namespace.get('testnamespace'), 'noversions'), 'testprovider')

        # Ensure browse URL hasn't been modified
        assert module_provider._get_db_row()['repo_browse_url_template'] == 'old-value'

    @pytest.mark.parametrize('url', [
        None,
        '',
        'https://github.com/example/blah',
        'http://github.com/example/blah',
        'https://github.com:7999/{namespace}/{provider}-{module}',
    ])
    def test_update_repo_base_url_template(self, url):
        """Ensure update_repo_base_url_template successfully updates path"""
        module_provider = ModuleProvider.get(Module(Namespace.get('testnamespace'), 'noversions'), 'testprovider')
        module_provider.update_attributes(repo_base_url_template=None)

        module_provider.update_repo_base_url_template(url)

        # Create new module provider object
        module_provider = ModuleProvider.get(Module(Namespace.get('testnamespace'), 'noversions'), 'testprovider')
        assert module_provider._get_db_row()['repo_base_url_template'] == url

    @pytest.mark.parametrize('url, expected_exception, expected_message', [
        ('://github.com/example/blah',
         terrareg.errors.RepositoryUrlDoesNotContainValidSchemeError,
         'URL does not contain a scheme (e.g. https://)'),
        ('ftp://github.com/example/blah',
         terrareg.errors.RepositoryUrlContainsInvalidSchemeError,
         'URL contains an unknown scheme (e.g. https/http)'),
        ('https://github.com:example/blah',
         terrareg.errors.RepositoryUrlContainsInvalidPortError,
         'URL contains a invalid port. Only use a colon to for specifying a port, otherwise a forward slash should be used.'),
        ('https://github.com',
         terrareg.errors.RepositoryUrlDoesNotContainPathError,
         'URL does not contain a path'),
        ('https:///example/blah',
         terrareg.errors.RepositoryUrlDoesNotContainHostError,
         'URL does not contain a host/domain'),
        ('https://{invalidvalue}/example',
         terrareg.errors.RepositoryUrlContainsInvalidTemplateError,
         'URL contains invalid template value. Only the following template values are allowed: {namespace}, {module}, {provider}'),
        ('https://{path}/example',
         terrareg.errors.RepositoryUrlContainsInvalidTemplateError,
         'URL contains invalid template value. Only the following template values are allowed: {namespace}, {module}, {provider}'),
        ('https://{tag}/example',
         terrareg.errors.RepositoryUrlContainsInvalidTemplateError,
         'URL contains invalid template value. Only the following template values are allowed: {namespace}, {module}, {provider}'),
    ])
    def test_update_repo_base_url_invalid_url(self, url, expected_exception, expected_message):
        """Ensure update_repo_base_url with invalid URLs"""
        module_provider = ModuleProvider.get(Module(Namespace.get('testnamespace'), 'noversions'), 'testprovider')
        module_provider.update_attributes(repo_base_url_template='old-value')

        with pytest.raises(expected_exception) as exc:
            module_provider.update_repo_base_url_template(url)
        assert str(exc.value) == expected_message

        # Create new module provider object
        module_provider = ModuleProvider.get(Module(Namespace.get('testnamespace'), 'noversions'), 'testprovider')

        # Ensure base URL hasn't been modified
        assert module_provider._get_db_row()['repo_base_url_template'] == 'old-value'

    @pytest.mark.parametrize('format', [
        '{version}',
        'v{version}',
        '{major}',
        '{minor}',
        '{patch}',
        '{major}.{minor}',
        '{major}.{patch}',
        '{minor}.{patch}',
        'releases/v{minor}.{patch}-testing',
        # Unsetting value
        None,
        ''
    ])
    def test_update_git_tag_format_valid(self, format):
        """Test update_git_tag_format with valid values"""
        module_provider = ModuleProvider.get(Module(Namespace.get('testnamespace'), 'noversions'), 'testprovider')
        module_provider.update_git_tag_format(git_tag_format=format)

        module_provider = ModuleProvider.get(Module(Namespace.get('testnamespace'), 'noversions'), 'testprovider')

        # If format has been provided, assert that it's now set
        if format:
            assert module_provider.git_tag_format == format
        else:
            # Otherwise, check that the default value is returned
            assert module_provider.git_tag_format == '{version}'

    @pytest.mark.parametrize('format', [
        # Invalid placeholder
        '{notversion}',
        # No placeholders
        'v',
        # Mixture of valid and invalid placeholders
        '{test}.{version}'
    ])
    def test_update_git_tag_format_invalid(self, format):
        """Test update_git_tag_format with valid values"""
        module_provider = ModuleProvider.get(Module(Namespace.get('testnamespace'), 'noversions'), 'testprovider')
        with pytest.raises(terrareg.errors.InvalidGitTagFormatError):
            module_provider.update_git_tag_format(git_tag_format=format)

    @pytest.mark.parametrize('git_tag_format, tag, expected_version', [
        # Test version placeholder
        ## On it's own
        ('{version}', '1.2.3', '1.2.3'),
        # With prefix and suffix
        ('v{version}a', 'v1.2.3a', '1.2.3'),
        # Invalid match - missing part of version
        ('{version}', '1.2', None),
        ('{version}', '1.2.', None),
        ('{version}', '.1.2', None),
        ('{version}', 'a1.2.3', None),
        ('{version}', '1.2.3a', None),

        # Test major placeholder
        ## On it's own
        ('{major}', '1', '1.0.0'),
        # With prefix and suffix
        ('release/v{major}a', 'release/v5a', '5.0.0'),
        # No match
        ('{major}', 'abc', None),
        ('{major}', '1.5.2', None),

        # Test minor placeholder
        ## On it's own
        ('{minor}', '1', '0.1.0'),
        # With prefix and suffix
        ('release/v{minor}a', 'release/v5a', '0.5.0'),
        # No match
        ('{minor}', 'abc', None),
        ('{minor}', '1.5.2', None),

        # Test patch placeholder
        ## On it's own
        ('{patch}', '1', '0.0.1'),
        # With prefix and suffix
        ('release/v{patch}a', 'release/v5a', '0.0.5'),
        # No match
        ('{patch}', 'abc', None),
        ('{patch}', '1.5.2', None),

        # Test combination of major, minor and patch
        ('release/v{major}.{minor}', 'release/v4.9', '4.9.0'),
        ('release/v{major}.{patch}', 'release/v4.9', '4.0.9'),
        ('release/v{minor}.{patch}', 'release/v4.9', '0.4.9'),
        ('release/v{major}.{minor}-{patch}', 'release/v4.9-5', '4.9.5'),
    ])
    def test_get_version_from_tag(self, git_tag_format, tag, expected_version):
        """Test get_version_from_tag and get_version_from_tag_ref"""
        module_provider = ModuleProvider.create(Module(Namespace.get('testnamespace'), 'noversions'), 'testproviderversionimport')
        try:

            module_provider.update_git_tag_format(git_tag_format)

            # test get_version_from_tag
            version = module_provider.get_version_from_tag(tag=tag)
            assert version == expected_version

            # Test get_version_from_tag_ref
            tag_ref_version = module_provider.get_version_from_tag_ref(tag_ref=f'refs/tags/{tag}')
            assert tag_ref_version == expected_version

        finally:
            module_provider.delete()

    @pytest.mark.parametrize("new_namespace_name", [
        # Original namespace
        "testnamespace",

        # New namespace
        "moduleextraction"
    ])
    @pytest.mark.parametrize("new_module_name", [
        # Original module name
        "torename",

        # New module name
        "newname",

        # Case change
        "NewName",
    ])
    @pytest.mark.parametrize("new_provider_name", [
        # Original provider name
        "test",

        # New provider name
        "newprovider",
    ])
    def test_update_name(self, new_namespace_name, new_module_name, new_provider_name):
        """Test update_name method in successful case"""
        original_namespace_name = "testnamespace"
        original_module_name = "torename"
        original_provider_name = "test"

        # Skip test where all attributes are the same
        if new_namespace_name == original_namespace_name and \
                new_module_name == original_module_name and \
                new_provider_name == original_provider_name:
            pytest.skip('Ignore test case with all unchanged params')

        try:
            # Create new module provider for test
            namespace = Namespace.get(original_namespace_name)
            module_provider = ModuleProvider.create(module=Module(namespace=namespace, name=original_module_name), name=original_provider_name)
            original_module_pk = module_provider.pk

            # Remove all audit events
            db = Database.get()
            with db.get_connection() as conn:
                conn.execute(db.audit_history.delete())

            # Remove all module provider redirects
            db = Database.get()
            with db.get_connection() as conn:
                conn.execute(db.module_provider_redirect.delete())

            # Update attributes
            new_namespace = Namespace.get(name=new_namespace_name)
            new_module_provider = module_provider.update_name(namespace=new_namespace, module_name=new_module_name, provider_name=new_provider_name)

            # Ensure attributes of the new module provider are correct
            assert isinstance(new_module_provider, ModuleProvider)
            assert new_module_provider.name == new_provider_name
            assert new_module_provider.module.name == new_module_name
            assert new_module_provider.module.namespace.name == new_namespace_name

            # Attempt to obtain via new name
            test_provider_from_new_name = ModuleProvider.get(Module(Namespace(name=new_namespace_name), name=new_module_name), name=new_provider_name)
            assert test_provider_from_new_name is not None
            assert test_provider_from_new_name.pk == original_module_pk
            assert test_provider_from_new_name.name == new_provider_name
            assert test_provider_from_new_name.module.name == new_module_name
            assert test_provider_from_new_name.module.namespace.name == new_namespace_name

            # Attempt to obtain using redirect from old name
            test_provider_from_new_name = ModuleProvider.get(Module(Namespace(name=original_namespace_name), name=original_module_name), name=original_provider_name)
            assert test_provider_from_new_name is not None
            assert test_provider_from_new_name.pk == original_module_pk
            assert test_provider_from_new_name.name == new_provider_name
            assert test_provider_from_new_name.module.name == new_module_name
            assert test_provider_from_new_name.module.namespace.name == new_namespace_name

            # Check audit event
            unprocessed_audit_events, _, _ = AuditEvent.get_events(limit=5, descending=True, order_by="timestamp")
            for audit_action, original_value, new_value in [
                    [terrareg.audit_action.AuditAction.MODULE_PROVIDER_UPDATE_NAMESPACE, original_namespace_name, new_namespace_name],
                    [terrareg.audit_action.AuditAction.MODULE_PROVIDER_UPDATE_MODULE_NAME, original_module_name, new_module_name],
                    [terrareg.audit_action.AuditAction.MODULE_PROVIDER_UPDATE_PROVIDER_NAME, original_provider_name, new_provider_name]]:
                filtered_events = [e for e in filter(lambda x: x['action'] == audit_action, unprocessed_audit_events)]
                if original_value != new_value:
                    assert len(filtered_events) == 1
                    audit_event = filtered_events[0]
                    assert audit_event['action'] == audit_action
                    assert audit_event['object_type'] == "ModuleProvider"
                    assert audit_event['object_id'] == "testnamespace/torename/test"
                    assert audit_event['old_value'] == original_value
                    assert audit_event['new_value'] == new_value
                    # Remove audit event from unprocessed events
                    unprocessed_audit_events.remove(audit_event)
                else:
                    assert len(filtered_events) == 0

            # Ensure all audit events have been processed
            assert len(unprocessed_audit_events) == 0

        finally:
            # Attempt to delete original
            namespace = Namespace.get(original_namespace_name)
            module_provider = ModuleProvider.get(Module(namespace=namespace, name=original_module_name), name=original_provider_name)
            if module_provider:
                module_provider.delete()

            # Attempt to deleted renamed
            namespace = Namespace.get(new_namespace_name)
            module_provider = ModuleProvider.get(Module(namespace=namespace, name=new_module_name), name=new_provider_name)
            if module_provider:
                module_provider.delete()

    @pytest.mark.parametrize("duplicate_namespace_name, duplicate_module_name, duplicate_provider_name", [
        # When provider is changed to duplicate
        ("testnamespace", "torename", "duplicate"),

        # When module name is changed to duplicate
        ("testnamespace", "duplicate", "test"),
        # Duplicate with different case
        ("testnamespace", "DuplicatE", "test"),

        # When namespace is changed to duplicate
        ("moduleextraction", "torename", "test"),

        # When changing all parameters
        ("moduleextraction", "duplicate", "duplicate"),
        ("moduleextraction", "Duplicate", "duplicate"),
    ])
    def test_update_name_duplicate(self, duplicate_namespace_name, duplicate_module_name, duplicate_provider_name):
        """Test using update_name with duplicate resulting module provider"""
        # Create new module provider for test
        original_namespace_name = "testnamespace"
        original_module_name = "torename"
        original_provider_name = "test"

        try:
            module_provider = ModuleProvider.create(module=Module(namespace=Namespace.get(original_namespace_name), name=original_module_name), name=original_provider_name)
            original_module_provider_pk = module_provider.pk

            duplicate_module_provider = ModuleProvider.create(module=Module(namespace=Namespace.get(duplicate_namespace_name), name=duplicate_module_name), name=duplicate_provider_name)
            duplicate_module_provider_pk = duplicate_module_provider.pk

            # Remove all audit events
            db = Database.get()
            with db.get_connection() as conn:
                conn.execute(db.audit_history.delete())

            # Remove all module provider redirects
            db = Database.get()
            with db.get_connection() as conn:
                conn.execute(db.module_provider_redirect.delete())

            # Update attributes
            new_namespace = Namespace.get(name=duplicate_namespace_name)
            with pytest.raises(terrareg.errors.DuplicateModuleProviderError):
                module_provider.update_name(namespace=new_namespace, module_name=duplicate_module_name, provider_name=duplicate_provider_name)

            # Ensure no audit events were created
            audit_events, _, _ = AuditEvent.get_events(limit=5, descending=True, order_by="timestamp")
            assert len(audit_events) == 0

            # Ensure no redirects exist
            db = Database.get()
            with db.get_connection() as conn:
                res = conn.execute(db.module_provider_redirect.select()).all()
                assert len(res) == 0

            # Check PK of olds modules
            assert ModuleProvider.get(
                module=Module(
                    namespace=Namespace.get(original_namespace_name),
                    name=original_module_name
                ), name=original_provider_name
            ).pk == original_module_provider_pk
            assert ModuleProvider.get(
                module=Module(
                    namespace=Namespace.get(duplicate_namespace_name),
                    name=duplicate_module_name
                ),
                name=duplicate_provider_name
            ).pk == duplicate_module_provider_pk

        finally:
            # Attempt to delete original
            namespace = Namespace.get(original_namespace_name)
            module_provider = ModuleProvider.get(Module(namespace=namespace, name=original_module_name), name=original_provider_name)
            if module_provider:
                module_provider.delete()

            # Attempt to delete duplicate
            namespace = Namespace.get(duplicate_namespace_name)
            module_provider = ModuleProvider.get(Module(namespace=namespace, name=duplicate_module_name), name=duplicate_provider_name)
            if module_provider:
                module_provider.delete()

    @pytest.mark.parametrize("new_namespace_name, new_module_name, new_provider_name", [
        # When provider is changed to duplicate
        ("testnamespace", "torename", "duplicate"),

        # When module name is changed to duplicate
        ("testnamespace", "duplicate", "test"),

        # When namespace is changed to duplicate
        ("moduleextraction", "torename", "test"),

        # When changing all parameters
        ("moduleextraction", "duplicate", "duplicate"),
    ])
    def test_create_override_redirect(self, new_namespace_name, new_module_name, new_provider_name):
        """Test creating module provider that clashes with redirect"""
        # Create new module provider for test
        original_namespace_name = "testnamespace"
        original_module_name = "torename"
        original_provider_name = "test"

        try:
            module_provider = ModuleProvider.create(module=Module(namespace=Namespace.get(original_namespace_name), name=original_module_name), name=original_provider_name)
            original_module_provider_pk = module_provider.pk

            # Update name of original module provider
            module_provider.update_name(namespace=Namespace.get(new_namespace_name), module_name=new_module_name, provider_name=new_provider_name)

            # Attempt to create provider overriding original
            with pytest.raises(terrareg.errors.DuplicateModuleProviderError):
                ModuleProvider.create(module=Module(namespace=Namespace.get(original_namespace_name), name=original_module_name), name=original_provider_name)

            # Obtain old module provider using new name
            test_old_module_provider = ModuleProvider.get(module=Module(namespace=Namespace.get(new_namespace_name), name=new_module_name), name=new_provider_name)
            assert test_old_module_provider.pk == original_module_provider_pk

            test_new_module_provider = ModuleProvider.get(module=Module(namespace=Namespace.get(original_namespace_name), name=original_module_name), name=original_provider_name)
            assert test_new_module_provider.pk == original_module_provider_pk

        finally:
            # Attempt to delete original
            namespace = Namespace.get(original_namespace_name)
            module_provider = ModuleProvider.get(Module(namespace=namespace, name=original_module_name), name=original_provider_name)
            if module_provider:
                module_provider.delete()

            # Attempt to delete override
            namespace = Namespace.get(new_namespace_name)
            module_provider = ModuleProvider.get(Module(namespace=namespace, name=original_module_name), name=original_provider_name)
            if module_provider:
                module_provider.delete()

