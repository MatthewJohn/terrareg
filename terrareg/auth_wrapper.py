
from functools import wraps

from flask import abort

import terrareg.auth


def auth_wrapper(auth_check_method, *wrapper_args, request_kwarg_map={}, **wrapper_kwargs):
    """
    Wrapper to custom authentication decorators.
    An authentication checking method should be passed with args/kwargs, which will be
    used to check authentication and authorisation.
    """
    def decorator_wrapper(func):
        """Check user is authenticated as admin and either call function or return 401, if not."""
        @wraps(func)
        def wrapper(*args, **kwargs):
            auth_method = terrareg.auth.AuthFactory().get_current_auth_method()

            auth_kwargs = wrapper_kwargs.copy()
            for request_kwarg in request_kwarg_map:
                if request_kwarg in kwargs:
                    auth_kwargs[request_kwarg_map[request_kwarg]] = kwargs[request_kwarg]

            if (status := getattr(auth_method, auth_check_method)(*wrapper_args, **auth_kwargs)) == False:
                if auth_method.is_authenticated():
                    abort(403)
                else:
                    abort(401)
            elif status == True:
                return func(*args, **kwargs)
            else:
                raise Exception('Invalid response from auth check method')
        return wrapper
    return decorator_wrapper