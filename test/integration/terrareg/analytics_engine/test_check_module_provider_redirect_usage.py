
from datetime import datetime, timedelta
import json
from unittest import mock

import pytest
import sqlalchemy

import terrareg.analytics
from terrareg.database import Database
from terrareg.models import Example, ExampleFile, Module, ModuleDetails, ModuleProviderRedirect, Namespace, ModuleProvider, ModuleVersion
import terrareg.errors
from test.integration.terrareg import TerraregIntegrationTest


class TestCheckModuleProviderRedirectUsage(TerraregIntegrationTest):
    """Test analytics check_module_provider_redirect_usage method"""

    def _create_analytics(self, namespace, module, provider, version, token, timestamp):
        """Create analytics entry"""
        mock_get_datetime_now = mock.MagicMock(return_value=timestamp)
        with mock.patch('terrareg.analytics.AnalyticsEngine.get_datetime_now', mock_get_datetime_now):
            terrareg.analytics.AnalyticsEngine.record_module_version_download(
                namespace_name=namespace,
                module_name=module,
                provider_name=provider,
                module_version=version,
                analytics_token=token,
                terraform_version='1.1.1',
                user_agent='',
                auth_token=None
            )

    def teardown_method(self, method):
        # Remove module provider, namespace and analytics
        namespace = Namespace.get('testredirect')
        provider = ModuleProvider.get(Module(namespace, 'newname'), 'testprovider')
        provider.delete()
        namespace.delete()

        db = Database.get()
        with db.get_connection() as conn:
            conn.execute(db.analytics.delete())

        return super().teardown_method(method)

    def _setup_test_analytics(self):
        """Setup test data"""
        namespace = Namespace.create('testredirect')
        module = Module(namespace, 'testredirectdelete')
        provider = ModuleProvider.create(module, 'testprovider')
        version = ModuleVersion(provider, '1.1.1')
        version.prepare_module()

        # Record analytics token against old name before name change
        ## Older token
        self._create_analytics(
            namespace='testredirect',
            module='testredirectdelete',
            provider='testprovider',
            version=version,
            token='beforemove',
            timestamp=datetime(year=2023, month=3, day=1, hour=5, minute=5, second=0)
        )
        ## Newer call to token, to ensure latest version of the token is provided
        self._create_analytics(
            namespace='testredirect',
            module='testredirectdelete',
            provider='testprovider',
            version=version,
            token='beforemove',
            timestamp=datetime(year=2023, month=3, day=5, hour=5, minute=5, second=0)
        )

        # Rename provider
        provider = provider.update_name(namespace=namespace, module_name='newname', provider_name='testprovider')

        # Record analytics token against old name
        self._create_analytics(
            namespace='testredirect',
            module='testredirectdelete',
            provider='testprovider',
            version=version,
            token='oldnameaftermove',
            timestamp=datetime(year=2023, month=3, day=7, hour=5, minute=5, second=0)
        )
        self._create_analytics(
            namespace='testredirect',
            module='testredirectdelete',
            provider='testprovider',
            version=version,
            token='testmigrate',
            timestamp=datetime(year=2023, month=3, day=9, hour=5, minute=5, second=0)
        )

        # Re-record one pre-existing analytics token against new name
        # Record new analytics token against new name
        self._create_analytics(
            namespace='testredirect',
            module='newname',
            provider='testprovider',
            version=version,
            token='testmigrate',
            timestamp=datetime(year=2023, month=3, day=11, hour=5, minute=5, second=0)
        )
        self._create_analytics(
            namespace='testredirect',
            module='newname',
            provider='testprovider',
            version=version,
            token='onlynewname',
            timestamp=datetime(year=2023, month=3, day=13, hour=5, minute=5, second=0)
        )
        return provider

    def test_basic_use(self):
        """Test with variety of analytics, checking that the correct rows are returned."""
        provider = self._setup_test_analytics()

        redirects = ModuleProviderRedirect.get_by_module_provider(provider)
        assert len(redirects) == 1
        redirect = redirects[0]

        with mock.patch('terrareg.config.Config.REDIRECT_DELETION_LOOKBACK_DAYS', -1):
            res = terrareg.analytics.AnalyticsEngine.check_module_provider_redirect_usage(redirect)
        
        expected_results = {
            # Latest instance of 'beforemove' token
            'beforemove': datetime(year=2023, month=3, day=5, hour=5, minute=5, second=0),
            'oldnameaftermove': datetime(year=2023, month=3, day=7, hour=5, minute=5, second=0)
        }

        assert len(res) == len(expected_results)

        for row in res:
            assert row[0] in expected_results
            assert (expected_results[row[0]] - timedelta(minutes=1)) < row[1] < (expected_results[row[0]] + timedelta(minutes=1))
            del expected_results[row[0]]

    @pytest.mark.parametrize('lookback_days, expected_results', [
        # Expect all results using 14 day's worth of look-back
        (14, {
            'beforemove': datetime(year=2023, month=3, day=5, hour=5, minute=5, second=0),
            'oldnameaftermove': datetime(year=2023, month=3, day=7, hour=5, minute=5, second=0)
        }),
        # Expect all results from 10 days, as only the old version of 'before' remove should be
        # excluded
        (10, {
            'beforemove': datetime(year=2023, month=3, day=5, hour=5, minute=5, second=0),
            'oldnameaftermove': datetime(year=2023, month=3, day=7, hour=5, minute=5, second=0)
        }),
        # Boundary test for 'beforeremove', which should not appear
        (9, {
            'oldnameaftermove': datetime(year=2023, month=3, day=7, hour=5, minute=5, second=0)
        }),
        # Boundary test for 'oldnameaftermove', which should not appear
        (8, {
            'oldnameaftermove': datetime(year=2023, month=3, day=7, hour=5, minute=5, second=0)
        }),
        (7, {
            # No results
        }),
    ])
    def test_lookback_period(self, lookback_days, expected_results):
        """Test look-back period to determine that the correct results are returned."""

        provider = self._setup_test_analytics()

        redirects = ModuleProviderRedirect.get_by_module_provider(provider)
        assert len(redirects) == 1
        redirect = redirects[0]

        # Mock look-back period to 
        with mock.patch('terrareg.config.Config.REDIRECT_DELETION_LOOKBACK_DAYS', lookback_days), \
                mock.patch('terrareg.analytics.AnalyticsEngine.get_datetime_now',
                           mock.MagicMock(return_value=datetime(year=2023, month=3, day=15, hour=4, minute=0, second=0))):
            res = terrareg.analytics.AnalyticsEngine.check_module_provider_redirect_usage(redirect)

        assert len(res) == len(expected_results)

        for row in res:
            assert row[0] in expected_results
            assert (expected_results[row[0]] - timedelta(minutes=1)) < row[1] < (expected_results[row[0]] + timedelta(minutes=1))
            del expected_results[row[0]]

