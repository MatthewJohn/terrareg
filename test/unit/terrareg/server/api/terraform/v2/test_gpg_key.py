
import unittest.mock

from test.integration.terrareg import TerraregIntegrationTest
from test import client, app_context, test_request_context
import terrareg.user_group_namespace_permission_type
import terrareg.models


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
            },
            {
                # 54C1377D52F39966A222C4B37C0591F6E7E2D514
                "ascii_armor": """
-----BEGIN PGP PUBLIC KEY BLOCK-----

mI0EZXWEJAEEALe5CZ3ZcQO+9sa08dB+yOYrDP4F2WX9pkDZKGZQUSKFfwpHmBdy
ck1pz9GDGGdovH3ryTTxcqqI4VQOIlw6AlxiCnZ/4VPxCfEMKpl220uj316fvGDS
HJwv0643kvS7TguWIsva8ry9y2uauTKA+ndltmSapkmPyQ9NL4HhWHUFABEBAAG0
JEF1dG9nZW5lcmF0ZWQgS2V5IDxtYXR0aGV3QGxhcHRvcDIxPojOBBMBCgA4FiEE
VME3fVLzmWaiIsSzfAWR9ufi1RQFAmV1hCQCGy8FCwkIBwIGFQoJCAsCBBYCAwEC
HgECF4AACgkQfAWR9ufi1RShgwP9H8/TNBSbfqLrJMfo8igJpjL/QzGLUiCaYVde
PjT9ZYjszupV4i+PyRKydSzR+/6Nu7eZNb+lG6T+J8VlYdk2G+w8SzFgb7Bf98LT
+Hp49OXPf6HPrGZ41jW/5O1wtO/jolHgNkmtiSuxDXdr49gBOsGFWcujVE1ZNGEW
jlqqRaU=
=Hmce
-----END PGP PUBLIC KEY BLOCK-----
""".strip(),
            }
        ]
    },
    "second-namespace": {
        "gpg_keys": [
            # Test real DELETE
            {
                # C86CB378D2FE185EF6B129511EA75F18C926FB95
                "ascii_armor": """
-----BEGIN PGP PUBLIC KEY BLOCK-----

mI0EZXWIZgEEAMJNDQ63LSxip8G+oM4AKAfnYpWmtfrOXdDs/3Fv7LXwhSLDjzO+
vXT946LS3bsiNE+mbTqKm+fx5saOpVwK7mWzCgdLVuZWSCTkVJcNdyIDJe15Yk35
/GI4o1+0ICULPIGCRqXnPnLr0AcDV67pP1xEujAab2TmiFEGu7TmiEBRABEBAAG0
JEF1dG9nZW5lcmF0ZWQgS2V5IDxtYXR0aGV3QGxhcHRvcDIxPojOBBMBCgA4FiEE
yGyzeNL+GF72sSlRHqdfGMkm+5UFAmV1iGYCGy8FCwkIBwIGFQoJCAsCBBYCAwEC
HgECF4AACgkQHqdfGMkm+5UEsgP+Mu9gTe/VFmE+nwGTfK9a5d5f4zdIZDxhEse+
G6ZDLIVHSvsPQXLZNEDxNFFeoW47QJ5kYQjK0TKt+FdZoOeEcvWOeCxpFa1h1hsX
FdV/6+HIZM36mVdIobbUEkhKmp/TMLF6+i+7QkJ1EBfdkguu7Twe+ymrhWJ1kycl
8q67TGU=
=jd5w
-----END PGP PUBLIC KEY BLOCK-----

""".strip(),
            },
            # Test failed delete attempts
            {
                # 932E6646677B15817B515242671CF4D4386E1642
                "ascii_armor": """
-----BEGIN PGP PUBLIC KEY BLOCK-----

mI0EZXWIcQEEAOSL2IwgBL59gGDg9bSQNVt++Pw9Wvl4qZ7zxpEOJBBRbYe9pk13
qlZbaR3EKa+VH2k/T3BAh58MyF9Tb9BgJ2iXPaPekm2STPAMXqlvJ1XkN7v8qJk1
vaM8tHHx+/NoZcBfRK6BUPadW3KnD+d59vbREGm8bpPjesfOUFD8rY7tABEBAAG0
JEF1dG9nZW5lcmF0ZWQgS2V5IDxtYXR0aGV3QGxhcHRvcDIxPojOBBMBCgA4FiEE
ky5mRmd7FYF7UVJCZxz01DhuFkIFAmV1iHECGy8FCwkIBwIGFQoJCAsCBBYCAwEC
HgECF4AACgkQZxz01DhuFkI/zwQAil4loUns0zyuPy0nkLKLNnGEOcSvMb42I1ys
N8nmfJBjR6hXsaBVD7RRwNB5Q5hI77lPLUsdssNFnwV7Mch6aWsRYgUrdAcAywFB
imBSul7EbQ6LViv/1J/vN6ew2pd1ssnz0u8HzZvKZfAUqC9dj8pRZpr1moGMZ0zi
qaXczEQ=
=FwgT
-----END PGP PUBLIC KEY BLOCK-----

""".strip(),
            },
            {
                # 4703BD2FD3F62F9EBED9936D91F8403784C7F127
                "ascii_armor": """
-----BEGIN PGP PUBLIC KEY BLOCK-----

mI0EZXv5UgEEAOXiZ+GWol1SJrokFhwtyyt9xsSvJVWfj7PDRoM4XFyWM6SWpITK
38wjPKPl5xSRYCaY7eJwLJLuqepMi2hRZCi7tvgYSqu0uVtVuEkGpRXIRlHgLRZ5
bkKuHo7ClSDDfkMHaNz2KyG8H6OFJCf7W59ktbvkYtk1avgFyAL5+RovABEBAAG0
JEF1dG9nZW5lcmF0ZWQgS2V5IDxtYXR0aGV3QGxhcHRvcDIxPojOBBMBCgA4FiEE
RwO9L9P2L56+2ZNtkfhAN4TH8ScFAmV7+VICGy8FCwkIBwIGFQoJCAsCBBYCAwEC
HgECF4AACgkQkfhAN4TH8SfcowQAsLaG15wW8ZQ2zHh/IgyVku3r8BKI3IV22rkC
6IhwujdBcUd/FqRmPwCJ+fgbVJjt1HNAFdlQyX1YUJhOeq6sfgDFgu1o53j1R1aZ
3z1djSJoM5H9MzAOdki/HYVVQu8vqH/r/0HbIjuuG2bhkrbE6RMbAZkCSWN/cwBW
1/Ht1b0=
=uFHw
-----END PGP PUBLIC KEY BLOCK-----

""".strip(),
            }
        ],
        "providers": {
            "in-use": {
                "repository": {
                    "provider_source": "Test Github Autogenerate",
                    "provider_id": "second-namespace/terraform-provider-in-use",
                    "name": "terraform-provider-in-use",
                    "description": "Empty Provider Publish",
                    "owner": "second-namespace",
                    "clone_url": "https://git.example.com/initalproviders/terraform-provider-in-use.git",
                    "logo_url": "https://git.example.com/initalproviders/terraform-provider-test-initial.png"
                },
                "category_slug": "visible-monitoring",
                "use_default_provider_source_auth": True,
                "tier": "community",
                "versions": {
                    "1.0.0": {
                        "git_tag": "v1.0.0",
                        "gpg_key_fingerprint": "4703BD2FD3F62F9EBED9936D91F8403784C7F127"
                    }
                }
            },
        }
    }
}


class TestApiGpgKeyGet(TerraregIntegrationTest):
    """Test ApiGpgKey get endpoint"""

    _TEST_DATA = test_data

    def test_endpoint(self, client, test_request_context):
        """Test endpoint."""
        with test_request_context:
            gpg_key = terrareg.models.GpgKey.get_by_fingerprint("54C1377D52F39966A222C4B37C0591F6E7E2D514")
            db_row = gpg_key._get_db_row()
        res = client.get(f'/v2/gpg-keys/test-namespace/{gpg_key.pk}')
        assert res.status_code == 200
        assert res.json == {
            'data': {
                'attributes': {
                    'ascii-armor': '-----BEGIN PGP PUBLIC KEY BLOCK-----\n'
                                   '\n'
                                   'mI0EZXWEJAEEALe5CZ3ZcQO+9sa08dB+yOYrDP4F2WX9pkDZKGZQUSKFfwpHmBdy\n'
                                   'ck1pz9GDGGdovH3ryTTxcqqI4VQOIlw6AlxiCnZ/4VPxCfEMKpl220uj316fvGDS\n'
                                   'HJwv0643kvS7TguWIsva8ry9y2uauTKA+ndltmSapkmPyQ9NL4HhWHUFABEBAAG0\n'
                                   'JEF1dG9nZW5lcmF0ZWQgS2V5IDxtYXR0aGV3QGxhcHRvcDIxPojOBBMBCgA4FiEE\n'
                                   'VME3fVLzmWaiIsSzfAWR9ufi1RQFAmV1hCQCGy8FCwkIBwIGFQoJCAsCBBYCAwEC\n'
                                   'HgECF4AACgkQfAWR9ufi1RShgwP9H8/TNBSbfqLrJMfo8igJpjL/QzGLUiCaYVde\n'
                                   'PjT9ZYjszupV4i+PyRKydSzR+/6Nu7eZNb+lG6T+J8VlYdk2G+w8SzFgb7Bf98LT\n'
                                   '+Hp49OXPf6HPrGZ41jW/5O1wtO/jolHgNkmtiSuxDXdr49gBOsGFWcujVE1ZNGEW\n'
                                   'jlqqRaU=\n'
                                   '=Hmce\n'
                                   '-----END PGP PUBLIC KEY BLOCK-----',
                    'created-at': db_row['created_at'].isoformat(),
                    'key-id': '7C0591F6E7E2D514',
                    'namespace': 'test-namespace',
                    'source': '',
                    'source-url': None,
                    'trust-signature': '',
                    'updated-at': db_row['updated_at'].isoformat()
                },
                'id': '2',
                'type': 'gpg-keys'
            }
        }

    def test_non_matching_key_id_and_namespace(self, client, test_request_context):
        """Test endpoint with non matching key ID and namespace."""
        with test_request_context:
            gpg_key = terrareg.models.GpgKey.get_by_fingerprint("54C1377D52F39966A222C4B37C0591F6E7E2D514")
        res = client.get(f'/v2/gpg-keys/second-namespace/{gpg_key.pk}')
        assert res.status_code == 404
        assert res.json == {'errors': ['Not Found']}

    def test_non_existent_key(self, client):
        """Test endpoint with non matching key ID and namespace."""
        res = client.get(f'/v2/gpg-keys/second-namespace/6246246246')
        assert res.status_code == 404
        assert res.json == {'errors': ['Not Found']}

    def test_endpoint_non_existent_namespace(self, client, test_request_context):
        """Test endpoint with non-existent namespace."""
        with test_request_context:
            gpg_key = terrareg.models.GpgKey.get_by_fingerprint("54C1377D52F39966A222C4B37C0591F6E7E2D514")
        res = client.get(f'/v2/gpg-keys/doesnotexist/{gpg_key.pk}')
        assert res.status_code == 404
        assert res.json == {'errors': ['Not Found']}

    def test_unauthenticated(self, client, test_request_context):
        """Test unauthenticated call to API"""
        with test_request_context:
            gpg_key = terrareg.models.GpgKey.get_by_fingerprint("54C1377D52F39966A222C4B37C0591F6E7E2D514")
        def call_endpoint():
            return client.get(f'/v2/gpg-keys/test-namespace/{gpg_key.pk}')

        self._test_unauthenticated_read_api_endpoint_test(call_endpoint)


class TestApiGpgKeyDelete(TerraregIntegrationTest):
    """Test ApiGpgKey DELETE endpoint"""

    _TEST_DATA = test_data

    def test_endpoint(self, client, test_request_context):
        """Test endpoint."""
        with test_request_context:
            gpg_key = terrareg.models.GpgKey.get_by_fingerprint("C86CB378D2FE185EF6B129511EA75F18C926FB95")
        mock_check_csrf_token = unittest.mock.MagicMock()
        with unittest.mock.patch('terrareg.csrf.check_csrf_token', mock_check_csrf_token):
            res = client.delete(f'/v2/gpg-keys/second-namespace/{gpg_key.pk}', json={'csrf_token': 'testcsrf'})

        mock_check_csrf_token.assert_called_once_with('testcsrf')
        assert res.status_code == 201
        assert res.json == {}

        with test_request_context:
            gpg_key = terrareg.models.GpgKey.get_by_fingerprint("C86CB378D2FE185EF6B129511EA75F18C926FB95")
            assert gpg_key is None

    def test_delete_in_use_gpg_key(self, client, test_request_context):
        """Test deleting a GPG key that is in use."""
        with test_request_context:
            gpg_key = terrareg.models.GpgKey.get_by_fingerprint("4703BD2FD3F62F9EBED9936D91F8403784C7F127")
        mock_check_csrf_token = unittest.mock.MagicMock()
        with unittest.mock.patch('terrareg.csrf.check_csrf_token', mock_check_csrf_token):
            res = client.delete(f'/v2/gpg-keys/second-namespace/{gpg_key.pk}', json={'csrf_token': 'testcsrf'})

        mock_check_csrf_token.assert_called_once_with('testcsrf')
        assert res.status_code == 500
        assert res.json == {"status": "Error", "message": "Cannot delete GPG key as it is used by provider versions"}

        with test_request_context:
            gpg_key = terrareg.models.GpgKey.get_by_fingerprint("4703BD2FD3F62F9EBED9936D91F8403784C7F127")
            assert gpg_key is not None

    def test_non_matching_key_id_and_namespace(self, client, test_request_context):
        """Test endpoint with non matching key ID and namespace."""
        with test_request_context:
            gpg_key = terrareg.models.GpgKey.get_by_fingerprint("932E6646677B15817B515242671CF4D4386E1642")
        with unittest.mock.patch('terrareg.csrf.check_csrf_token', unittest.mock.MagicMock()) as mock_check_csrf_token:
            res = client.delete(f'/v2/gpg-keys/test-namespace/{gpg_key.pk}', json={'csrf_token': 'testcsrf'})
        assert res.status_code == 404
        assert res.json == {'errors': ['Not Found']}

    def test_non_existent_key(self, client):
        """Test endpoint with non matching key ID and namespace."""
        res = client.delete(f'/v2/gpg-keys/second-namespace/6246246246', json={'csrf_token': 'testcsrf'})
        assert res.status_code == 404
        assert res.json == {'errors': ['Not Found']}

    def test_endpoint_non_existent_namespace(self, client, test_request_context):
        """Test endpoint with non-existent namespace."""
        with test_request_context:
            gpg_key = terrareg.models.GpgKey.get_by_fingerprint("932E6646677B15817B515242671CF4D4386E1642")
        with unittest.mock.patch('terrareg.csrf.check_csrf_token', unittest.mock.MagicMock()) as mock_check_csrf_token:
            res = client.delete(f'/v2/gpg-keys/doesnotexist/{gpg_key.pk}', json={'csrf_token': 'testcsrf'})
        assert res.status_code == 404
        assert res.json == {'errors': ['Not Found']}

    def test_unauthenticated(self, client, test_request_context):
        """Test unauthenticated call to API"""
        with test_request_context:
            gpg_key = terrareg.models.GpgKey.get_by_fingerprint("932E6646677B15817B515242671CF4D4386E1642")

        mock_auth_method = unittest.mock.MagicMock()
        mock_auth_method.check_namespace_access = unittest.mock.MagicMock(return_value=False)
        mock_auth_method.get_username.return_value = 'unauthenticated user'
        mock_get_current_auth_method = unittest.mock.MagicMock(return_value=mock_auth_method)

        with unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', mock_get_current_auth_method):
            res = client.delete(f'/v2/gpg-keys/test-namespace/{gpg_key.pk}', json={'csrf_token': 'testcsrf'})
            assert res.status_code == 403
            assert res.json == {
                'message': "You don't have the permission to access the requested resource. It is either read-protected or not readable by the server."
            }

            mock_auth_method.check_namespace_access.assert_called_once_with(
                terrareg.user_group_namespace_permission_type.UserGroupNamespacePermissionType.FULL,
                namespace='test-namespace'
            )
