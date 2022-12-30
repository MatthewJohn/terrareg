
from flask import session

from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.auth_wrapper
import terrareg.audit
import terrareg.audit_action
import terrareg.auth
import terrareg.models


class ApiTerraregAdminAuthenticate(ErrorCatchingResource):
    """Interface to perform authentication as an admin and set appropriate cookie."""

    method_decorators = [terrareg.auth_wrapper.auth_wrapper('is_built_in_admin')]

    def _post(self):
        """Handle POST requests to the authentication endpoint."""

        session_obj = self.create_session()
        if not isinstance(session_obj, terrareg.models.Session):
            return {'message': 'Sessions not enabled in configuration'}, 403

        session['is_admin_authenticated'] = True
        session['authentication_type'] = terrareg.auth.AuthenticationType.SESSION_PASSWORD.value
        session.modified = True

        # Create audit event
        terrareg.audit.AuditEvent.create_audit_event(
            action=terrareg.audit_action.AuditAction.USER_LOGIN,
            object_type=None, object_id=None,
            old_value=None, new_value=None
        )

        return {'authenticated': True}
