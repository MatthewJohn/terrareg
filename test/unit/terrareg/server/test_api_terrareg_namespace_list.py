
from unittest import mock

import pytest

import terrareg.registry_resource_type
from test.unit.terrareg import TerraregUnitTest, setup_test_data, mock_models
from test import client
import terrareg.result_data


class TestApiTerraregNamespaceList(TerraregUnitTest):
    """Test ApiTerraregNamespaceList resource."""

    def test_with_no_namespaces(self, client):
        """Test endpoint when no namespaces are present."""
        with mock.patch('terrareg.models.Namespace.get_all') as mocked_namespace_get_all:
            mocked_namespace_get_all.return_value = terrareg.result_data.ResultData(offset=0, limit=10, rows=[], count=0)

            res = client.get('/v1/terrareg/namespaces')

            assert res.status_code == 200
            assert res.json == []

            mocked_namespace_get_all.assert_called_once_with(
                only_published=False, limit=None, offset=0,
                resource_type=terrareg.registry_resource_type.RegistryResourceType.MODULE
            )

    def test_with_no_namespaces_and_limit_offset(self, client):
        """Test endpoint when no namespaces are present with limit and offset."""
        with mock.patch('terrareg.models.Namespace.get_all') as mocked_namespace_get_all:
            mocked_namespace_get_all.return_value = terrareg.result_data.ResultData(offset=0, limit=10, rows=[], count=0)

            res = client.get('/v1/terrareg/namespaces?only_published=true&offset=12&limit=14')

            assert res.status_code == 200
            assert res.json == {
                'meta': {'current_offset': 0, 'limit': 10},
                'namespaces': []
            }

            mocked_namespace_get_all.assert_called_once_with(
                only_published=True, limit=14, offset=12,
                resource_type=terrareg.registry_resource_type.RegistryResourceType.MODULE
            )

    @pytest.mark.parametrize('query_string, expected_type', [
        ('', terrareg.registry_resource_type.RegistryResourceType.MODULE),
        ('&type=module', terrareg.registry_resource_type.RegistryResourceType.MODULE),
        ('&type=provider', terrareg.registry_resource_type.RegistryResourceType.PROVIDER),
    ])
    def test_with_namespace_types(self, query_string, expected_type, client):
        """Test endpoint when no namespaces are present with limit and offset."""
        with mock.patch('terrareg.models.Namespace.get_all') as mocked_namespace_get_all:
            mocked_namespace_get_all.return_value = terrareg.result_data.ResultData(offset=0, limit=10, rows=[], count=0)

            res = client.get(f'/v1/terrareg/namespaces?only_published=true&offset=12&limit=14{query_string}')

            assert res.status_code == 200
            assert res.json == {
                'meta': {'current_offset': 0, 'limit': 10},
                'namespaces': []
            }

            mocked_namespace_get_all.assert_called_once_with(
                only_published=True, limit=14, offset=12,
                resource_type=expected_type
            )

    @setup_test_data()
    def test_with_namespaces_present(self, client, mock_models):
        """Test endpoint with existing namespaces."""
        res = client.get('/v1/terrareg/namespaces')
        assert res.status_code == 200
        assert res.json == [
            {'name': 'testnamespace', 'view_href': '/modules/testnamespace', 'display_name': None},
            {'name': 'moduledetails', 'view_href': '/modules/moduledetails', 'display_name': None},
            {'name': 'secondtestnamespace', 'view_href': '/modules/secondtestnamespace', 'display_name': None},
            {'name': 'smallernamespacelist', 'view_href': '/modules/smallernamespacelist', 'display_name': 'Smaller Namespace List'},
            {'name': 'moduleextraction', 'view_href': '/modules/moduleextraction', 'display_name': None},
            {'name': 'emptynamespace', 'view_href': '/modules/emptynamespace', 'display_name': None}
        ]

    def test_unauthenticated(self, client, mock_models):
        """Test unauthenticated call to API"""
        def call_endpoint():
            return client.get('/v1/terrareg/namespaces')

        self._test_unauthenticated_read_api_endpoint_test(call_endpoint)

    @setup_test_data()
    @pytest.mark.parametrize('offset, limit, prev_offset, next_offset, expected_namespaces', [
        (0, 1, None, 1, [{'display_name': None, 'name': 'testnamespace', 'view_href': '/modules/testnamespace'}]),
        (1, 1, 0, 2, [{'display_name': None, 'name': 'moduledetails', 'view_href': '/modules/moduledetails'}])
    ])
    def test_with_offset_and_limit(self, offset, limit, prev_offset, next_offset, expected_namespaces, client, mock_models):
        """Test calling endpoint with offset and limits"""
        res = client.get(f'/v1/terrareg/namespaces?offset={offset}&limit={limit}')
        assert res.status_code == 200

        expected_meta = {'current_offset': offset, 'limit': limit}
        if prev_offset is not None:
            expected_meta['prev_offset'] = prev_offset
        if next_offset is not None:
            expected_meta['next_offset'] = next_offset

        assert res.json == {
            'meta': expected_meta,
            'namespaces': expected_namespaces
        }
