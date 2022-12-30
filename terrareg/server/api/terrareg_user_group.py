
from terrareg.server.error_catching_resource import ErrorCatchingResource
from terrareg.auth_wrapper import auth_wrapper
from terrareg.models import UserGroup


class ApiTerraregAuthUserGroup(ErrorCatchingResource):
    """Interface to interact with single user group."""

    method_decorators = [auth_wrapper('is_admin')]

    def _delete(self, user_group):
        """Delete user group."""
        user_group_obj = UserGroup.get_by_group_name(user_group)
        if not user_group_obj:
            return {'message': 'User group does not exist.'}, 400

        user_group_obj.delete()
        return {}, 200