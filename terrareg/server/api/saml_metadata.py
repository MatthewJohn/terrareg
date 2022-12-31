
from flask import request, make_response

from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.saml


class ApiSamlMetadata(ErrorCatchingResource):
    """Meta-data endpoint for SAML"""

    def _get(self):
        """Return SAML SP metadata."""
        auth = terrareg.saml.Saml2.initialise_request_auth_object(request)
        settings = auth.get_settings()
        metadata = settings.get_sp_metadata()
        errors = settings.validate_metadata(metadata)

        if len(errors) == 0:
            resp = make_response(metadata, 200)
            resp.headers['Content-Type'] = 'text/xml'
        else:
            resp = make_response(', '.join(errors), 500)
        return resp
