"""Integration tests for provider source GitHub App authentication in module extraction"""

import json
import unittest.mock
import pytest

import terrareg.config
import terrareg.database
import terrareg.models
import terrareg.module_extractor
import terrareg.namespace_type
import terrareg.provider_source.factory
import terrareg.provider_source_type
from test.integration.terrareg import TerraregIntegrationTest


class TestProviderSourceGitAuthentication(TerraregIntegrationTest):
    """Test provider source GitHub App authentication for module extraction"""

    def test_module_extraction_with_provider_source_github_app_auth(self):
        """Test end-to-end module extraction using provider source GitHub App authentication"""
        # Create GitHub organization namespace
        namespace = terrareg.models.Namespace.create(
            name="test-org-auth",
            display_name="Test Organization Auth",
            type_=terrareg.namespace_type.NamespaceType.GITHUB_ORGANISATION
        )

        # Create provider_source with GitHub App config
        provider_source_name = "test-github-provider-source-auth"
        db = terrareg.database.Database.get()
        with terrareg.database.Database.get_connection() as conn:
            conn.execute(db.provider_source.insert().values(
                name=provider_source_name,
                api_name="test-github-provider-source-auth",
                provider_source_type=terrareg.provider_source_type.ProviderSourceType.GITHUB,
                config=db.encode_blob(json.dumps({
                    "base_url": "https://github.example.com",
                    "api_url": "https://api.github.example.com",
                    "client_id": "test-client-id",
                    "client_secret": "test-client-secret",
                    "login_button_text": "Test Login",
                    "private_key_path": "./test_key.pem",
                    "app_id": "test-app-id",
                    "default_installation_id": "test-default-installation-id",
                    "auto_generate_github_organisation_namespaces": False
                }))
            ))

        provider_source = terrareg.provider_source.factory.ProviderSourceFactory.get().get_provider_source_by_name(provider_source_name)

        # Create module and module_provider with provider_source linked
        module = terrareg.models.Module(namespace=namespace, name="test-module")
        module_provider = terrareg.models.ModuleProvider.get(module=module, name="aws", create=True)
        module_provider.update_attributes(provider_source_name=provider_source.name)

        # Set a git clone URL template
        mock_repo_url = "https://github.example.com/{namespace}/{module}-{provider}.git"
        with unittest.mock.patch('terrareg.config.Config.ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER', True):
            module_provider.update_attributes(repo_clone_url_template=mock_repo_url)

        # Create module version
        module_version = terrareg.models.ModuleVersion(module_provider=module_provider, version='1.0.0')

        # Mock GitHub API and git clone
        mock_installation_id = 123456789
        mock_access_token = "ghs_test_token_abc123"

        try:
            with unittest.mock.patch('terrareg.provider_source.github.requests.get') as mock_get, \
                 unittest.mock.patch('terrareg.provider_source.github.requests.post') as mock_post, \
                 unittest.mock.patch('terrareg.module_extractor.subprocess.check_output') as mock_clone:

                # Mock GitHub API calls for installation ID lookup
                mock_get.return_value.status_code = 200
                mock_get.return_value.json.return_value = {"id": mock_installation_id}

                # Mock token generation
                mock_post.return_value.status_code = 201
                mock_post.return_value.json.return_value = {"token": mock_access_token}

                # Mock git clone to succeed
                mock_clone.return_value = b""

                # Execute module extraction
                with terrareg.module_extractor.GitModuleExtractor(module_version=module_version) as extractor:
                    extractor._clone_repository()

                # Verify GitHub API was called to get installation ID
                assert mock_get.called
                call_args = mock_get.call_args
                assert f"/orgs/{namespace.name}/installation" in str(call_args)

                # Verify token generation was called
                assert mock_post.called
                post_args = mock_post.call_args
                assert f"/app/installations/{mock_installation_id}/access_tokens" in str(post_args)

                # Verify git clone was called with authenticated URL
                assert mock_clone.called
                clone_cmd = mock_clone.call_args[0][0]
                # Find the URL in the git command
                git_url = None
                for i, arg in enumerate(clone_cmd):
                    if "github.example.com" in arg:
                        git_url = arg
                        break
                assert git_url is not None, "GitHub URL not found in clone command"
                assert f"x-access-token:{mock_access_token}" in git_url
        finally:
            # Cleanup
            # ModuleVersion not in DB, no deletion needed
            module_provider.delete()
            # Module not in DB, no deletion needed
            namespace.delete()
            with terrareg.database.Database.get_connection() as conn:
                conn.execute(db.provider_source.delete(db.provider_source.c.name==provider_source_name))

    def test_module_extraction_fallback_to_basic_credentials_no_installation(self):
        """Test fallback to basic credentials when provider source has no GitHub installation"""
        # Create namespace
        namespace = terrareg.models.Namespace.create(
            name="test-org-fallback",
            display_name="Test Organization Fallback",
            type_=terrareg.namespace_type.NamespaceType.GITHUB_ORGANISATION
        )

        # Create provider_source with GitHub App config
        provider_source_name = "test-github-provider-source-fallback"
        db = terrareg.database.Database.get()
        with terrareg.database.Database.get_connection() as conn:
            conn.execute(db.provider_source.insert().values(
                name=provider_source_name,
                api_name="test-github-provider-source-fallback",
                provider_source_type=terrareg.provider_source_type.ProviderSourceType.GITHUB,
                config=db.encode_blob(json.dumps({
                    "base_url": "https://github.example.com",
                    "api_url": "https://api.github.example.com",
                    "client_id": "test-client-id",
                    "client_secret": "test-client-secret",
                    "login_button_text": "Test Login",
                    "private_key_path": "./test_key.pem",
                    "app_id": "test-app-id",
                    "default_installation_id": "test-default-installation-id",
                    "auto_generate_github_organisation_namespaces": False
                }))
            ))

        provider_source = terrareg.provider_source.factory.ProviderSourceFactory.get().get_provider_source_by_name(provider_source_name)

        # Create module and module_provider with provider_source linked
        module = terrareg.models.Module(namespace=namespace, name="test-module")
        module_provider = terrareg.models.ModuleProvider.get(module=module, name="aws", create=True)
        module_provider.update_attributes(provider_source_name=provider_source.name)

        # Set a git clone URL template
        mock_repo_url = "https://github.example.com/{namespace}/{module}-{provider}.git"
        with unittest.mock.patch('terrareg.config.Config.ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER', True):
            module_provider.update_attributes(repo_clone_url_template=mock_repo_url)

        # Create module version
        module_version = terrareg.models.ModuleVersion(module_provider=module_provider, version='1.0.0')

        try:
            with unittest.mock.patch('terrareg.provider_source.github.requests.get') as mock_get, \
                 unittest.mock.patch('terrareg.module_extractor.subprocess.check_output') as mock_clone, \
                 unittest.mock.patch('terrareg.config.Config.UPSTREAM_GIT_CREDENTIALS_USERNAME', 'fallback_user'), \
                 unittest.mock.patch('terrareg.config.Config.UPSTREAM_GIT_CREDENTIALS_PASSWORD', 'fallback_pass'):

                # Mock GitHub API to return no installation
                mock_get.return_value.status_code = 404
                mock_get.return_value.json.return_value = {"message": "Not found"}

                # Mock git clone to succeed
                mock_clone.return_value = b""

                # Execute module extraction
                with terrareg.module_extractor.GitModuleExtractor(module_version=module_version) as extractor:
                    extractor._clone_repository()

                # Verify GitHub API was called to get installation ID
                assert mock_get.called

                # Verify git clone was called with basic auth fallback
                assert mock_clone.called
                clone_cmd = mock_clone.call_args[0][0]
                # Find the URL in the git command
                git_url = None
                for i, arg in enumerate(clone_cmd):
                    if "github.example.com" in arg:
                        git_url = arg
                        break
                assert git_url is not None, "GitHub URL not found in clone command"
                assert "fallback_user:fallback_pass@" in git_url
        finally:
            # Cleanup
            module_provider.delete()
            namespace.delete()
            with terrareg.database.Database.get_connection() as conn:
                conn.execute(db.provider_source.delete(db.provider_source.c.name==provider_source_name))

    def test_module_extraction_github_user_namespace(self):
        """Test module extraction with GitHub user namespace"""
        # Create GitHub user namespace
        namespace = terrareg.models.Namespace.create(
            name="test-github-user",
            display_name="Test GitHub User",
            type_=terrareg.namespace_type.NamespaceType.GITHUB_USER
        )

        # Create provider_source with GitHub App config
        provider_source_name = "test-github-provider-source-user"
        db = terrareg.database.Database.get()
        with terrareg.database.Database.get_connection() as conn:
            conn.execute(db.provider_source.insert().values(
                name=provider_source_name,
                api_name="test-github-provider-source-user",
                provider_source_type=terrareg.provider_source_type.ProviderSourceType.GITHUB,
                config=db.encode_blob(json.dumps({
                    "base_url": "https://github.example.com",
                    "api_url": "https://api.github.example.com",
                    "client_id": "test-client-id",
                    "client_secret": "test-client-secret",
                    "login_button_text": "Test Login",
                    "private_key_path": "./test_key.pem",
                    "app_id": "test-app-id",
                    "default_installation_id": "test-default-installation-id",
                    "auto_generate_github_organisation_namespaces": False
                }))
            ))

        provider_source = terrareg.provider_source.factory.ProviderSourceFactory.get().get_provider_source_by_name(provider_source_name)

        # Create module and module_provider with provider_source linked
        module = terrareg.models.Module(namespace=namespace, name="test-module")
        module_provider = terrareg.models.ModuleProvider.get(module=module, name="aws", create=True)
        module_provider.update_attributes(provider_source_name=provider_source.name)

        # Set a git clone URL template
        mock_repo_url = "https://github.example.com/{namespace}/{module}-{provider}.git"
        with unittest.mock.patch('terrareg.config.Config.ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER', True):
            module_provider.update_attributes(repo_clone_url_template=mock_repo_url)

        # Create module version
        module_version = terrareg.models.ModuleVersion(module_provider=module_provider, version='1.0.0')

        # Mock GitHub API and git clone
        mock_installation_id = 987654321
        mock_access_token = "ghs_test_token_xyz789"

        try:
            with unittest.mock.patch('terrareg.provider_source.github.requests.get') as mock_get, \
                 unittest.mock.patch('terrareg.provider_source.github.requests.post') as mock_post, \
                 unittest.mock.patch('terrareg.module_extractor.subprocess.check_output') as mock_clone:

                # Mock GitHub API calls for installation ID lookup
                mock_get.return_value.status_code = 200
                mock_get.return_value.json.return_value = {"id": mock_installation_id}

                # Mock token generation
                mock_post.return_value.status_code = 201
                mock_post.return_value.json.return_value = {"token": mock_access_token}

                # Mock git clone to succeed
                mock_clone.return_value = b""

                # Execute module extraction
                with terrareg.module_extractor.GitModuleExtractor(module_version=module_version) as extractor:
                    extractor._clone_repository()

                # Verify GitHub API was called to get installation ID for user
                assert mock_get.called
                call_args = mock_get.call_args
                assert f"/users/{namespace.name}/installation" in str(call_args)

                # Verify token generation was called
                assert mock_post.called

                # Verify git clone was called with authenticated URL
                assert mock_clone.called
                clone_cmd = mock_clone.call_args[0][0]
                # Find the URL in the git command
                git_url = None
                for i, arg in enumerate(clone_cmd):
                    if "github.example.com" in arg:
                        git_url = arg
                        break
                assert git_url is not None, "GitHub URL not found in clone command"
                assert f"x-access-token:{mock_access_token}" in git_url
        finally:
            # Cleanup
            module_provider.delete()
            namespace.delete()
            with terrareg.database.Database.get_connection() as conn:
                conn.execute(db.provider_source.delete(db.provider_source.c.name==provider_source_name))

    def test_module_extraction_without_provider_source(self):
        """Test module extraction without provider source uses basic credentials"""
        # Create namespace without GitHub type
        namespace = terrareg.models.Namespace.create(
            name="test-plain-namespace",
            display_name="Test Plain Namespace"
        )

        # Create module and module_provider WITHOUT provider_source
        module = terrareg.models.Module(namespace=namespace, name="test-module")
        module_provider = terrareg.models.ModuleProvider.get(module=module, name="aws", create=True)

        # Set a git clone URL template
        mock_repo_url = "https://github.example.com/{namespace}/{module}-{provider}.git"
        with unittest.mock.patch('terrareg.config.Config.ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER', True):
            module_provider.update_attributes(repo_clone_url_template=mock_repo_url)

        # Create module version
        module_version = terrareg.models.ModuleVersion(module_provider=module_provider, version='1.0.0')

        try:
            with unittest.mock.patch('terrareg.module_extractor.subprocess.check_output') as mock_clone, \
                 unittest.mock.patch('terrareg.config.Config.UPSTREAM_GIT_CREDENTIALS_USERNAME', 'basic_user'), \
                 unittest.mock.patch('terrareg.config.Config.UPSTREAM_GIT_CREDENTIALS_PASSWORD', 'basic_pass'):

                # Mock git clone to succeed
                mock_clone.return_value = b""

                # Execute module extraction
                with terrareg.module_extractor.GitModuleExtractor(module_version=module_version) as extractor:
                    extractor._clone_repository()

                # Verify git clone was called with basic auth
                assert mock_clone.called
                clone_cmd = mock_clone.call_args[0][0]
                # Find the URL in the git command
                git_url = None
                for i, arg in enumerate(clone_cmd):
                    if "github.example.com" in arg:
                        git_url = arg
                        break
                assert git_url is not None, "GitHub URL not found in clone command"
                assert "basic_user:basic_pass@" in git_url
        finally:
            # Cleanup
            module_provider.delete()
            namespace.delete()


    def test_namespace_default_provider_source_used_for_module_extraction(self):
        """
        When a ModuleProvider has no provider_source but its Namespace has a default_provider_source,
        verify that module extraction uses the namespace's default provider source authentication.
        """
        # Create namespace with default_provider_source
        namespace = terrareg.models.Namespace.create(
            name="test-namespace-default",
            display_name="Test Namespace Default",
            type_=terrareg.namespace_type.NamespaceType.GITHUB_ORGANISATION
        )

        # Create provider_source for namespace
        provider_source_name = "test-namespace-default-ps"
        db = terrareg.database.Database.get()
        with terrareg.database.Database.get_connection() as conn:
            conn.execute(db.provider_source.insert().values(
                name=provider_source_name,
                api_name="test-namespace-default-ps",
                provider_source_type=terrareg.provider_source_type.ProviderSourceType.GITHUB,
                config=db.encode_blob(json.dumps({
                    "base_url": "https://github.example.com",
                    "api_url": "https://api.github.example.com",
                    "client_id": "test-client-id",
                    "client_secret": "test-client-secret",
                    "login_button_text": "Test Login",
                    "private_key_path": "./test_key.pem",
                    "app_id": "test-app-id",
                    "default_installation_id": "test-default-installation-id",
                    "auto_generate_github_organisation_namespaces": False
                }))
            ))

        # Set namespace default provider source
        with terrareg.database.Database.get_connection() as conn:
            conn.execute(db.namespace.update().where(
                db.namespace.c.namespace == namespace.name
            ).values(
                default_provider_source_name=provider_source_name
            ))

        # Create module_provider WITHOUT provider_source
        module = terrareg.models.Module(namespace=namespace, name="test-module")
        module_provider = terrareg.models.ModuleProvider.get(module=module, name="aws", create=True)

        mock_repo_url = "https://github.example.com/{namespace}/{module}-{provider}.git"
        with unittest.mock.patch('terrareg.config.Config.ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER', True):
            module_provider.update_attributes(repo_clone_url_template=mock_repo_url)

        module_version = terrareg.models.ModuleVersion(module_provider=module_provider, version='1.0.0')

        mock_installation_id = 111111111
        mock_access_token = "ghs_namespace_default_token"

        try:
            with unittest.mock.patch('terrareg.provider_source.github.requests.get') as mock_get, \
                 unittest.mock.patch('terrareg.provider_source.github.requests.post') as mock_post, \
                 unittest.mock.patch('terrareg.module_extractor.subprocess.check_output') as mock_clone:

                mock_get.return_value.status_code = 200
                mock_get.return_value.json.return_value = {"id": mock_installation_id}

                mock_post.return_value.status_code = 201
                mock_post.return_value.json.return_value = {"token": mock_access_token}

                mock_clone.return_value = b""

                with terrareg.module_extractor.GitModuleExtractor(module_version=module_version) as extractor:
                    extractor._clone_repository()

                # Verify namespace's provider source was used
                assert mock_get.called
                call_args = mock_get.call_args
                assert f"/orgs/{namespace.name}/installation" in str(call_args)

                assert mock_clone.called
                clone_cmd = mock_clone.call_args[0][0]
                git_url = None
                for i, arg in enumerate(clone_cmd):
                    if "github.example.com" in arg:
                        git_url = arg
                        break
                assert git_url is not None
                assert f"x-access-token:{mock_access_token}" in git_url
        finally:
            module_provider.delete()
            with terrareg.database.Database.get_connection() as conn:
                conn.execute(db.namespace.update().where(
                    db.namespace.c.namespace == namespace.name
                ).values(default_provider_source_name=None))
                namespace.delete()
                conn.execute(db.provider_source.delete(db.provider_source.c.name==provider_source_name))

    def test_module_provider_provider_source_overrides_namespace(self):
        """
        When both ModuleProvider.provider_source and Namespace.default_provider_source exist,
        verify that module extraction uses the ModuleProvider's provider source (not namespace).
        """
        # Create namespace with default_provider_source
        namespace = terrareg.models.Namespace.create(
            name="test-override-namespace",
            display_name="Test Override Namespace",
            type_=terrareg.namespace_type.NamespaceType.GITHUB_ORGANISATION
        )

        # Create two provider_sources
        namespace_ps_name = "test-namespace-override-ps"
        module_ps_name = "test-module-override-ps"
        db = terrareg.database.Database.get()

        with terrareg.database.Database.get_connection() as conn:
            # Namespace provider source
            conn.execute(db.provider_source.insert().values(
                name=namespace_ps_name,
                api_name="test-namespace-override-ps",
                provider_source_type=terrareg.provider_source_type.ProviderSourceType.GITHUB,
                config=db.encode_blob(json.dumps({
                    "base_url": "https://github.example.com",
                    "api_url": "https://api.github.example.com",
                    "client_id": "test-client-id",
                    "client_secret": "test-client-secret",
                    "login_button_text": "Test Login",
                    "private_key_path": "./test_key.pem",
                    "app_id": "test-app-id",
                    "default_installation_id": "test-default-installation-id",
                    "auto_generate_github_organisation_namespaces": False
                }))
            ))
            # Module provider source
            conn.execute(db.provider_source.insert().values(
                name=module_ps_name,
                api_name="test-module-override-ps",
                provider_source_type=terrareg.provider_source_type.ProviderSourceType.GITHUB,
                config=db.encode_blob(json.dumps({
                    "base_url": "https://github.example.com",
                    "api_url": "https://api.github.example.com",
                    "client_id": "test-client-id-2",
                    "client_secret": "test-client-secret-2",
                    "login_button_text": "Test Login 2",
                    "private_key_path": "./test_key.pem",
                    "app_id": "test-app-id-2",
                    "default_installation_id": "test-default-installation-id-2",
                    "auto_generate_github_organisation_namespaces": False
                }))
            ))

        # Set namespace default
        with terrareg.database.Database.get_connection() as conn:
            conn.execute(db.namespace.update().where(
                db.namespace.c.namespace == namespace.name
            ).values(
                default_provider_source_name=namespace_ps_name
            ))

        # Create module_provider WITH different provider_source
        module = terrareg.models.Module(namespace=namespace, name="test-module")
        module_provider = terrareg.models.ModuleProvider.get(module=module, name="aws", create=True)
        module_provider.update_attributes(provider_source_name=module_ps_name)

        mock_repo_url = "https://github.example.com/{namespace}/{module}-{provider}.git"
        with unittest.mock.patch('terrareg.config.Config.ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER', True):
            module_provider.update_attributes(repo_clone_url_template=mock_repo_url)

        module_version = terrareg.models.ModuleVersion(module_provider=module_provider, version='1.0.0')

        mock_installation_id = 222222222
        mock_access_token = "ghs_module_override_token"

        try:
            with unittest.mock.patch('terrareg.provider_source.github.requests.get') as mock_get, \
                 unittest.mock.patch('terrareg.provider_source.github.requests.post') as mock_post, \
                 unittest.mock.patch('terrareg.module_extractor.subprocess.check_output') as mock_clone:

                mock_get.return_value.status_code = 200
                mock_get.return_value.json.return_value = {"id": mock_installation_id}

                mock_post.return_value.status_code = 201
                mock_post.return_value.json.return_value = {"token": mock_access_token}

                mock_clone.return_value = b""

                with terrareg.module_extractor.GitModuleExtractor(module_version=module_version) as extractor:
                    extractor._clone_repository()

                # Verify module provider's source was used (not namespace's)
                assert mock_get.called
                # Should be called with module provider source's app_id
                assert mock_post.called

                assert mock_clone.called
                clone_cmd = mock_clone.call_args[0][0]
                git_url = None
                for i, arg in enumerate(clone_cmd):
                    if "github.example.com" in arg:
                        git_url = arg
                        break
                assert git_url is not None
                assert f"x-access-token:{mock_access_token}" in git_url
        finally:
            module_provider.delete()
            with terrareg.database.Database.get_connection() as conn:
                conn.execute(db.namespace.update().where(
                    db.namespace.c.namespace == namespace.name
                ).values(default_provider_source_name=None))
                namespace.delete()
                conn.execute(db.provider_source.delete(db.provider_source.c.name==namespace_ps_name))
                conn.execute(db.provider_source.delete(db.provider_source.c.name==module_ps_name))

    def test_fallback_to_basic_credentials_when_no_provider_source(self):
        """
        When neither ModuleProvider.provider_source nor Namespace.default_provider_source exist,
        verify that module extraction falls back to basic credentials.
        """
        # Create namespace WITHOUT default_provider_source
        namespace = terrareg.models.Namespace.create(
            name="test-fallback-namespace",
            display_name="Test Fallback Namespace"
        )

        # Create module_provider WITHOUT provider_source
        module = terrareg.models.Module(namespace=namespace, name="test-module")
        module_provider = terrareg.models.ModuleProvider.get(module=module, name="aws", create=True)

        mock_repo_url = "https://github.example.com/{namespace}/{module}-{provider}.git"
        with unittest.mock.patch('terrareg.config.Config.ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER', True):
            module_provider.update_attributes(repo_clone_url_template=mock_repo_url)

        module_version = terrareg.models.ModuleVersion(module_provider=module_provider, version='1.0.0')

        try:
            with unittest.mock.patch('terrareg.module_extractor.subprocess.check_output') as mock_clone, \
                 unittest.mock.patch('terrareg.config.Config.UPSTREAM_GIT_CREDENTIALS_USERNAME', 'fallback_user'), \
                 unittest.mock.patch('terrareg.config.Config.UPSTREAM_GIT_CREDENTIALS_PASSWORD', 'fallback_pass'):

                mock_clone.return_value = b""

                with terrareg.module_extractor.GitModuleExtractor(module_version=module_version) as extractor:
                    extractor._clone_repository()

                assert mock_clone.called
                clone_cmd = mock_clone.call_args[0][0]
                git_url = None
                for i, arg in enumerate(clone_cmd):
                    if "github.example.com" in arg:
                        git_url = arg
                        break
                assert git_url is not None
                assert "fallback_user:fallback_pass@" in git_url
        finally:
            module_provider.delete()
            namespace.delete()

    def test_backward_compatibility_existing_module_provider_provider_source(self):
        """
        Verify that existing ModuleProviders with provider_source_name continue to work
        exactly as before, regardless of namespace default.
        """
        # Create namespace WITHOUT default_provider_source
        namespace = terrareg.models.Namespace.create(
            name="test-compat-namespace",
            display_name="Test Compatibility Namespace",
            type_=terrareg.namespace_type.NamespaceType.GITHUB_ORGANISATION
        )

        # Create provider_source
        provider_source_name = "test-compat-ps"
        db = terrareg.database.Database.get()
        with terrareg.database.Database.get_connection() as conn:
            conn.execute(db.provider_source.insert().values(
                name=provider_source_name,
                api_name="test-compat-ps",
                provider_source_type=terrareg.provider_source_type.ProviderSourceType.GITHUB,
                config=db.encode_blob(json.dumps({
                    "base_url": "https://github.example.com",
                    "api_url": "https://api.github.example.com",
                    "client_id": "test-client-id",
                    "client_secret": "test-client-secret",
                    "login_button_text": "Test Login",
                    "private_key_path": "./test_key.pem",
                    "app_id": "test-app-id",
                    "default_installation_id": "test-default-installation-id",
                    "auto_generate_github_organisation_namespaces": False
                }))
            ))

        # Create module_provider WITH provider_source (traditional setup)
        module = terrareg.models.Module(namespace=namespace, name="test-module")
        module_provider = terrareg.models.ModuleProvider.get(module=module, name="aws", create=True)
        module_provider.update_attributes(provider_source_name=provider_source_name)

        mock_repo_url = "https://github.example.com/{namespace}/{module}-{provider}.git"
        with unittest.mock.patch('terrareg.config.Config.ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER', True):
            module_provider.update_attributes(repo_clone_url_template=mock_repo_url)

        module_version = terrareg.models.ModuleVersion(module_provider=module_provider, version='1.0.0')

        mock_installation_id = 333333333
        mock_access_token = "ghs_compat_token"

        try:
            with unittest.mock.patch('terrareg.provider_source.github.requests.get') as mock_get, \
                 unittest.mock.patch('terrareg.provider_source.github.requests.post') as mock_post, \
                 unittest.mock.patch('terrareg.module_extractor.subprocess.check_output') as mock_clone:

                mock_get.return_value.status_code = 200
                mock_get.return_value.json.return_value = {"id": mock_installation_id}

                mock_post.return_value.status_code = 201
                mock_post.return_value.json.return_value = {"token": mock_access_token}

                mock_clone.return_value = b""

                with terrareg.module_extractor.GitModuleExtractor(module_version=module_version) as extractor:
                    extractor._clone_repository()

                assert mock_get.called
                assert mock_clone.called
                clone_cmd = mock_clone.call_args[0][0]
                git_url = None
                for i, arg in enumerate(clone_cmd):
                    if "github.example.com" in arg:
                        git_url = arg
                        break
                assert git_url is not None
                assert f"x-access-token:{mock_access_token}" in git_url
        finally:
            module_version.delete()
            module_provider.delete()
            module.delete()
            namespace.delete()
            with terrareg.database.Database.get_connection() as conn:
                conn.execute(db.provider_source.delete(db.provider_source.c.name==provider_source_name))

    def test_provider_source_failure_falls_back_to_namespace(self):
        """
        When ModuleProvider.provider_source authentication fails (e.g., no installation),
        verify that it falls back to Namespace.default_provider_source.
        """
        # Create namespace with default_provider_source
        namespace = terrareg.models.Namespace.create(
            name="test-fallback-src-namespace",
            display_name="Test Fallback Source Namespace",
            type_=terrareg.namespace_type.NamespaceType.GITHUB_ORGANISATION
        )

        # Create two provider_sources
        namespace_ps_name = "test-fallback-namespace-ps"
        module_ps_name = "test-fallback-module-ps"
        db = terrareg.database.Database.get()

        with terrareg.database.Database.get_connection() as conn:
            conn.execute(db.provider_source.insert().values(
                name=namespace_ps_name,
                api_name="test-fallback-namespace-ps",
                provider_source_type=terrareg.provider_source_type.ProviderSourceType.GITHUB,
                config=db.encode_blob(json.dumps({
                    "base_url": "https://github.example.com",
                    "api_url": "https://api.github.example.com",
                    "client_id": "test-client-id",
                    "client_secret": "test-client-secret",
                    "login_button_text": "Test Login",
                    "private_key_path": "./test_key.pem",
                    "app_id": "test-app-id",
                    "default_installation_id": "test-default-installation-id",
                    "auto_generate_github_organisation_namespaces": False
                }))
            ))
            conn.execute(db.provider_source.insert().values(
                name=module_ps_name,
                api_name="test-fallback-module-ps",
                provider_source_type=terrareg.provider_source_type.ProviderSourceType.GITHUB,
                config=db.encode_blob(json.dumps({
                    "base_url": "https://github.example.com",
                    "api_url": "https://api.github.example.com",
                    "client_id": "test-client-id-fail",
                    "client_secret": "test-client-secret-fail",
                    "login_button_text": "Test Login Fail",
                    "private_key_path": "./test_key.pem",
                    "app_id": "test-app-id-fail",
                    "default_installation_id": "test-default-installation-id-fail",
                    "auto_generate_github_organisation_namespaces": False
                }))
            ))

        # Set namespace default
        with terrareg.database.Database.get_connection() as conn:
            conn.execute(db.namespace.update().where(
                db.namespace.c.namespace == namespace.name
            ).values(
                default_provider_source_name=namespace_ps_name
            ))

        # Create module_provider with provider_source that has NO installation
        module = terrareg.models.Module(namespace=namespace, name="test-module")
        module_provider = terrareg.models.ModuleProvider.get(module=module, name="aws", create=True)
        module_provider.update_attributes(provider_source_name=module_ps_name)

        mock_repo_url = "https://github.example.com/{namespace}/{module}-{provider}.git"
        with unittest.mock.patch('terrareg.config.Config.ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER', True):
            module_provider.update_attributes(repo_clone_url_template=mock_repo_url)

        module_version = terrareg.models.ModuleVersion(module_provider=module_provider, version='1.0.0')

        mock_installation_id = 444444444
        mock_access_token = "ghs_fallback_token"

        try:
            with unittest.mock.patch('terrareg.provider_source.github.requests.get') as mock_get, \
                 unittest.mock.patch('terrareg.provider_source.github.requests.post') as mock_post, \
                 unittest.mock.patch('terrareg.module_extractor.subprocess.check_output') as mock_clone:

                # First call (module provider source) returns 404, second call (namespace) returns 200
                call_count = [0]
                def mock_get_side_effect(*args, **kwargs):
                    call_count[0] += 1
                    if call_count[0] == 1:
                        # Module provider source - no installation
                        mock_resp = unittest.mock.MagicMock()
                        mock_resp.status_code = 404
                        mock_resp.json.return_value = {"message": "Not found"}
                        return mock_resp
                    else:
                        # Namespace default - has installation
                        mock_resp = unittest.mock.MagicMock()
                        mock_resp.status_code = 200
                        mock_resp.json.return_value = {"id": mock_installation_id}
                        return mock_resp

                mock_get.side_effect = mock_get_side_effect

                mock_post.return_value.status_code = 201
                mock_post.return_value.json.return_value = {"token": mock_access_token}

                mock_clone.return_value = b""

                with terrareg.module_extractor.GitModuleExtractor(module_version=module_version) as extractor:
                    extractor._clone_repository()

                # Verify both attempts were made
                assert mock_get.call_count == 2

                # Verify git clone was called with namespace's token (fallback worked)
                assert mock_clone.called
                clone_cmd = mock_clone.call_args[0][0]
                git_url = None
                for i, arg in enumerate(clone_cmd):
                    if "github.example.com" in arg:
                        git_url = arg
                        break
                assert git_url is not None
                assert f"x-access-token:{mock_access_token}" in git_url
        finally:
            module_version.delete()
            module_provider.delete()
            module.delete()
            with terrareg.database.Database.get_connection() as conn:
                conn.execute(db.namespace.update().where(
                    db.namespace.c.namespace == namespace.name
                ).values(default_provider_source_name=None))
                namespace.delete()
                conn.execute(db.provider_source.delete(db.provider_source.c.name==namespace_ps_name))
                conn.execute(db.provider_source.delete(db.provider_source.c.name==module_ps_name))

    def test_ssh_urls_unmodified_by_authentication_logic(self):
        """
        Verify that SSH URLs (ssh://git@...) are never modified, regardless of
        provider source configuration at any level.
        """
        # Create namespace with default_provider_source
        namespace = terrareg.models.Namespace.create(
            name="test-ssh-namespace",
            display_name="Test SSH Namespace"
        )

        provider_source_name = "test-ssh-ps"
        db = terrareg.database.Database.get()
        with terrareg.database.Database.get_connection() as conn:
            conn.execute(db.provider_source.insert().values(
                name=provider_source_name,
                api_name="test-ssh-ps",
                provider_source_type=terrareg.provider_source_type.ProviderSourceType.GITHUB,
                config=db.encode_blob(json.dumps({
                    "base_url": "https://github.example.com",
                    "api_url": "https://api.github.example.com",
                    "client_id": "test-client-id",
                    "client_secret": "test-client-secret",
                    "login_button_text": "Test Login",
                    "private_key_path": "./test_key.pem",
                    "app_id": "test-app-id",
                    "default_installation_id": "test-default-installation-id",
                    "auto_generate_github_organisation_namespaces": False
                }))
            ))

        with terrareg.database.Database.get_connection() as conn:
            conn.execute(db.namespace.update().where(
                db.namespace.c.namespace == namespace.name
            ).values(
                default_provider_source_name=provider_source_name
            ))

        module = terrareg.models.Module(namespace=namespace, name="test-module")
        module_provider = terrareg.models.ModuleProvider.get(module=module, name="aws", create=True)
        module_provider.update_attributes(provider_source_name=provider_source_name)

        # Use SSH URL template
        ssh_url_template = "ssh://git@github.example.com/{namespace}/{module}-{provider}.git"
        with unittest.mock.patch('terrareg.config.Config.ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER', True):
            module_provider.update_attributes(repo_clone_url_template=ssh_url_template)

        module_version = terrareg.models.ModuleVersion(module_provider=module_provider, version='1.0.0')

        try:
            with unittest.mock.patch('terrareg.provider_source.github.requests.get') as mock_get, \
                 unittest.mock.patch('terrareg.module_extractor.subprocess.check_output') as mock_clone:

                mock_clone.return_value = b""

                with terrareg.module_extractor.GitModuleExtractor(module_version=module_version) as extractor:
                    extractor._clone_repository()

                # Verify GitHub API was NOT called for SSH URLs
                assert not mock_get.called

                # Verify git clone was called with unmodified SSH URL
                assert mock_clone.called
                clone_cmd = mock_clone.call_args[0][0]
                git_url = None
                for i, arg in enumerate(clone_cmd):
                    if "github.example.com" in arg:
                        git_url = arg
                        break
                assert git_url is not None
                # SSH URL should be unchanged
                assert "ssh://git@github.example.com/" in git_url
                assert "x-access-token:" not in git_url
        finally:
            module_provider.delete()
            with terrareg.database.Database.get_connection() as conn:
                conn.execute(db.namespace.update().where(
                    db.namespace.c.namespace == namespace.name
                ).values(default_provider_source_name=None))
                namespace.delete()
                conn.execute(db.provider_source.delete(db.provider_source.c.name==provider_source_name))
