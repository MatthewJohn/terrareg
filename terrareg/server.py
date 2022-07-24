
import hmac
import os
import re
import datetime
from functools import wraps
import json
import urllib.parse
import hashlib
from enum import Enum

from flask import (
    Flask, request, render_template,
    redirect, make_response, send_from_directory,
    session, g
)
from flask_restful import Resource, Api, reqparse, inputs, abort

import terrareg.config
from terrareg.database import Database
from terrareg.errors import (
    InvalidModuleNameError, InvalidModuleProviderNameError, InvalidNamespaceNameError, InvalidVersionError, RepositoryUrlParseError, TerraregError, UploadError, NoModuleVersionAvailableError,
    NoSessionSetError, IncorrectCSRFTokenError
)
from terrareg.models import (
    Example, ExampleFile, Namespace, Module, ModuleProvider,
    ModuleVersion, ProviderLogo, Session, Submodule,
    GitProvider
)
from terrareg.module_search import ModuleSearch
from terrareg.module_extractor import ApiUploadModuleExtractor, GitModuleExtractor
from terrareg.analytics import AnalyticsEngine
from terrareg.filters import NamespaceTrustFilter


def catch_name_exceptions(f):
    """Wrapper method to catch name validation errors."""
    @wraps(f)
    def decorated_function(self, *args, **kwargs):
        try:
            return f(self, *args, **kwargs)

        # Handle invalid namespace name
        except InvalidNamespaceNameError:
            return self._render_template(
                'error.html',
                error_title='Invalid namespace name',
                error_description="The namespace name '{}' is invalid".format(kwargs['namespace'])
            ), 400

        # Handle invalid module name exceptions
        except InvalidModuleNameError:
            namespace = None
            if 'namespace' in kwargs:
                namespace = Namespace(name=kwargs['namespace'])
            return self._render_template(
                'error.html',
                error_title='Invalid module name',
                error_description="The module name '{}' is invalid".format(kwargs['name']),
                namespace=namespace
            ), 400

        # Handle invalid provider name exceptions
        except InvalidModuleProviderNameError:
            namespace = None
            module = None
            if 'namespace' in kwargs:
                namespace = Namespace(name=kwargs['namespace'])
                if 'name' in kwargs:
                    module = Module(namespace=namespace, name=kwargs['name'])
            return self._render_template(
                'error.html',
                error_title='Invalid provider name',
                error_description="The provider name '{}' is invalid".format(kwargs['provider']),
                namespace=namespace,
                module=module
            ), 400

        # Handle invalid version number error
        except InvalidVersionError:
            namespace = None
            module = None
            module_provider_name = None
            if 'namespace' in kwargs:
                namespace = Namespace(name=kwargs['namespace'])
                if 'name' in kwargs:
                    module = Module(namespace=namespace, name=kwargs['name'])
                    if 'provider' in kwargs:
                        module_provider_name = kwargs['provider']
            version = None
            if 'version' in kwargs:
                version = kwargs['version']
            return self._render_template(
                'error.html',
                error_title='Invalid version number',
                error_description=("The version number '{}' is invalid".format(version) if version else ''),
                namespace=namespace,
                module=module,
                module_provider_name=module_provider_name
            ), 400
    return decorated_function


class Server(object):
    """Manage web server and route requests"""

    def __init__(self, ssl_public_key=None, ssl_private_key=None):
        """Create flask app and store member variables"""
        self._app = Flask(
            __name__,
            static_folder='static',
            template_folder='templates'
        )
        self._api = Api(
            self._app,
            #prefix='v1'
        )

        self.host = '0.0.0.0'
        self.port = terrareg.config.Config().LISTEN_PORT
        self.ssl_public_key = ssl_public_key
        self.ssl_private_key = ssl_private_key

        if not os.path.isdir(terrareg.config.Config().DATA_DIRECTORY):
            os.mkdir(terrareg.config.Config().DATA_DIRECTORY)
        if not os.path.isdir(self._get_upload_directory()):
            os.mkdir(self._get_upload_directory())
        if not os.path.isdir(os.path.join(terrareg.config.Config().DATA_DIRECTORY, 'modules')):
            os.mkdir(os.path.join(terrareg.config.Config().DATA_DIRECTORY, 'modules'))

        self._app.config['UPLOAD_FOLDER'] = self._get_upload_directory()

        # Initialise database
        Database.get().initialise()
        GitProvider.initialise_from_config()

        self._register_routes()

    def _get_upload_directory(self):
        return os.path.join(terrareg.config.Config().DATA_DIRECTORY, 'upload')

    def _register_routes(self):
        """Register routes with flask."""

        # Terraform registry routes
        self._api.add_resource(
            ApiTerraformWellKnown,
            '/.well-known/terraform.json'
        )

        self._api.add_resource(
            ApiModuleList,
            '/v1/modules',
            '/v1/modules/'
        )
        self._api.add_resource(
            ApiModuleSearch,
            '/v1/modules/search',
            '/v1/modules/search/'
        )
        self._api.add_resource(
            ApiNamespaceModules,
            '/v1/modules/<string:namespace>',
            '/v1/modules/<string:namespace>/'
        )
        self._api.add_resource(
            ApiModuleDetails,
            '/v1/modules/<string:namespace>/<string:name>',
            '/v1/modules/<string:namespace>/<string:name>/'
        )
        self._api.add_resource(
            ApiModuleProviderDetails,
            '/v1/modules/<string:namespace>/<string:name>/<string:provider>',
            '/v1/modules/<string:namespace>/<string:name>/<string:provider>/')
        self._api.add_resource(
            ApiModuleVersions,
            '/v1/modules/<string:namespace>/<string:name>/<string:provider>/versions',
            '/v1/modules/<string:namespace>/<string:name>/<string:provider>/versions/'
        )
        self._api.add_resource(
            ApiModuleVersionDetails,
            '/v1/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>',
            '/v1/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/'
        )
        self._api.add_resource(
            ApiModuleVersionDownload,
            '/v1/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/download'
        )
        self._api.add_resource(
            ApiModuleProviderDownloadsSummary,
            '/v1/modules/<string:namespace>/<string:name>/<string:provider>/downloads/summary'
        )

        # Views
        self._app.route('/')(self._view_serve_static_index)
        self._app.route(
            '/login'
        )(self._view_serve_login)
        self._app.route(
            '/logout'
        )(self._logout)
        self._app.route(
            '/create-module'
        )(self._view_serve_create_module)
        self._app.route(
            '/initial-setup'
        )(self._view_serve_initial_setup)
        self._app.route(
            '/modules'
        )(self._view_serve_namespace_list)
        self._app.route(
            '/modules/'
        )(self._view_serve_namespace_list)
        self._app.route(
            '/modules/search'
        )(self._view_serve_module_search)
        self._app.route(
            '/modules/search/'
        )(self._view_serve_module_search)
        self._app.route(
            '/modules/<string:namespace>'
        )(self._view_serve_namespace)
        self._app.route(
            '/modules/<string:namespace>/'
        )(self._view_serve_namespace)
        self._app.route(
            '/modules/<string:namespace>/<string:name>'
        )(self._view_serve_module)
        self._app.route(
            '/modules/<string:namespace>/<string:name>/'
        )(self._view_serve_module)
        self._app.route(
            '/modules/<string:namespace>/<string:name>/<string:provider>'
        )(self._view_serve_module_provider)
        self._app.route(
            '/modules/<string:namespace>/<string:name>/<string:provider>/'
        )(self._view_serve_module_provider)
        self._app.route(
            '/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>'
        )(self._view_serve_module_provider)
        self._app.route(
            '/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/'
        )(self._view_serve_module_provider)
        self._app.route(
            '/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/submodule/<path:submodule_path>'
        )(self._view_serve_submodule)
        self._app.route(
            '/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/example/<path:submodule_path>'
        )(self._view_serve_example)

        # Terrareg APIs
        ## Config endpoint
        self._api.add_resource(
            ApiTerraregConfig,
            '/v1/terrareg/config'
        )
        self._api.add_resource(
            ApiTerraregGitProviders,
            '/v1/terrareg/git_providers'
        )
        ## Analytics URLs /v1/terrareg/analytics
        self._api.add_resource(
            ApiTerraregGlobalStatsSummary,
            '/v1/terrareg/analytics/global/stats_summary'
        )
        self._api.add_resource(
            ApiTerraregMostRecentlyPublishedModuleVersion,
            '/v1/terrareg/analytics/global/most_recently_published_module_version'
        )
        self._api.add_resource(
            ApiTerraregGlobalUsageStats,
            '/v1/terrareg/analaytics/global/usage_stats'
        )
        self._api.add_resource(
            ApiTerraregModuleProviderAnalyticsTokenVersions,
            '/v1/terrareg/analytics/<string:namespace>/<string:name>/<string:provider>/token_versions'
        )
        self._api.add_resource(
            ApiTerraregMostDownloadedModuleProviderThisWeek,
            '/v1/terrareg/analytics/global/most_downloaded_module_provider_this_week'
        )

        # Initial setup
        self._api.add_resource(
            ApiTerraregInitialSetupData,
            '/v1/terrareg/initial_setup'
        )

        ## namespaces endpoint
        self._api.add_resource(
            ApiTerraregNamespaces,
            '/v1/terrareg/namespaces'
        )
        self._api.add_resource(
            ApiTerraregNamespaceDetails,
            '/v1/terrareg/namespaces/<string:namespace>'
        )

        ## Module endpoints /v1/terreg/modules
        self._api.add_resource(
            ApiTerraregNamespaceModules,
            '/v1/terrareg/modules/<string:namespace>'
        )
        self._api.add_resource(
            ApiTerraregModuleProviderDetails,
            '/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>'
        )
        self._api.add_resource(
            ApiTerraregModuleVersionDetails,
            '/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>'
        )
        self._api.add_resource(
            ApiTerraregModuleProviderCreate,
            '/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/create'
        )
        self._api.add_resource(
            ApiTerraregModuleProviderDelete,
            '/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/delete'
        )
        self._api.add_resource(
            ApiTerraregModuleProviderSettings,
            '/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/settings'
        )
        self._api.add_resource(
            ApiTerraregModuleProviderIntegrations,
            '/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/integrations'
        )
        self._api.add_resource(
            ApiModuleVersionCreateBitBucketHook,
            '/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/hooks/bitbucket'
        )
        self._api.add_resource(
            ApiTerraregModuleVersionVariableTemplate,
            '/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/variable_template'
        )
        self._api.add_resource(
            ApiTerraregModuleVersionReadmeHtml,
            '/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/readme_html'
        )
        self._api.add_resource(
            ApiModuleVersionUpload,
            '/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/upload'
        )
        self._api.add_resource(
            ApiModuleVersionCreate,
            '/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/import'
        )
        self._api.add_resource(
            ApiModuleVersionSourceDownload,
            '/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/source.zip'
        )
        self._api.add_resource(
            ApiTerraregModuleVersionPublish,
            '/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/publish'
        )
        self._api.add_resource(
            ApiTerraregModuleVersionDelete,
            '/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/delete'
        )
        self._api.add_resource(
            ApiTerraregModuleVerisonSubmodules,
            '/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/submodules'
        )
        self._api.add_resource(
            ApiTerraregSubmoduleDetails,
            '/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/submodules/details/<path:submodule>'
        )
        self._api.add_resource(
            ApiTerraregSubmoduleReadmeHtml,
            '/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/submodules/readme_html/<path:submodule>'
        )
        self._api.add_resource(
            ApiTerraregModuleVersionExamples,
            '/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/examples'
        )
        self._api.add_resource(
            ApiTerraregExampleDetails,
            '/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/examples/details/<path:example>'
        )
        self._api.add_resource(
            ApiTerraregExampleReadmeHtml,
            '/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/examples/readme_html/<path:example>'
        )
        self._api.add_resource(
            ApiTerraregExampleFileList,
            '/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/examples/filelist/<path:example>'
        )
        self._api.add_resource(
            ApiTerraregExampleFile,
            '/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/examples/file/<path:example_file>'
        )

        self._api.add_resource(
            ApiTerraregProviderLogos,
            '/v1/terrareg/provider_logos'
        )

        self._api.add_resource(
            ApiTerraregModuleSearchFilters,
            '/v1/terrareg/search_filters'
        )

        ## Auth endpoints /v1/terrareg/auth
        self._api.add_resource(
            ApiTerraregAdminAuthenticate,
            '/v1/terrareg/auth/admin/login'
        )
        self._api.add_resource(
            ApiTerraregIsAuthenticated,
            '/v1/terrareg/auth/admin/is_authenticated'
        )

        # Healthcheck endpoint
        self._api.add_resource(
            ApiTerraregHealth,
            '/v1/terrareg/health'
        )

    def _render_template(self, *args, **kwargs):
        """Override render_template, passing in base variables."""
        return render_template(
            *args, **kwargs,
            terrareg_application_name=terrareg.config.Config().APPLICATION_NAME,
            terrareg_logo_url=terrareg.config.Config().LOGO_URL,
            ALLOW_MODULE_HOSTING=terrareg.config.Config().ALLOW_MODULE_HOSTING,
            TRUSTED_NAMESPACE_LABEL=terrareg.config.Config().TRUSTED_NAMESPACE_LABEL,
            CONTRIBUTED_NAMESPACE_LABEL=terrareg.config.Config().CONTRIBUTED_NAMESPACE_LABEL,
            VERIFIED_MODULE_LABEL=terrareg.config.Config().VERIFIED_MODULE_LABEL,
            csrf_token=get_csrf_token()
        )

    def run(self, debug=None):
        """Run flask server."""
        kwargs = {
            'host': self.host,
            'port': self.port,
            'debug': terrareg.config.Config().DEBUG if debug is None else debug
        }
        if self.ssl_public_key and self.ssl_private_key:
            kwargs['ssl_context'] = (self.ssl_public_key, self.ssl_private_key)

        self._app.secret_key = terrareg.config.Config().SECRET_KEY

        self._app.run(**kwargs)

    def _module_provider_404(self, namespace: Namespace, module: Module,
                             module_provider_name: str):
        return self._render_template(
            'error.html',
            error_title='Module/Provider does not exist',
            error_description='The module {namespace}/{module}/{module_provider_name} does not exist'.format(
                namespace=namespace.name,
                module=module.name,
                module_provider_name=module_provider_name
            ),
            namespace=namespace,
            module=module,
            module_provider_name=module_provider_name
        ), 404

    def _view_serve_static_index(self):
        """Serve static index"""
        return self._render_template('index.html')

    def _view_serve_login(self):
        """Serve static login page."""
        return self._render_template('login.html')

    def _logout(self):
        """Remove cookie and redirect."""
        # Check if session exists in database and, if so,
        # delete it
        session_obj = Session.check_session(session_id=session.get('session_id', None))
        if session_obj:
            session_obj.delete()
        session['session_id'] = None

        session['is_admin_authenticated'] = False
        return redirect('/')

    def _view_serve_create_module(self):
        """Provide view to create module provider."""
        return self._render_template(
            'create_module_provider.html',
            git_providers=GitProvider.get_all(),
            ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER=terrareg.config.Config().ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER,
            ALLOW_CUSTOM_GIT_URL_MODULE_VERSION=terrareg.config.Config().ALLOW_CUSTOM_GIT_URL_MODULE_VERSION
        )


    def _view_serve_initial_setup(self):
        """Rendew view for initial setup."""
        return self._render_template('initial_setup.html')

    def _view_serve_namespace_list(self):
        """Render view for display module."""
        return self._render_template(
            'namespace_list.html'
        )

    @catch_name_exceptions
    def _view_serve_namespace(self, namespace):
        """Render view for namespace."""

        return self._render_template(
            'namespace.html',
            namespace=namespace
        )

    @catch_name_exceptions
    def _view_serve_module(self, namespace, name):
        """Render view for display module."""
        namespace = Namespace(namespace)
        module = Module(namespace=namespace, name=name)
        module_providers = module.get_providers()

        # If only one provider for module, redirect to it.
        if len(module_providers) == 1:
            return redirect(module_providers[0].get_view_url())
        else:
            return self._render_template(
                'module.html',
                namespace=namespace,
                module=module,
                module_providers=module_providers
            )

    @catch_name_exceptions
    def _view_serve_module_provider(self, namespace, name, provider, version=None):
        """Render view for displaying module provider information"""
        namespace = Namespace(namespace)
        module = Module(namespace=namespace, name=name)
        module_provider = ModuleProvider.get(module=module, name=provider)
        if module_provider is None:
            return self._module_provider_404(
                namespace=namespace,
                module=module,
                module_provider_name=provider)

        if version is None:
            module_version = module_provider.get_latest_version()

        else:
            module_version = ModuleVersion.get(module_provider=module_provider, version=version)

            if module_version is None:
                # If a version number was provided and it does not exist,
                # redirect to the module provider
                return redirect(module_provider.get_view_url())

        return self._render_template('module_provider.html')

    @catch_name_exceptions
    def _view_serve_submodule(self, namespace, name, provider, version, submodule_path):
        """Review view for displaying submodule"""
        namespace = Namespace(namespace)
        module = Module(namespace=namespace, name=name)
        module_provider = ModuleProvider(module=module, name=provider)
        module_version = ModuleVersion.get(module_provider=module_provider, version=version)

        if module_version is None:
            return redirect(module_provider.get_view_url())

        submodule = Submodule(module_version=module_version, module_path=submodule_path)

        return self._render_template('module_provider.html')

    @catch_name_exceptions
    def _view_serve_example(self, namespace, name, provider, version, submodule_path):
        """Review view for displaying example"""
        namespace = Namespace(namespace)
        module = Module(namespace=namespace, name=name)
        module_provider = ModuleProvider(module=module, name=provider)
        module_version = ModuleVersion.get(module_provider=module_provider, version=version)

        if module_version is None:
            return redirect(module_provider.get_view_url())

        submodule = Example(module_version=module_version, module_path=submodule_path)

        return self._render_template('module_provider.html')

    def _view_serve_module_search(self):
        """Search modules based on input."""
        return self._render_template('module_search.html')


class AuthenticationType(Enum):
    """Determine the method of authentication."""
    NOT_CHECKED = 0
    NOT_AUTHENTICATED = 1
    AUTHENTICATION_TOKEN = 2
    SESSION = 3


def get_csrf_token():
    """Return current session CSRF token."""
    return session.get('csrf_token', '')


def check_csrf_token(csrf_token):
    """Check CSRF token."""
    # If user is authenticated using authentication token,
    # do not required CSRF token
    if get_current_authentication_type() is AuthenticationType.AUTHENTICATION_TOKEN:
        return False

    session_token = get_csrf_token()
    if not session_token:
        raise NoSessionSetError('No session is presesnt to check CSRF token')
    elif session_token != csrf_token:
        raise IncorrectCSRFTokenError('CSRF token is incorrect')
    else:
        return True


def get_current_authentication_type():
    """Return the current authentication method of the user."""
    return g.get('authentication_type', AuthenticationType.NOT_CHECKED)


def check_admin_authentication():
    """Check authorization header is present or authenticated session"""
    authenticated = False
    g.authentication_type = AuthenticationType.NOT_AUTHENTICATED

    # Check that:
    # - An admin authentication token has been setup
    # - A token has neeif valid authorisation header has been passed
    if (terrareg.config.Config().ADMIN_AUTHENTICATION_TOKEN and
            request.headers.get('X-Terrareg-ApiKey', '') ==
            terrareg.config.Config().ADMIN_AUTHENTICATION_TOKEN):
        authenticated = True
        g.authentication_type = AuthenticationType.AUTHENTICATION_TOKEN

    # Check if authenticated via session
    # - Ensure session key has been setup
    if (terrareg.config.Config().SECRET_KEY and
            Session.check_session(session.get('session_id', None)) and
            session.get('is_admin_authenticated', False)):
        authenticated = True
        g.authentication_type = AuthenticationType.SESSION

    return authenticated


def require_admin_authentication(func):
    """Check user is authenticated as admin and either call function or return 401, if not."""
    @wraps(func)
    def wrapper(*args, **kwargs):
        if not check_admin_authentication():
            abort(401)
        else:
            return func(*args, **kwargs)
    return wrapper


def check_api_key_authentication(api_keys):
    """Check API key authentication."""
    # If user is authenticated as admin, allow
    if check_admin_authentication():
        return True
    # Check if no API keys have been configured
    # and allow request
    if not api_keys:
        return True

    # Check header against list of allowed API keys
    provided_api_key = request.headers.get('X-Terrareg-ApiKey', '')
    return provided_api_key and provided_api_key in api_keys


def require_api_authentication(api_keys):
    """Check user is authenticated using API key or as admin and either call function or return 401, if not."""
    def outer_wrapper(func):
        @wraps(func)
        def wrapper(*args, **kwargs):

            if not check_api_key_authentication(api_keys):
                abort(401)
            else:
                return func(*args, **kwargs)
        return wrapper
    return outer_wrapper


class ApiTerraformWellKnown(Resource):
    """Terraform .well-known discovery"""

    def get(self):
        """Return wellknown JSON"""
        return {
            "modules.v1": "/v1/modules/"
        }


class ErrorCatchingResource(Resource):
    """Provide resource that catches terrareg errors."""

    def _get(self, *args, **kwargs):
        """Placeholder for overridable get method."""
        return {'message': 'The method is not allowed for the requested URL.'}, 405

    def get(self, *args, **kwargs):
        """Run subclasses get in error handling fashion."""
        try:
            return self._get(*args, **kwargs)
        except TerraregError as exc:
            return {
                "status": "Error",
                "message": str(exc)
            }, 500

    def _post(self, *args, **kwargs):
        """Placeholder for overridable post method."""
        return {'message': 'The method is not allowed for the requested URL.'}, 405

    def post(self, *args, **kwargs):
        """Run subclasses post in error handling fashion."""
        try:
            return self._post(*args, **kwargs)
        except TerraregError as exc:
            return {
                "status": "Error",
                "message": str(exc)
            }, 500

    def _delete(self, *args, **kwargs):
        """Placeholder for overridable delete method."""
        return {'message': 'The method is not allowed for the requested URL.'}, 405

    def delete(self, *args, **kwargs):
        """Run subclasses delete in error handling fashion."""
        try:
            return self._delete(*args, **kwargs)
        except TerraregError as exc:
            return {
                "status": "Error",
                "message": str(exc)
            }, 500

    def _get_404_response(self):
        """Return common 404 error"""
        return {'errors': ['Not Found']}, 404

    def _get_401_response(self):
        """Return standardised 401."""
        return {'message': ('The server could not verify that you are authorized to access the URL requested. '
                            'You either supplied the wrong credentials (e.g. a bad password), '
                            'or your browser doesn\'t understand how to supply the credentials required.')
        }, 401


class ApiTerraregHealth(ErrorCatchingResource):
    """Endpoint to return 200 when healthy."""

    def _get(self):
        """Return static 200"""
        return {
            "message": "Ok"
        }


class ApiTerraregConfig(ErrorCatchingResource):
    """Endpoint to return config used by UI."""

    def _get(self):
        """Return config."""
        return {
            'TRUSTED_NAMESPACE_LABEL': terrareg.config.Config().TRUSTED_NAMESPACE_LABEL,
            'CONTRIBUTED_NAMESPACE_LABEL': terrareg.config.Config().CONTRIBUTED_NAMESPACE_LABEL,
            'VERIFIED_MODULE_LABEL': terrareg.config.Config().VERIFIED_MODULE_LABEL,
            'ANALYTICS_TOKEN_PHRASE': terrareg.config.Config().ANALYTICS_TOKEN_PHRASE,
            'ANALYTICS_TOKEN_DESCRIPTION': terrareg.config.Config().ANALYTICS_TOKEN_DESCRIPTION,
            'EXAMPLE_ANALYTICS_TOKEN': terrareg.config.Config().EXAMPLE_ANALYTICS_TOKEN,
            'ALLOW_MODULE_HOSTING': terrareg.config.Config().ALLOW_MODULE_HOSTING,
            'UPLOAD_API_KEYS_ENABLED': bool(terrareg.config.Config().UPLOAD_API_KEYS),
            'PUBLISH_API_KEYS_ENABLED': bool(terrareg.config.Config().PUBLISH_API_KEYS),
            'DISABLE_TERRAREG_EXCLUSIVE_LABELS': terrareg.config.Config().DISABLE_TERRAREG_EXCLUSIVE_LABELS,
            'ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER': terrareg.config.Config().ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER,
            'ALLOW_CUSTOM_GIT_URL_MODULE_VERSION': terrareg.config.Config().ALLOW_CUSTOM_GIT_URL_MODULE_VERSION,
            'ADMIN_AUTHENTICATION_TOKEN_ENABLED': bool(terrareg.config.Config().ADMIN_AUTHENTICATION_TOKEN),
            'SECRET_KEY_SET': bool(terrareg.config.Config().SECRET_KEY)
        }


class ApiTerraregInitialSetupData(ErrorCatchingResource):
    """Interface to provide data to the initial setup page."""

    def _get(self):
        """Return information for steps for setting up Terrareg."""
        # Get first namespace, if present
        namespace = None
        module = None
        module_provider = None
        namespaces = Namespace.get_all(only_published=False)
        version = None
        integrations = {}
        if namespaces:
            namespace = namespaces[0]
        if namespace:
            modules = namespace.get_all_modules()
            if modules:
                module = modules[0]
        if module:
            providers = module.get_providers()
            if providers:
                module_provider = providers[0]
                integrations = module_provider.get_integrations()

        if module_provider:
            versions = module_provider.get_versions(include_beta=True, include_unpublished=True)
            if versions:
                version = versions[0]

        return {
            "module_created": bool(module_provider),
            "version_indexed": bool(version),
            "version_published": bool(version.published) if version else False,
            "module_configured_with_git": bool(module_provider.get_git_clone_url()) if module_provider else False,
            "module_view_url": module_provider.get_view_url() if module_provider else None,
            "module_upload_endpoint": integrations['upload']['url'] if 'upload' in integrations else None,
            "module_publish_endpoint": integrations['publish']['url'] if 'publish' in integrations else None
        }


class ApiTerraregGitProviders(ErrorCatchingResource):
    """Interface to obtain git provider configurations."""

    def _get(self):
        """Return list of git providers"""
        return [
            {
                'id': git_provider.pk,
                'name': git_provider.name
            }
            for git_provider in GitProvider.get_all()
        ]


class ApiTerraregModuleProviderIntegrations(ErrorCatchingResource):
    """Intereface to provide list of integration URLs"""

    def _get(self, namespace, name, provider):
        """Return list of integration URLs"""
        namespace = Namespace(namespace)
        module = Module(namespace=namespace, name=name)
        # Get module provider and, optionally create, if it doesn't exist
        module_provider = ModuleProvider.get(module=module, name=provider)

        integrations = module_provider.get_integrations()

        return [
            integrations[integration]
            for integration in ['upload', 'import', 'hooks_bitbucket', 'hooks_github', 'hooks_gitlab', 'publish']
            if integration in integrations
        ]


class ApiModuleVersionUpload(ErrorCatchingResource):

    ALLOWED_EXTENSIONS = ['zip']

    method_decorators = [require_api_authentication(terrareg.config.Config().UPLOAD_API_KEYS)]

    def allowed_file(self, filename):
        """Check if file has allowed file-extension"""
        return '.' in filename and \
               filename.rsplit('.', 1)[1].lower() in self.ALLOWED_EXTENSIONS

    def _post(self, namespace, name, provider, version):
        """Handle module version upload."""

        # If module hosting is disabled,
        # refuse the upload of new modules
        if not terrareg.config.Config().ALLOW_MODULE_HOSTING:
            return {'message': 'Module upload is disabled.'}, 400

        with Database.start_transaction():
            namespace = Namespace(namespace)
            module = Module(namespace=namespace, name=name)
            # Get module provider and, optionally create, if it doesn't exist
            module_provider = ModuleProvider.get(module=module, name=provider, create=True)
            module_version = ModuleVersion(module_provider=module_provider, version=version)

            if len(request.files) != 1:
                raise UploadError('One file can be uploaded')

            file = request.files[[f for f in request.files.keys()][0]]

            # If the user does not select a file, the browser submits an
            # empty file without a filename.
            if file.filename == '':
                raise UploadError('No selected file')

            if not file or not self.allowed_file(file.filename):
                raise UploadError('Error occurred - unknown file extension')

            module_version.prepare_module()
            with ApiUploadModuleExtractor(upload_file=file, module_version=module_version) as me:
                me.process_upload()

            return {
                'status': 'Success'
            }


class ApiModuleVersionCreate(ErrorCatchingResource):
    """Provide interface to create release for git-backed modules."""

    method_decorators = [require_api_authentication(terrareg.config.Config().UPLOAD_API_KEYS)]

    def _post(self, namespace, name, provider, version):
        """Handle creation of module version."""
        with Database.start_transaction():
            namespace = Namespace(name=namespace)
            module = Module(namespace=namespace, name=name)
            # Get module provider and optionally create, if it doesn't exist
            module_provider = ModuleProvider.get(module=module, name=provider, create=True)

            # Ensure module provider exists
            if not module_provider:
                return {'message': 'Module provider does not exist'}, 400

            # Ensure that the module provider has a repository url configured.
            if not module_provider.get_git_clone_url():
                return {'message': 'Module provider is not configured with a repository'}, 400

            module_version = ModuleVersion(module_provider=module_provider, version=version)

            module_version.prepare_module()
            with GitModuleExtractor(module_version=module_version) as me:
                me.process_upload()

            return {
                'status': 'Success'
            }


class ApiModuleVersionCreateBitBucketHook(ErrorCatchingResource):
    """Provide interface for bitbucket hook to detect pushes of new tags."""

    def _post(self, namespace, name, provider):
        """Create new version based on bitbucket hooks."""
        with Database.start_transaction():
            namespace = Namespace(name=namespace)
            module = Module(namespace=namespace, name=name)
            # Get module provider and optionally create, if it doesn't exist
            module_provider = ModuleProvider.get(module=module, name=provider, create=True)

            # Validate signature
            if terrareg.config.Config().UPLOAD_API_KEYS:
                # Get signature from request
                request_signature = request.headers.get('X-Hub-Signature', '')
                # Remove 'sha256=' from beginning of header
                request_signature = re.sub(r'^sha256=', '', request_signature)
                # Iterate through each of the keys and test
                for test_key in terrareg.config.Config().UPLOAD_API_KEYS:
                    # Generate
                    valid_signature = hmac.new(bytes(test_key, 'utf8'), b'', hashlib.sha256)
                    valid_signature.update(request.data)
                    # If the signatures match, break from loop
                    if hmac.compare_digest(valid_signature.hexdigest(), request_signature):
                        break
                # If a valid signature wasn't found with one of the configured keys,
                # return 401
                else:
                    return self._get_401_response()

            if not module_provider.get_git_clone_url():
                return {'message': 'Module provider is not configured with a repository'}, 400

            bitbucket_data = request.json

            if not ('changes' in bitbucket_data and type(bitbucket_data['changes']) == list):
                return {'message': 'List of changes not found in payload'}, 400

            imported_versions = {}
            error = False

            for change in bitbucket_data['changes']:

                # Check that change is a dict
                if not type(change) is dict:
                    continue

                # Check if change type is tag
                if not ('ref' in change and
                        type(change['ref']) is dict and
                        'type' in change['ref'] and
                        type(change['ref']['type']) == str and
                        change['ref']['type'] == 'TAG'):
                    continue

                # Check type of change is an ADD or UPDATE
                if not ('type' in change and
                        type(change['type']) is str and
                        change['type'] in ['ADD', 'UPDATE']):
                    continue

                # Obtain tag name
                tag_ref = change['ref']['id'] if 'id' in change['ref'] else None

                # Attempt to match version against regex
                version = module_provider.get_version_from_tag_ref(tag_ref)

                if not version:
                    continue

                # Create module version
                module_version = ModuleVersion(module_provider=module_provider, version=version)

                # Perform import from git
                try:
                    module_version.prepare_module()
                    with GitModuleExtractor(module_version=module_version) as me:
                        me.process_upload()
                except TerraregError as exc:
                    imported_versions[version] = {
                        'status': 'Failed',
                        'message': str(exc)
                    }
                    continue

                imported_versions[version] = {
                    'status': 'Success'
                }

            if error:
                return {
                    'status': 'Error',
                    'message': 'One or more tags failed to import',
                    'tags': imported_versions
                }, 500
            return {
                'status': 'Success',
                'message': 'Imported all provided tags',
                'tags': imported_versions
            }


class ApiModuleList(ErrorCatchingResource):
    def _get(self):
        """Return list of modules."""
        parser = reqparse.RequestParser()
        parser.add_argument(
            'offset', type=int,
            default=0, help='Pagination offset')
        parser.add_argument(
            'limit', type=int,
            default=10, help='Pagination limit'
        )
        parser.add_argument(
            'provider', type=str,
            default=None, help='Limits modules to a specific provider.',
            action='append', dest='providers'
        )
        parser.add_argument(
            'verified', type=inputs.boolean,
            default=False, help='Limits modules to only verified modules.'
        )

        args = parser.parse_args()

        search_results = ModuleSearch.search_module_providers(
            providers=args.providers,
            verified=args.verified,
            offset=args.offset,
            limit=args.limit
        )

        return {
            "meta": search_results.meta,
            "modules": [
                module_provider.get_latest_version().get_api_outline()
                for module_provider in search_results.module_providers
            ]
        }


class ApiModuleSearch(ErrorCatchingResource):

    def _get(self):
        """Search for modules, given query string, namespace or provider."""
        parser = reqparse.RequestParser()
        parser.add_argument(
            'q', type=str,
            required=True,
            help='The search string.'
        )
        parser.add_argument(
            'offset', type=int,
            default=0, help='Pagination offset')
        parser.add_argument(
            'limit', type=int,
            default=10, help='Pagination limit'
        )
        parser.add_argument(
            'provider', type=str,
            default=None, help='Limits modules to a specific provider.',
            action='append', dest='providers'
        )
        parser.add_argument(
            'namespace', type=str,
            default=None, help='Limits modules to a specific namespace.',
            action='append', dest='namespaces'
        )
        parser.add_argument(
            'verified', type=inputs.boolean,
            default=False, help='Limits modules to only verified modules.'
        )

        parser.add_argument(
            'trusted_namespaces', type=inputs.boolean,
            default=None, help='Limits modules to include trusted namespaces.'
        )
        parser.add_argument(
            'contributed', type=inputs.boolean,
            default=None, help='Limits modules to include contributed modules.'
        )

        args = parser.parse_args()

        namespace_trust_filters = NamespaceTrustFilter.UNSPECIFIED
        # If either trusted namepsaces or contributed have been provided
        # (irrelevant of whether they are set to true or false),
        # setup the filter to no longer be unspecified.
        if args.trusted_namespaces is not None or args.contributed is not None:
            namespace_trust_filters = []

        if args.trusted_namespaces:
            namespace_trust_filters.append(NamespaceTrustFilter.TRUSTED_NAMESPACES)
        if args.contributed:
            namespace_trust_filters.append(NamespaceTrustFilter.CONTRIBUTED)

        search_results = ModuleSearch.search_module_providers(
            query=args.q,
            namespaces=args.namespaces,
            providers=args.providers,
            verified=args.verified,
            namespace_trust_filters=namespace_trust_filters,
            offset=args.offset,
            limit=args.limit
        )

        return {
            "meta": search_results.meta,
            "modules": [
                module_provider.get_latest_version().get_api_outline()
                for module_provider in search_results.module_providers
            ]
        }

class ApiNamespaceModules(ErrorCatchingResource):
    """Interface to obtain list of modules in namespace."""

    def _get(self, namespace):
        """Return list of modules in namespace"""

        parser = reqparse.RequestParser()
        parser.add_argument(
            'offset', type=int,
            default=0, help='Pagination offset')
        parser.add_argument(
            'limit', type=int,
            default=10, help='Pagination limit'
        )
        args = parser.parse_args()

        search_results = ModuleSearch.search_module_providers(
            offset=args.offset,
            limit=args.limit,
            namespaces=[namespace],
            include_internal=True
        )

        if not search_results.module_providers:
            return self._get_404_response()

        return {
            "meta": search_results.meta,
            "modules": [
                module_provider.get_latest_version().get_api_outline()
                for module_provider in search_results.module_providers
            ]
        }

class ApiModuleDetails(ErrorCatchingResource):
    def _get(self, namespace, name):
        """Return latest version for each module provider."""

        parser = reqparse.RequestParser()
        parser.add_argument(
            'offset', type=int,
            default=0, help='Pagination offset')
        parser.add_argument(
            'limit', type=int,
            default=10, help='Pagination limit'
        )
        args = parser.parse_args()

        namespace, _ = Namespace.extract_analytics_token(namespace)

        search_results = ModuleSearch.search_module_providers(
            offset=args.offset,
            limit=args.limit,
            namespaces=[namespace],
            modules=[name]
        )

        if not search_results.module_providers:
            return self._get_404_response()

        return {
            "meta": search_results.meta,
            "modules": [
                module_provider.get_latest_version().get_api_outline()
                for module_provider in search_results.module_providers
            ]
        }


class ApiModuleProviderDetails(ErrorCatchingResource):

    def _get(self, namespace, name, provider):
        """Return list of version."""

        namespace, _ = Namespace.extract_analytics_token(namespace)
        namespace = Namespace(namespace)
        module = Module(namespace=namespace, name=name)
        module_provider = ModuleProvider(module=module, name=provider)
        module_version = module_provider.get_latest_version()

        if not module_version:
            return self._get_404_response()

        return module_version.get_api_details()


class ApiModuleVersionDetails(ErrorCatchingResource):

    def _get(self, namespace, name, provider, version):
        """Return list of version."""

        namespace, _ = Namespace.extract_analytics_token(namespace)
        namespace = Namespace(namespace)
        module = Module(namespace=namespace, name=name)
        module_provider = ModuleProvider(module=module, name=provider)
        module_version = ModuleVersion.get(module_provider=module_provider, version=version)

        if module_version is None:
            return self._get_404_response()

        return module_version.get_api_details()


class ApiModuleVersions(ErrorCatchingResource):

    def _get(self, namespace, name, provider):
        """Return list of version."""

        namespace, _ = Namespace.extract_analytics_token(namespace)
        namespace = Namespace(namespace)
        module = Module(namespace=namespace, name=name)
        module_provider = ModuleProvider(module=module, name=provider)
        return {
            "modules": [
                {
                    "source": "{namespace}/{module}/{provider}".format(
                        namespace=namespace.name,
                        module=module.name,
                        provider=module_provider.name
                    ),
                    "versions": [
                        {
                            "version": v.version,
                            "root": {
                                # @TODO: Add providers/depdencies
                                "providers": [],
                                "dependencies": []
                            },
                            # @TODO: Add submodule information
                            "submodules": []
                        }
                        for v in module_provider.get_versions()
                    ]
                }
            ]
        }


class ApiModuleVersionDownload(ErrorCatchingResource):
    """Provide download endpoint."""

    def _get(self, namespace, name, provider, version):
        """Provide download header for location to download source."""
        namespace, analytics_token = Namespace.extract_analytics_token(namespace)
        namespace = Namespace(namespace)
        module = Module(namespace=namespace, name=name)
        module_provider = ModuleProvider(module=module, name=provider)
        module_version = ModuleVersion.get(module_provider=module_provider, version=version)

        if module_version is None:
            return self._get_404_response()

        # Determine if module download should be rejected due to
        # non-existent analytics token
        if not analytics_token and not terrareg.config.Config().ALLOW_UNIDENTIFIED_DOWNLOADS:
            return make_response(
                ("\nAn {analytics_token_phrase} must be provided.\n"
                 "Please update module source to include {analytics_token_phrase}.\n"
                 "\nFor example:\n  source = \"{host}/{example_analytics_token}__{namespace}/{module_name}/{provider}\"").format(
                    analytics_token_phrase=terrareg.config.Config().ANALYTICS_TOKEN_PHRASE,
                    host=request.host,
                    example_analytics_token=terrareg.config.Config().EXAMPLE_ANALYTICS_TOKEN,
                    namespace=namespace.name,
                    module_name=module.name,
                    provider=module_provider.name
                ),
                401
            )

        auth_token = None
        auth_token_match = re.match(r'Bearer (.*)', request.headers.get('Authorization', ''))
        if auth_token_match:
            auth_token = auth_token_match.group(1)

        # Record download
        AnalyticsEngine.record_module_version_download(
            module_version=module_version,
            analytics_token=analytics_token,
            terraform_version=request.headers.get('X-Terraform-Version', None),
            user_agent=request.headers.get('User-Agent', None),
            auth_token=auth_token
        )

        resp = make_response('', 204)
        resp.headers['X-Terraform-Get'] = module_version.get_source_download_url()
        return resp


class ApiModuleVersionSourceDownload(ErrorCatchingResource):
    """Return source package of module version"""

    def _get(self, namespace, name, provider, version):
        """Return static file."""
        if not terrareg.config.Config().ALLOW_MODULE_HOSTING:
            return {'message': 'Module hosting is disbaled'}, 500

        namespace = Namespace(namespace)
        module = Module(namespace=namespace, name=name)
        module_provider = ModuleProvider(module=module, name=provider)
        module_version = ModuleVersion(module_provider=module_provider, version=version)
        return send_from_directory(module_version.base_directory, module_version.archive_name_zip)


class ApiModuleProviderDownloadsSummary(ErrorCatchingResource):
    """Provide download summary for module provider."""

    def _get(self, namespace, name, provider):
        """Return list of download counts for module provider."""
        namespace = Namespace(namespace)
        module = Module(namespace=namespace, name=name)
        module_provider = ModuleProvider.get(module=module, name=provider)
        if module_provider is None:
            return self._get_404_response()

        return {
            "data": {
                "type": "module-downloads-summary",
                "id": module_provider.id,
                "attributes": AnalyticsEngine.get_module_provider_download_stats(module_provider)
            }
        }


class ApiTerraregNamespaces(ErrorCatchingResource):
    """Provide interface to obtain namespaces."""

    def _get(self):
        """Return list of namespaces."""
        namespaces = Namespace.get_all(only_published=False)

        return [
            {
                "name": namespace.name,
                "view_href": namespace.get_view_url()
            }
            for namespace in namespaces
        ]



class ApiTerraregProviderLogos(ErrorCatchingResource):
    """Provide interface to obtain all provider logo details"""

    def _get(self):
        """Return all details about provider logos."""
        return {
            provider_logo.provider: {
                'source': provider_logo.source,
                'alt': provider_logo.alt,
                'tos': provider_logo.tos,
                'link': provider_logo.link
            }
            for provider_logo in ProviderLogo.get_all()
        }


class ApiTerraregGlobalStatsSummary(ErrorCatchingResource):
    """Provide global download stats for homepage."""

    def _get(self):
        """Return number of namespaces, modules, module versions and downloads"""
        return {
            'namespaces': Namespace.get_total_count(),
            'modules': ModuleProvider.get_total_count(),
            'module_versions': ModuleVersion.get_total_count(),
            'downloads': AnalyticsEngine.get_total_downloads()
        }


class ApiTerraregGlobalUsageStats(ErrorCatchingResource):
    """Provide interface to obtain statistics about global module usage."""

    def _get(self):
        """
        Return stats on total module providers,
        total unique analytics tokens per module
        (with and without auth token).
        """
        module_usage_with_auth_token = AnalyticsEngine.get_global_module_usage()
        module_usage_including_empty_auth_token = AnalyticsEngine.get_global_module_usage(include_empty_auth_token=True)
        total_analytics_token_with_auth_token = sum(module_usage_with_auth_token.values())
        total_analytics_token_including_empty_auth_token = sum(module_usage_including_empty_auth_token.values())
        return {
            'module_provider_count': ModuleProvider.get_total_count(),
            'module_provider_usage_breakdown_with_auth_token': module_usage_with_auth_token,
            'module_provider_usage_count_with_auth_token': total_analytics_token_with_auth_token,
            'module_provider_usage_including_empty_auth_token': module_usage_including_empty_auth_token,
            'module_provider_usage_count_including_empty_auth_token': total_analytics_token_including_empty_auth_token
        }


class ApiTerraregMostRecentlyPublishedModuleVersion(ErrorCatchingResource):
    """Return data for most recently published module version."""

    def _get(self):
        """Return number of namespaces, modules, module versions and downloads"""
        module_version = ModuleSearch.get_most_recently_published()
        if not module_version:
            return {}, 404
        return module_version.get_api_outline()


class ApiTerraregMostDownloadedModuleProviderThisWeek(ErrorCatchingResource):
    """Return data for most downloaded module provider this week."""

    def _get(self):
        """Return most downloaded module this week"""
        module_provider = ModuleSearch.get_most_downloaded_module_provider_this_Week()
        if not module_provider:
            return {}, 404

        return module_provider.get_latest_version().get_api_outline()


class ApiTerraregModuleProviderAnalyticsTokenVersions(ErrorCatchingResource):
    """Provide download summary for module provider."""

    def _get(self, namespace, name, provider):
        """Return list of download counts for module provider."""
        namespace = Namespace(namespace)
        module = Module(namespace=namespace, name=name)
        module_provider = ModuleProvider(module=module, name=provider)
        return AnalyticsEngine.get_module_provider_token_versions(module_provider)


class ApiTerraregNamespaceDetails(ErrorCatchingResource):
    """Interface to obtain custom terrareg namespace details."""

    def _get(self, namespace):
        """Return custom terrareg config for namespace."""
        namespace = Namespace(namespace)
        return namespace.get_details()


class ApiTerraregNamespaceModules(ErrorCatchingResource):
    """Interface to obtain list of modules in namespace."""

    def _get(self, namespace):
        """Return list of modules in namespace"""

        parser = reqparse.RequestParser()
        parser.add_argument(
            'offset', type=int,
            default=0, help='Pagination offset')
        parser.add_argument(
            'limit', type=int,
            default=10, help='Pagination limit'
        )
        args = parser.parse_args()

        namespace_obj = Namespace(name=namespace)
        module_providers = [
            module_provider
            for module in namespace_obj.get_all_modules()
            for module_provider in module.get_providers()
        ]
        if not module_providers:
            return self._get_404_response()

        meta = {
            'limit': args.limit,
            'current_offset': args.offset
        }
        if len(module_providers) > (args.offset + args.limit):
            meta['next_offset'] = (args.offset + args.limit)
        if args.offset > 0:
            meta['prev_offset'] = max(args.offset - args.limit, 0)

        return {
            "meta": meta,
            "modules": [
                module_provider.get_api_outline()
                if module_provider.get_latest_version() is None else
                module_provider.get_latest_version().get_api_outline()
                for module_provider in module_providers[args.offset:args.offset + args.limit]
            ]
        }


class ApiTerraregModuleProviderDetails(ErrorCatchingResource):
    """Interface to obtain module provider details."""

    def _get(self, namespace, name, provider):
        """Return details about module version."""
        namespace = Namespace(namespace)
        module = Module(namespace=namespace, name=name)
        module_provider = ModuleProvider.get(module=module, name=provider)

        if module_provider is None:
            return self._get_404_response()

        # If a version exists, obtain the details for that
        latest_version = module_provider.get_latest_version()
        if latest_version is not None:
            return latest_version.get_terrareg_api_details()

        # Otherwise, return module provider details
        return module_provider.get_terrareg_api_details()


class ApiTerraregModuleVersionDetails(ErrorCatchingResource):
    """Interface to obtain module verison details."""

    def _get(self, namespace, name, provider, version=None):
        """Return details about module version."""
        namespace = Namespace(namespace)
        module = Module(namespace=namespace, name=name)
        module_provider = ModuleProvider.get(module=module, name=provider)
        if module_provider is None:
            return self._get_404_response()

        if version is not None:
            module_version = ModuleVersion.get(module_provider=module_provider, version=version)
        else:
            module_version = module_provider.get_latest_version()

        if module_version is None:
            return self._get_404_response()

        return module_version.get_terrareg_api_details()


class ApiTerraregModuleVersionVariableTemplate(ErrorCatchingResource):
    """Provide variable template for module version."""

    def _get(self, namespace, name, provider, version):
        """Return variable template."""
        namespace = Namespace(namespace)
        module = Module(namespace=namespace, name=name)
        module_provider = ModuleProvider(module=module, name=provider)
        module_version = ModuleVersion(module_provider=module_provider, version=version)
        return module_version.variable_template


class ApiTerraregModuleVersionReadmeHtml(ErrorCatchingResource):
    """Provide variable template for module version."""

    def _get(self, namespace, name, provider, version):
        """Return variable template."""
        namespace = Namespace(namespace)
        module = Module(namespace=namespace, name=name)
        module_provider = ModuleProvider(module=module, name=provider)
        module_version = ModuleVersion(module_provider=module_provider, version=version)
        return module_version.get_readme_html(server_hostname=request.host)


class ApiTerraregModuleSearchFilters(ErrorCatchingResource):
    """Return list of filters availabe for search."""

    def _get(self):
        """Return list of available filters and filter counts for search query."""
        parser = reqparse.RequestParser()
        parser.add_argument(
            'q', type=str,
            required=True,
            help='The search string.'
        )
        args = parser.parse_args()

        return ModuleSearch.get_search_filters(query=args.q)


class ApiTerraregIsAuthenticated(ErrorCatchingResource):
    """Interface to teturn whether user is authenticated as an admin."""

    method_decorators = [require_admin_authentication]

    def _get(self):
        return {'authenticated': True}


class ApiTerraregAdminAuthenticate(ErrorCatchingResource):
    """Interface to perform authentication as an admin and set appropriate cookie."""

    method_decorators = [require_admin_authentication]

    def _post(self):
        """Handle POST requests to the authentication endpoint."""

        if not terrareg.config.Config().SECRET_KEY:
            return {'message': 'Sessions not enabled in configuration'}, 403

        # Check if a session already exists and delete it
        if session.get('session_id', None):
            session_obj = Session.check_session(session.get('session_id', None))
            if session_obj:
                session_obj.delete()

        session['is_admin_authenticated'] = True
        session['csrf_token'] = hashlib.sha1(os.urandom(64)).hexdigest()
        session_obj = Session.create_session()
        session['session_id'] = session_obj.id
        session.modified = True

        # Whilst authenticating a user, piggyback the request
        # to take the opportunity to delete old sessions
        Session.cleanup_old_sessions()

        return {'authenticated': True}


class ApiTerraregModuleProviderCreate(ErrorCatchingResource):
    """Provide interface to create module provider."""

    method_decorators = [require_admin_authentication]

    def _post(self, namespace, name, provider):
        """Handle update to settings."""
        parser = reqparse.RequestParser()
        parser.add_argument(
            'git_provider_id', type=str,
            required=False,
            default=None,
            help='ID of the git provider to associate to module provider.',
            location='json'
        )
        parser.add_argument(
            'repo_base_url_template', type=str,
            required=False,
            default=None,
            help='Templated base git URL.',
            location='json'
        )
        parser.add_argument(
            'repo_clone_url_template', type=str,
            required=False,
            default=None,
            help='Templated git clone URL.',
            location='json'
        )
        parser.add_argument(
            'repo_browse_url_template', type=str,
            required=False,
            default=None,
            help='Templated URL for browsing repository.',
            location='json'
        )
        parser.add_argument(
            'git_tag_format', type=str,
            required=False,
            default=None,
            help='Module provider git tag format.',
            location='json'
        )
        parser.add_argument(
            'git_path', type=str,
            required=False,
            default=None,
            help='Path within git repository that the module exists.',
            location='json'
        )
        parser.add_argument(
            'csrf_token', type=str,
            required=False,
            help='CSRF token',
            location='json',
            default=None
        )

        args = parser.parse_args()

        check_csrf_token(args.csrf_token)

        # Update repository URL of module version
        namespace = Namespace(name=namespace)
        module = Module(namespace=namespace, name=name)

        # Check if module provider already exists
        module_provider = ModuleProvider.get(module=module, name=provider)
        if module_provider is not None:
            return {'message': 'Module provider already exists'}, 400

        with Database.start_transaction() as transaction_context:
            module_provider = ModuleProvider.create(module=module, name=provider)

            # If git provider ID has been specified,
            # validate it and update attribute of module provider.
            if args.git_provider_id is not None:
                git_provider = GitProvider.get(id=args.git_provider_id)
                # If a non-empty git provider ID was provided and none
                # were returned, return an error about invalid
                # git provider ID
                if args.git_provider_id and git_provider is None:
                    transaction_context.transaction.rollback()
                    return {'message': 'Git provider does not exist.'}, 400

                module_provider.update_git_provider(git_provider=git_provider)

            # Ensure base repository URL is parsable
            repo_base_url_template = args.repo_base_url_template
            # If the argument is None, assume it's not being updated,
            # as this is the default value for the arg parser.
            if repo_base_url_template is not None:
                if repo_base_url_template == '':
                    # If repository URL is empty, set to None
                    repo_base_url_template = None

                try:
                    module_provider.update_repo_base_url_template(repo_base_url_template=repo_base_url_template)
                except RepositoryUrlParseError as exc:
                    transaction_context.transaction.rollback()
                    return {'message': 'Repo base URL: {}'.format(str(exc))}, 400

            # Ensure repository URL is parsable
            repo_clone_url_template = args.repo_clone_url_template
            # If the argument is None, assume it's not being updated,
            # as this is the default value for the arg parser.
            if repo_clone_url_template is not None:
                if repo_clone_url_template == '':
                    # If repository URL is empty, set to None
                    repo_clone_url_template = None

                try:
                    module_provider.update_repo_clone_url_template(repo_clone_url_template=repo_clone_url_template)
                except RepositoryUrlParseError as exc:
                    transaction_context.transaction.rollback()
                    return {'message': 'Repo clone URL: {}'.format(str(exc))}, 400

            # Ensure repository URL is parsable
            repo_browse_url_template = args.repo_browse_url_template
            if repo_browse_url_template is not None:
                if repo_browse_url_template == '':
                    # If repository URL is empty, set to None
                    repo_browse_url_template = None

                try:
                    module_provider.update_repo_browse_url_template(repo_browse_url_template=repo_browse_url_template)
                except RepositoryUrlParseError as exc:
                    transaction_context.transaction.rollback()
                    return {'message': 'Repo browse URL: {}'.format(str(exc))}, 400

            # Update git tag format of object
            git_tag_format = args.git_tag_format
            if git_tag_format is not None:
                module_provider.update_git_tag_format(git_tag_format=git_tag_format)

            # Update git path
            git_path = args.git_path
            if git_path is not None:
                module_provider.update_git_path(git_path=git_path)

        return {
            'id': module_provider.id
        }


class ApiTerraregModuleProviderDelete(ErrorCatchingResource):
    """Provide interface to delete module provider."""

    method_decorators = [require_admin_authentication]

    def _delete(self, namespace, name, provider):
        """Delete module provider."""
        parser = reqparse.RequestParser()
        parser.add_argument(
            'csrf_token', type=str,
            required=False,
            help='CSRF token',
            location='json',
            default=None
        )

        args = parser.parse_args()

        check_csrf_token(args.csrf_token)

        # Update repository URL of module version
        namespace = Namespace(name=namespace)
        module = Module(namespace=namespace, name=name)

        # Check if module provider already exists
        module_provider = ModuleProvider.get(module=module, name=provider)
        if module_provider is None:
            return {'message': 'Module provider does not exist'}, 400

        module_provider.delete()

class ApiTerraregModuleVersionDelete(ErrorCatchingResource):
    """Provide interface to delete module version."""

    method_decorators = [require_admin_authentication]

    def _delete(self, namespace, name, provider, version):
        """Delete module version."""
        parser = reqparse.RequestParser()
        parser.add_argument(
            'csrf_token', type=str,
            required=False,
            help='CSRF token',
            location='json',
            default=None
        )

        args = parser.parse_args()

        check_csrf_token(args.csrf_token)

        # Update repository URL of module version
        namespace_obj = Namespace(name=namespace)
        module = Module(namespace=namespace_obj, name=name)

        # Check if module provider already exists
        module_provider_obj = ModuleProvider.get(module=module, name=provider)
        if module_provider_obj is None:
            return {'message': 'Module provider does not exist'}, 400

        # Check if module version already exists
        version_obj = ModuleVersion.get(module_provider=module_provider_obj, version=version)
        if version_obj is None:
            return {'message': 'Module version does not exist'}, 400

        version_obj.delete()

        return {
            'status': 'Success'
        }


class ApiTerraregModuleProviderSettings(ErrorCatchingResource):
    """Provide interface to update module provider settings."""

    method_decorators = [require_admin_authentication]

    def _post(self, namespace, name, provider):
        """Handle update to settings."""
        parser = reqparse.RequestParser()
        parser.add_argument(
            'git_provider_id', type=str,
            required=False,
            default=None,
            help='ID of the git provider to associate to module provider.',
            location='json'
        )
        parser.add_argument(
            'repo_base_url_template', type=str,
            required=False,
            default=None,
            help='Templated base git repository URL.',
            location='json'
        )
        parser.add_argument(
            'repo_clone_url_template', type=str,
            required=False,
            default=None,
            help='Templated git clone URL.',
            location='json'
        )
        parser.add_argument(
            'repo_browse_url_template', type=str,
            required=False,
            default=None,
            help='Templated URL for browsing repository.',
            location='json'
        )
        parser.add_argument(
            'git_tag_format', type=str,
            required=False,
            default=None,
            help='Module provider git tag format.',
            location='json'
        )
        parser.add_argument(
            'git_path', type=str,
            required=False,
            default=None,
            help='Path within git repository that the module exists.',
            location='json'
        )
        parser.add_argument(
            'verified', type=inputs.boolean,
            required=False,
            default=None,
            help='Whether module provider is marked as verified.',
            location='json'
        )
        parser.add_argument(
            'csrf_token', type=str,
            required=False,
            help='CSRF token',
            location='json',
            default=None
        )

        args = parser.parse_args()

        check_csrf_token(args.csrf_token)

        # Update repository URL of module version
        namespace = Namespace(name=namespace)
        module = Module(namespace=namespace, name=name)
        module_provider = ModuleProvider.get(module=module, name=provider)
    
        if not module_provider:
            return {'message': 'Module provider does not exist'}, 400

        # If git provider ID has been specified,
        # validate it and update attribute of module provider.
        if args.git_provider_id is not None:
            git_provider = GitProvider.get(id=args.git_provider_id)
            # If a non-empty git provider ID was provided and none
            # were returned, return an error about invalid
            # git provider ID
            if args.git_provider_id and git_provider is None:
                return {'message': 'Git provider does not exist.'}, 400

            module_provider.update_git_provider(git_provider=git_provider)

        # Ensure base URL is parsable
        repo_base_url_template = args.repo_base_url_template
        # If the argument is None, assume it's not being updated,
        # as this is the default value for the arg parser.
        if repo_base_url_template is not None:
            if repo_base_url_template == '':
                # If repository URL is empty, set to None
                repo_base_url_template = None

            try:
                module_provider.update_repo_base_url_template(repo_base_url_template=repo_base_url_template)
            except RepositoryUrlParseError as exc:
                return {'message': 'Repo base URL: {}'.format(str(exc))}, 400

        # Ensure repository URL is parsable
        repo_clone_url_template = args.repo_clone_url_template
        # If the argument is None, assume it's not being updated,
        # as this is the default value for the arg parser.
        if repo_clone_url_template is not None:
            if repo_clone_url_template == '':
                # If repository URL is empty, set to None
                repo_clone_url_template = None

            try:
                module_provider.update_repo_clone_url_template(repo_clone_url_template=repo_clone_url_template)
            except RepositoryUrlParseError as exc:
                return {'message': 'Repo clone URL: {}'.format(str(exc))}, 400

        # Ensure repository URL is parsable
        repo_browse_url_template = args.repo_browse_url_template
        if repo_browse_url_template is not None:
            if repo_browse_url_template == '':
                # If repository URL is empty, set to None
                repo_browse_url_template = None

            try:
                module_provider.update_repo_browse_url_template(repo_browse_url_template=repo_browse_url_template)
            except RepositoryUrlParseError as exc:
                return {'message': 'Repo browse URL: {}'.format(str(exc))}, 400

        git_tag_format = args.git_tag_format
        if git_tag_format is not None:
            module_provider.update_git_tag_format(git_tag_format)

        # Update git path
        git_path = args.git_path
        if git_path is not None:
            module_provider.update_git_path(git_path=git_path)

        if args.verified is not None:
            module_provider.update_attributes(verified=args.verified)

        return {}


class ApiTerraregModuleVersionPublish(ErrorCatchingResource):
    """Provide interface to publish module version."""

    method_decorators = [require_api_authentication(terrareg.config.Config().PUBLISH_API_KEYS)]

    def _post(self, namespace, name, provider, version):
        """Publish module."""
        namespace = Namespace(name=namespace)
        module = Module(namespace=namespace, name=name)
        module_provider = ModuleProvider.get(module=module, name=provider)

        if not module_provider:
            return {'message': 'Module provider does not exist'}, 400

        module_version = ModuleVersion.get(module_provider=module_provider, version=version)
        if not module_version:
            return {'message': 'Module version does not exist'}, 400

        module_version.publish()
        return {
            'status': 'Success'
        }


class ApiTerraregModuleVerisonSubmodules(ErrorCatchingResource):
    """Interface to obtain list of submodules in module version."""

    def _get(self, namespace, name, provider, version):
        """Return list of submodules."""
        namespace = Namespace(name=namespace)
        module = Module(namespace=namespace, name=name)
        module_provider = ModuleProvider.get(module=module, name=provider)

        if not module_provider:
            return {'message': 'Module provider does not exist'}, 400

        module_version = ModuleVersion.get(module_provider=module_provider, version=version)
        if not module_version:
            return {'message': 'Module version does not exist'}, 400

        return [
            {
                'path': submodule.path,
                'href': submodule.get_view_url()
            }
            for submodule in module_version.get_submodules()
        ]


class ApiTerraregSubmoduleDetails(ErrorCatchingResource):
    """Interface to obtain submodule details."""

    def _get(self, namespace, name, provider, version, submodule):
        """Return details of submodule."""
        namespace_obj = Namespace(name=namespace)
        module_obj = Module(namespace=namespace_obj, name=name)
        module_provider = ModuleProvider.get(module=module_obj, name=provider)

        if not module_provider:
            return {'message': 'Module provider does not exist'}, 400

        module_version = ModuleVersion.get(module_provider=module_provider, version=version)
        if not module_version:
            return {'message': 'Module version does not exist'}, 400

        submodule_obj = Submodule.get(module_version=module_version, module_path=submodule)

        return submodule_obj.get_terrareg_api_details()


class ApiTerraregSubmoduleReadmeHtml(ErrorCatchingResource):
    """Interface to obtain submodule REAMDE in HTML format."""

    def _get(self, namespace, name, provider, version, submodule):
        """Return HTML formatted README of submodule."""
        namespace_obj = Namespace(name=namespace)
        module_obj = Module(namespace=namespace_obj, name=name)
        module_provider = ModuleProvider.get(module=module_obj, name=provider)

        if not module_provider:
            return {'message': 'Module provider does not exist'}, 400

        module_version = ModuleVersion.get(module_provider=module_provider, version=version)
        if not module_version:
            return {'message': 'Module version does not exist'}, 400

        submodule_obj = Submodule.get(module_version=module_version, module_path=submodule)

        return submodule_obj.get_readme_html(server_hostname=request.host)


class ApiTerraregModuleVersionExamples(ErrorCatchingResource):
    """Interface to obtain list of examples in module version."""

    def _get(self, namespace, name, provider, version):
        """Return list of examples."""
        namespace = Namespace(name=namespace)
        module = Module(namespace=namespace, name=name)
        module_provider = ModuleProvider.get(module=module, name=provider)

        if not module_provider:
            return {'message': 'Module provider does not exist'}, 400

        module_version = ModuleVersion.get(module_provider=module_provider, version=version)
        if not module_version:
            return {'message': 'Module version does not exist'}, 400

        return [
            {
                'path': example.path,
                'href': example.get_view_url()
            }
            for example in module_version.get_examples()
        ]


class ApiTerraregExampleDetails(ErrorCatchingResource):
    """Interface to obtain example details."""

    def _get(self, namespace, name, provider, version, example):
        """Return details of example."""
        namespace_obj = Namespace(name=namespace)
        module_obj = Module(namespace=namespace_obj, name=name)
        module_provider = ModuleProvider.get(module=module_obj, name=provider)

        if not module_provider:
            return {'message': 'Module provider does not exist'}, 400

        module_version = ModuleVersion.get(module_provider=module_provider, version=version)
        if not module_version:
            return {'message': 'Module version does not exist'}, 400

        example_obj = Example.get(module_version=module_version, module_path=example)

        return example_obj.get_terrareg_api_details()


class ApiTerraregExampleReadmeHtml(ErrorCatchingResource):
    """Interface to obtain example REAMDE in HTML format."""

    def _get(self, namespace, name, provider, version, example):
        """Return HTML formatted README of example."""
        namespace_obj = Namespace(name=namespace)
        module_obj = Module(namespace=namespace_obj, name=name)
        module_provider = ModuleProvider.get(module=module_obj, name=provider)

        if not module_provider:
            return {'message': 'Module provider does not exist'}, 400

        module_version = ModuleVersion.get(module_provider=module_provider, version=version)
        if not module_version:
            return {'message': 'Module version does not exist'}, 400

        example_obj = Example.get(module_version=module_version, module_path=example)

        return example_obj.get_readme_html(server_hostname=request.host)


class ApiTerraregExampleFileList(ErrorCatchingResource):
    """Interface to obtain list of example files."""

    def _get(self, namespace, name, provider, version, example):
        """Return list of files available in example."""
        namespace_obj = Namespace(name=namespace)
        module_obj = Module(namespace=namespace_obj, name=name)
        module_provider = ModuleProvider.get(module=module_obj, name=provider)

        if not module_provider:
            return {'message': 'Module provider does not exist'}, 400

        module_version = ModuleVersion.get(module_provider=module_provider, version=version)
        if not module_version:
            return {'message': 'Module version does not exist'}, 400

        example_obj = Example(module_version=module_version, module_path=example)

        return [
            {
                'filename': example_file.file_name,
                'path': example_file.path,
                'content_href': '/v1/terrareg/modules/{module_version_id}/example/files/{file_path}'.format(
                    module_version_id=module_version.id,
                    file_path=example_file.path)
            }
            for example_file in sorted(example_obj.get_files())
        ]


class ApiTerraregExampleFile(ErrorCatchingResource):
    """Interface to obtain content of example file."""

    def _get(self, namespace, name, provider, version, example_file):
        """Return conent of example file in example module."""
        namespace_obj = Namespace(name=namespace)
        module_obj = Module(namespace=namespace_obj, name=name)
        module_provider = ModuleProvider.get(module=module_obj, name=provider)

        if not module_provider:
            return {'message': 'Module provider does not exist'}, 400

        module_version = ModuleVersion.get(module_provider=module_provider, version=version)
        if not module_version:
            return {'message': 'Module version does not exist'}, 400

        example_file_obj = ExampleFile.get_by_path(module_version=module_version, file_path=example_file)

        if example_file_obj is None:
            return {'message': 'Example file object does not exist.'}

        return example_file_obj.get_content(server_hostname=request.host)

