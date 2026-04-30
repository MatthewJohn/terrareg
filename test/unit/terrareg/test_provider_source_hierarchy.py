"""Unit tests for provider source hierarchical lookup"""

import json
import unittest.mock

import terrareg.config
import terrareg.database
import terrareg.models
import terrareg.namespace_type
import terrareg.provider_source.factory
import terrareg.provider_source_type
from test.integration.terrareg import TerraregIntegrationTest


class TestProviderSourceHierarchy(TerraregIntegrationTest):
    """Test provider source hierarchical lookup methods"""

    def test_namespace_default_provider_source_property_returns_none_when_not_set(self):
        """Verify Namespace.default_provider_source returns None when not set"""
        namespace = terrareg.models.Namespace.create(
            name="test-no-default",
            display_name="Test No Default"
        )

        try:
            # Verify property returns None
            result = namespace.default_provider_source
            assert result is None, "Expected None when default_provider_source_name is not set"
        finally:
            namespace.delete()

    def test_namespace_default_provider_source_property_returns_provider_source_when_set(self):
        """Verify Namespace.default_provider_source returns provider source when set"""
        namespace = terrareg.models.Namespace.create(
            name="test-with-default",
            display_name="Test With Default",
            type_=terrareg.namespace_type.NamespaceType.GITHUB_ORGANISATION
        )

        # Create provider_source
        provider_source_name = "test-default-ps"
        db = terrareg.database.Database.get()
        with terrareg.database.Database.get_connection() as conn:
            conn.execute(db.provider_source.insert().values(
                name=provider_source_name,
                api_name="test-default-ps",
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

        try:
            # Verify property returns provider source
            result = namespace.default_provider_source
            assert result is not None, "Expected provider source when default_provider_source_name is set"
            assert result.name == provider_source_name, f"Expected {provider_source_name}, got {result.name}"
        finally:
            with terrareg.database.Database.get_connection() as conn:
                conn.execute(db.namespace.update().where(
                    db.namespace.c.namespace == namespace.name
                ).values(default_provider_source_name=None))
                namespace.delete()
                conn.execute(db.provider_source.delete(db.provider_source.c.name==provider_source_name))

    def test_module_provider_get_effective_provider_source_returns_module_provider_source(self):
        """Verify get_effective_provider_source returns module provider source when set"""
        namespace = terrareg.models.Namespace.create(
            name="test-module-priority",
            display_name="Test Module Priority"
        )

        # Create provider_source
        provider_source_name = "test-module-ps"
        db = terrareg.database.Database.get()
        with terrareg.database.Database.get_connection() as conn:
            conn.execute(db.provider_source.insert().values(
                name=provider_source_name,
                api_name="test-module-ps",
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

        module = terrareg.models.Module(namespace=namespace, name="test-module")
        module_provider = terrareg.models.ModuleProvider.get(module=module, name="aws", create=True)
        module_provider.update_attributes(provider_source_name=provider_source_name)

        try:
            # Verify get_effective_provider_source returns module provider's source
            result = module_provider.get_effective_provider_source()
            assert result is not None, "Expected provider source from module provider"
            assert result.name == provider_source_name, f"Expected {provider_source_name}, got {result.name}"
        finally:
            module_provider.delete()
            namespace.delete()
            with terrareg.database.Database.get_connection() as conn:
                conn.execute(db.provider_source.delete(db.provider_source.c.name==provider_source_name))

    def test_module_provider_get_effective_provider_source_falls_back_to_namespace(self):
        """Verify get_effective_provider_source falls back to namespace default when module provider has none"""
        namespace = terrareg.models.Namespace.create(
            name="test-namespace-fallback",
            display_name="Test Namespace Fallback"
        )

        # Create provider_source
        provider_source_name = "test-namespace-fallback-ps"
        db = terrareg.database.Database.get()
        with terrareg.database.Database.get_connection() as conn:
            conn.execute(db.provider_source.insert().values(
                name=provider_source_name,
                api_name="test-namespace-fallback-ps",
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

        # Set namespace default
        with terrareg.database.Database.get_connection() as conn:
            conn.execute(db.namespace.update().where(
                db.namespace.c.namespace == namespace.name
            ).values(
                default_provider_source_name=provider_source_name
            ))

        module = terrareg.models.Module(namespace=namespace, name="test-module")
        module_provider = terrareg.models.ModuleProvider.get(module=module, name="aws", create=True)
        # Note: NOT setting provider_source_name on module_provider

        try:
            # Verify get_effective_provider_source returns namespace default
            result = module_provider.get_effective_provider_source()
            assert result is not None, "Expected provider source from namespace default"
            assert result.name == provider_source_name, f"Expected {provider_source_name}, got {result.name}"
        finally:
            module_provider.delete()
            with terrareg.database.Database.get_connection() as conn:
                conn.execute(db.namespace.update().where(
                    db.namespace.c.namespace == namespace.name
                ).values(default_provider_source_name=None))
                namespace.delete()
                conn.execute(db.provider_source.delete(db.provider_source.c.name==provider_source_name))

    def test_module_provider_get_effective_provider_source_returns_none_when_no_source(self):
        """Verify get_effective_provider_source returns None when neither module provider nor namespace has source"""
        namespace = terrareg.models.Namespace.create(
            name="test-no-source",
            display_name="Test No Source"
        )

        module = terrareg.models.Module(namespace=namespace, name="test-module")
        module_provider = terrareg.models.ModuleProvider.get(module=module, name="aws", create=True)

        try:
            # Verify get_effective_provider_source returns None
            result = module_provider.get_effective_provider_source()
            assert result is None, "Expected None when no provider source is configured"
        finally:
            module_provider.delete()
            namespace.delete()

    def test_module_provider_get_effective_provider_source_module_priority_over_namespace(self):
        """Verify get_effective_provider_source prioritizes module provider source over namespace default"""
        namespace = terrareg.models.Namespace.create(
            name="test-priority-namespace",
            display_name="Test Priority Namespace"
        )

        # Create two provider_sources
        namespace_ps_name = "test-priority-namespace-ps"
        module_ps_name = "test-priority-module-ps"
        db = terrareg.database.Database.get()

        with terrareg.database.Database.get_connection() as conn:
            conn.execute(db.provider_source.insert().values(
                name=namespace_ps_name,
                api_name="test-priority-namespace-ps",
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
                api_name="test-priority-module-ps",
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

        module = terrareg.models.Module(namespace=namespace, name="test-module")
        module_provider = terrareg.models.ModuleProvider.get(module=module, name="aws", create=True)
        module_provider.update_attributes(provider_source_name=module_ps_name)

        try:
            # Verify get_effective_provider_source returns module provider's source (not namespace's)
            result = module_provider.get_effective_provider_source()
            assert result is not None, "Expected provider source"
            assert result.name == module_ps_name, f"Expected {module_ps_name} (module provider), got {result.name}"
            assert result.name != namespace_ps_name, "Should not return namespace's provider source"
        finally:
            module_provider.delete()
            with terrareg.database.Database.get_connection() as conn:
                conn.execute(db.namespace.update().where(
                    db.namespace.c.namespace == namespace.name
                ).values(default_provider_source_name=None))
                namespace.delete()
                conn.execute(db.provider_source.delete(db.provider_source.c.name==namespace_ps_name))
                conn.execute(db.provider_source.delete(db.provider_source.c.name==module_ps_name))
