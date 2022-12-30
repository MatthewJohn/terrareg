
from flask import session, redirect, request, make_response, render_template

from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.config
import terrareg.auth
import terrareg.saml
from terrareg.audit_action import AuditAction
from terrareg.audit import AuditEvent


class ApiSamlInitiate(ErrorCatchingResource):
    """Interface to initiate authentication via OpenID connect"""

    def _get(self):
        """Setup authentication request to redirect user to SAML provider."""
        auth = terrareg.saml.Saml2.initialise_request_auth_object(request)

        errors = None

        if 'sso' in request.args:
            session_obj = self.create_session()
            if session_obj is None:
                return {"Error", "Could not create session"}, 500

            idp_url = auth.login()

            session['AuthNRequestID'] = auth.get_last_request_id()
            session.modified = True

            return redirect(idp_url)

        elif 'acs' in request.args:
            request_id = session.get('AuthNRequestID')

            if request_id is None:
                return {"Error": "No request ID"}, 500

            auth.process_response(request_id=request_id)
            errors = auth.get_errors()
            if not errors and auth.is_authenticated():
                if 'AuthNRequestID' in session:
                    del session['AuthNRequestID']

                # Setup Authentcation session
                session['samlUserdata'] = auth.get_attributes()
                session['samlNameId'] = auth.get_nameid()
                session['samlNameIdFormat'] = auth.get_nameid_format()
                session['samlNameIdNameQualifier'] = auth.get_nameid_nq()
                session['samlNameIdSPNameQualifier'] = auth.get_nameid_spnq()
                session['samlSessionIndex'] = auth.get_session_index()
                session['is_admin_authenticated'] = True
                session['authentication_type'] = terrareg.auth.AuthenticationType.SESSION_SAML.value
                if terrareg.config.Config().SAML2_DEBUG:
                    print(f"""
Successul SAML2 authentication response:
User data:
{session['samlUserdata']}
User groups:
{session['samlUserdata'].get(terrareg.config.Config().SAML2_GROUP_ATTRIBUTE, [])}""")

                session.modified = True

        if errors:
            if terrareg.config.Config().SAML2_DEBUG:
                print(f"""
Error during SAML2 authentication response:
Errors:
{errors}""")
            res = make_response(render_template(
                'error.html',
                error_title='Login error',
                error_description='An error occured whilst processing SAML login request'
            ))
            res.headers['Content-Type'] = 'text/html'
            return res

        # Create audit event
        AuditEvent.create_audit_event(
            action=AuditAction.USER_LOGIN,
            object_type=None, object_id=None,
            old_value=None, new_value=None
        )

        return redirect('/', code=302)

    def _post(self):
        """Handle POST request."""
        return self._get()
