
import os
from functools import wraps

from flask import Flask, session, redirect
from flask_restful import Api

import terrareg.config
from terrareg.database import Database
from terrareg.models import (
    Namespace, Module, GitProvider, Session,
    ModuleVersion, ModuleProvider, Submodule,
    Example
)
from terrareg.errors import (
    InvalidNamespaceNameError,
    InvalidModuleNameError,
    InvalidModuleProviderNameError,
    InvalidVersionError
)
from .base_handler import BaseHandler
from terrareg.server.api import *


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
                namespace = Namespace.get(name=kwargs['namespace'])
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
                namespace = Namespace.get(name=kwargs['namespace'])
                if namespace is not None and 'name' in kwargs:
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
                namespace = Namespace.get(name=kwargs['namespace'])
                if namespace is not None and 'name' in kwargs:
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


class Server(BaseHandler):
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

        # Prometheus metrics
        self._api.add_resource(
            PrometheusMetrics,
            '/metrics'
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
            '/create-namespace',
        )(self._view_serve_create_namespace)
        self._app.route(
            '/create-module'
        )(self._view_serve_create_module)
        self._app.route(
            '/initial-setup'
        )(self._view_serve_initial_setup)
        self._app.route(
            '/user-groups'
        )(self._view_serve_user_groups)
        self._app.route(
            '/audit-history'
        )(self._view_serve_audit_history)
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


        # OpenID connect endpoints
        self._api.add_resource(
            ApiOpenIdInitiate,
            '/openid/login'
        )
        self._api.add_resource(
            ApiOpenIdCallback,
            '/openid/callback'
        )

        # Saml2 endpoints
        self._api.add_resource(
            ApiSamlInitiate,
            '/saml/login'
        )
        self._api.add_resource(
            ApiSamlMetadata,
            '/saml/metadata'
        )

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
            '/v1/terrareg/analytics/global/usage_stats'
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
            ApiTerraregModuleProviderVersions,
            '/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/versions'
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
            ApiModuleVersionCreateGitHubHook,
            '/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/hooks/github'
        )
        self._api.add_resource(
            ApiTerraregModuleVersionVariableTemplate,
            '/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/variable_template'
        )
        self._api.add_resource(
            ApiTerraregModuleVersionFile,
            '/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/files/<string:path>'
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
        self._api.add_resource(
            ApiTerraregAuditHistory,
            '/v1/terrareg/audit-history'
        )
        self._api.add_resource(
            ApiTerraregAuthUserGroups,
            '/v1/terrareg/user-groups'
        )
        self._api.add_resource(
            ApiTerraregAuthUserGroup,
            '/v1/terrareg/user-groups/<string:user_group>'
        )
        self._api.add_resource(
            ApiTerraregAuthUserGroupNamespacePermissions,
            '/v1/terrareg/user-groups/<string:user_group>/permissions/<string:namespace>'
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

    def run(self, debug=None):
        """Run flask server."""
        kwargs = {
            'host': self.host,
            'port': self.port,
            'debug': terrareg.config.Config().DEBUG if debug is None else debug,
            'threaded': terrareg.config.Config().THREADED
        }
        if self.ssl_public_key and self.ssl_private_key:
            kwargs['ssl_context'] = (self.ssl_public_key, self.ssl_private_key)

        self._app.secret_key = terrareg.config.Config().SECRET_KEY

        self._app.run(**kwargs)

    def _namespace_404(self, namespace_name: str):
        """Return 404 page for non-existent namespace"""
        return self._render_template(
            'error.html',
            error_title='Namespace does not exist',
            error_description='The namespace {namespace} does not exist'.format(
                namespace=namespace_name
            )
        ), 404

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
        session['openid_connect_state'] = None
        session['openid_connect_access_token'] = None
        session['openid_connect_id_token'] = None
        session['openid_connect_expires_at'] = None
        session.modified = True
        return redirect('/')

    def _view_serve_create_module(self):
        """Provide view to create module provider."""
        return self._render_template(
            'create_module_provider.html',
            git_providers=GitProvider.get_all(),
            ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER=terrareg.config.Config().ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER,
            ALLOW_CUSTOM_GIT_URL_MODULE_VERSION=terrareg.config.Config().ALLOW_CUSTOM_GIT_URL_MODULE_VERSION
        )

    def _view_serve_create_namespace(self):
        """Provide view to create namespace."""
        return self._render_template(
            'create_namespace.html'
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
        namespace_obj = Namespace.get(namespace)
        if namespace_obj is None:
            return self._namespace_404(
                namespace_name=namespace
            )
        module = Module(namespace=namespace_obj, name=name)
        module_providers = module.get_providers()

        # If only one provider for module, redirect to it.
        if len(module_providers) == 1:
            return redirect(module_providers[0].get_view_url())
        else:
            return self._render_template(
                'module.html',
                namespace=namespace_obj,
                module=module,
                module_providers=module_providers
            )

    @catch_name_exceptions
    def _view_serve_module_provider(self, namespace, name, provider, version=None):
        """Render view for displaying module provider information"""
        namespace_obj = Namespace.get(namespace)
        if namespace_obj is None:
            return self._namespace_404(
                namespace_name=namespace
            )

        module = Module(namespace=namespace_obj, name=name)
        module_provider = ModuleProvider.get(module=module, name=provider)
        if module_provider is None:
            return self._module_provider_404(
                namespace=namespace_obj,
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
        namespace_obj = Namespace.get(namespace)
        if namespace_obj is None:
            return self._namespace_404(namespace_name=namespace)

        module = Module(namespace=namespace_obj, name=name)
        module_provider = ModuleProvider.get(module=module, name=provider)
        if module_provider is None:
            return self._module_provider_404(
                namespace=namespace_obj,
                module=name,
                module_provider_name=provider)

        module_version = ModuleVersion.get(module_provider=module_provider, version=version)

        if module_version is None:
            return redirect(module_provider.get_view_url())

        submodule = Submodule(module_version=module_version, module_path=submodule_path)

        return self._render_template('module_provider.html')

    @catch_name_exceptions
    def _view_serve_example(self, namespace, name, provider, version, submodule_path):
        """Review view for displaying example"""
        namespace_obj = Namespace.get(namespace)
        if namespace_obj is None:
            return self._namespace_404(namespace_name=namespace)

        module = Module(namespace=namespace_obj, name=name)
        module_provider = ModuleProvider.get(module=module, name=provider)
        if module_provider is None:
            return self._module_provider_404(
                namespace=namespace_obj,
                module=name,
                module_provider_name=provider)

        module_version = ModuleVersion.get(module_provider=module_provider, version=version)

        if module_version is None:
            return redirect(module_provider.get_view_url())

        submodule = Example(module_version=module_version, module_path=submodule_path)

        return self._render_template('module_provider.html')

    def _view_serve_module_search(self):
        """Search modules based on input."""
        return self._render_template('module_search.html')

    def _view_serve_user_groups(self):
        """Page to view/modify user groups and permissions."""
        if not terrareg.auth.AuthFactory().get_current_auth_method().is_admin():
            return self._render_template(
                'error.html',
                root_bread_brumb='User groups',
                error_title='Permission denied',
                error_description="You are not logged in or do not have permssion to view this page"
            ), 403
        return self._render_template('user_groups.html')

    def _view_serve_audit_history(self):
        """Page to view/modify user groups and permissions."""
        if not terrareg.auth.AuthFactory().get_current_auth_method().is_admin():
            return self._render_template(
                'error.html',
                root_bread_brumb='Audit History',
                error_title='Permission denied',
                error_description="You are not logged in or do not have permssion to view this page"
            ), 403
        return self._render_template('audit_history.html')
