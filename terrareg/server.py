

import os
import re
import datetime
from functools import wraps

from flask import Flask, request, render_template, redirect, make_response, send_from_directory, session
from flask_restful import Resource, Api, reqparse, inputs, abort

import terrareg.config
from terrareg.database import Database
from terrareg.errors import TerraregError, UploadError, NoModuleVersionAvailableError
from terrareg.models import Namespace, Module, ModuleProvider, ModuleVersion
from terrareg.module_search import ModuleSearch
from terrareg.module_extractor import ApiUploadModuleExtractor
from terrareg.analytics import AnalyticsEngine
from terrareg.filters import NamespaceTrustFilter
from terrareg.config import APPLICATION_NAME, LOGO_URL


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
        self.port = 5000
        self.ssl_public_key = ssl_public_key
        self.ssl_private_key = ssl_private_key

        if not os.path.isdir(terrareg.config.DATA_DIRECTORY):
            os.mkdir(terrareg.config.DATA_DIRECTORY)
        if not os.path.isdir(self._get_upload_directory()):
            os.mkdir(self._get_upload_directory())
        if not os.path.isdir(os.path.join(terrareg.config.DATA_DIRECTORY, 'modules')):
            os.mkdir(os.path.join(terrareg.config.DATA_DIRECTORY, 'modules'))

        self._app.config['UPLOAD_FOLDER'] = self._get_upload_directory()

        # Initialise database
        Database.get().initialise()

        self._register_routes()

    def _get_upload_directory(self):
        return os.path.join(terrareg.config.DATA_DIRECTORY, 'upload')

    def _register_routes(self):
        """Register routes with flask."""

        # Upload module
        self._api.add_resource(
            ApiUploadModule,
            '/v1/<string:namespace>/<string:name>/<string:provider>/<string:version>/upload'
        )

        # Download module tar
        self._api.add_resource(
            ApiModuleVersionSourceDownload,
            '/static/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/source.zip'
        )

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

        # Terrareg APIs
        self._api.add_resource(
            ApiTerraregGlobalStatsSummary,
            '/v1/terrareg/analytics/global/stats_summary'
        )
        self._api.add_resource(
            ApiTerraregMostRecentlyPublishedModuleVersion,
            '/v1/terrareg/analytics/global/most_recently_published_module_version'
        )
        self._api.add_resource(
            ApiTerraregModuleProviderAnalyticsTokenVersions,
            '/v1/terrareg/analytics/<string:namespace>/<string:name>/<string:provider>/token_versions'
        )
        self._api.add_resource(
            ApiTerraregMostDownloadedModuleProviderThisWeek,
            '/v1/terrareg/analytics/global/most_downloaded_module_provider_this_week'
        )
        self._api.add_resource(
            ApiTerraregModuleVersionVariableTemplate,
            '/v1/terrareg/<string:namespace>/<string:name>/<string:provider>/<string:version>/variable_template'
        )
        self._api.add_resource(
            ApiTerraregModuleSearchFilters,
            '/v1/terrareg/search_filters'
        )
        self._api.add_resource(
            ApiTerraregAdminAuthenticate,
            '/v1/terrareg/auth/admin/login'
        )
        self._api.add_resource(
            ApiTerraregIsAuthenticated,
            '/v1/terrareg/auth/admin/is_authenticated'
        )

    def _render_template(self, *args, **kwargs):
        """Override render_template, passing in base variables."""
        return render_template(
            *args, **kwargs,
            terrareg_application_name=APPLICATION_NAME,
            terrareg_logo_url=LOGO_URL
        )


    def run(self):
        """Run flask server."""
        kwargs = {
            'host': self.host,
            'port': self.port,
            'debug': terrareg.config.DEBUG
        }
        if self.ssl_public_key and self.ssl_private_key:
            kwargs['ssl_context'] = (self.ssl_public_key, self.ssl_private_key)

        self._app.secret_key = terrareg.config.SECRET_KEY

        self._app.run(**kwargs)

    def _view_serve_static_index(self):
        """Serve static index"""
        return self._render_template('index.html')

    def _view_serve_login(self):
        """Serve static login page."""
        return self._render_template('login.html')

    def _logout(self):
        """Remove cookie and redirect."""
        session['is_admin_authenticated'] = False
        return redirect('/')

    def _view_serve_namespace_list(self):
        """Render view for display module."""
        namespaces = Namespace.get_all()

        # If only one provider for module, redirect to it.
        if len(namespaces) == 1:
            return redirect(namespaces[0].get_view_url())
        else:
            return self._render_template(
                'namespace_list.html',
                namespaces=namespaces
            )

    def _view_serve_namespace(self, namespace):
        """Render view for namespace."""
        namespace = Namespace(namespace)
        modules = namespace.get_all_modules()

        return self._render_template(
            'namespace.html',
            namespace=namespace,
            modules=modules
        )

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

    def _view_serve_module_provider(self, namespace, name, provider, version=None):
        """Render view for displaying module provider information"""
        namespace = Namespace(namespace)
        module = Module(namespace=namespace, name=name)
        module_provider = ModuleProvider(module=module, name=provider)
        if version is None:
            try:
                module_version = module_provider.get_latest_version()
            except NoModuleVersionAvailableError:
                # If no version was provided and no versions were found for the module provider,
                # redirect to the module
                return redirect(module.get_view_url())

        else:
            module_version = ModuleVersion.get(module_provider=module_provider, version=version)

            if module_version is None:
                # If a version number was provided and it does not exist,
                # redirect to the module provider
                return redirect(module_provider.get_view_url())

        return self._render_template(
            'module_provider.html',
            namespace=namespace,
            module=module,
            module_provider=module_provider,
            module_version=module_version,
            server_hostname=request.host
        )

    def _view_serve_module_search(self):
        """Search modules based on input."""
        return self._render_template('module_search.html')


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
        raise NotImplementedError

    def get(self, *args, **kwargs):
        """Run subclasses get in error handling fashion."""
        try:
            return self._get(*args, **kwargs)
        except TerraregError as exc:
            return {
                "status": "Error",
                "message": str(exc)
            }

    def _post(self, *args, **kwargs):
        """Placeholder for overridable post method."""
        raise NotImplementedError

    def post(self, *args, **kwargs):
        """Run subclasses post in error handling fashion."""
        try:
            return self._post(*args, **kwargs)
        except TerraregError as exc:
            return {
                "status": "Error",
                "message": str(exc)
            }

    def _get_404_response(self):
        """Return common 404 error"""
        return {'errors': ['Not Found']}, 404


class ApiUploadModule(ErrorCatchingResource):

    ALLOWED_EXTENSIONS = ['zip']

    def allowed_file(self, filename):
        """Check if file has allowed file-extension"""
        return '.' in filename and \
               filename.rsplit('.', 1)[1].lower() in self.ALLOWED_EXTENSIONS

    def _post(self, namespace, name, provider, version):
        """Handle module version upload."""

        namespace = Namespace(namespace)
        module = Module(namespace=namespace, name=name)
        module_provider = ModuleProvider(module=module, name=provider)
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
            default=None, help='Limits modules to a specific provider.'
        )
        parser.add_argument(
            'verified', type=inputs.boolean,
            default=False, help='Limits modules to only verified modules.'
        )

        args = parser.parse_args()

        # Limit the limits
        limit = 50 if args.limit > 50 else args.limit
        limit = 1 if limit < 1 else limit
        current_offset = 0 if args.offset < 0 else args.offset

        module_providers = ModuleSearch.search_module_providers(
            provider=args.provider,
            verified=args.verified,
            offset=current_offset,
            limit=limit
        )

        return {
            "meta": {
                "limit": limit,
                "current_offset": current_offset,
                "next_offset": (current_offset + limit),
                "prev_offset": (current_offset - limit) if (current_offset >= limit) else 0
            },
            "modules": [
                module_provider.get_latest_version().get_api_outline()
                for module_provider in module_providers
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
            default=None, help='Limits modules to a specific provider.'
        )
        parser.add_argument(
            'namespace', type=str,
            default=None, help='Limits modules to a specific namespace.'
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

        # Limit the limits
        limit = 50 if args.limit > 50 else args.limit
        limit = 1 if limit < 1 else limit
        current_offset = 0 if args.offset < 0 else args.offset

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

        module_providers = ModuleSearch.search_module_providers(
            query=args.q,
            namespace=args.namespace,
            provider=args.provider,
            verified=args.verified,
            namespace_trust_filters=namespace_trust_filters,
            offset=current_offset,
            limit=limit
        )

        return {
            "meta": {
                "limit": limit,
                "current_offset": current_offset,
                "next_offset": (current_offset + limit),
                "prev_offset": (current_offset - limit) if (current_offset >= limit) else 0
            },
            "modules": [
                module_provider.get_latest_version().get_api_outline()
                for module_provider in module_providers
            ]
        }


class ApiModuleDetails(ErrorCatchingResource):
    def _get(self, namespace, name):
        """Return latest version for each module provider."""

        namespace, _ = Namespace.extract_analytics_token(namespace)
        namespace = Namespace(namespace)
        module = Module(namespace=namespace, name=name)
        modules = [
            module_provider.get_latest_version().get_api_outline()
            for module_provider in module.get_providers()
        ]

        if not modules:
            return self._get_404_response()

        return {
            "meta": {
                "limit": 5,
                "offset": 0
            },
            "modules": [
                module_provider.get_latest_version().get_api_outline()
                for module_provider in module.get_providers()
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
        print(request.headers)
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
        if not analytics_token and not terrareg.config.ALLOW_UNIDENTIFIED_DOWNLOADS:
            return make_response(
                ("\nAn {analytics_token_phrase} must be provided.\n"
                 "Please update module source to include {analytics_token_phrase}.\n"
                 "\nFor example:\n  source = \"{host}/{example_analytics_token}__{namespace}/{module_name}/{provider}\"").format(
                    analytics_token_phrase=terrareg.config.ANALYTICS_TOKEN_PHRASE,
                    host=request.host,
                    example_analytics_token=terrareg.config.EXAMPLE_ANALYTICS_TOKEN,
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
        print(dict(request.headers))

        resp = make_response('', 204)
        resp.headers['X-Terraform-Get'] = module_version.get_source_download_url()
        return resp


class ApiModuleVersionSourceDownload(ErrorCatchingResource):
    """Return source package of module version"""

    def _get(self, namespace, name, provider, version):
        """Return static file."""
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
        module_provider = ModuleProvider(module=module, name=provider)
        return {
            "data": {
                "type": "module-downloads-summary",
                "id": module_provider.id,
                "attributes": AnalyticsEngine.get_module_provider_download_stats(module_provider)
            }
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


class ApiTerraregModuleVersionVariableTemplate(ErrorCatchingResource):
    """Provide variable template for module version."""

    def _get(self, namespace, name, provider, version):
        """Return variable template."""
        namespace = Namespace(namespace)
        module = Module(namespace=namespace, name=name)
        module_provider = ModuleProvider(module=module, name=provider)
        module_version = ModuleVersion(module_provider=module_provider, version=version)
        return module_version.variable_template


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


def check_admin_authentication():
    """Check authorization header is present or authenticated session"""
    authenticated = False
    # Check that:
    # - An admin authentication token has been setup
    # - A token has neeif valid authorisation header has been passed
    if (terrareg.config.ADMIN_AUTHENTICATION_TOKEN and
            request.headers.get('X-Terrareg-ApiKey', '') ==
            terrareg.config.ADMIN_AUTHENTICATION_TOKEN):
        authenticated = True

    # Check if authenticated via session
    # - Ensure session key has been setup
    if (terrareg.config.SECRET_KEY and
            session.get('is_admin_authenticated', False) and
            'expires' in session and
            session.get('expires').timestamp() > datetime.datetime.now().timestamp()):
        authenticated = True
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

        if not terrareg.config.SECRET_KEY:
            return {'error': 'Sessions not enabled in configuration'}, 403

        session['is_admin_authenticated'] = True
        session['expires'] = (
            datetime.datetime.now() +
            datetime.timedelta(minutes=terrareg.config.ADMIN_SESSION_EXPIRY_MINS)
        )
        return {'authenticated': True}
