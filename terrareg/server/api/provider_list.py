from terrareg.server.error_catching_resource import ErrorCatchingResource


class ApiProviderList(ErrorCatchingResource):
    def _get(self, namespace, provider_type):

        return {
            "id": f"{namespace}/{provider_type}",
            "versions": [
                {
                    "version": "0.8.0",
                    "protocols": [
                        "5.0"
                    ],
                    "platforms": [
                        {
                            "os": "darwin",
                            "arch": "arm64"
                        },
                    ]
                }
            ]
        }
