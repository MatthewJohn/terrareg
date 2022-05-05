
from test.unit.terrareg import client, TerraregUnitTest


class TestApiTerraregHealth(TerraregUnitTest):
    """Test ApiTerraregHealth resource."""

    def test_with_no_params(self, client):
        """Test endpoint without query parameters"""
        res = client.get('/v1/terrareg/health')
        assert res.status_code == 200
        assert res.json == {
            'message': 'Ok'
        }

    def test_with_post(self, client):
        """Test invalid post request"""
        res = client.post('/v1/terrareg/health')
        assert res.status_code == 405
