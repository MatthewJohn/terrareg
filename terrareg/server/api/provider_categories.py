
from flask_restful import reqparse

from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.provider_category_model
import terrareg.auth_wrapper


class ApiProviderCategories(ErrorCatchingResource):
    """Interface to obtain list of provider categories."""

    method_decorators = [terrareg.auth_wrapper.auth_wrapper('can_access_read_api')]

    def _get(self):
        """Return list of all provider categories"""
        return {
            "data": [
                {
                    "type": "categories",
                    "id": str(provider_category.pk),
                    "attributes": {
                        "name": provider_category.name,
                        "slug": provider_category.slug,
                        "user-selectable": provider_category.user_selectable
                    },
                    "links": {
                        "self": f"/v2/categories/{provider_category.pk}"
                    }
                }
                for provider_category in terrareg.provider_category_model.ProviderCategoryFactory.get().get_all_provider_categories()
            ]
        }
