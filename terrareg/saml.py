
from onelogin.saml2.auth import OneLogin_Saml2_Auth, OneLogin_Saml2_Settings
from onelogin.saml2.idp_metadata_parser import OneLogin_Saml2_IdPMetadataParser
from onelogin.saml2.utils import OneLogin_Saml2_Utils

import terrareg.config

class Saml2:

    @classmethod
    def get_settings(cls):
        """Create settings for saml2"""
        config = terrareg.config.Config()

        settings = {
            "strict": True,
            "debug": config.DEBUG,
            "sp": {
                # "entityId": f"https://{config.DOMAIN_NAME}/metadata/",
                "entityId": config.SAML2_ENTITY_ID,
                "assertionConsumerService": {
                    "url": f"https://{config.DOMAIN_NAME}/saml/login?acs",
                    "binding": "urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST"
                },
                # "attributeConsumingService": {
                #         "serviceName": "SP test",
                #         "serviceDescription": "Test Service",
                #         "requestedAttributes": [
                #             # {
                #             #     "name": "",
                #             #     "isRequired": false,
                #             #     "nameFormat": "",
                #             #     "friendlyName": "",
                #             #     "attributeValue": []
                #             # }
                #         ]
                # },
                "singleLogoutService": {
                    "url": f"https://{config.DOMAIN_NAME}/saml/login?sls",
                    #"responseUrl": "https://<sp_domain>/?sls",
                    "binding": "urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect"
                },
                "NameIDFormat": "urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified",
                "x509cert": config.SAML2_PUBLIC_KEY,
                "privateKey": config.SAML2_PRIVATE_KEY
                # 'x509certNew': '',
            },

            # "idp": {
            #     "entityId": "https://app.onelogin.com/saml/metadata/<onelogin_connector_id>",
            #     "singleSignOnService": {
            #         "url": "https://app.onelogin.com/trust/saml2/http-post/sso/<onelogin_connector_id>",
            #         "binding": "urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect"
            #     },
            #     "singleLogoutService": {
            #         "url": "https://app.onelogin.com/trust/saml2/http-redirect/slo/<onelogin_connector_id>",
            #         "responseUrl": "https://app.onelogin.com/trust/saml2/http-redirect/slo_return/<onelogin_connector_id>",
            #         "binding": "urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect"
            #     },
            #     "x509cert": "<onelogin_connector_cert>"
            #     # 'certFingerprint': '',
            #     # 'certFingerprintAlgorithm': 'sha1',
            #     # 'x509certMulti': {
            #     #      'signing': [
            #     #          '<cert1-string>'
            #     #      ],
            #     #      'encryption': [
            #     #          '<cert2-string>'
            #     #      ]
            #     # }
            # }
        }
        idp_metadata = cls.get_idp_metadata()
        if 'idp' in idp_metadata:
            settings['idp'] = idp_metadata['idp']
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
        auth = OneLogin_Saml2_Auth(
            request,
            cls.get_settings())

    @classmethod
    def get_idp_metadata(cls):
        """Obtain metadata from IdP"""
        config = terrareg.config.Config()

        args = {}
        if config.SAML2_ISSUER_ENTITY_ID:
            args['entity_id'] = config.SAML2_ISSUER_ENTITY_ID
        print(config.SAML2_IDP_METADATA_URL)
        idp_data = OneLogin_Saml2_IdPMetadataParser.parse_remote(
            config.SAML2_IDP_METADATA_URL,
            **args)
        return idp_data