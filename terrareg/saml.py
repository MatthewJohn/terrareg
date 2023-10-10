
import datetime

import onelogin.saml2.auth
import onelogin.saml2.idp_metadata_parser
import onelogin.saml2.utils

import terrareg.config
from terrareg.utils import get_public_url_details

class Saml2:

    _IDP_METADATA = None
    _IDP_METADATA_REFRESH_DATE = None
    # Retain IdP keys cache for 12 hours
    _IDP_METADATA_REFRESH_INTERVAL = datetime.timedelta(hours=12)

    @classmethod
    def is_enabled(cls):
        """Whether SAML authentication is enabled"""
        config = terrareg.config.Config()
        _, domain, _ = get_public_url_details()
        return (domain is not None and
                config.SAML2_ENTITY_ID is not None and
                config.SAML2_IDP_METADATA_URL is not None and
                config.SAML2_PUBLIC_KEY is not None and
                config.SAML2_PRIVATE_KEY is not None)

    @classmethod
    def get_settings(cls):
        """Create settings for saml2"""
        config = terrareg.config.Config()

        _, domain, _ = get_public_url_details()
        settings = {
            "strict": True,
            "debug": config.DEBUG,
            "sp": {
                "entityId": config.SAML2_ENTITY_ID,
                "assertionConsumerService": {
                    "url": f"https://{domain}/saml/login?acs",
                    "binding": "urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST"
                },
                "singleLogoutService": {
                    "url": f"https://{domain}/saml/login?sls",
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
        _, domain, _ = get_public_url_details()
        return {
            'http_host': domain,
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
        auth = onelogin.saml2.auth.OneLogin_Saml2_Auth(
            request_data,
            cls.get_settings())

        security_data = auth.get_settings().get_security_data()
        security_data['authnRequestsSigned'] = True
        security_data['logoutRequestSigned'] = True
        security_data['logoutResponseSigned'] = True
        security_data['signMetadata'] = True
        security_data['wantMessagesSigned'] = True
        security_data['wantAssertionsSigned'] = True
        security_data['wantAssertionsEncrypted'] = False
        security_data['wantNameIdEncrypted'] = False
        security_data['rejectDeprecatedAlgorithm'] = True
        security_data['failOnAuthnContextMismatch'] = True

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
            cls._IDP_METADATA = onelogin.saml2.idp_metadata_parser.OneLogin_Saml2_IdPMetadataParser.parse_remote(
                config.SAML2_IDP_METADATA_URL,
                **args)
            cls._IDP_METADATA_REFRESH_DATE = datetime.datetime.now() + cls._IDP_METADATA_REFRESH_INTERVAL

        return cls._IDP_METADATA

    @classmethod
    def get_self_url(cls, request):
        """Return self URL."""
        request_data = cls.get_request_data(request)
        return onelogin.saml2.utils.OneLogin_Saml2_Utils.get_self_url(request_data)
