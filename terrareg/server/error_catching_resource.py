
from typing import Tuple, Union, Dict

from flask_restful import Resource

from terrareg.server.base_handler import BaseHandler
import terrareg.errors
import terrareg.models
import terrareg.provider_model


def api_error(msg):
    """Return API error message"""
    return {
        "status": "Error",
        "message": msg
    }


class ErrorCatchingResource(Resource, BaseHandler):
    """Provide resource that catches terrareg errors."""

    def _get(self, *args, **kwargs):
        """Placeholder for overridable get method."""
        return {'message': 'The method is not allowed for the requested URL.'}, 405

    def _get_arg_parser(self):
        """Return arg parser for GET requests"""
        raise NotImplementedError

    def get(self, *args, **kwargs):
        """Run subclasses get in error handling fashion."""
        try:
            return self._get(*args, **kwargs)
        except terrareg.errors.TerraregError as exc:
            return {
                "status": "Error",
                "message": str(exc)
            }, 500

    def _post(self, *args, **kwargs):
        """Placeholder for overridable post method."""
        return {'message': 'The method is not allowed for the requested URL.'}, 405

    def _post_arg_parser(self):
        """Return arg parser for POST requests"""
        raise NotImplementedError

    def post(self, *args, **kwargs):
        """Run subclasses post in error handling fashion."""
        try:
            return self._post(*args, **kwargs)
        except terrareg.errors.TerraregError as exc:
            return {
                "status": "Error",
                "message": str(exc)
            }, 500

    def _delete(self, *args, **kwargs):
        """Placeholder for overridable delete method."""
        return {'message': 'The method is not allowed for the requested URL.'}, 405

    def _delete_arg_parser(self):
        """Return arg parser for DELETE requests"""
        raise NotImplementedError

    def delete(self, *args, **kwargs):
        """Run subclasses delete in error handling fashion."""
        try:
            return self._delete(*args, **kwargs)
        except terrareg.errors.TerraregError as exc:
            return api_error(str(exc)), 500

    def _get_404_response(self) -> Tuple[Dict[str, str], int]:
        """Return common 404 error"""
        return {'errors': ['Not Found']}, 404

    def _get_401_response(self) -> Tuple[Dict[str, str], int]:
        """Return standardised 401."""
        return {'message': ('The server could not verify that you are authorized to access the URL requested. '
                            'You either supplied the wrong credentials (e.g. a bad password), '
                            'or your browser doesn\'t understand how to supply the credentials required.')
        }, 401

    def get_module_provider_by_names(self, namespace: str, name: str, provider: str, create: bool=False) -> Union[
            Tuple[terrareg.models.Namespace, terrareg.models.Module, terrareg.models.ModuleProvider, None],
            Tuple[None, None, None, Tuple[Dict[str, str], int]],
        ]:
        """Obtain namespace, module, provider objects by name"""
        namespace_obj = terrareg.models.Namespace.get(namespace, create=create)
        if namespace_obj is None:
            return None, None, None, ({'message': 'Namespace does not exist'}, 400)

        module_obj = terrareg.models.Module(namespace=namespace_obj, name=name)

        module_provider_obj = terrareg.models.ModuleProvider.get(module=module_obj, name=provider, create=create)
        if module_provider_obj is None:
            return None, None, None, ({'message': 'Module provider does not exist'}, 400)

        return namespace_obj, module_obj, module_provider_obj, None

    def get_module_version_by_name(self, namespace: str, name: str, provider: str, version: str, create: bool=False) -> Union[
            Tuple[terrareg.models.Namespace, terrareg.models.Module, terrareg.models.ModuleProvider, terrareg.models.ModuleVersion, None],
            Tuple[None, None, None, None, Tuple[Dict[str, str], int]]
        ]:
        """Obtain namespace, module, provider and module version by names"""
        namespace_obj, module_obj, module_provider_obj, error = self.get_module_provider_by_names(namespace, name, provider, create=create)
        if error:
            return namespace_obj, module_obj, module_provider_obj, None, error

        module_version_obj = terrareg.models.ModuleVersion.get(module_provider=module_provider_obj, version=version)
        if module_version_obj is None:
            return namespace_obj, module_obj, module_provider_obj, None, ({'message': 'Module version does not exist'}, 400)

        return namespace_obj, module_obj, module_provider_obj, module_version_obj, None

    def get_provider_by_names(self, namespace: str, provider: str) -> Union[
            Tuple[terrareg.models.Namespace, terrareg.provider_model.Provider, None],
            Tuple[None, None, Tuple[Dict[str, str], int]],
        ]:
        """Obtain namespace, provider objects by name"""
        namespace_obj = terrareg.models.Namespace.get(namespace)
        if namespace_obj is None:
            return None, None, ({'message': 'Namespace does not exist'}, 400)

        provider_obj = terrareg.provider_model.Provider.get(namespace=namespace_obj, name=provider)
        if provider_obj is None:
            return None, None, ({'message': 'Provider does not exist'}, 400)

        return namespace_obj, provider_obj, None
