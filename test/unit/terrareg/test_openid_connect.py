
import re
from unittest import mock
import pytest

import terrareg.config
from terrareg.openid_connect import OpenidConnect
from test.unit.terrareg import TerraregUnitTest

class TestOpenidConnect(TerraregUnitTest):

    _METADATA_CONFIG = None

    @pytest.mark.parametrize('config_values,expected_result', [
        ({'OPENID_CONNECT_CLIENT_ID': 'testclientid',
          'OPENID_CONNECT_CLIENT_SECRET': 'testclientsecret',
          'OPENID_CONNECT_ISSUER': 'https://testissuer',
          'DOMAIN_NAME': 'unittest.local'},
         True),
        ({'OPENID_CONNECT_CLIENT_ID': None,
          'OPENID_CONNECT_CLIENT_SECRET': 'testclientsecret',
          'OPENID_CONNECT_ISSUER': 'https://testissuer',
          'DOMAIN_NAME': 'unittest.local'},
         False),
        ({'OPENID_CONNECT_CLIENT_ID': 'testclientid',
          'OPENID_CONNECT_CLIENT_SECRET': None,
          'OPENID_CONNECT_ISSUER': 'https://testissuer',
          'DOMAIN_NAME': 'unittest.local'},
         False),
        ({'OPENID_CONNECT_CLIENT_ID': 'testclientid',
          'OPENID_CONNECT_CLIENT_SECRET': None,
          'OPENID_CONNECT_ISSUER': 'https://testissuer',
          'DOMAIN_NAME': 'unittest.local'},
         False),
        ({'OPENID_CONNECT_CLIENT_ID': 'testclientid',
          'OPENID_CONNECT_CLIENT_SECRET': 'testclientsecret',
          'OPENID_CONNECT_ISSUER': 'https://testissuer',
          'DOMAIN_NAME': None},
         False)
    ])
    def test_is_enabled(self, config_values, expected_result):
        """Test is_enabled method"""
        class MockConfig:
            pass
        for key in config_values:
            setattr(MockConfig, key, config_values[key])
        with mock.patch('terrareg.config.Config', MockConfig):
            assert OpenidConnect.is_enabled() == expected_result

    def test_get_client(self):
        """Test get_client method"""
        mock_client = mock.MagicMock()
        with mock.patch('oauthlib.oauth2.WebApplicationClient', mock.MagicMock(return_value=mock_client)) as mock_web_application_class, \
                mock.patch('terrareg.config.Config.OPENID_CONNECT_CLIENT_ID', 'unittestclientid'):
            client = OpenidConnect.get_client()
        
        assert client is mock_client
        mock_web_application_class.assert_called_once_with('unittestclientid')

    def test_get_redirect_url(self):
        """test get_redirect_url"""
        with mock.patch('terrareg.config.Config.DOMAIN_NAME', 'unittest.domain'):
            assert OpenidConnect.get_redirect_url() == 'https://unittest.domain/openid/callback'

    def test_obtain_issuer_metadata(self):
        """Obtain wellknown metadata from issuer"""

        mock_request_repsonse = mock.MagicMock()
        mock_request_repsonse.json = mock.MagicMock(return_value={'example': 'metadata'})

        with mock.patch('requests.get', mock.MagicMock(return_value=mock_request_repsonse)) as mock_requests_get, \
                mock.patch('terrareg.config.Config.OPENID_CONNECT_ISSUER', 'https://idpprovider.com/default'), \
                mock.patch('terrareg.openid_connect.OpenidConnect.is_enabled', mock.MagicMock(return_value=True)):
            config = OpenidConnect.obtain_issuer_metadata()

            assert config == {'example': 'metadata'}

            mock_requests_get.assert_called_once_with('https://idpprovider.com/default/.well-known/openid-configuration')

            mock_requests_get.reset_mock()

            # Call a second time, to assert that requests.get isn't called again,
            # as the value should be cached
            assert OpenidConnect.obtain_issuer_metadata() == {'example': 'metadata'}

            mock_requests_get.assert_not_called()

    def test_generate_state(self):
        """Test generate_state"""
        # Test function several times, asserting length and characters
        previous_vals = []
        for i in range(5):
            state = OpenidConnect.generate_state()

            assert state not in previous_vals
            previous_vals.append(state)
            
            assert len(state) == 24
            assert re.match(r'^[a-zA-Z0-9]+$', state)

    @pytest.mark.parametrize('is_enabled,auth_endpoint,expect_called', [
        (True, 'https://testprovider.com/default/auth', True),
        (False, 'https://testprovider.com/default/auth', False),
        (True, None, False)
    ])
    def test_get_authorize_redirect_url(self, is_enabled, auth_endpoint, expect_called):
        """Get authorize URL to redirect user to for authentication"""
        mock_client = mock.MagicMock()
        mock_client.prepare_request_uri = mock.MagicMock(return_value='https://redirecturls.com/goes?here=true')
        with mock.patch('terrareg.openid_connect.OpenidConnect.is_enabled', is_enabled), \
                mock.patch('terrareg.openid_connect.OpenidConnect.obtain_issuer_metadata',
                           mock.MagicMock(return_value={'authorization_endpoint': auth_endpoint})), \
                mock.patch('terrareg.openid_connect.OpenidConnect.generate_state',
                           mock.MagicMock(return_value='unitteststate')), \
                mock.patch('terrareg.openid_connect.OpenidConnect.get_client',
                           mock.MagicMock(return_value=mock_client)), \
                mock.patch('terrareg.openid_connect.OpenidConnect.get_redirect_url',
                           mock.MagicMock(return_value='https://examplecallback.url')):

            url, state = OpenidConnect.get_authorize_redirect_url()

            if expect_called:
                assert url == 'https://redirecturls.com/goes?here=true'
                assert state == 'unitteststate'

                mock_client.prepare_request_uri.assert_called_once_with(
                    auth_endpoint,
                    redirect_uri='https://examplecallback.url',
                    scope=['openid', 'profile'],
                    state='unitteststate')
            else:
                assert url == None
                assert state == None

    @pytest.mark.parametrize('uri,valid_state,metadata,valid_response', [
        ('https://localhost.com/callback?token=something&state=teststate', 'teststate', {'token_endpoint': 'https://token.com/endpoint'}, True),
        ('https://localhost.com/callback?token=something&state=teststate', 'teststate', None, False),
        ('https://localhost.com/callback?token=something&state=teststate', 'teststate', {'notatokenendpoint': 'somethingelse'}, False)
    ])
    def test_fetch_access_token(cls, uri, valid_state, metadata, valid_response):
        """Fetch access token from OpenID issuer"""

        mock_client = mock.MagicMock()
        mock_client.parse_request_uri_response = mock.MagicMock(return_value={'code': 'unittestcode'})
        mock_client.prepare_request_body = mock.MagicMock(return_value='testbodyfortokenendpointrequest')
        mock_client.parse_request_body_response = mock.MagicMock(return_value={
            'expires_at': 1853432, 'access_token': 'unittestaccesstoken',
            'id_token': 'unittestidtoken', 'expires_in': 13432})

        mock_post_response = mock.MagicMock()
        mock_post_response.text = "TestTokenEndpointResponse"


        with mock.patch('terrareg.openid_connect.OpenidConnect.get_client',
                        mock.MagicMock(return_value=mock_client)), \
                mock.patch('terrareg.config.Config.OPENID_CONNECT_CLIENT_ID', 'unittestclientid'), \
                mock.patch('terrareg.config.Config.OPENID_CONNECT_CLIENT_SECRET', 'unittestclientsecret'), \
                mock.patch('terrareg.openid_connect.OpenidConnect.obtain_issuer_metadata',
                           mock.MagicMock(return_value=metadata)), \
                mock.patch('terrareg.openid_connect.OpenidConnect.get_redirect_url',
                           mock.MagicMock(return_value='https://examplecallback.url')), \
                mock.patch('requests.post', mock.MagicMock(return_value=mock_post_response)) as mock_post:

            response = OpenidConnect.fetch_access_token(uri=uri, valid_state=valid_state)

            mock_client.parse_request_uri_response.assert_called_once_with(uri=uri, state=valid_state)

            mock_client.prepare_request_body.assert_called_once_with(
                code='unittestcode',
                client_id='unittestclientid',
                client_secret='unittestclientsecret',
                redirect_uri='https://examplecallback.url'
            )


            if valid_response:

                mock_post.assert_called_once_with(
                    'https://token.com/endpoint',
                    'testbodyfortokenendpointrequest',
                    headers={'Content-Type': 'application/x-www-form-urlencoded'})

                mock_client.parse_request_body_response.assert_called_once_with(
                    "TestTokenEndpointResponse"
                )

                assert response == {
                    'expires_at': 1853432, 'access_token': 'unittestaccesstoken',
                    'id_token': 'unittestidtoken', 'expires_in': 13432
                }
            else:
                mock_post.assert_not_called()
                mock_client.parse_request_body_response.assert_not_called()
                assert response == None
