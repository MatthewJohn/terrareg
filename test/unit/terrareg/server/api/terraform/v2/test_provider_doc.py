
import pytest

from test.integration.terrareg import TerraregIntegrationTest
from test import client, app_context, test_request_context


class TestApiProviderDoc(TerraregIntegrationTest):
    """Test ApiV2ProviderDoc endpoint"""

    @pytest.mark.parametrize('query_string', [
        '',
        '?output=markdown'
    ])
    def test_endpoint(self, query_string, client):
        """Test endpoint."""

        res = client.get(f'/v2/provider-docs/6347{query_string}')
        assert res.status_code == 200
        assert res.json == {
            'data': {
                'attributes': {
                    'category': 'resources',
                    'content': """
# Some Title!

## Second title

This module:

 * Creates something
 * Does something else

and it _really_ *does* work!
""",
                    'language': 'hcl',
                    'path': 'resources/new-thing.md',
                    'slug': 'some_new_resource',
                    'subcategory': 'some-second-subcategory',
                    'title': 'multiple_versions_thing_new',
                    'truncated': False
                },
                'id': '6347',
                'links': {'self': '/v2/provider-docs/6347'},
                'type': 'provider-docs'
            }
        }

    def test_endpoint_with_html(self, client):
        """Test endpoint with HTML output."""

        res = client.get(f'/v2/provider-docs/6347?output=html')
        assert res.status_code == 200
        assert res.json == {
            'data': {
                'attributes': {
                    'category': 'resources',
                    'content': """
<h1 id="terrareg-anchor-resourcesnew-thingmd-some-title">Some Title!</h1>
<h2 id="terrareg-anchor-resourcesnew-thingmd-second-title">Second title</h2>
<p>This module:</p>
<ul>
<li>Creates something</li>
<li>Does something else</li>
</ul>
<p>and it <em>really</em> <em>does</em> work!</p>
""".strip(),
                    'language': 'hcl',
                    'path': 'resources/new-thing.md',
                    'slug': 'some_new_resource',
                    'subcategory': 'some-second-subcategory',
                    'title': 'multiple_versions_thing_new',
                    'truncated': False
                },
                'id': '6347',
                'links': {'self': '/v2/provider-docs/6347'},
                'type': 'provider-docs'
            }
        }

    def test_endpoint_with_non_existent(self, client):
        """Test endpoint with namespace."""

        res = client.get(f'/v2/provider-docs/9999999')
        assert res.status_code == 404
        assert res.json == {'errors': ['Not Found']}

    def test_unauthenticated(self, client):
        """Test unauthenticated call to API"""
        def call_endpoint():
            return client.get('/v2/provider-docs/6347')

        self._test_unauthenticated_read_api_endpoint_test(call_endpoint)