
import re
from unittest import mock
import pytest
from requests import request

from werkzeug.datastructures import ImmutableMultiDict

import terrareg.config
from terrareg.saml import Saml2
from test.unit.terrareg import TerraregUnitTest

class TestSaml(TerraregUnitTest):

    @classmethod
    def test_is_enabled(cls):
        """Whether SAML auithentication is enabled"""
        config = terrareg.config.Config()
        return (config.DOMAIN_NAME is not None and
                config.SAML2_ENTITY_ID is not None and
                config.SAML2_IDP_METADATA_URL is not None and
                config.SAML2_PUBLIC_KEY is not None and
                config.SAML2_PRIVATE_KEY is not None)

    @pytest.mark.parametrize('config_values,expected_result', [
        ({'SAML2_ENTITY_ID': 'testclientid',
          'SAML2_PUBLIC_KEY': '--- saml2publickey ---',
          'SAML2_PRIVATE_KEY': '--- SAML2PRIVATE KEY! ---',
          'SAML2_IDP_METADATA_URL': 'https://testissuer',
          'DOMAIN_NAME': 'unittest.local'},
         True),
        ({'SAML2_ENTITY_ID': None,
          'SAML2_PUBLIC_KEY': '--- saml2publickey ---',
          'SAML2_PRIVATE_KEY': '--- SAML2PRIVATE KEY! ---',
          'SAML2_IDP_METADATA_URL': 'https://testissuer',
          'DOMAIN_NAME': 'unittest.local'},
         False),
        ({'SAML2_ENTITY_ID': 'testclientid',
          'SAML2_PUBLIC_KEY': None,
          'SAML2_PRIVATE_KEY': '--- SAML2PRIVATE KEY! ---',
          'SAML2_IDP_METADATA_URL': 'https://testissuer',
          'DOMAIN_NAME': 'unittest.local'},
         False),
        ({'SAML2_ENTITY_ID': 'testclientid',
          'SAML2_PUBLIC_KEY': '--- saml2publickey ---',
          'SAML2_PRIVATE_KEY': '--- SAML2PRIVATE KEY! ---',
          'SAML2_IDP_METADATA_URL': None,
          'DOMAIN_NAME': 'unittest.local'},
         False),
        ({'SAML2_ENTITY_ID': 'testclientid',
          'SAML2_PUBLIC_KEY': '--- saml2publickey ---',
          'SAML2_PRIVATE_KEY': None,
          'SAML2_IDP_METADATA_URL': 'https://testissuer',
          'DOMAIN_NAME': 'unittest.local'},
         False),
        ({'SAML2_ENTITY_ID': 'testclientid',
          'SAML2_PUBLIC_KEY': '--- saml2publickey ---',
          'SAML2_PRIVATE_KEY': '--- SAML2PRIVATE KEY! ---',
          'SAML2_IDP_METADATA_URL': 'https://testissuer',
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
            assert Saml2.is_enabled() == expected_result

    def test_get_settings(self):
        """Test get_settings"""
        with mock.patch('terrareg.saml.Saml2.get_idp_metadata', return_value={
                    'idp': {'some': 'key'}, 'sp': {'is': 'ignored', 'entityId': 'somethingelse'}}), \
                mock.patch('terrareg.config.Config.DOMAIN_NAME', 'unittest-domain.com'), \
                mock.patch('terrareg.config.Config.SAML2_PUBLIC_KEY', '--- SAML PUBLIC KEY ---'), \
                mock.patch('terrareg.config.Config.SAML2_PRIVATE_KEY', '!--- SAML PRIVATE KEY ---!'), \
                mock.patch('terrareg.config.Config.SAML2_ENTITY_ID', 'mock-saml2-entity-id'), \
                mock.patch('terrareg.config.Config.DEBUG', False):

            settings = Saml2.get_settings()

            assert settings == {
                "strict": True,
                "debug": False,
                "sp": {
                    "entityId": 'mock-saml2-entity-id',
                    "assertionConsumerService": {
                        "url": "https://unittest-domain.com/saml/login?acs",
                        "binding": "urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST"
                    },
                    "singleLogoutService": {
                        "url": "https://unittest-domain.com/saml/login?sls",
                        "binding": "urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect"
                    },
                    "NameIDFormat": "urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified",
                    "x509cert": '--- SAML PUBLIC KEY ---',
                    "privateKey": '!--- SAML PRIVATE KEY ---!'
                },
                "idp": {
                    'some': 'key'
                }
            }

    @pytest.mark.parametrize('domain_name,path,args,form,expected_result', [
        ('unittest-domain.com', '/saml/login',
         ImmutableMultiDict([('sso', ''), ('another-arg', 'a value')]),
         ImmutableMultiDict([('a-form-attribute', 'some-value')]),
         
         {'http_host': 'unittest-domain.com', 'server_port': 443, 'https': True,
          'script_name': '/saml/login',
          'get_data': ImmutableMultiDict([('sso', ''), ('another-arg', 'a value')]),
          'post_data': ImmutableMultiDict([('a-form-attribute', 'some-value')])
         }
        )
    ])
    def test_get_request_data(self, domain_name, path, args, form, expected_result):
        """test get_request_data"""
        class TestRequest:
            def __init__(self, path, args, form):
                self.path = path
                self.args = args
                self.form = form
        test_request = TestRequest(path, args, form)

        with mock.patch('terrareg.config.Config.DOMAIN_NAME', domain_name):
            request_data = Saml2.get_request_data(test_request)

        assert request_data == expected_result

    def test_initialise_request_auth_object(self):
        """test initialise_request_auth_object."""

        mock_security_data = {
            'mockSecurityData': 'here'
        }
        mock_auth_settings = mock.MagicMock()
        mock_auth_settings.get_security_data.return_value = mock_security_data
        mock_onelogin_saml2_auth_object = mock.MagicMock()
        mock_onelogin_saml2_auth_object.get_settings.return_value = mock_auth_settings
        mock_onelogin_saml2_auth = mock.MagicMock(return_value=mock_onelogin_saml2_auth_object)

        mock_request_data = mock.MagicMixin()
        mock_get_request_data = mock.MagicMock(return_value=mock_request_data)

        mock_settings = mock.MagicMock()
        mock_get_settings = mock.MagicMock(return_value=mock_settings)

        with mock.patch('terrareg.saml.Saml2.get_request_data', mock_get_request_data), \
                mock.patch('terrareg.saml.Saml2.get_settings', mock_get_settings), \
                mock.patch('onelogin.saml2.auth.OneLogin_Saml2_Auth', mock_onelogin_saml2_auth):
            mock_request = object()
            auth_object = Saml2.initialise_request_auth_object(mock_request)

            assert auth_object == mock_onelogin_saml2_auth_object

            mock_get_request_data.assert_called_once_with(mock_request)
            mock_onelogin_saml2_auth.assert_called_once_with(mock_request_data, mock_settings)
            mock_get_settings.assert_called_once()
            mock_auth_settings.get_security_data.assert_called_once()

            assert mock_security_data == {
                'mockSecurityData': 'here',
                'authnRequestsSigned': True,
                'failOnAuthnContextMismatch': True,
                'logoutRequestSigned': True,
                'logoutResponseSigned': True,
                'rejectDeprecatedAlgorithm': True,
                'signMetadata': True,
                'wantAssertionsEncrypted': False,
                'wantAssertionsSigned': True,
                'wantMessagesSigned': True,
                'wantNameIdEncrypted': False,
            }

    def test_get_idp_metadata(self):
        """Obtain metadata from IdP"""

        mock_idp_metadata = mock.MagicMock()
        with mock.patch('onelogin.saml2.idp_metadata_parser.OneLogin_Saml2_IdPMetadataParser.parse_remote',
                        mock.MagicMock(return_value=mock_idp_metadata)) as mock_parse_remote, \
                mock.patch('terrareg.config.Config.SAML2_IDP_METADATA_URL', 'https://unittestmetadata.com/endpoint'):

            with mock.patch('terrareg.config.Config.SAML2_ISSUER_ENTITY_ID', None):
                meta_data = Saml2.get_idp_metadata()

                assert meta_data == mock_idp_metadata
                mock_parse_remote.assert_called_once_with('https://unittestmetadata.com/endpoint')

                # Call a second time, setting, ensuring that the parse remote function is not called,
                # as the value should be cached
                mock_parse_remote.reset_mock()

                assert Saml2.get_idp_metadata() == meta_data
                mock_parse_remote.assert_not_called()

            with mock.patch('terrareg.config.Config.SAML2_ISSUER_ENTITY_ID', 'unittest-entity'):

                # Call with entity URL and assert that it is passed to parse remote method
                Saml2._IDP_METADATA = None
                assert Saml2.get_idp_metadata() == mock_idp_metadata
                mock_parse_remote.assert_called_once_with('https://unittestmetadata.com/endpoint', entity_id='unittest-entity')
