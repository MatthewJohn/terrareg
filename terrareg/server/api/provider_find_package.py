from terrareg.server.error_catching_resource import ErrorCatchingResource


class ApiProviderFindPackage(ErrorCatchingResource):
    def _get(self, namespace, provider_type, version, os, arch):

        return {
            "protocols": ["4.0", "5.1"],
            "os": f"{os}",
            "arch": f"{arch}",
            "filename": "terraform-provider-random_2.0.0_linux_amd64.zip",
            "download_url": "https://releases.hashicorp.com/terraform-provider-random/2.0.0/terraform-provider-random_2.0.0_linux_amd64.zip",
            "shasums_url": "https://releases.hashicorp.com/terraform-provider-random/2.0.0/terraform-provider-random_2.0.0_SHA256SUMS",
            "shasums_signature_url": "https://releases.hashicorp.com/terraform-provider-random/2.0.0/terraform-provider-random_2.0.0_SHA256SUMS.sig",
            "shasum": "5f9c7aa76b7c34d722fc9123208e26b22d60440cb47150dd04733b9b94f4541a"
        }

