

import os

from flask import Flask, request, render_template, redirect, make_response, send_from_directory
from flask_restful import Resource, Api, reqparse

from terrareg.config import DATA_DIRECTORY
from terrareg.database import Database
from terrareg.models import Namespace, Module, ModuleProvider, ModuleVersion
from terrareg.module_search import ModuleSearch
from terrareg.module_extractor import ModuleExtractor
from terrareg.analytics import AnalyticsEngine


class Server(object):
    """Manage web server and route requests"""

    ALLOWED_EXTENSIONS = {'zip'}

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
        self.debug = True
        self.ssl_public_key = ssl_public_key
        self.ssl_private_key = ssl_private_key

        if not os.path.isdir(DATA_DIRECTORY):
            os.mkdir(DATA_DIRECTORY)
        if not os.path.isdir(self._get_upload_directory()):
            os.mkdir(self._get_upload_directory())
        if not os.path.isdir(os.path.join(DATA_DIRECTORY, 'modules')):
            os.mkdir(os.path.join(DATA_DIRECTORY, 'modules'))

        self._app.config['UPLOAD_FOLDER'] = self._get_upload_directory()

        # Initialise database
        Database.get().initialise()

        self._register_routes()

    def _get_upload_directory(self):
        return os.path.join(DATA_DIRECTORY, 'upload')

    def _register_routes(self):
        """Register routes with flask."""

        # Upload module
        self._app.route(
            '/v1/<string:namespace>/<string:name>/<string:provider>/<string:version>/upload',
            methods=['POST']
        )(self._upload_module_version)

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

        # Views
        self._app.route('/')(self._view_serve_static_index)
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

    def run(self):
        """Run flask server."""
        kwargs = {
            'host': self.host,
            'port': self.port,
            'debug': self.debug
        }
        if self.ssl_public_key and self.ssl_private_key:
            kwargs['ssl_context'] = (self.ssl_public_key, self.ssl_private_key)

        self._app.run(**kwargs)

    def allowed_file(self, filename):
        return '.' in filename and \
               filename.rsplit('.', 1)[1].lower() in self.ALLOWED_EXTENSIONS


    def _upload_module_version(self, namespace, name, provider, version):
        """Handle module version upload."""

        namespace = Namespace(namespace)
        module = Module(namespace=namespace, name=name)
        module_provider = ModuleProvider(module=module, name=provider)
        module_version = ModuleVersion(module_provider=module_provider, version=version)

        if len(request.files) != 1:
            return 'One file can be uploaded'

        file = request.files[[f for f in request.files.keys()][0]]

        # If the user does not select a file, the browser submits an
        # empty file without a filename.
        if file.filename == '':
            return 'No selected file'
        if file and self.allowed_file(file.filename):
            module_version.prepare_module()
            with ModuleExtractor(upload_file=file, module_version=module_version) as me:
                me.process_upload()
            return 'Upload sucessful'

        return 'Error occurred - unknown file extension'

    def _view_serve_static_index(self):
        """Serve static index"""
        return render_template('index.html')

    def _view_serve_namespace_list(self):
        """Render view for display module."""
        namespaces = Namespace.get_all()

        # If only one provider for module, redirect to it.
        if len(namespaces) == 1:
            return redirect(namespaces[0].get_view_url())
        else:
            return render_template(
                'namespace_list.html',
                namespaces=namespaces
            )

    def _view_serve_namespace(self, namespace):
        """Render view for namespace."""
        namespace = Namespace(namespace)
        modules = namespace.get_all_modules()

        return render_template(
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
            return render_template(
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
            module_version = module_provider.get_latest_version()
        else:
            module_version = ModuleVersion(module_provider=module_provider, version=version)

        return render_template(
            'module_provider.html',
            namespace=namespace,
            module=module,
            module_provider=module_provider,
            module_version=module_version
        )

    def _view_serve_module_search(self):
        """Search modules based on input."""
        return render_template('module_search.html')


class ApiTerraformWellKnown(Resource):
    """Terraform .well-known discovery"""

    def get(self):
        """Return wellknown JSON"""
        return {
            "modules.v1": "/v1/modules/"
        }


class ApiModuleList(Resource):
    def get(self):
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
            'verified', type=bool,
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


class ApiModuleSearch(Resource):

    def get(self):
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
            'verified', type=bool,
            default=False, help='Limits modules to only verified modules.'
        )

        args = parser.parse_args()

        # Limit the limits
        limit = 50 if args.limit > 50 else args.limit
        limit = 1 if limit < 1 else limit
        current_offset = 0 if args.offset < 0 else args.offset

        module_providers = ModuleSearch.search_module_providers(
            query=args.q,
            namespace=args.namespace,
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

class ApiModuleDetails(Resource):
    def get(self, namespace, name):
        """Return latest version for each module provider."""

        namespace, _ = Namespace.extract_analytics_token(namespace)
        namespace = Namespace(namespace)
        module = Module(namespace=namespace, name=name)
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

class ApiModuleProviderDetails(Resource):

    def get(self, namespace, name, provider):
        """Return list of version."""

        namespace, _ = Namespace.extract_analytics_token(namespace)
        namespace = Namespace(namespace)
        module = Module(namespace=namespace, name=name)
        module_provider = ModuleProvider(module=module, name=provider)
        module_version = module_provider.get_latest_version()
        return module_version.get_api_details()


class ApiModuleVersionDetails(Resource):

    def get(self, namespace, name, provider, version):
        """Return list of version."""

        namespace, _ = Namespace.extract_analytics_token(namespace)
        namespace = Namespace(namespace)
        module = Module(namespace=namespace, name=name)
        module_provider = ModuleProvider(module=module, name=provider)
        module_version = ModuleVersion(module_provider=module_provider, version=version)
        return module_version.get_api_details()


class ApiModuleVersions(Resource):

    def get(self, namespace, name, provider):
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

class ApiModuleVersionDownload(Resource):
    """Provide download endpoint."""

    def get(self, namespace, name, provider, version):
        """Provide download header for location to download source."""
        namespace, analytics_token = Namespace.extract_analytics_token(namespace)
        namespace = Namespace(namespace)
        module = Module(namespace=namespace, name=name)
        module_provider = ModuleProvider(module=module, name=provider)
        module_version = ModuleVersion(module_provider=module_provider, version=version)

        # Record download
        AnalyticsEngine.record_module_version_download(
            module_version=module_version,
            analytics_token=analytics_token,
            terraform_version=request.headers.get('X-Terraform-Version', None),
            user_agent=request.headers.get('User-Agent', None)
        )

        resp = make_response('', 204)
        print(request.headers)
        resp.headers['X-Terraform-Get'] = module_version.get_source_download_url()
        return resp


class ApiModuleVersionSourceDownload(Resource):
    """Return source package of module version"""

    def get(self, namespace, name, provider, version):
        """Return static file."""
        namespace = Namespace(namespace)
        module = Module(namespace=namespace, name=name)
        module_provider = ModuleProvider(module=module, name=provider)
        module_version = ModuleVersion(module_provider=module_provider, version=version)
        return send_from_directory(module_version.base_directory, module_version.archive_name_zip)
