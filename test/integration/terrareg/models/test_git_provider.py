
from contextlib import contextmanager
import json
import unittest.mock

import pytest

import terrareg.models
from test.integration.terrareg import TerraregIntegrationTest
import terrareg.database
import terrareg.errors


@pytest.fixture
def set_config():

    @contextmanager
    def inner(config):
        with unittest.mock.patch('terrareg.config.Config.GIT_PROVIDER_CONFIG', json.dumps(config)):
            yield
    return inner


@pytest.fixture
def delete_existing_git_providers():
    db = terrareg.database.Database.get()
    with db.get_connection() as conn:
        conn.execute(db.git_provider.delete())


class TestGitProvider(TerraregIntegrationTest):
    """Test GitProvider model"""

    _GIT_PROVIDER_DATA = {}

    def test_initialise_from_config(self, set_config, delete_existing_git_providers):
        """Test initialise_from_config"""
        config = [
            {
                "name": "Test One",
                "base_url": "https://example.com/{namespace}/{module}",
                "clone_url": "ssh://git@example.com/{namespace}/{module}.git",
                "browse_url": "https://example.com/{namespace}/{module}/tree/{tag}/{path}"
            },
            {
                "name": "Test Two",
                "base_url": "https://example.com/{namespace}/modules",
                "clone_url": "ssh://git@example.com/{namespace}/modules.git",
                "browse_url": "https://example.com/{namespace}/modules/tree/{tag}/{path}",
                "git_path": "/{module}/{provider}"
            },
        ]
        with set_config(config):
            terrareg.models.GitProvider.initialise_from_config()

            providers = terrareg.models.GitProvider.get_all()
            assert len(providers) == 2

            # Sort by name
            providers.sort(key=lambda x: x.name)

            # Check first provider
            assert providers[0].name == "Test One"
            assert providers[0].base_url_template == "https://example.com/{namespace}/{module}"
            assert providers[0].clone_url_template == "ssh://git@example.com/{namespace}/{module}.git"
            assert providers[0].browse_url_template == "https://example.com/{namespace}/{module}/tree/{tag}/{path}"
            assert providers[0].git_path_template is None

            # Check second
            assert providers[1].name == "Test Two"
            assert providers[1].base_url_template == "https://example.com/{namespace}/modules"
            assert providers[1].clone_url_template == "ssh://git@example.com/{namespace}/modules.git"
            assert providers[1].browse_url_template == "https://example.com/{namespace}/modules/tree/{tag}/{path}"
            assert providers[1].git_path_template == "/{module}/{provider}"

    @pytest.mark.parametrize('url_suffix, git_path, valid', [
        # With/without provider
        ('{namespace}/{module}', '', True),
        ('{namespace}/{module}/{provider}', '', True),
        ('{namespace}/{module}', '{provider}', True),

        # Without module
        ('{namespace}', '', False),
        # Module in git path
        ('{namespace}', '{module}', True),

        # Without namespace
        ('{module}', '', False),
        # Namespace in git path
        ('{module}', '{namespace}', True),

        # Without any
        ('blah', '', False),
        # module and namespace in git path
        ('blah', '{namespace}-{module}', True),

        # Invalid placeholders
        ('{namespace}-{module}-{somethingelse}', '', False),
        ('{namespace}-{module}', '{somethingelse}', False),
    ])
    def test_initialise_from_config_missing_placeholders(self, url_suffix, git_path, valid, delete_existing_git_providers, set_config):
        """Test initialise_from_config with missing placeholders"""
        config = [
            {
                "name": "Test One",
                "base_url": f"https://example.com/{url_suffix}",
                "clone_url": f"ssh://git@example.com/{url_suffix}",
                "browse_url": f"https://example.com/{url_suffix}/tree/{{tag}}/{{path}}",
                "git_path": git_path
            }
        ]
        with set_config(config):
            if valid:
                terrareg.models.GitProvider.initialise_from_config()

                providers = terrareg.models.GitProvider.get_all()
                assert len(providers) == 1

                # Check first provider
                assert providers[0].name == "Test One"
                assert providers[0].base_url_template == f"https://example.com/{url_suffix}"
                assert providers[0].clone_url_template == f"ssh://git@example.com/{url_suffix}"
                assert providers[0].browse_url_template == f"https://example.com/{url_suffix}/tree/{{tag}}/{{path}}"
                assert providers[0].git_path_template == (git_path or None)
        
            else:
                with pytest.raises(terrareg.errors.RepositoryUrlParseError):
                    terrareg.models.GitProvider.initialise_from_config()
