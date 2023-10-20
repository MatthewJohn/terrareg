
from typing import Dict, Union, List
from urllib.parse import parse_qs

import requests

from terrareg.errors import InvalidProviderSourceConfigError
from .base import BaseProviderSource
import terrareg.provider_source_type
import terrareg.repository_model
import terrareg.provider_version_model
import terrareg.provider_model
import terrareg.provider_source.repository_release_metadata


class GithubProviderSource(BaseProviderSource):

    TYPE = terrareg.provider_source_type.ProviderSourceType.GITHUB

    @classmethod
    def generate_db_config_from_source_config(cls, config: Dict[str, str]) -> Dict[str, Union[str, bool]]:
        """Generate DB config from config"""
        db_config = {}
        for required_attr in ["base_url", "api_url", "client_id", "client_secret", "login_button_text"]:
            if not (val := config.get(required_attr)) or not isinstance(val, str):
                raise InvalidProviderSourceConfigError(f"Missing required Github provider source config: {required_attr}")

            db_config[required_attr] = val

        for bool_attr in ["auto_generate_github_organisation_namespaces"]:
            val = config.get(bool_attr)
            if not isinstance(val, bool):
                raise InvalidProviderSourceConfigError(f"Missing required Github provider source config: {bool_attr}")
            db_config[bool_attr] = val

        return db_config

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
        return self._config["login_button_text"]

    def is_enabled(self) -> bool:
        """Whether github authentication is enabled"""
        return bool(self._client_id and self._client_secret and self._base_url and self._api_url)

    def get_login_redirect_url(self) -> str:
        """Generate login redirect URL"""
        return f"{self._base_url}/login/oauth/authorize?client_id={self._client_id}"

    def get_access_token(self, code: str) -> Union[None, str]:
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
                if (not (repo_id := repository.get("id")) or
                        not (repo_name := repository.get("name")) or
                        not (owner_name := repository.get("owner", {}).get("login")) or
                        not (clone_url := repository.get("clone_url"))):
                    continue

                terrareg.repository_model.Repository.create(
                    provider_source=self,
                    provider_id=repo_id,
                    name=repo_name,
                    description=repository.get("description"),
                    owner=owner_name,
                    clone_url=clone_url,
                    authentication_key=access_token
                )

            if len(results) < 100:
                break

            page += 1

    def get_new_releases(self, provider: 'terrareg.provider_model.Provider', access_token: str) -> List['terrareg.provider_source.repository_release_metadata.RepositoryReleaseMetadata']:
        """Obtain all repository releases that aren't associated with a pre-existing release"""
        repository = provider.repository
        page = 1
        releases = []
        obtain_results = True

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
                        not (tag_name := release.get("tag_name"))):
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
                        name=release_name, tag=tag_name, provider_id=release_id,
                        release_artifacts=release_artifacts
                    )
                )

            if len(results) < 100:
                obtain_results = False
            
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
                             artifact_metadata: 'terrareg.provider_source.repository_release_metadata.ReleaseArtifactMetadata',
                             release_metadata: 'terrareg.provider_source.repository_release_metadata.RepositoryReleaseMetadata',
                             repository: 'terrareg.repository_model.Repository',
                             access_token: str):
        """Return release artifact file content"""
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
