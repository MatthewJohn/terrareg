
from operator import mod
import unittest.mock

import pytest

from test.unit.terrareg import (
    TEST_MODULE_DATA, MockModule, MockModuleProvider, MockNamespace, mocked_server_namespace_fixture,
    setup_test_data, TerraregUnitTest
)
from test import client, app_context, test_request_context


class TestApiTerraregGlobalUsageStats(TerraregUnitTest):
    """Test global usage stats endpoint"""

    def test_global_usage_stats(
            self, app_context,
            test_request_context,
            client
        ):
        """Test update of repository URL."""
        with client, \
                unittest.mock.patch('terrareg.analytics.AnalyticsEngine.get_global_module_usage') as mocked_get_global_module_usage, \
                unittest.mock.patch('terrareg.models.ModuleProvider.get_total_count') as mocked_get_total_count:

            def get_global_module_usage(include_empty_auth_token=False):
                mock_data = {
                    'namespace/testmodule1/provider': 5,
                    'namespace/testmodule2/anotherprovider': 2,
                    'anothernamespace/testmodule2/anotherprovider': 8
                }
                if include_empty_auth_token:
                    mock_data['namespace/testmodule2/anotherprovider'] = 9
                    mock_data['namespace/emptyauthtoken/provider'] = 3
                return mock_data

            mocked_get_global_module_usage.side_effect = get_global_module_usage
            mocked_get_total_count.return_value = 23

            res = client.get('/v1/terrareg/analaytics/global/usage_stats')

            assert res.json == {
                'module_provider_count': 23,
                'module_provider_usage_breakdown_with_auth_token': {
                    'namespace/testmodule1/provider': 5,
                    'namespace/testmodule2/anotherprovider': 2,
                    'anothernamespace/testmodule2/anotherprovider': 8
                },
                'module_provider_usage_count_with_auth_token': 15,
                'module_provider_usage_including_empty_auth_token': {
                    'namespace/testmodule1/provider': 5,
                    'namespace/testmodule2/anotherprovider': 9,
                    'anothernamespace/testmodule2/anotherprovider': 8,
                    'namespace/emptyauthtoken/provider': 3
                },
                'module_provider_usage_count_including_empty_auth_token': 25
            }
            assert res.status_code == 200
        
            mocked_get_total_count.assert_called_once_with()
            mocked_get_global_module_usage.assert_called()
