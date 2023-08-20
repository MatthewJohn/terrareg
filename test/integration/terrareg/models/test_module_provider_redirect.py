
from datetime import datetime
import json
from unittest import mock

import pytest
import sqlalchemy

import terrareg.analytics
from terrareg.audit import AuditEvent
from terrareg.audit_action import AuditAction
from terrareg.database import Database
from terrareg.models import Example, ExampleFile, Module, ModuleDetails, ModuleProviderRedirect, Namespace, ModuleProvider, ModuleVersion
import terrareg.errors
from test.integration.terrareg import TerraregIntegrationTest


class TestModuleProviderRedirect(TerraregIntegrationTest):

    def _setup_offending_analytics_token(self):
        """Setup namespace, module provider and analytics tokens that will cause redirect deletion failure"""
        namespace = Namespace.create('testredirect')
        module = Module(namespace, 'testredirectdelete')
        provider = ModuleProvider.create(module, 'testprovider')
        version = ModuleVersion(provider, '1.1.1')
        version.prepare_module()

        # Record analytics token against old name before name change
        terrareg.analytics.AnalyticsEngine.record_module_version_download(
            namespace_name='testredirect',
            module_name='testredirectdelete',
            provider_name='testprovider',
            module_version=version,
            analytics_token='beforemove',
            terraform_version='1.1.1',
            user_agent='',
            auth_token=None
        )

        # Rename provider
        provider = provider.update_name(namespace=namespace, module_name='newname', provider_name='testprovider')

        # Record analytics token against old name
        terrareg.analytics.AnalyticsEngine.record_module_version_download(
            namespace_name='testredirect',
            module_name='testredirectdelete',
            provider_name='testprovider',
            module_version=version,
            analytics_token='oldnameaftermove',
            terraform_version='1.1.1',
            user_agent='',
            auth_token=None
        )
        terrareg.analytics.AnalyticsEngine.record_module_version_download(
            namespace_name='testredirect',
            module_name='testredirectdelete',
            provider_name='testprovider',
            module_version=version,
            analytics_token='testmigrate',
            terraform_version='1.1.1',
            user_agent='',
            auth_token=None
        )

        # Re-record one pre-existing analytics token against new name
        # Record new analytics token against new name
        terrareg.analytics.AnalyticsEngine.record_module_version_download(
            namespace_name='testredirect',
            module_name='newname',
            provider_name='testprovider',
            module_version=version,
            analytics_token='testmigrate',
            terraform_version='1.1.1',
            user_agent='',
            auth_token=None,
        )
        terrareg.analytics.AnalyticsEngine.record_module_version_download(
            namespace_name='testredirect',
            module_name='newname',
            provider_name='testprovider',
            module_version=version,
            analytics_token='onlynewname',
            terraform_version='1.1.1',
            user_agent='',
            auth_token=None,
        )

        return provider, namespace, version

    def setup_method(self, method):
        """Delete any analytics data"""
        db = Database.get()
        with db.get_connection() as conn:
            conn.execute(db.analytics.delete())
        return super().setup_method(method)

    def teardown_method(self, method):
        """Delete any analytics data"""
        db = Database.get()
        with db.get_connection() as conn:
            conn.execute(db.analytics.delete())

        return super().teardown_method(method)

    def test_delete_with_offending_analytics_migrated(self):
        """Test deleting a ModuleProviderRedirect with analytics that have all migrated to new name"""
        # Create namespace/module provider
        provider, namespace, version = self._setup_offending_analytics_token()

        try:
            # Add analytics for remaining 2 tokens to new name
            for token in ['oldnameaftermove', 'beforemove']:
                terrareg.analytics.AnalyticsEngine.record_module_version_download(
                    namespace_name='testredirect',
                    module_name='newname',
                    provider_name='testprovider',
                    module_version=version,
                    analytics_token=token,
                    terraform_version='1.1.1',
                    user_agent='',
                    auth_token=None,
                )

            # Delete any pre-existing audit events
            self._delete_audit_events()

            redirects = ModuleProviderRedirect.get_by_module_provider(provider)
            assert len(redirects) == 1
            redirect = redirects[0]

            redirect.delete()

            # Ensure redirect was not deleted
            redirects = ModuleProviderRedirect.get_by_module_provider(provider)
            assert len(redirects) == 0

            # Ensure audit event is created
            audit_events, _, _ = AuditEvent.get_events()
            assert len(audit_events) == 1
            assert audit_events[0][3] == AuditAction.MODULE_PROVIDER_REDIRECT_DELETE
            assert audit_events[0][4] == "ModuleProviderRedirect"
            # Check redirect ID is correct
            assert audit_events[0][5] == "testredirect/testredirectdelete/testprovider"
            # Check ID of module provider is correct
            assert audit_events[0][6] == str(provider.pk)

        finally:
            provider.delete()
            namespace.delete()

    def test_delete_with_offending_analytics(self):
        """Test deleting a ModuleProviderRedirect with analytics that are still using the redirect"""
        # Create namespace/module provider
        provider, namespace, _ = self._setup_offending_analytics_token()

        try:
            redirects = ModuleProviderRedirect.get_by_module_provider(provider)
            assert len(redirects) == 1
            redirect = redirects[0]
            redirect_pk = redirect._pk

            self._delete_audit_events()

            with pytest.raises(terrareg.errors.ModuleProviderRedirectInUseError):
                redirect.delete()

            # Ensure redirect was not deleted
            redirects = ModuleProviderRedirect.get_by_module_provider(provider)
            assert len(redirects) == 1
            redirect = redirects[0]

            assert redirect._pk == redirect_pk

            # Ensure no audit events were created
            audit_events, _, _ = AuditEvent.get_events()
            assert len(audit_events) == 0
        finally:
            provider.delete()
            namespace.delete()

    def test_delete_with_force_with_offending_analytics_with_force_disabled(self):
        """
        Test deleting a ModuleProviderRedirect with analytics that are still
        using the redirect using force, with forceful deletion disabled by config
        """
        with mock.patch("terrareg.config.Config.ALLOW_FORCEFUL_MODULE_PROVIDER_REDIRECT_DELETION", False):
            # Create namespace/module provider
            provider, namespace, _ = self._setup_offending_analytics_token()

            try:
                redirects = ModuleProviderRedirect.get_by_module_provider(provider)
                assert len(redirects) == 1
                redirect = redirects[0]

                with pytest.raises(terrareg.errors.ModuleProviderRedirectForceDeletionNotAllowedError):
                    redirect.delete(force=True)

                # Ensure redirect has been removed
                redirects = ModuleProviderRedirect.get_by_module_provider(provider)
                assert len(redirects) == 1
            finally:
                provider.delete()
                namespace.delete()

    def test_delete_with_internal_force_with_offending_analytics(self):
        """
        Test deleting a ModuleProviderRedirect with analytics that are still
        using the redirect using internal force, with configuration disallowing normal force
        """

        with mock.patch("terrareg.config.Config.ALLOW_FORCEFUL_MODULE_PROVIDER_REDIRECT_DELETION", False):
            # Create namespace/module provider
            provider, namespace, _ = self._setup_offending_analytics_token()

            try:
                redirects = ModuleProviderRedirect.get_by_module_provider(provider)
                assert len(redirects) == 1
                redirect = redirects[0]

                redirect.delete(internal_force=True)

                # Ensure redirect has been removed
                redirects = ModuleProviderRedirect.get_by_module_provider(provider)
                assert len(redirects) == 0
            finally:
                provider.delete()
                namespace.delete()


    def test_delete_with_force_with_offending_analytics(self):
        """Test deleting a ModuleProviderRedirect with analytics that are still using the redirect using force, with configuration allowing forceful deletion"""
        with mock.patch("terrareg.config.Config.ALLOW_FORCEFUL_MODULE_PROVIDER_REDIRECT_DELETION", True):

            # Create namespace/module provider
            provider, namespace, _ = self._setup_offending_analytics_token()

            try:
                redirects = ModuleProviderRedirect.get_by_module_provider(provider)
                assert len(redirects) == 1
                redirect = redirects[0]

                # Forceful deletion should be allowed
                redirect.delete(force=True)

                # Ensure redirect has been removed
                redirects = ModuleProviderRedirect.get_by_module_provider(provider)
                assert len(redirects) == 0
            finally:
                provider.delete()
                namespace.delete()
