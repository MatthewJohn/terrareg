
from flask_restful import reqparse

from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.auth_wrapper
import terrareg.provider_version_model
import terrareg.provider_version_documentation_model
import terrareg.provider_documentation_type


class ApiV2ProviderDocs(ErrorCatchingResource):
    """Interface for querying provider docs"""

    method_decorators = [terrareg.auth_wrapper.auth_wrapper('can_access_read_api')]

    def _get_arg_parser(self):
        """Return argument parser for querying docs"""
        parser = reqparse.RequestParser()
        parser.add_argument(
            'filter[provider-version]',
            type=int,
            location='args',
            required=True,
            dest='provider_version_id',
            help='Provider version ID to query documenation from'
        )
        parser.add_argument(
            'filter[category]',
            type=str,
            location='args',
            required=True,
            dest='category',
            help='Provider documentation category'
        )
        parser.add_argument(
            'filter[slug]',
            type=str,
            location='args',
            required=True,
            dest='slug',
            help='Slug of documentation to query for'
        )
        parser.add_argument(
            'filter[language]',
            type=str,
            location='args',
            required=True,
            dest='language',
            help='Documentation language to filter results'
        )
        parser.add_argument(
            'page[size]',
            type=int,
            location='args',
            required=True,
            dest='page_size',
            help='Result page size'
        )
        return parser

    def _get(self):
        """
        Query provider version documentation.
        
        This API is very static and requires all arguments to be passed.
        Page size, is effectively unused, as the query filters will result in 0 or 1 result.
        """
        args = self._get_arg_parser().parse_args()

        # Obtain provider version from ID
        provider_version = terrareg.provider_version_model.ProviderVersion.get_by_pk(args.provider_version_id)
        if not provider_version:
            return self._get_v2_error("Invalid Provider Version ID")

        try:
            category = terrareg.provider_documentation_type.ProviderDocumentationType(args.category)
        except ValueError:
            return self._get_v2_error("Invalid category type")

        documents = terrareg.provider_version_documentation_model.ProviderVersionDocumentation.search(
            provider_version=provider_version,
            category=category,
            language=args.language,
            slug=args.slug
        )

        return {
            "data": [
                document.get_v2_api_outline()
                for document in documents
            ]
        }
