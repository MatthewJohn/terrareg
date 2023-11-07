
from flask import request, redirect, make_response, render_template, session

from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.provider_source.factory
import terrareg.github
import terrareg.auth
import terrareg.audit
import terrareg.audit_action
import terrareg.config
import terrareg.namespace_type


class GithubLoginCallback(ErrorCatchingResource):
    """Interface to handle call-back from Github login"""

    def _github_login_error(self, error):
        """Return github login error"""
        return make_response(self._render_template(
            'error.html',
            error_title='Login error',
            error_description=error,
            root_bread_brumb='Login'
        ))

    def _get(self, provider_source):
        """Handle callback from Github auth."""
        code = request.args.get("code")

        # Obtain provider source
        provider_source_factory = terrareg.provider_source.factory.ProviderSourceFactory.get()
        provider_source_obj = provider_source_factory.get_provider_source_by_api_name(provider_source)
        if provider_source_obj is None:
            return self._get_404_response()

        # Obtain access token, purely to ensure that the code is valid
        access_token = provider_source_obj.get_user_access_token(code)
        if access_token is None:
            return self._github_login_error(f"Invalid code returned from {provider_source}")

        # Store access token in server-side session
        current_session = terrareg.auth.AuthFactory.get_current_session()
        if current_session:
            current_session.provider_source_auth = {
                "github_access_token": access_token
            }

        # If user is authenticated, update session
        user_id = provider_source_obj.get_username(access_token)
        if user_id is None:
            return self._github_login_error(f"Invalid user data returned from {provider_source}")
        session['github_username'] = user_id

        # Obtain list of organisations that the user is an owner of and add the user's user ID to the list
        organisations = {
            org: terrareg.namespace_type.NamespaceType.GITHUB_ORGANISATION.value
            for org in provider_source_obj.get_user_organisations(access_token)
        }
        organisations[user_id] = terrareg.namespace_type.NamespaceType.GITHUB_USER.value
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

        provider_source_obj.update_repositories(access_token)

        # Redirect to homepage
        return redirect("/", code=302)
