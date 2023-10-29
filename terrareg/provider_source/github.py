
import os
from re import A
import time
from typing import Dict, Union, List, Tuple
from urllib.parse import parse_qs

from cryptography.hazmat.primitives import serialization
from cryptography.hazmat.backends import default_backend
import jwt

import requests

from terrareg.errors import InvalidProviderSourceConfigError
from .base import BaseProviderSource
import terrareg.provider_source_type
import terrareg.repository_model
import terrareg.provider_version_model
import terrareg.provider_model
import terrareg.provider_source.repository_release_metadata
import terrareg.models
import terrareg.namespace_type


class GithubProviderSource(BaseProviderSource):

    TYPE = terrareg.provider_source_type.ProviderSourceType.GITHUB

    # TEMPORARY CACHE FOR INSTALLATION ACCESS TOKENS
    # @TODO REMOVE
    INSTALLATION_ID_TOKENS = {}

    @classmethod
    def generate_db_config_from_source_config(cls, config: Dict[str, str]) -> Dict[str, Union[str, bool]]:
        """Generate DB config from config"""
        db_config = {}
        for required_attr in ["base_url", "api_url", "client_id", "client_secret", "login_button_text", "private_key_path", "app_id"]:
            if not (val := config.get(required_attr)) or not isinstance(val, str):
                raise InvalidProviderSourceConfigError(f"Missing required Github provider source config: {required_attr}")

            db_config[required_attr] = val

        for optional_attr in ["default_access_token", "default_installation_id"]:
            if optional_attr in config:
                db_config[optional_attr] = config[optional_attr]

        for bool_attr in ["auto_generate_github_organisation_namespaces"]:
            val = config.get(bool_attr)
            if not isinstance(val, bool):
                raise InvalidProviderSourceConfigError(f"Missing required Github provider source config: {bool_attr}")
            db_config[bool_attr] = val

        return db_config

    def __init__(self, *args, **kwargs):
        """Store member variables"""
        super().__init__(*args, **kwargs)
        self._private_key_content = None

    @property
    def _client_id(self) -> Union[None, str]:
        """Return client ID"""
        return self._config.get("client_id")

    @property
    def _client_secret(self) -> Union[None, str]:
        """Return client secret"""
        return self._config.get("client_secret")

    @property
    def _base_url(self) -> Union[None, str]:
        """Return base Github URL"""
        return self._config.get("base_url")

    @property
    def _api_url(self) -> Union[None, str]:
        """Return Github API URL"""
        return self._config.get("api_url")

    @property
    def auto_generate_github_organisation_namespaces(self) -> bool:
        """Whether to namespaces should be automatically generated for each github organisation membership"""
        return self._config.get("auto_generate_github_organisation_namespaces", False)

    @property
    def login_button_text(self) -> str:
        """Return login buton text"""
        return self._config.get("login_button_text")

    @property
    def _private_key_path(self) -> str:
        """Return github app private key path"""
        return self._config.get("private_key_path")

    @property
    def _private_key(self) -> bytes:
        """Return content of private key"""
        if self._private_key_content is None:
            if not self._private_key_path or not os.path.isfile(self._private_key_path):
                return None

            with open(self._private_key_path, "r") as pem_file:
                self._private_key_content = pem_file.read().encode('utf-8')

        return self._private_key_content

    @property
    def github_app_id(self) -> int:
        """Return github app ID"""
        return self._config.get("app_id")

    def is_enabled(self) -> bool:
        """Whether github authentication is enabled"""
        return bool(self._client_id and self._client_secret and self._base_url and self._api_url)

    def get_login_redirect_url(self) -> str:
        """Generate login redirect URL"""
        return f"{self._base_url}/login/oauth/authorize?client_id={self._client_id}"

    def _get_default_access_token(self):
        """Return default access token, when communicating with Github outside the context of an authenticated user"""
        if installation_id := self._config.get("default_installation_id"):
            return self.generate_app_installation_token(installation_id)
        else:
            return self._config.get("default_access_token")

    def get_user_access_token(self, code: str) -> Union[None, str]:
        """Obtain access token from code"""
        if not code:
            return None

        res = requests.post(
            f"{self._base_url}/login/oauth/access_token",
            data={
                "client_id": self._client_id,
                "client_secret": self._client_secret,
                "code": code
            }
        )
        if res.status_code == 200:
            data = parse_qs(res.text)
            if (access_tokens := data.get("access_token")) and len(access_tokens) == 1:
                return access_tokens[0]

    def get_username(self, access_token) -> Union[None, str]:
        """Get username of authenticated user"""
        if not access_token:
            return None

        res = requests.get(
            f"{self._api_url}/user",
            headers={
                "X-GitHub-Api-Version": "2022-11-28",
                "Accept": "application/vnd.github+json",
                "Authorization": f"Bearer {access_token}"
            }
        )
        if res.status_code == 200:
            return res.json().get("login")

    def get_user_organisations(self, access_token) -> List[str]:
        """Get username of authenticated user"""
        if not access_token:
            return []

        res = requests.get(
            f"{self._api_url}/user/memberships/orgs",
            headers={
                "X-GitHub-Api-Version": "2022-11-28",
                "Accept": "application/vnd.github+json",
                "Authorization": f"Bearer {access_token}"
            }
        )

        if res.status_code == 200 and (response_data := res.json()):
            # Iterate over memberships, only get active memberships
            # that where the user is admin
            return [
                org_membership.get("organization", {}).get("login")
                for org_membership in response_data
                if (
                    org_membership.get("organization", {}).get("login") and
                    org_membership.get("state") == "active" and
                    org_membership.get("role") == "admin"
                )
            ]
        return []

    def _add_repository(self, repository_metadat: dict) -> None:
        """Create repository using metadata from github"""
        # @TODO Indicate if repo already exists to stop processing additional repos
        if (not (repo_id := repository_metadat.get("id")) or
                not (repo_name := repository_metadat.get("name")) or
                not (owner_name := repository_metadat.get("owner", {}).get("login")) or
                not (clone_url := repository_metadat.get("clone_url"))):
            return None

        terrareg.repository_model.Repository.create(
            provider_source=self,
            provider_id=repo_id,
            name=repo_name,
            description=repository_metadat.get("description"),
            owner=owner_name,
            clone_url=clone_url,
            logo_url=repository_metadat.get("owner", {}).get("avatar_url"),
        )


    def update_repositories(self, access_token: str) -> None:
        """Refresh list of repositories"""
        page = 1
        while True:
            res = requests.get(
                f"{self._api_url}/user/repos",
                params={
                    "visibility": "public",
                    "affiliation": "owner,organization_member",
                    "sort": "created",
                    "direction": "desc",
                    "per_page": "100",
                    "page": str(page)
                },
                headers={
                    "X-GitHub-Api-Version": "2022-11-28",
                    "Accept": "application/vnd.github+json",
                    "Authorization": f"Bearer {access_token}"
                }
            )
            if res.status_code != 200:
                print(f"Invalid response code from github: {res.status_code}")
                return

            results = res.json()

            for repository in results:
                self._add_repository(repository_metadat=repository)

            if len(results) < 100:
                break

            page += 1

    def _get_commit_hash_by_release(self,
                             repository: 'terrareg.repository_model.Repository',
                             tag_name: str,
                             access_token: str):
        """Return commit hash for tag name"""
        res = requests.get(
            f"{self._api_url}/repos/{repository.owner}/{repository.name}/git/ref/tags/{tag_name}",
            headers={
                "X-GitHub-Api-Version": "2022-11-28",
                "Accept": "application/vnd.github+json",
                "Authorization": f"Bearer {access_token}"
            }
        )
        if res.status_code != 200:
            return None

        return res.json().get("object", {}).get("sha")

    def get_new_releases(self, provider: 'terrareg.provider_model.Provider') -> List['terrareg.provider_source.repository_release_metadata.RepositoryReleaseMetadata']:
        """Obtain all repository releases that aren't associated with a pre-existing release"""
        repository = provider.repository
        page = 1
        releases = []
        obtain_results = True

        installation_id = self.get_github_app_installation_id(namespace=provider.namespace)
        access_token = self.generate_app_installation_token(installation_id=installation_id)

        while obtain_results:
            res = requests.get(
                f"{self._api_url}/repos/{repository.owner}/{repository.name}/releases",
                params={
                    "per_page": "100",
                    "page": str(page)
                },
                headers={
                    "X-GitHub-Api-Version": "2022-11-28",
                    "Accept": "application/vnd.github+json",
                    "Authorization": f"Bearer {access_token}"
                }
            )

            if res.status_code != 200:
                print(f"Invalid response code from github: {res.status_code}")
                return []

            results = res.json()
            for release in results:
                if (not (release_id := release.get("id")) or
                        not (release_name := release.get("name")) or
                        not (tag_name := release.get("tag_name")) or
                        not (archive_url := release.get("tarball_url")) or
                        not (commit_hash := self._get_commit_hash_by_release(
                            repository=repository,
                            tag_name=tag_name,
                            access_token=access_token))):
                    print("Could not obtain one of: release name, tag name, archive url or commit hash for release")
                    continue

                # Obtain version from tag and skip if it's invalid
                version = terrareg.provider_source.repository_release_metadata.RepositoryReleaseMetadata.tag_to_version(tag_name)
                if not version:
                    continue

                # If a provider version exists for the release,
                # exit early
                pre_existing_provider_version = terrareg.provider_version_model.ProviderVersion(provider=provider, version=version)
                if pre_existing_provider_version.exists:
                    obtain_results = False
                    break

                # Obtain release artifacts
                release_artifacts = self._get_release_artifacts_metadata(release_id=release_id, repository=repository, access_token=access_token)

                releases.append(
                    terrareg.provider_source.repository_release_metadata.RepositoryReleaseMetadata(
                        name=release_name,
                        tag=tag_name,
                        provider_id=release_id,
                        commit_hash=commit_hash,
                        archive_url=archive_url,
                        release_artifacts=release_artifacts,
                    )
                )

            if len(results) < 100:
                obtain_results = False
            else:
                page += 1
            
        return releases

    def _get_release_artifacts_metadata(self, repository: 'terrareg.repository_model.Repository',
                                        release_id: int, access_token: str) -> List['terrareg.provider_source.repository_release_metadata.ReleaseArtifactMetadata']:
        """Obtain list of release artifact metdata for a given release"""
        res = requests.get(
            f"{self._api_url}/repos/{repository.owner}/{repository.name}/releases/{release_id}/assets",
            params={
                "per_page": "100",
                "page": "1"
            },
            headers={
                "X-GitHub-Api-Version": "2022-11-28",
                "Accept": "application/vnd.github+json",
                "Authorization": f"Bearer {access_token}"
            }
        )

        if res.status_code != 200:
            print(f"_get_release_artifacts_metadata: Invalid response code from github assets list: {res.status_code}")
            return []

        return [
            terrareg.provider_source.repository_release_metadata.ReleaseArtifactMetadata(
                name=asset.get("name"),
                provider_id=asset.get("id")
            )
            for asset in res.json()
            if asset.get("name") and asset.get("id")
        ]

    def get_release_artifact(self,
                             provider: 'terrareg.provider_model.Provider',
                             artifact_metadata: 'terrareg.provider_source.repository_release_metadata.ReleaseArtifactMetadata',
                             release_metadata: 'terrareg.provider_source.repository_release_metadata.RepositoryReleaseMetadata') -> str:
        """Return release artifact file content"""
        repository = provider.repository

        installation_id = self.get_github_app_installation_id(namespace=provider.namespace)
        access_token = self.generate_app_installation_token(installation_id=installation_id)

        res = requests.get(
            f"{self._api_url}/repos/{repository.owner}/{repository.name}/releases/assets/{artifact_metadata.provider_id}",
            headers={
                "X-GitHub-Api-Version": "2022-11-28",
                "Accept": "application/octet-stream",
                "Authorization": f"Bearer {access_token}"
            },
            allow_redirects=True
        )
        if res.status_code == 404:
            print("get_release_artifact returned 404")
            return None
        return res.content

    def get_release_archive(self,
                            provider: 'terrareg.provider_model.Provider',
                            release_metadata: 'terrareg.provider_source.repository_release_metadata.RepositoryReleaseMetadata') -> Tuple[bytes, Union[None, str]]:
        """Obtain release archive, returning bytes of archive"""
        repository = provider.repository

        installation_id = self.get_github_app_installation_id(namespace=provider.namespace)
        access_token = self.generate_app_installation_token(installation_id=installation_id)

        res = requests.get(
            f"{self._api_url}/repos/{repository.owner}/{repository.name}/tarball/{release_metadata.tag}",
            headers={
                "X-GitHub-Api-Version": "2022-11-28",
                "Accept": "application/json",
                "Authorization": f"Bearer {access_token}"
            },
            allow_redirects=True
        )
        content = None
        if res.status_code == 404:
            print("get_release_artifact returned 404")
        else:
            content = res.content

        return content, f"{repository.owner}-{repository.name}-{release_metadata.commit_hash[0:7]}"

    def get_public_source_url(self, repository: 'terrareg.repository_model.Repository'):
        """Return public URL for source"""
        return f"{self._base_url}/{repository.owner}/{repository.name}"

    def get_public_artifact_download_url(self,
                                         provider_version: 'terrareg.provider_version_model.ProviderVersion',
                                         artifact_name: str):
        """Return public URL for source"""
        return f"{self.get_public_source_url(provider_version.provider.repository)}/releases/download/{provider_version.git_tag}/{artifact_name}"

    def generate_app_installation_token(self, installation_id: str):
        """Generate app access token"""
        if installation_id is None:
            return None

        if installation_id not in self.__class__.INSTALLATION_ID_TOKENS:

            res = requests.post(
                f"{self._api_url}/app/installations/{installation_id}/access_tokens",
                headers={
                    "X-GitHub-Api-Version": "2022-11-28",
                    "Accept": "application/vnd.github+json",
                    "Authorization": f"Bearer {self._generate_jwt()}"
                }
            )
            if res.status_code != 201:
                raise Exception(f"Unable to generate app installation token: {res.status_code}: {res.content}")
            self.__class__.INSTALLATION_ID_TOKENS[installation_id] = res.json().get("token")

        return self.__class__.INSTALLATION_ID_TOKENS[installation_id]

    def _generate_jwt(self):
        """Generate app installation JWT"""
        pem = self._private_key
        if not pem:
            return None

        payload = {
            # Issued at time
            'iat': int(time.time()),
            # JWT expiration time (10 minutes maximum)
            'exp': int(time.time()) + 600,
            # GitHub App's identifier
            'iss': self.github_app_id
        }

        private_key = serialization.load_pem_private_key(
            pem, password=None, backend=default_backend()
        )
        return jwt.encode(payload, private_key, algorithm="RS256")

    def _get_app_metadata(self) -> dict:
        """Return app metadata"""
        res = requests.get(
            f"{self._api_url}/app",
            headers={
                "X-GitHub-Api-Version": "2022-11-28",
                "Accept": "application/json",
                "Authorization": f"Bearer {self._generate_jwt()}"
            }
        )
        if res.status_code == 200:
            return res.json()
        raise Exception(f"Could not obtain app metadata: {res.status_code}: {res.content}")

    def get_app_installation_url(self):
        """Generate app installation URL"""
        metadata = self._get_app_metadata()
        return f"{metadata.get('html_url')}/installations/new"

    def get_github_app_installation_id(self, namespace: 'terrareg.models.Namespace') -> Union[int, None]:
        """Obtain a github org/user's app installation status"""
        # Determine URL of installation based on namespace type
        url = self._api_url
        if namespace.namespace_type is terrareg.namespace_type.NamespaceType.GITHUB_ORGANISATION:
            url += f"/orgs/{namespace.name}/installation"
        elif namespace.namespace_type is terrareg.namespace_type.NamespaceType.GITHUB_USER:
            url += f"/users/{namespace.name}/installation"
        else:
            # Otherwise is namespace is not based on a github user/org, return None
            return None

        res = requests.get(
            url,
            headers={
                "X-GitHub-Api-Version": "2022-11-28",
                "Accept": "application/json",
                "Authorization": f"Bearer {self._generate_jwt()}"
            }
        )

        if res.status_code == 404:
            return None
        elif res.status_code == 200:
            return res.json().get("id")
        else:
            raise Exception(f"Unrecognised response code from github installation check: {res.status_code}")

    def _is_entity_org_or_user(self, identity: str, access_token: str):
        """Determine if an entity is a user or organisation"""
        res = requests.get(
            f"{self._api_url}/users/{identity}",
            headers={
                "X-GitHub-Api-Version": "2022-11-28",
                "Accept": "application/json",
                "Authorization": f"Bearer {access_token}"
            }
        )
        if res.status_code != 200:
            return None
        if type_ := res.json().get("type"):
            if type_ == "User":
                return terrareg.namespace_type.NamespaceType.GITHUB_USER
            elif type_ == "Organization":
                return terrareg.namespace_type.NamespaceType.GITHUB_ORGANISATION

    def refresh_namespace_repositories(self, namespace: 'terrareg.models.Namespace') -> None:
        """Refresh list of repositories for namespace using default installation"""
        access_token = self._get_default_access_token()
        if not access_token:
            raise Exception("Provider source default access token/installation has not been configured")

        type_ = self._is_entity_org_or_user(namespace.name, access_token=access_token)
        if not type_:
            raise Exception("Could not find namespace entity in provider")

        url = f"{self._api_url}/{'orgs' if type_ is terrareg.namespace_type.NamespaceType.GITHUB_ORGANISATION else 'users'}/{namespace.name}/repos"

        page = 1
        while True:
            res = requests.get(
                url,
                params={
                    "sort": "created",
                    "direction": "desc",
                    "per_page": "100",
                    "page": str(page)
                },
                headers={
                    "X-GitHub-Api-Version": "2022-11-28",
                    "Accept": "application/vnd.github+json",
                    "Authorization": f"Bearer {access_token}"
                }
            )
            if res.status_code != 200:
                print(f"Invalid response code from github: {res.status_code}")
                return

            results = res.json()

            for repository in results:
                self._add_repository(repository_metadat=repository)

            if len(results) < 100:
                break

            page += 1
