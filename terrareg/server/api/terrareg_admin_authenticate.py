
from flask import session

from terrareg.server.error_catching_resource import ErrorCatchingResource
from terrareg.auth_wrapper import auth_wrapper
from terrareg.audit import AuditEvent
from terrareg.audit_action import AuditAction
import terrareg.auth
from terrareg.models import Session


class ApiTerraregAdminAuthenticate(ErrorCatchingResource):
    """Interface to perform authentication as an admin and set appropriate cookie."""

    method_decorators = [auth_wrapper('is_built_in_admin')]

    def _post(self):
        """Handle POST requests to the authentication endpoint."""

        session_obj = self.create_session()
        if not isinstance(session_obj, Session):
            return {'message': 'Sessions not enabled in configuration'}, 403

        session['is_admin_authenticated'] = True
        session['authentication_type'] = terrareg.auth.AuthenticationType.SESSION_PASSWORD.value
        session.modified = True

        # Create audit event
        AuditEvent.create_audit_event(
            action=AuditAction.USER_LOGIN,
            object_type=None, object_id=None,
            old_value=None, new_value=None
        )

        return {'authenticated': True}
