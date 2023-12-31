
from datetime import datetime
import unittest.mock

import pytest

from test.integration.terrareg import TerraregIntegrationTest
from test import client, app_context, test_request_context
import terrareg.provider_search


provider_categories = [
    {
        "id": 523,
        "name": "Visible Monitoring",
        "slug": "visible-monitoring",
        "user-selectable": True
    }
]

provider_sources = [
    {
        "name": "Test Github Autogenerate",
        "type": "github",
        "base_url": "https://github.example.com",
        "api_url": "https://api.github.example.com",
        "client_id": "unittest-client-id",
        "client_secret": "unittest-client-secret",
        "login_button_text": "Login via Github using this unit test",
        "private_key_path": "./path/to/key.pem",
        "app_id": "1234appid",
        "default_access_token": "pa-test-personal-access-token",
        "default_installation_id": "ut-default-installation-id-here",
        "auto_generate_github_organisation_namespaces": True
    }
]

test_data = {
    "test-namespace": {
        "gpg_keys": [
            {
                "ascii_armor": """
-----BEGIN PGP PUBLIC KEY BLOCK-----

mI0EZUHt7QEEAKgSXXCkqShvE54omLsE0Gzu/Es2Nelwnps8ETlcHPKag0VlZch/
0HPyF3hGsdZM7GB1il7fGCGw6Urkmci7XkRj2M09QtAvE2YPOqfNfMvHQrLIAkBV
lP/4xIBnGMmsUYVMAeo0DiDdFf3Q3pIbWDhd7+OCPKh80F/pYM1Rm4qnABEBAAG0
UVRlc3QgVGVycmFyZWcgVGVzdHMgKFRlc3QgS2V5IGZvciB0ZXJyYXJlZyBUZXN0
cykgPHRlcnJhcmVnLXRlc3RzQGNvbGFtYWlsLmNvLnVrPojOBBMBCgA4FiEEIadO
Tj/f5DhTK9WENN43SsNkDNsFAmVB7e0CGwMFCwkIBwIGFQoJCAsCBBYCAwECHgEC
F4AACgkQNN43SsNkDNtkywP/SR8U/c3gzAY4w0KF3ZG5sBJqrBfdA2d2R//Bsjvz
jRCpGdaXVBJG2FFyfl5QLLhC56rS6nsX6vcXkrRGQtYG6Bhroo6eWjVnyT1RMM+A
wD5uwCijPlSdl82q91aFQk3jwqNoe4/gr9ERHagx3MAgMTEhIzPaKpGHtL7TPM+B
nOi4jQRlQe3tAQQAxCeKNhBAv13aXeSvPI1JKW9pcg5g9Hfd4s/qj82/0hE/Kfgt
4u7RGOEe7q1WgKirtoiv/XSpwKMSlXtt9AH8lbgkveiJ3V+DqJxdzCm42Zlyvg9Z
9sqLz6XOAyMkv44U1x182KMipuuethRmSemN8jthc4Bh5iEM/l7460IyRk8AEQEA
AYi2BBgBCgAgFiEEIadOTj/f5DhTK9WENN43SsNkDNsFAmVB7e0CGwwACgkQNN43
SsNkDNtn+AP+Pm3+u+if0BExYTMKJ0/dU4ICWBkyuuMDkQlz8oOn9/w9EYvkqR/r
QypRou1K0KbLxBCz0vqAM7KLXe0rKwZZ3eWSThiwTJkFlkJsUgwMqqROteYmWm3S
MK0hMLszB/mfN0Q2DW4U0tWslehdEA+aaccwA5PVFKdkA12ImK500TY=
=EL4W
-----END PGP PUBLIC KEY BLOCK-----
""".strip()
            }
        ],
        "providers": {
            "test-one": {
                "repository": {
                    "provider_source": "Test Github Autogenerate",
                    "provider_id": "initial-providers/terraform-provider-test-one",
                    "name": "terraform-provider-test-one",
                    "description": "Empty Provider Publish",
                    "owner": "test-namespace",
                    "clone_url": "https://git.example.com/initalproviders/terraform-provider-test-one.git",
                    "logo_url": "https://git.example.com/initalproviders/terraform-provider-test-initial.png"
                },
                "category_slug": "visible-monitoring",
                "use_default_provider_source_auth": True,
                "tier": "community",
                "versions": {
                    "1.6.0": {
                        "git_tag": "v1.6.0",
                        "gpg_key_fingerprint": "21A74E4E3FDFE438532BD58434DE374AC3640CDB",
                        "published_at": datetime(year=2022, month=5, day=6, hour=12, minute=0, second=50),
                    }
                }
            },
            "empty-provider": {
                "repository": {
                    "provider_source": "Test Github Autogenerate",
                    "provider_id": "initial-providers/terraform-provider-empty-provider",
                    "name": "terraform-provider-empty-provider",
                    "description": "Empty Provider Publish",
                    "owner": "test-namespace",
                    "clone_url": "https://git.example.com/initalproviders/terraform-provider-empty-provider-publish.git",
                    "logo_url": "https://git.example.com/initalproviders/terraform-provider-test-initial.png"
                },
                "category_slug": "visible-monitoring",
                "use_default_provider_source_auth": True,
                "tier": "community",
                "versions": {
                }
            },
            "test-two": {
                "repository": {
                    "provider_source": "Test Github Autogenerate",
                    "provider_id": "initial-providers/terraform-provider-test-two",
                    "name": "terraform-provider-test-two",
                    "description": "Test Multiple Versions",
                    "owner": "test-namespace",
                    "clone_url": "https://git.example.com/initalproviders/terraform-provider-test-two.git",
                    "logo_url": "https://git.example.com/initalproviders/terraform-provider-test-initial.png"
                },
                "category_slug": "visible-monitoring",
                "use_default_provider_source_auth": True,
                "tier": "community",
                "versions": {
                    "1.0.0": {
                        "git_tag": "v1.0.0",
                        "gpg_key_fingerprint": "21A74E4E3FDFE438532BD58434DE374AC3640CDB"
                    },
                    "2.0.1": {
                        "git_tag": "v2.0.1",
                        "gpg_key_fingerprint": "21A74E4E3FDFE438532BD58434DE374AC3640CDB",
                        "published_at": datetime(year=2023, month=10, day=3, hour=20, minute=2, second=5),
                    }
                }
            },
        }
    },
    "second-namespace": {
        "gpg_keys": [
            {
                "ascii_armor": """
-----BEGIN PGP PUBLIC KEY BLOCK-----

mI0EZVD0zwEEAJtjkOHz5pFnNw80L4qtKU98+/IVvyEEvQyOreGHdB+E5E6rtVFk
buaF7FrzJzaRj+I4hL6QB8ApkdwRdc+gaZL9KsrY6RI5WyYr8jJ/pANoxFkkIwd0
5Q2U6rkxI2SlWuHYuEmtjhJ8rFGPDRnpkTuQMxgUkUxFoHWMFIiprqmnABEBAAG0
LFRlc3QgR1BHIEtleSAyIChUZXN0KSA8dGVzdGdwZzJAZXhhbXBsZS5jb20+iM4E
EwEKADgWIQSUynK3ovRgamwYIRrpSk8q1ijZJgUCZVD0zwIbAwULCQgHAgYVCgkI
CwIEFgIDAQIeAQIXgAAKCRDpSk8q1ijZJkluBACKoMBoW4QO0d6H/h+8Ucx6/eHj
h5c9R/e7IxSJwB6lKxJGc/YkmHniP742O9opwovbxso7CrzHvdoiEoqdUJApwkk6
k2F6FxWgcZGUpFQVPTFc6iueumXsFu24gHHHiCE+106zN8YW72/lORFulVwLfo2d
Gux4McQ/g3qsP2X217iNBGVQ9M8BBADDrdRUG4mZ2cGLfhEAKDQo8f5ezuAIM2Ja
61m9jjAdRkMYwhrq5+tiVmSrVoqueaxE8cbj5C5XoOomfFOMsD4GVkzHE3t/LPdw
A0iu1usXu0rjImNnlMCVaMpIQGFJrf/EtgUPqVMGSQdNHb8ezeztodPP4gqKDB+f
2O2W0j1cxwARAQABiLYEGAEKACAWIQSUynK3ovRgamwYIRrpSk8q1ijZJgUCZVD0
zwIbDAAKCRDpSk8q1ijZJvtVA/978o0EI/lPuSoUO7EuhFyHpX2xVRL9lNwGXsyk
JDGiJXTZ7vi8pwC5GNknF2eOH1ZPXeeJxJZFh3GAa/Zk0C/BuuugnFwd6j3vxKAq
g122uE3VRxyt2bk9hDQg4rI0Y6nqexCt939GLG+Y3sG4/aEGvk3fd9a/fqn2CayP
Olm9bg==
=/0jI
-----END PGP PUBLIC KEY BLOCK-----
""".strip()
            }
        ],
        "providers": {
            "test-three": {
                "repository": {
                    "provider_source": "Test Github Autogenerate",
                    "provider_id": "second-namespace/terraform-provider-test-three",
                    "name": "terraform-provider-test-three",
                    "description": "Empty Provider Publish",
                    "owner": "second-namespace",
                    "clone_url": "https://git.example.com/initalproviders/terraform-provider-test-one.git",
                    "logo_url": "https://git.example.com/initalproviders/terraform-provider-test-initial.png"
                },
                "category_slug": "visible-monitoring",
                "use_default_provider_source_auth": True,
                "tier": "community",
                "versions": {
                    "1.6.0": {
                        "git_tag": "v1.6.0",
                        "gpg_key_fingerprint": "94CA72B7A2F4606A6C18211AE94A4F2AD628D926",
                        "published_at": datetime(year=2023, month=11, day=1, hour=23, minute=2, second=5),
                    }
                }
            }
        }
    }
}


class TestApiNamespaceProviders(TerraregIntegrationTest):
    """Test ApiNamespaceProviders endpoint"""

    _PROVIDER_CATEGORIES = provider_categories
    _PROVIDER_SOURCES = provider_sources
    _TEST_DATA = test_data

    def test_endpoint(self, client):
        """Test endpoint."""
        res = client.get('/v1/providers/test-namespace')
        assert res.status_code == 200
        assert res.json == {
            'meta': {
                'current_offset': 0,
                'limit': 10
            },
            'providers': [
                {
                    'alias': None,
                    'description': 'Test Multiple Versions',
                    'downloads': 0,
                    'id': 'test-namespace/test-two/2.0.1',
                    'logo_url': 'https://git.example.com/initalproviders/terraform-provider-test-initial.png',
                    'name': 'test-two',
                    'namespace': 'test-namespace',
                    'owner': 'test-namespace',
                    'published_at': '2023-10-03T20:02:05',
                    'source': 'https://github.example.com/test-namespace/terraform-provider-test-two',
                    'tag': 'v2.0.1',
                    'tier': 'community',
                    'version': '2.0.1'
                },
                {
                    'alias': None,
                    'description': 'Empty Provider Publish',
                    'downloads': 0,
                    'id': 'test-namespace/test-one/1.6.0',
                    'logo_url': 'https://git.example.com/initalproviders/terraform-provider-test-initial.png',
                    'name': 'test-one',
                    'namespace': 'test-namespace',
                    'owner': 'test-namespace',
                    'published_at': '2022-05-06T12:00:50',
                    'source': 'https://github.example.com/test-namespace/terraform-provider-test-one',
                    'tag': 'v1.6.0',
                    'tier': 'community',
                    'version': '1.6.0'
                },
            ]
        }

    @pytest.mark.parametrize('offset, expected_call_offset, expected_applied_offset', [
        (None, 0, 0),
        (0, 0, 0),
        (21, 21, 21)
    ])
    @pytest.mark.parametrize('limit, expected_call_limit, expected_applied_limit', [
        (None, 10, 10),
        (5, 5, 5),
        (10, 10, 10),
        (150, 150, 50),
    ])
    def test_endpoint_with_params(self, client, offset, expected_call_offset, expected_applied_offset, limit, expected_call_limit, expected_applied_limit):
        """Test endpoint with parameters"""
        params = ''
        if offset is not None:
            params += f'&offset={offset}'
        if limit is not None:
            params += f'&limit={limit}'

        mock_search_providers = unittest.mock.MagicMock(side_effect=terrareg.provider_search.ProviderSearch.search_providers)

        with unittest.mock.patch('terrareg.provider_search.ProviderSearch.search_providers', mock_search_providers):
            res = client.get(f'/v1/providers/test-namespace?{params}')

        assert res.status_code == 200
        assert res.json["meta"]["current_offset"] == expected_applied_offset
        assert res.json["meta"]["limit"] == expected_applied_limit

        mock_search_providers.assert_called_once_with(
            offset=expected_call_offset,
            limit=expected_call_limit,
            namespaces=['test-namespace']
        )

    def test_endpoint_non_existent_namespace(self, client):
        """Test endpoint with non-existent namespace."""
        res = client.get('/v1/providers/does-not-exist')
        assert res.status_code == 404

    def test_unauthenticated(self, client):
        """Test unauthenticated call to API"""
        def call_endpoint():
            return client.get('/v1/providers/test-namespace')

        self._test_unauthenticated_read_api_endpoint_test(call_endpoint)
