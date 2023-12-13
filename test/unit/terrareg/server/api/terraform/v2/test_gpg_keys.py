
import unittest.mock
import datetime

import pytest

from test.integration.terrareg import TerraregIntegrationTest
from test import client, app_context, test_request_context
import terrareg.user_group_namespace_permission_type
import terrareg.models


class AnyDateString:

    def __eq__(self, __o):
        if isinstance(__o, str):
            try:
                datetime.datetime.fromisoformat(__o)
                return True
            except:
                pass
        return False



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
            }

        ]
    },
    "test-namespace3": {
        "gpg_keys": [
{
                # BB9E07F3ACAB9C5B7ECC6C268F8450E7137572A9
                "ascii_armor": """
-----BEGIN PGP PUBLIC KEY BLOCK-----

mI0EZXgOKgEEAK8rzJSWZ6arT7BmsAYOd1XGoQS9mXe0LUbkEforluXw7S4ycdxy
SxHQlKqDUnF7KHi1jGBD+qKmAPliSVM+HiUb2v9NF2n72h8x2/xpeDBQtCemoyh/
Dg0JC4KtBXvY65wf77axAcPVcJsOok89zB02bRLt0Ul38kDTkOvzqpczABEBAAG0
JEF1dG9nZW5lcmF0ZWQgS2V5IDxtYXR0aGV3QGxhcHRvcDIxPojOBBMBCgA4FiEE
u54H86yrnFt+zGwmj4RQ5xN1cqkFAmV4DioCGy8FCwkIBwIGFQoJCAsCBBYCAwEC
HgECF4AACgkQj4RQ5xN1cqmJhQP/f9VvYlHUb+bOI0IKS0ZT1Ei7QVtpOe20IGK8
eLts6l9aDbNZ8dBMIm4N1JxoBZ7jC9f/rx3cEYu/IV2qVuvCIg2KpnbA0hBh00/c
ITgfvu6Gpot4OVBV6zHlKTT6We06onto9/sfEPjqtQxje2mQA8f9TFwYS7ZWaiMb
yemtm3E=
=gJbO
-----END PGP PUBLIC KEY BLOCK-----
""".strip(),
            }
        ]
    }
}


class TestApiGpgKeysGet(TerraregIntegrationTest):
    """Test ApiGpgKey get endpoint"""

    _TEST_DATA = test_data

    def test_endpoint_with_namespace(self, client, test_request_context):
        """Test endpoint with namespace."""

        res = client.get('/v2/gpg-keys?filter[namespace]=second-namespace')
        assert res.status_code == 200
        assert res.json == {
            'data': [{'attributes': {'ascii-armor': '-----BEGIN PGP PUBLIC KEY '
                                          'BLOCK-----\n'
                                          '\n'
                                          'mI0EZXWIZgEEAMJNDQ63LSxip8G+oM4AKAfnYpWmtfrOXdDs/3Fv7LXwhSLDjzO+\n'
                                          'vXT946LS3bsiNE+mbTqKm+fx5saOpVwK7mWzCgdLVuZWSCTkVJcNdyIDJe15Yk35\n'
                                          '/GI4o1+0ICULPIGCRqXnPnLr0AcDV67pP1xEujAab2TmiFEGu7TmiEBRABEBAAG0\n'
                                          'JEF1dG9nZW5lcmF0ZWQgS2V5IDxtYXR0aGV3QGxhcHRvcDIxPojOBBMBCgA4FiEE\n'
                                          'yGyzeNL+GF72sSlRHqdfGMkm+5UFAmV1iGYCGy8FCwkIBwIGFQoJCAsCBBYCAwEC\n'
                                          'HgECF4AACgkQHqdfGMkm+5UEsgP+Mu9gTe/VFmE+nwGTfK9a5d5f4zdIZDxhEse+\n'
                                          'G6ZDLIVHSvsPQXLZNEDxNFFeoW47QJ5kYQjK0TKt+FdZoOeEcvWOeCxpFa1h1hsX\n'
                                          'FdV/6+HIZM36mVdIobbUEkhKmp/TMLF6+i+7QkJ1EBfdkguu7Twe+ymrhWJ1kycl\n'
                                          '8q67TGU=\n'
                                          '=jd5w\n'
                                          '-----END PGP PUBLIC KEY BLOCK-----',
                           'created-at': AnyDateString(),
                           'key-id': '1EA75F18C926FB95',
                           'namespace': 'second-namespace',
                           'source': '',
                           'source-url': None,
                           'trust-signature': '',
                           'updated-at': AnyDateString()},
            'id': '3',
            'type': 'gpg-keys'},
           {'attributes': {'ascii-armor': '-----BEGIN PGP PUBLIC KEY '
                                          'BLOCK-----\n'
                                          '\n'
                                          'mI0EZXWIcQEEAOSL2IwgBL59gGDg9bSQNVt++Pw9Wvl4qZ7zxpEOJBBRbYe9pk13\n'
                                          'qlZbaR3EKa+VH2k/T3BAh58MyF9Tb9BgJ2iXPaPekm2STPAMXqlvJ1XkN7v8qJk1\n'
                                          'vaM8tHHx+/NoZcBfRK6BUPadW3KnD+d59vbREGm8bpPjesfOUFD8rY7tABEBAAG0\n'
                                          'JEF1dG9nZW5lcmF0ZWQgS2V5IDxtYXR0aGV3QGxhcHRvcDIxPojOBBMBCgA4FiEE\n'
                                          'ky5mRmd7FYF7UVJCZxz01DhuFkIFAmV1iHECGy8FCwkIBwIGFQoJCAsCBBYCAwEC\n'
                                          'HgECF4AACgkQZxz01DhuFkI/zwQAil4loUns0zyuPy0nkLKLNnGEOcSvMb42I1ys\n'
                                          'N8nmfJBjR6hXsaBVD7RRwNB5Q5hI77lPLUsdssNFnwV7Mch6aWsRYgUrdAcAywFB\n'
                                          'imBSul7EbQ6LViv/1J/vN6ew2pd1ssnz0u8HzZvKZfAUqC9dj8pRZpr1moGMZ0zi\n'
                                          'qaXczEQ=\n'
                                          '=FwgT\n'
                                          '-----END PGP PUBLIC KEY BLOCK-----',
                           'created-at': AnyDateString(),
                           'key-id': '671CF4D4386E1642',
                           'namespace': 'second-namespace',
                           'source': '',
                           'source-url': None,
                           'trust-signature': '',
                           'updated-at': AnyDateString()},
            'id': '4',
            'type': 'gpg-keys'}],

        }

    def test_endpoint_with_multiple_namespaces(self, client, test_request_context):
        """Test endpoint with multiple namespace."""

        res = client.get('/v2/gpg-keys?filter[namespace]=second-namespace,test-namespace3')
        assert res.status_code == 200
        assert res.json == {
            'data': [{'attributes': {'ascii-armor': '-----BEGIN PGP PUBLIC KEY '
                                          'BLOCK-----\n'
                                          '\n'
                                          'mI0EZXWIZgEEAMJNDQ63LSxip8G+oM4AKAfnYpWmtfrOXdDs/3Fv7LXwhSLDjzO+\n'
                                          'vXT946LS3bsiNE+mbTqKm+fx5saOpVwK7mWzCgdLVuZWSCTkVJcNdyIDJe15Yk35\n'
                                          '/GI4o1+0ICULPIGCRqXnPnLr0AcDV67pP1xEujAab2TmiFEGu7TmiEBRABEBAAG0\n'
                                          'JEF1dG9nZW5lcmF0ZWQgS2V5IDxtYXR0aGV3QGxhcHRvcDIxPojOBBMBCgA4FiEE\n'
                                          'yGyzeNL+GF72sSlRHqdfGMkm+5UFAmV1iGYCGy8FCwkIBwIGFQoJCAsCBBYCAwEC\n'
                                          'HgECF4AACgkQHqdfGMkm+5UEsgP+Mu9gTe/VFmE+nwGTfK9a5d5f4zdIZDxhEse+\n'
                                          'G6ZDLIVHSvsPQXLZNEDxNFFeoW47QJ5kYQjK0TKt+FdZoOeEcvWOeCxpFa1h1hsX\n'
                                          'FdV/6+HIZM36mVdIobbUEkhKmp/TMLF6+i+7QkJ1EBfdkguu7Twe+ymrhWJ1kycl\n'
                                          '8q67TGU=\n'
                                          '=jd5w\n'
                                          '-----END PGP PUBLIC KEY BLOCK-----',
                           'created-at': AnyDateString(),
                           'key-id': '1EA75F18C926FB95',
                           'namespace': 'second-namespace',
                           'source': '',
                           'source-url': None,
                           'trust-signature': '',
                           'updated-at': AnyDateString()},
            'id': '3',
            'type': 'gpg-keys'},
           {'attributes': {'ascii-armor': '-----BEGIN PGP PUBLIC KEY '
                                          'BLOCK-----\n'
                                          '\n'
                                          'mI0EZXWIcQEEAOSL2IwgBL59gGDg9bSQNVt++Pw9Wvl4qZ7zxpEOJBBRbYe9pk13\n'
                                          'qlZbaR3EKa+VH2k/T3BAh58MyF9Tb9BgJ2iXPaPekm2STPAMXqlvJ1XkN7v8qJk1\n'
                                          'vaM8tHHx+/NoZcBfRK6BUPadW3KnD+d59vbREGm8bpPjesfOUFD8rY7tABEBAAG0\n'
                                          'JEF1dG9nZW5lcmF0ZWQgS2V5IDxtYXR0aGV3QGxhcHRvcDIxPojOBBMBCgA4FiEE\n'
                                          'ky5mRmd7FYF7UVJCZxz01DhuFkIFAmV1iHECGy8FCwkIBwIGFQoJCAsCBBYCAwEC\n'
                                          'HgECF4AACgkQZxz01DhuFkI/zwQAil4loUns0zyuPy0nkLKLNnGEOcSvMb42I1ys\n'
                                          'N8nmfJBjR6hXsaBVD7RRwNB5Q5hI77lPLUsdssNFnwV7Mch6aWsRYgUrdAcAywFB\n'
                                          'imBSul7EbQ6LViv/1J/vN6ew2pd1ssnz0u8HzZvKZfAUqC9dj8pRZpr1moGMZ0zi\n'
                                          'qaXczEQ=\n'
                                          '=FwgT\n'
                                          '-----END PGP PUBLIC KEY BLOCK-----',
                           'created-at': AnyDateString(),
                           'key-id': '671CF4D4386E1642',
                           'namespace': 'second-namespace',
                           'source': '',
                           'source-url': None,
                           'trust-signature': '',
                           'updated-at': AnyDateString()},
            'id': '4',
            'type': 'gpg-keys'},
           {'attributes': {'ascii-armor': '-----BEGIN PGP PUBLIC KEY '
                                          'BLOCK-----\n'
                                          '\n'
                                          'mI0EZXgOKgEEAK8rzJSWZ6arT7BmsAYOd1XGoQS9mXe0LUbkEforluXw7S4ycdxy\n'
                                          'SxHQlKqDUnF7KHi1jGBD+qKmAPliSVM+HiUb2v9NF2n72h8x2/xpeDBQtCemoyh/\n'
                                          'Dg0JC4KtBXvY65wf77axAcPVcJsOok89zB02bRLt0Ul38kDTkOvzqpczABEBAAG0\n'
                                          'JEF1dG9nZW5lcmF0ZWQgS2V5IDxtYXR0aGV3QGxhcHRvcDIxPojOBBMBCgA4FiEE\n'
                                          'u54H86yrnFt+zGwmj4RQ5xN1cqkFAmV4DioCGy8FCwkIBwIGFQoJCAsCBBYCAwEC\n'
                                          'HgECF4AACgkQj4RQ5xN1cqmJhQP/f9VvYlHUb+bOI0IKS0ZT1Ei7QVtpOe20IGK8\n'
                                          'eLts6l9aDbNZ8dBMIm4N1JxoBZ7jC9f/rx3cEYu/IV2qVuvCIg2KpnbA0hBh00/c\n'
                                          'ITgfvu6Gpot4OVBV6zHlKTT6We06onto9/sfEPjqtQxje2mQA8f9TFwYS7ZWaiMb\n'
                                          'yemtm3E=\n'
                                          '=gJbO\n'
                                          '-----END PGP PUBLIC KEY BLOCK-----',
                           'created-at': AnyDateString(),
                           'key-id': '8F8450E7137572A9',
                           'namespace': 'test-namespace3',
                           'source': '',
                           'source-url': None,
                           'trust-signature': '',
                           'updated-at': AnyDateString()},
            'id': '5',
            'type': 'gpg-keys'}],
        }

    def test_endpoint_with_non_existent_namespace(self, client, test_request_context):
        """Test endpoint with non-existent namespace."""

        res = client.get('/v2/gpg-keys?filter[namespace]=doesnotexist')
        assert res.status_code == 200
        assert res.json == {'data': []}

    def test_endpoint_without_namespace_filter(self, client, test_request_context):
        """Test endpoint without namespace filter."""

        res = client.get('/v2/gpg-keys')
        assert res.status_code == 400
        assert res.json == {
            'message': {
                'filter[namespace]': 'Comma-separated list of namespaces to obtain GPG keys for'
            }
        }

    def test_unauthenticated(self, client):
        """Test unauthenticated call to API"""
        def call_endpoint():
            return client.get('/v2/gpg-keys?filter[namespace]=doesnotexist')

        self._test_unauthenticated_read_api_endpoint_test(call_endpoint)



class TestApiGpgKeysPost(TerraregIntegrationTest):
    """Test ApiGpgKey POST endpoint"""

    _TEST_DATA = test_data

    def test_endpoint(self, client, test_request_context):
        """Test endpoint with namespace."""

        with unittest.mock.patch('terrareg.csrf.check_csrf_token', unittest.mock.MagicMock()) as mock_check_csrf_token:
            res = client.post(
                '/v2/gpg-keys',
                json={
                    "data": {
                        "type": "gpg-keys",
                        "attributes": {
                            "namespace": "test-namespace3",
                            "ascii-armor": """
-----BEGIN PGP PUBLIC KEY BLOCK-----

mI0EZXgRngEEAKaYmbZK/MZIy0jjn3pX8Mt0tfxrQTp6exH7bZs5Ny1oJ4QTl3ac
4vYXvbfC8Yj7DFfs1vq98eNcOnM/USQ/07TZbZfg+4xbY15zLoejtQdAHdsTBAUH
kBj4U2OQmFV+ojdy7w0Qps9tlI7YPk8OwszmdlgvOLoT9v2PMGwv+jNzABEBAAG0
JEF1dG9nZW5lcmF0ZWQgS2V5IDxtYXR0aGV3QGxhcHRvcDIxPojOBBMBCgA4FiEE
ECijmYXVXj+iT3wDn9e147tCsnEFAmV4EZ4CGy8FCwkIBwIGFQoJCAsCBBYCAwEC
HgECF4AACgkQn9e147tCsnHHKAP/S3mojxFe3ja6wQBM//pDgr949O8BoInk239V
uhJu/UvrdqyXmjh5gJ0x2YVFZjhPOll334E4W6CMHi9IR0flZMIcTIjWTUjlTY74
xUgTTj2yKpD5Ia6fWrxOvMq/bnfA8Ai2Kr+AkulvmBQtInTpbUWllXVoCxtzFIck
IuzCt+Q=
=P0YS
-----END PGP PUBLIC KEY BLOCK-----
"""
                        },
                        "csrf_token": "testcsrftoken"
                    }
                }
            )

        assert res.status_code == 200
        assert res.json == {
            'data': {'attributes': {'ascii-armor': '-----BEGIN PGP PUBLIC KEY BLOCK-----\n'
                                         '\n'
                                         'mI0EZXgRngEEAKaYmbZK/MZIy0jjn3pX8Mt0tfxrQTp6exH7bZs5Ny1oJ4QTl3ac\n'
                                         '4vYXvbfC8Yj7DFfs1vq98eNcOnM/USQ/07TZbZfg+4xbY15zLoejtQdAHdsTBAUH\n'
                                         'kBj4U2OQmFV+ojdy7w0Qps9tlI7YPk8OwszmdlgvOLoT9v2PMGwv+jNzABEBAAG0\n'
                                         'JEF1dG9nZW5lcmF0ZWQgS2V5IDxtYXR0aGV3QGxhcHRvcDIxPojOBBMBCgA4FiEE\n'
                                         'ECijmYXVXj+iT3wDn9e147tCsnEFAmV4EZ4CGy8FCwkIBwIGFQoJCAsCBBYCAwEC\n'
                                         'HgECF4AACgkQn9e147tCsnHHKAP/S3mojxFe3ja6wQBM//pDgr949O8BoInk239V\n'
                                         'uhJu/UvrdqyXmjh5gJ0x2YVFZjhPOll334E4W6CMHi9IR0flZMIcTIjWTUjlTY74\n'
                                         'xUgTTj2yKpD5Ia6fWrxOvMq/bnfA8Ai2Kr+AkulvmBQtInTpbUWllXVoCxtzFIck\n'
                                         'IuzCt+Q=\n'
                                         '=P0YS\n'
                                         '-----END PGP PUBLIC KEY BLOCK-----',
                          'created-at': AnyDateString(),
                          'key-id': '9FD7B5E3BB42B271',
                          'namespace': 'test-namespace3',
                          'source': '',
                          'source-url': None,
                          'trust-signature': '',
                          'updated-at': AnyDateString()},
           'id': '6',
           'type': 'gpg-keys'},
        }

        mock_check_csrf_token.assert_called_once_with('testcsrftoken')

        with test_request_context:
            assert terrareg.models.GpgKey.get_by_fingerprint("1028A39985D55E3FA24F7C039FD7B5E3BB42B271") is not None


    def test_non_existent_namespace(self, client, test_request_context):
        """Test endpoint with non-existent namespace."""

        with unittest.mock.patch('terrareg.csrf.check_csrf_token', unittest.mock.MagicMock()) as mock_check_csrf_token:
            res = client.post(
                '/v2/gpg-keys',
                json={
                    "data": {
                        "type": "gpg-keys",
                        "attributes": {
                            "namespace": "doesnotexist",
                            "ascii-armor": """
-----BEGIN PGP PUBLIC KEY BLOCK-----

mI0EZXgSVAEEAOmSYSJ79eO6nyIiTLrgfCngRJQ8tMLufkmyYCg69si9ZlrDVeYg
NapSCFNmi+1OtLhQhJXviJh1MeuQmQt+nyidmXYIUPoW9dT8Dh0zsC1VL1MCS8nI
/wNyDP6vbjuFyChEILeUkVuZWj+eWiF0d+aZpNObuP8HbIgj4VdDwvVpABEBAAG0
JEF1dG9nZW5lcmF0ZWQgS2V5IDxtYXR0aGV3QGxhcHRvcDIxPojOBBMBCgA4FiEE
JyIrrgaxqZ65WtdFPaKFhUx+PAgFAmV4ElQCGy8FCwkIBwIGFQoJCAsCBBYCAwEC
HgECF4AACgkQPaKFhUx+PAgJ0wQAwE0+WcqKBDLM1+XVjnEngPgIq1/iWT9yLcpx
2N2A5Oe8iNu+MUN5ag+TDWDFV62cGsHr+6e+kLC8SV9mWkwNT6a94WDrhr7ntwYR
EN3wkwn+pMHaJ9lX/KVKL9gseJaBM5rVHI93qBNbueqS/UJMa+h44M78q3m0TO4H
/WbFn7M=
=g7gZ
-----END PGP PUBLIC KEY BLOCK-----

"""
                        }
                    },
                    "csrf_token": "testcsrftoken"
                }
            )

        assert res.status_code == 400
        assert res.json == {'message': 'Namespace does not exist'}

        # Ensure GPG key does not exist
        with test_request_context:
            assert terrareg.models.GpgKey.get_by_fingerprint("27222BAE06B1A99EB95AD7453DA285854C7E3C08") is None

    def test_pre_existing_gpg_key(self, client, test_request_context):
        """Test endpoint with duplicate GPG key."""
        with test_request_context:
            # Get DB row of original key
            db_row = terrareg.models.GpgKey.get_by_fingerprint("BB9E07F3ACAB9C5B7ECC6C268F8450E7137572A9")._get_db_row()

        with unittest.mock.patch('terrareg.csrf.check_csrf_token', unittest.mock.MagicMock()) as mock_check_csrf_token:
            res = client.post(
                '/v2/gpg-keys',
                json={
                    "data": {
                        "type": "gpg-keys",
                        "attributes": {
                            "namespace": "second-namespace",
                            "ascii-armor": """
-----BEGIN PGP PUBLIC KEY BLOCK-----

mI0EZXgOKgEEAK8rzJSWZ6arT7BmsAYOd1XGoQS9mXe0LUbkEforluXw7S4ycdxy
SxHQlKqDUnF7KHi1jGBD+qKmAPliSVM+HiUb2v9NF2n72h8x2/xpeDBQtCemoyh/
Dg0JC4KtBXvY65wf77axAcPVcJsOok89zB02bRLt0Ul38kDTkOvzqpczABEBAAG0
JEF1dG9nZW5lcmF0ZWQgS2V5IDxtYXR0aGV3QGxhcHRvcDIxPojOBBMBCgA4FiEE
u54H86yrnFt+zGwmj4RQ5xN1cqkFAmV4DioCGy8FCwkIBwIGFQoJCAsCBBYCAwEC
HgECF4AACgkQj4RQ5xN1cqmJhQP/f9VvYlHUb+bOI0IKS0ZT1Ei7QVtpOe20IGK8
eLts6l9aDbNZ8dBMIm4N1JxoBZ7jC9f/rx3cEYu/IV2qVuvCIg2KpnbA0hBh00/c
ITgfvu6Gpot4OVBV6zHlKTT6We06onto9/sfEPjqtQxje2mQA8f9TFwYS7ZWaiMb
yemtm3E=
=gJbO
-----END PGP PUBLIC KEY BLOCK-----
"""
                        }
                    },
                    "csrf_token": "testcsrftoken"
                }
            )

        assert res.status_code == 500
        assert res.json == {'message': 'A duplicate GPG key exists with the same fingerprint', 'status': 'Error'}

        # Ensure GPG key row has not been modified
        with test_request_context:
            # Get DB row of original key
            assert terrareg.models.GpgKey.get_by_fingerprint("BB9E07F3ACAB9C5B7ECC6C268F8450E7137572A9")._get_db_row() == db_row

    @pytest.mark.parametrize('gpg_key', [
        'Invalid Key',
        '',
        None,
        1234
    ])
    def test_invalid_gpg_key(self, client, gpg_key):
        """Test endpoint with invalid GPG key."""
        with unittest.mock.patch('terrareg.csrf.check_csrf_token', unittest.mock.MagicMock()) as mock_check_csrf_token:
            res = client.post(
                '/v2/gpg-keys',
                json={
                    "data": {
                        "type": "gpg-keys",
                        "attributes": {
                            "namespace": "second-namespace",
                            "ascii-armor": gpg_key
                        }
                    },
                    "csrf_token": "testcsrftoken"
                }
            )

        assert res.status_code == 500
        assert res.json == {'message': 'GPG key provided is invalid or could not be read', 'status': 'Error'}
