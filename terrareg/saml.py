
import datetime

from onelogin.saml2.auth import OneLogin_Saml2_Auth, OneLogin_Saml2_Settings
from onelogin.saml2.idp_metadata_parser import OneLogin_Saml2_IdPMetadataParser
from onelogin.saml2.utils import OneLogin_Saml2_Utils

import terrareg.config

class Saml2:

    _IDP_METADATA = None
    _IDP_METADATA_REFRESH_DATE = None
    # Retain IdP keys cache for 12 hours
    _IDP_METADATA_REFRESH_INTERVAL = datetime.timedelta(hours=12)

    @classmethod
    def is_enabled(cls):
        """Whether SAML auithentication is enabled"""
        config = terrareg.config.Config()
        return (config.DOMAIN_NAME is not None and
                config.SAML2_ENTITY_ID is not None and
                config.SAML2_IDP_METADATA_URL is not None and
                config.SAML2_PUBLIC_KEY is not None and
                config.SAML2_PRIVATE_KEY is not None)

    @classmethod
    def get_settings(cls):
        """Create settings for saml2"""
        config = terrareg.config.Config()

        settings = {
            "strict": True,
            "debug": config.DEBUG,
            "sp": {
                "entityId": config.SAML2_ENTITY_ID,
                "assertionConsumerService": {
                    "url": f"https://{config.DOMAIN_NAME}/saml/login?acs",
                    "binding": "urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST"
                },
                "singleLogoutService": {
                    "url": f"https://{config.DOMAIN_NAME}/saml/login?sls",
                    "binding": "urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect"
                },
                "NameIDFormat": "urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified",
                "x509cert": config.SAML2_PUBLIC_KEY,
                "privateKey": config.SAML2_PRIVATE_KEY
            },
            "idp": cls.get_idp_metadata().get('idp', {})
        }

        return settings

    @classmethod
    def get_request_data(cls, request):
        """Obtain request data for saml2"""
        return {
            'http_host': terrareg.config.Config().DOMAIN_NAME,
            'server_port': 443,
            'https': True,
            'script_name': request.path,
            'get_data': request.args.copy(),
            'post_data': request.form.copy()
        }

    @classmethod
    def initialise_request_auth_object(cls, request):
        """Initialise auth object."""
        request_data = cls.get_request_data(request)
        auth = OneLogin_Saml2_Auth(
            request_data,
            cls.get_settings())

        return auth

    @classmethod
    def get_idp_metadata(cls):
        """Obtain metadata from IdP"""
        if (not cls._IDP_METADATA or
                cls._IDP_METADATA_REFRESH_DATE is None or
                cls._IDP_METADATA_REFRESH_DATE < datetime.datetime.now()):
            config = terrareg.config.Config()

            args = {}
            if config.SAML2_ISSUER_ENTITY_ID:
                args['entity_id'] = config.SAML2_ISSUER_ENTITY_ID
            cls._IDP_METADATA = OneLogin_Saml2_IdPMetadataParser.parse_remote(
                config.SAML2_IDP_METADATA_URL,
                **args)
            cls._IDP_METADATA_REFRESH_DATE = datetime.datetime.now() + cls._IDP_METADATA_REFRESH_INTERVAL

        return cls._IDP_METADATA

    @classmethod
    def get_self_url(cls, request):
        """Return self URL."""
        request_data = cls.get_request_data(request)
        return OneLogin_Saml2_Utils.get_self_url(request_data)
