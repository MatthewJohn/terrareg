
from flask import request, redirect, make_response, render_template, session

from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.github
import terrareg.auth
import terrareg.audit
import terrareg.audit_action
import terrareg.config


class GithubLoginCallback(ErrorCatchingResource):
    """Interface to handle callback from Github login"""

    def _github_login_error(self, error):
        """Return github login error"""
        return make_response(render_template(
            'error.html',
            error_title='Login error',
            error_description=error
        ))

    def _get(self):
        """Handle callback from github auth."""
        code = request.args.get("code")

        if not terrareg.github.Github.is_enabled():
            return self._github_login_error('Github authentication is not enabled')

        # Obtain access token, purely to ensure that the code is valid
        access_token = terrareg.github.Github.get_access_token(code)
        if access_token is None:
            return self._github_login_error("Invalid code returned from Github")

        # If user is authenticated, update session
        user_id = terrareg.github.Github.get_username(access_token)
        if user_id is None:
            return self._github_login_error("Invalid user data returned from Github")
        session['github_username'] = user_id

        # Obtain list of organisations that the user is an owner of and add the user's user ID to the list
        organisations = terrareg.github.Github.get_user_organisations(access_token)
        organisations.append(user_id)
        session['organisations'] = organisations
        session['is_admin_authenticated'] = True
        session['authentication_type'] = terrareg.auth.AuthenticationType.SESSION_GITHUB.value
        session.modified = True

        # Create audit event
        terrareg.audit.AuditEvent.create_audit_event(
            action=terrareg.audit_action.AuditAction.USER_LOGIN,
            object_type=None, object_id=None,
            old_value=None, new_value=None
        )

        # Redirect to homepage
        return redirect("/", code=302)
