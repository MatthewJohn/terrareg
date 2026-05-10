"""Unit tests for provider source hierarchical lookup"""

import json
import unittest.mock
import pytest

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


class TestProviderSourceUpdateMethods(TerraregIntegrationTest):
    """Test provider source update methods"""

    def test_namespace_update_default_provider_source_set_valid(self):
        """Test setting valid provider source on namespace"""
        namespace = terrareg.models.Namespace.create(
            name="test-update-valid",
            display_name="Test Update Valid"
        )

        # Create provider_source
        provider_source_name = "test-ps-valid"
        db = terrareg.database.Database.get()
        with terrareg.database.Database.get_connection() as conn:
            conn.execute(db.provider_source.insert().values(
                name=provider_source_name,
                api_name="test-ps-valid",
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

        try:
            # Set valid provider source
            namespace.update_default_provider_source(provider_source_name)

            # Verify it was set
            assert namespace.default_provider_source is not None
            assert namespace.default_provider_source.name == provider_source_name
        finally:
            with terrareg.database.Database.get_connection() as conn:
                conn.execute(db.namespace.update().where(
                    db.namespace.c.namespace == namespace.name
                ).values(default_provider_source_name=None))
                namespace.delete()
                conn.execute(db.provider_source.delete(db.provider_source.c.name==provider_source_name))

    def test_namespace_update_default_provider_source_unset_empty_string(self):
        """Test unsetting provider source with empty string"""
        namespace = terrareg.models.Namespace.create(
            name="test-unset",
            display_name="Test Unset"
        )

        provider_source_name = "test-ps-unset"
        db = terrareg.database.Database.get()
        with terrareg.database.Database.get_connection() as conn:
            conn.execute(db.provider_source.insert().values(
                name=provider_source_name,
                api_name="test-ps-unset",
                provider_source_type=terrareg.provider_source_type.ProviderSourceType.GITHUB,
                config=db.encode_blob(json.dumps({
                    "base_url": "https://github.example.com",
                    "api_url": "https://api.github.example.com",
                    "client_id": "test-client-id",
                    "client_secret": "test-client-secret"
                }))
            ))
            # Set provider source first
            conn.execute(db.namespace.update().where(
                db.namespace.c.namespace == namespace.name
            ).values(default_provider_source_name=provider_source_name))

        try:
            # Verify it's set
            assert namespace.default_provider_source is not None

            # Unset with empty string
            namespace.update_default_provider_source("")

            # Verify it was unset
            assert namespace.default_provider_source is None
        finally:
            namespace.delete()
            with terrareg.database.Database.get_connection() as conn:
                conn.execute(db.provider_source.delete(db.provider_source.c.name==provider_source_name))

    def test_namespace_update_default_provider_source_invalid_provider(self):
        """Test setting invalid provider source raises error"""
        namespace = terrareg.models.Namespace.create(
            name="test-invalid",
            display_name="Test Invalid"
        )

        try:
            # Try to set invalid provider source
            with pytest.raises(terrareg.errors.InvalidProviderSourceNameError):
                namespace.update_default_provider_source("nonexistent-provider")
        finally:
            namespace.delete()

    def test_namespace_update_default_provider_source_no_change(self):
        """Test None parameter doesn't change anything"""
        namespace = terrareg.models.Namespace.create(
            name="test-no-change",
            display_name="Test No Change"
        )

        try:
            # Initially None
            assert namespace.default_provider_source is None

            # Call with None - should not raise and should not change
            namespace.update_default_provider_source(None)

            # Verify still None
            assert namespace.default_provider_source is None
        finally:
            namespace.delete()

    def test_module_provider_update_provider_source_set_valid(self):
        """Test setting valid provider source on module provider"""
        namespace = terrareg.models.Namespace.create(
            name="test-mp-valid",
            display_name="Test MP Valid"
        )

        provider_source_name = "test-mp-ps-valid"
        db = terrareg.database.Database.get()
        with terrareg.database.Database.get_connection() as conn:
            conn.execute(db.provider_source.insert().values(
                name=provider_source_name,
                api_name="test-mp-ps-valid",
                provider_source_type=terrareg.provider_source_type.ProviderSourceType.GITHUB,
                config=db.encode_blob(json.dumps({
                    "base_url": "https://github.example.com",
                    "api_url": "https://api.github.example.com",
                    "client_id": "test-client-id",
                    "client_secret": "test-client-secret"
                }))
            ))

        module = terrareg.models.Module(namespace=namespace, name="test-module")
        module_provider = terrareg.models.ModuleProvider.get(module=module, name="aws", create=True)

        try:
            # Set valid provider source
            module_provider.update_provider_source(provider_source_name)

            # Verify it was set
            assert module_provider.provider_source is not None
            assert module_provider.provider_source.name == provider_source_name
        finally:
            module_provider.delete()
            namespace.delete()
            with terrareg.database.Database.get_connection() as conn:
                conn.execute(db.provider_source.delete(db.provider_source.c.name==provider_source_name))

    def test_module_provider_update_provider_source_unset_empty_string(self):
        """Test unsetting provider source with empty string"""
        namespace = terrareg.models.Namespace.create(
            name="test-mp-unset",
            display_name="Test MP Unset"
        )

        provider_source_name = "test-mp-ps-unset"
        db = terrareg.database.Database.get()
        with terrareg.database.Database.get_connection() as conn:
            conn.execute(db.provider_source.insert().values(
                name=provider_source_name,
                api_name="test-mp-ps-unset",
                provider_source_type=terrareg.provider_source_type.ProviderSourceType.GITHUB,
                config=db.encode_blob(json.dumps({
                    "base_url": "https://github.example.com",
                    "api_url": "https://api.github.example.com",
                    "client_id": "test-client-id",
                    "client_secret": "test-client-secret"
                }))
            ))

        module = terrareg.models.Module(namespace=namespace, name="test-module")
        module_provider = terrareg.models.ModuleProvider.get(module=module, name="aws", create=True)

        # Set provider source first
        module_provider.update_attributes(provider_source_name=provider_source_name)

        try:
            # Verify it's set
            assert module_provider.provider_source is not None

            # Unset with empty string
            module_provider.update_provider_source("")

            # Verify it was unset
            assert module_provider.provider_source is None
        finally:
            module_provider.delete()
            namespace.delete()
            with terrareg.database.Database.get_connection() as conn:
                conn.execute(db.provider_source.delete(db.provider_source.c.name==provider_source_name))

    def test_module_provider_update_provider_source_invalid_provider(self):
        """Test setting invalid provider source raises error"""
        namespace = terrareg.models.Namespace.create(
            name="test-mp-invalid",
            display_name="Test MP Invalid"
        )

        module = terrareg.models.Module(namespace=namespace, name="test-module")
        module_provider = terrareg.models.ModuleProvider.get(module=module, name="aws", create=True)

        try:
            # Try to set invalid provider source
            with pytest.raises(terrareg.errors.InvalidProviderSourceNameError):
                module_provider.update_provider_source("nonexistent-provider")
        finally:
            module_provider.delete()
            namespace.delete()

    def test_module_provider_update_inheritance_disabled_enable(self):
        """Test enabling inheritance (setting to false)"""
        namespace = terrareg.models.Namespace.create(
            name="test-inherit-enable",
            display_name="Test Inherit Enable"
        )

        module = terrareg.models.Module(namespace=namespace, name="test-module")
        module_provider = terrareg.models.ModuleProvider.get(module=module, name="aws", create=True)

        # Set to disabled first
        module_provider.update_attributes(provider_source_inheritance_disabled=True)

        try:
            # Verify it's disabled
            assert module_provider.provider_source_inheritance_disabled is True

            # Enable inheritance (set to False)
            module_provider.update_provider_source_inheritance_disabled(False)

            # Verify it's enabled
            assert module_provider.provider_source_inheritance_disabled is False
        finally:
            module_provider.delete()
            namespace.delete()

    def test_module_provider_update_inheritance_disabled_disable(self):
        """Test disabling inheritance (setting to true)"""
        namespace = terrareg.models.Namespace.create(
            name="test-inherit-disable",
            display_name="Test Inherit Disable"
        )

        module = terrareg.models.Module(namespace=namespace, name="test-module")
        module_provider = terrareg.models.ModuleProvider.get(module=module, name="aws", create=True)

        try:
            # Verify it's enabled by default
            assert module_provider.provider_source_inheritance_disabled is False

            # Disable inheritance
            module_provider.update_provider_source_inheritance_disabled(True)

            # Verify it's disabled
            assert module_provider.provider_source_inheritance_disabled is True
        finally:
            module_provider.delete()
            namespace.delete()

    def test_module_provider_update_inheritance_disabled_no_change(self):
        """Test None parameter doesn't change anything"""
        namespace = terrareg.models.Namespace.create(
            name="test-inherit-no-change",
            display_name="Test Inherit No Change"
        )

        module = terrareg.models.Module(namespace=namespace, name="test-module")
        module_provider = terrareg.models.ModuleProvider.get(module=module, name="aws", create=True)

        try:
            # Verify default is False
            assert module_provider.provider_source_inheritance_disabled is False

            # Call with None - should not change
            module_provider.update_provider_source_inheritance_disabled(None)

            # Verify still False
            assert module_provider.provider_source_inheritance_disabled is False
        finally:
            module_provider.delete()
            namespace.delete()

    def test_get_effective_provider_source_with_inheritance_disabled(self):
        """Test that disabled inheritance prevents namespace fallback"""
        namespace = terrareg.models.Namespace.create(
            name="test-inherit-effective",
            display_name="Test Inherit Effective"
        )

        namespace_ps_name = "test-namespace-ps-effective"
        db = terrareg.database.Database.get()

        with terrareg.database.Database.get_connection() as conn:
            conn.execute(db.provider_source.insert().values(
                name=namespace_ps_name,
                api_name="test-namespace-ps-effective",
                provider_source_type=terrareg.provider_source_type.ProviderSourceType.GITHUB,
                config=db.encode_blob(json.dumps({
                    "base_url": "https://github.example.com",
                    "api_url": "https://api.github.example.com",
                    "client_id": "test-client-id",
                    "client_secret": "test-client-secret"
                }))
            ))
            # Set namespace default
            conn.execute(db.namespace.update().where(
                db.namespace.c.namespace == namespace.name
            ).values(default_provider_source_name=namespace_ps_name))

        module = terrareg.models.Module(namespace=namespace, name="test-module")
        module_provider = terrareg.models.ModuleProvider.get(module=module, name="aws", create=True)

        try:
            # Without inheritance disabled, should get namespace default
            result = module_provider.get_effective_provider_source()
            assert result is not None
            assert result.name == namespace_ps_name

            # Disable inheritance
            module_provider.update_provider_source_inheritance_disabled(True)

            # With inheritance disabled, should return None (no source at any level)
            result = module_provider.get_effective_provider_source()
            assert result is None
        finally:
            module_provider.delete()
            with terrareg.database.Database.get_connection() as conn:
                conn.execute(db.namespace.update().where(
                    db.namespace.c.namespace == namespace.name
                ).values(default_provider_source_name=None))
                namespace.delete()
                conn.execute(db.provider_source.delete(db.provider_source.c.name==namespace_ps_name))
