# Changelog

## [2.75.2](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.75.1...v2.75.2) (2023-09-09)


### Bug Fixes

* Fix module version publishing when creating a module version (either by hooks or create/upload API endpoints) when AUTO_PUBLISH_MODULE_VERSIONS is enabled. ([29a04db](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/29a04dbf9231da92591f63fd7ea90f39ac1c8b17)), closes [#425](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/425)

## [2.75.1](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.75.0...v2.75.1) (2023-09-07)


### Bug Fixes

* Fix github webhook when using API key for Github secreet to use correct header for SHA256 encoded header. ([68043de](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/68043de1e1745e4e810f5ba7ad08b424765ef940)), closes [#424](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/424)

# [2.75.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.74.0...v2.75.0) (2023-08-21)


### Bug Fixes

* Avoid error when passing None value to updat_git_tag_format, which threw an exception in urllib quote method ([22a36aa](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/22a36aa98064ce1098e0c8e235f2f57a98f036ba)), closes [#412](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/412)
* Handle KeyError exceptions thrown in update_git_tag_format when an unmatched placeholder is present ([873bf47](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/873bf479849c61da93a111fce9964bab0812bbae)), closes [#412](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/412)


### Features

* Add handling of git tags with non-semantic versioning in git hooks. ([c0f4f7e](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/c0f4f7e0d1d5a30e1d6287fc8176358887696cb7)), closes [#412](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/412)
* Add new API endpoint for importing/indexing module versions by either version or git tag. ([89d734e](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/89d734e3ff88ccb47c070748b3ff93d84e2fe6e3)), closes [#412](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/412) [#416](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/416)
* Fix providing module provider git tag format with placeholders to allow tags that do not use a full semantic version. ([cff3da6](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/cff3da6a692a24673218747f3820f99cf97e490c)), closes [#412](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/412)

# [2.74.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.73.1...v2.74.0) (2023-08-20)


### Bug Fixes

* Replace link for "show resources depdencies" with buton to make it more obvious ([827d646](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/827d646d5ae88d234041f7172a7f1fb9b36974a4)), closes [#411](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/411)


### Features

* Add ability to delete module provider redirects in module provider settings in UI ([9441429](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/9441429c102de52600a4785bb5727423558813c8)), closes [#274](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/274)
* Add ability to delete namespace from UI ([1fb72a4](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/1fb72a4c55712ce1e43df502411bca86d9bd945c)), closes [#268](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/268)
* Add API endpoint to list module provider redirects and endpoint to delete module provider redirect. ([5473a4e](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/5473a4eb0d34fb7dd0422c84e759bb10acc95b4f)), closes [#274](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/274)
* Add arguments to module provider settings API to change module's name, provider and namespace. ([9215b22](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/9215b225a785c6bf01536149df82c39c7d0873ed)), closes [#274](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/274)
* Add configuration to determine how many days worth of module access to check whilst determining if redirect can be deleted ([7646546](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/76465460107bdbbb1f6c48f0a3c03d7ba8ce63b7)), closes [#274](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/274)
* Add delete method to module provider redirect and functionality to check if redirect is still in use by analytics tokens ([a65f5a0](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/a65f5a08338f21facbcdfb719a9dc904d694d513)), closes [#274](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/274)
* Add DELETE method to namespace details API to support deletion of namespace ([93bf63d](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/93bf63d28d03ada336e9f2461db6167a6dad2e4c)), closes [#268](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/268)
* Extract all Terraform provider versions and recursive module details during extraction. ([7af6fd0](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/7af6fd03dabad9075b8d441ece208f14f8d69a37)), closes [#237](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/237)
* Record namespace name, module name and provider name used in module download URL in analytics to be able to determine redirect usage ([3e87227](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/3e872274da88af8ba1c657bc20d945d5313199fc)), closes [#274](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/274)
* **ui:** Add form in module provider settings tab to change namespace, module name and provider ([4d6275c](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/4d6275c38f176782e36224cd82af80e99459ebce)), closes [#274](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/274)

## [2.73.1](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.73.0...v2.73.1) (2023-08-18)


### Bug Fixes

* Fix typo in error response in login page when SSO response is invalid ([9463749](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/94637498ad7043a2bcae986a7131b27dc9fda471))

# [2.73.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.72.0...v2.73.0) (2023-08-08)


### Features

* Add support for using "waitress" server in docker containers. ([ca01c84](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/ca01c84607123bb28b3186cb31153280037d896b)), closes [#408](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/408)

# [2.72.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.71.3...v2.72.0) (2023-08-01)


### Bug Fixes

* Only perfrom secondary ordering audit events by timestamp descending if the primary sort is not based on the timestamp ([d7f4a71](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/d7f4a717b9bb30595fcb69d026aa242c113ab747)), closes [#269](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/269)


### Features

* Allow renaming of namespace and changing of display name. ([ddfa200](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/ddfa2004d6773989c60666cf46f53f926861b610)), closes [#269](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/269)

## [2.71.3](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.71.2...v2.71.3) (2023-07-08)


### Bug Fixes

* Delete module provider data directory on deletion on module provider ([b44b560](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/b44b56075efe6847ca41090194aa01772eeb1451)), closes [#406](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/406)
* Handle deletion of module version directory when there are pre-existing files in the module version directory that aren't managed by Terrareg ([4d7a0c8](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/4d7a0c8996711574db8687cf2ebe00707a38d0c1)), closes [#406](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/406)
* Remove module version data files and directory when removing the module version ([257c6d7](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/257c6d7185ca58a9a13d7766e91eb3e29a716cb4)), closes [#406](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/406)

## [2.71.2](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.71.1...v2.71.2) (2023-06-23)


### Bug Fixes

* Fix bug where CWD of main process is changing during module extraction. ([fcf6325](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/fcf6325423d32a2e31107c25a0f4376f66a0c46c)), closes [#404](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/404)

## [2.71.1](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.71.0...v2.71.1) (2023-06-21)


### Bug Fixes

* Add handling of non-existent example/submodule in example/submodule details/README API endpoints and Example file list API endpoint ([62ea089](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/62ea08929f39d2bc20571e282a1777fcb55be855)), closes [#387](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/387)
* **ui:** Handle non-existent submodule/example in module provider page and showing a warning ([238ffa8](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/238ffa8704f49f84b19cd7a8af073dcdcb598140)), closes [#387](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/387)

# [2.71.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.70.1...v2.71.0) (2023-06-20)


### Features

* Add configuration to add additional file extensions that are extracted from examples to display in UI ([cf142af](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/cf142af07022c21744a357742027aba2a97bf87e)), closes [#303](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/303)
* Include additional syntax highlighting for javascript (json), bash, batch, pl/sql, powershell, python and Dockerfile ([c722c41](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/c722c414288f9093db52ce113b0fec77ebda33c9)), closes [#303](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/303)
* Update default config to extract .sh, .tfvars and .json files in module examples ([151f8e3](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/151f8e3b68fe38308afbc373c1f7d048958bd7c1)), closes [#303](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/303)

## [2.70.1](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.70.0...v2.70.1) (2023-06-18)


### Bug Fixes

* Allow line break (br) elements in sanitised HTML ([0d21e9b](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/0d21e9b58e603bee554d8e223394b15fb2a23a1a)), closes [#402](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/402)
* Fix erronious links in converted README markdown. ([21e1925](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/21e1925c7ce7033596b118c946fa151953fcb6ef)), closes [#402](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/402)
* Update descriptions in input/output tables in module provider page to respect line breaks in descriptions. ([0cf0b86](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/0cf0b86a57e964c89c61e8b54d7e90bd2c8ed418)), closes [#402](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/402)

# [2.70.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.69.0...v2.70.0) (2023-06-14)


### Features

* Add Cherry Dark Mode Theme ([31a5b88](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/31a5b88f385dcdccfc37d72f557d7a5a6926ade9))

# [2.69.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.68.3...v2.69.0) (2023-06-05)


### Features

* Add support for two character namespace/module names. ([3a32e73](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/3a32e736e9d63bfa80835200d41901a436f50fd7)), closes [#397](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/397)

## [2.68.3](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.68.2...v2.68.3) (2023-06-02)


### Bug Fixes

* Fix handling of terraform provider block with content before the block in the file. ([40b258c](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/40b258c93fb3d89ef38845ca263fede1518d0a2e)), closes [#395](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/395)

## [2.68.2](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.68.1...v2.68.2) (2023-06-02)


### Bug Fixes

* Attempt to find find terraform files with backend configuration and create override file during extraction. ([be76ec8](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/be76ec8b074ee52cb00124077a599fd70a8ff984)), closes [#395](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/395)

## [2.68.1](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.68.0...v2.68.1) (2023-06-01)


### Bug Fixes

* Disable backend when initialising terraform ([8198b3c](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/8198b3c58b86312839648d6a4dc191fdfa2c04b1)), closes [#395](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/395)

# [2.68.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.67.0...v2.68.0) (2023-05-09)


### Features

* Add target_terraform_version arguments to terrareg module provider/module version details API endpoints, which add version compatibility to the response ([975bbdc](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/975bbdce8a03a358f1546303893de2f1282e9fff)), closes [#295](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/295)
* Add Terraform constraint version to user preferences and update module search version constraint to get/set the user preference value ([b6a9c62](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/b6a9c62658b1137c1e099cf9ad91ca31d5d7eb9c)), closes [#295](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/295)
* Display Terraform compatibility result on module provider page ([368545d](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/368545d2d1b48d6726a769d7a2c0856e098e7fb2)), closes [#295](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/295)
* Show module compatibility with selected Terraform version in module search ([67c2f84](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/67c2f843c305b7432573d7f5560017cc20cc4c23)), closes [#295](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/295)

# [2.67.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.66.4...v2.67.0) (2023-04-25)


### Bug Fixes

* Populate "dependencies" attribute in API response, with a list of external module dependencies ([05553e8](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/05553e8583f6a823abce542381d1994f232f296f)), closes [#386](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/386)
* Remove html sanitisation of provider version to remove HTML entities from API response. ([1832b04](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/1832b04c6ad8b65b439ca6879e19f698d881fb8b)), closes [#385](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/385)


### Features

* Add tab to module provider page detailing modules called by current module ([f727a48](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/f727a486d6a5a2f08306a6acf23d178a24ec734e)), closes [#139](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/139)
* Display module calls in submodule/example pages ([17f1649](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/17f1649a4ab4341fada40c1e73f6c3fb4d497f68)), closes [#139](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/139)

## [2.66.4](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.66.3...v2.66.4) (2023-04-24)


### Bug Fixes

* Stop terraform graph failures from causing module extraction to fail ([0db5c8e](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/0db5c8ea4e77fe97de0348fb02919e19d99a003c)), closes [#380](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/380)

## [2.66.3](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.66.2...v2.66.3) (2023-04-21)


### Bug Fixes

* Hide 'Settings' tab from module provider page when user does not have write permission to the namespace ([5c9936b](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/5c9936b58198fa3b512670e7c07cd4d43c071899)), closes [#369](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/369)

## [2.66.2](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.66.1...v2.66.2) (2023-04-03)


### Bug Fixes

* Allow overriding of OpenID connect requested scopes. ([d8a1a9c](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/d8a1a9c16fef5a62bbd28bc8b2279a48809ac8a9)), closes [#365](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/365)

## [2.66.1](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.66.0...v2.66.1) (2023-03-28)


### Bug Fixes

* Reload current page when any user preferences are changed ([e911aa8](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/e911aa878897526d06dde4c8ad0de96f77d08d7b)), closes [#376](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/376)
* Reload current page when preferences are saved after the theme is changed ([de5510a](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/de5510aa1c61c59124eef6cd62ec9c5d54683530)), closes [#376](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/376)
* Update cookie set when selecting theme to specify a path, rather than the default, which is the user's current path ([6392aad](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/6392aad7bff5fe537c5b0df3027f138d8062f564)), closes [#375](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/375)

# [2.66.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.65.0...v2.66.0) (2023-03-27)


### Bug Fixes

* Fix bug where hostname in example replacement contains port of request. ([5690487](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/5690487ec1b1d9f6186965e7975f2ed42fd1db41)), closes [#372](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/372) [#347](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/347) [#348](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/348)


### Features

* Add config to disable analytics tokens entirely ([968cccb](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/968cccba562ad4bd6c0132baea525c580d348d2d)), closes [#372](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/372)
* Disable analytics tab in module provider page when analytics are disabled ([3da61e7](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/3da61e77429f0ac31ea28b1a388b1de37c679e81)), closes [#372](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/372)
* Disable analytics token in usage builder when analaytics is disabled globally ([b7a437d](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/b7a437d4c629a0f06e8aebcef69a5653c80592fd)), closes [#372](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/372)
* Update example usage and Terraform replaced in README/example files to no longer show analytics token is DISABLE_ANALYTICS is set ([9d09253](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/9d09253f6d97158ed9e27219e7c5e9428fb837d9)), closes [#372](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/372)
* Update module version download to ignore analytics tokens if DISABLE_ANALYTICS is set ([fa7e4ee](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/fa7e4ee37765482b3ab3b2401c8302e5579d4604)), closes [#372](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/372)

# [2.65.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.64.0...v2.65.0) (2023-03-25)


### Bug Fixes

* Add initial sorting of security issues by filename. This is usually the default outcome of the tfsec output, but test data did not include it. ([6eaf023](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/6eaf0232e117e02a8a95f1690020e7edc8258a20)), closes [#371](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/371)


### Features

* Show highest severity security issue in security issue badge. ([232b1a9](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/232b1a9b29b4b0e15fcf0e602a7e6d6b4178c90e)), closes [#320](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/320)
* Update security issues label to show breakdown of severities ([e3392e0](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/e3392e0acb7324566f005b9378be1d6a4fcc34f0)), closes [#320](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/320)

# [2.64.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.63.0...v2.64.0) (2023-03-24)


### Features

* Add support for external images in README and custom markdown tabs. ([4ed384f](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/4ed384f40a96b512e138ba7e50986a3506f2b9ff)), closes [#370](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/370)

# [2.63.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.62.3...v2.63.0) (2023-03-23)


### Bug Fixes

* Perform HTML sanitisation after conversion of markdown to HTML ([1132e24](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/1132e242562d9666171dc9c260c6ad68182bcc6b)), closes [#300](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/300)


### Features

* Add support for links to heading anchors in README/markdown tabs ([d2d6ea4](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/d2d6ea48ca3ad1b38c7a2dbb1252bd1e69c2fc86)), closes [#300](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/300)
* Automatically select tab containing anchor link in module provider page ([014dc9b](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/014dc9ba340bb1561116fe60e2c5430ad8b0d70f)), closes [#300](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/300)

## [2.62.3](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.62.2...v2.62.3) (2023-03-22)


### Bug Fixes

* Fix error when performing terraform init, which tries to use non-existent .terraform.d/plugin-cache directory. ([19213ee](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/19213ee4c23758316a83f0dc4657dd96d32d5d6f)), closes [#364](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/364)

## [2.62.2](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.62.1...v2.62.2) (2023-03-22)


### Bug Fixes

* Add catching of errors due to invalid placeholder in repo base URL. ([0119b36](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/0119b369406fa4e85650664e4aa748463ecb0dbb)), closes [#360](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/360)
* Add handling of invalid placeholder in repo browse URL template. ([e4d9cd1](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/e4d9cd17e0f2e6dddc53a801f6e246ad66452360)), closes [#360](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/360)
* Add validation of port in git clone URL, ensuring that if it exists, it's a valid port. ([dbcf808](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/dbcf808da5a296cfc787589950240e64909e6e2f)), closes [#360](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/360)
* Catch errors caused by invalid template placeholders in git clone URL. ([d522b23](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/d522b239e2b078504ef4a6a0a814d28adde6ac6b)), closes [#360](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/360)
* Update examples of SSH git clone URLs to use forward slash delimeter after domain, rathern than colon ([74c055d](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/74c055de7ea7fae6a922bab9edaddca95de05daf)), closes [#360](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/360)

## [2.62.1](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.62.0...v2.62.1) (2023-03-22)


### Bug Fixes

* Fix tfswitch when using Dockerfile ([7fcda64](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/7fcda64e3f345f47481059e5080a818406725c6f)), closes [#356](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/356)
* Show command output when command errors with debug mode enables ([82d7197](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/82d7197cb69e2c470157e2e07e4cd7121de48284)), closes [#356](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/356)
* Update module extraction methods to return exception details when DEBUG is enabled ([3c2dbdb](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/3c2dbdb58235e6d70b604734e6981a7be9763608)), closes [#356](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/356)

# [2.62.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.61.1...v2.62.0) (2023-03-21)


### Features

* Add example syntax highlighting ([c63ef25](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/c63ef252f5d93196dbead899befe4525327ba5c8)), closes [#304](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/304)
* Handle syntax highlighting of READMEs ([f49455d](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/f49455dc4378198ffb2eb7f51469a2545e201f95)), closes [#304](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/304)

## [2.61.1](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.61.0...v2.61.1) (2023-03-21)


### Bug Fixes

* Update initial setup to display commands to use API key when uploading/publishing module, if the API keys are set ([454e036](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/454e036c1c217ec02bcb0aa2b6005dce8327c73d)), closes [#351](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/351)
* Update upload_module script to support environment variables for setting upload/publish API keys ([8861214](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/8861214f3fe03f01af5b67af161831ff12a6709e)), closes [#351](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/351)

# [2.61.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.60.2...v2.61.0) (2023-03-17)


### Bug Fixes

* Add new 'PUBLIC_URL' config which is the source of protocol, domain and port used for end-user communication ([79baa0f](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/79baa0f5a8178623a66dc25cd40a3c4eb7a09302)), closes [#347](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/347)
* Add port to usage example if port is a non-standard port ([5c3c024](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/5c3c02429ed50e80d226f216d24f4c93a2ce321f)), closes [#348](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/348)


### Features

* Add support for using /modules/ endpoint in terraform, allowing the use of the registry without HTTPS ([4a3e3a2](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/4a3e3a2454c5c698c2c517c9f22f9b197fdd5af8)), closes [#347](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/347)
* Update usage example in module provider page to display "source" url with http download URL, if https is not being used ([b06d562](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/b06d56252cc3f925f0f2ff152ac63490baf946f7)), closes [#347](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/347)
* Use new http-accessible source URL in example files and README terraform. ([4a2f0b4](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/4a2f0b484fa65eceada854d77ca129e282b10215)), closes [#347](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/347) [#349](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/349)

## [2.60.2](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.60.1...v2.60.2) (2023-03-15)


### Bug Fixes

* Add default value to git tag format field in module provider create page. ([fc516b9](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/fc516b9d495e81a026073f86c45cf441f8dd6409)), closes [#311](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/311)

## [2.60.1](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.60.0...v2.60.1) (2023-03-14)


### Bug Fixes

* Update namespace module list endpoint to return empty module list rather than 404 when a namespace has no modules ([6ca3441](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/6ca3441f309c52d89ff8fe0f36e1c6da851cc8ea)), closes [#328](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/328)
* Update namespace page to display error about no modules when a namespace exists, rather than an eror that the namespace does not exist. ([ec0b3c9](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/ec0b3c9d362e8c3fa39a10d33f87199546270a8f)), closes [#328](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/328)

# [2.60.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.59.5...v2.60.0) (2023-03-14)


### Features

* Add support for 'latest' module provider download endpoint ([18f14e8](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/18f14e86cc6aa091f573f446a9d1f5428f14302d)), closes [#322](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/322)

## [2.59.5](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.59.4...v2.59.5) (2023-03-14)


### Bug Fixes

* Add steps to initial setup to disable AUTO_CREATE_NAMESPACE and AUTO_CREATE_MODULE_PROVIDER ([f18398b](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/f18398b78b0485cbe02e01a19a6bf9b400cfafd3)), closes [#337](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/337)

## [2.59.4](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.59.3...v2.59.4) (2023-03-14)


### Bug Fixes

* Disable autocreation of namespaces, modules and module providers in module provider details API endpoint ([e015a7f](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/e015a7f92248f533f436aeed933f49f05a44afb1)), closes [#338](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/338)

## [2.59.3](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.59.2...v2.59.3) (2023-03-13)


### Bug Fixes

* Update unauthenticated/unpriviledged access to module upload/publish endpoints when API keys are not set and access controls are enabled ([0afc6a8](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/0afc6a8e1aee289c9705f8f6575465c47935a907)), closes [#339](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/339)

## [2.59.2](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.59.1...v2.59.2) (2023-03-13)


### Bug Fixes

* Fallback to use openid connect user's email address, if username is not available ([c7e057a](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/c7e057a713ca12a7b5fd31a0641c4190d29ab176)), closes [#276](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/276)
* Fix authentication using OpenID connect with Azure ([29bcb1d](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/29bcb1d5365cdebb1367be27c5e62cf75ec15ff3)), closes [#276](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/276)

## [2.59.1](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.59.0...v2.59.1) (2023-03-11)


### Bug Fixes

* Disable validation of namespace name, allowing namespaces that have an invalid name to still be used. ([8443861](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/8443861bfe45bb0460a05c55b73313adf8980a6b)), closes [#330](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/330)
* Disallow namespace names with double underscores ([fe694ef](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/fe694ef7db1359817b708cd999b58a873bacf949)), closes [#330](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/330)
* Improve error message when namespace name is invalid, providing details of what is required for a valid name. ([400f6d1](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/400f6d11227b7e1339ba23f17a1a70f77e30a509)), closes [#330](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/330)
* Update namespace endpoint to return 400 on invalid namespace name/display name and duplicate names ([85e0010](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/85e00108e1bc11dc3fe774f44f244e4b999f5145)), closes [#326](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/326)

# [2.59.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.58.1...v2.59.0) (2023-03-10)


### Features

* Allow 'latest' in URLs instead of a specific version number. ([c347b03](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/c347b03ac72ac91ee8dfdd7fe28b711724fa9581)), closes [#333](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/333)

## [2.58.1](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.58.0...v2.58.1) (2023-03-09)


### Bug Fixes

* Fix text color of text inputs in pulse theme ([81f3982](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/81f398218463e022227fcf312d70830c6cd80fef)), closes [#336](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/336)
* Speed up page load and reduce content shift by rendering theme CSS path server-side ([6c96faf](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/6c96faff745bb1078bddc1961378110fe34d63d7)), closes [#336](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/336)

# [2.58.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.57.0...v2.58.0) (2023-02-14)


### Features

* Add caching of terraform providers in terrareg container. ([bf050d7](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/bf050d74f79cbd36aaeedae3a934f59318af1180)), closes [#316](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/316)

# [2.57.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.56.0...v2.57.0) (2023-01-31)


### Features

* Add sentry integration. ([6ea7e82](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/6ea7e825b25d23815411723bd7a6a209b8916b0a)), closes [#205](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/205)

# [2.56.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.55.0...v2.56.0) (2023-01-30)


### Features

* Add provider logos for hashicorp products: vault, nomad, vagrant and consul ([81c9d8a](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/81c9d8ac4eca6f2de07b9eb93d5256162bb6f661))

# [2.55.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.54.1...v2.55.0) (2023-01-30)


### Bug Fixes

* Add CSFR token validation to namespace create API ([bc68795](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/bc68795334f81dc60270a55057f0772e62541fb9)), closes [#325](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/325)
* Fix exception thrown when attempting to create a namespace with a None name or without provide a name attribute ([4574778](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/457477845d01dbffee662e566078780c9700ed9c)), closes [#327](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/327)
* Fix mis-aligned tags in create module provider page in description of templating for clone/browse URL templates ([83db99f](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/83db99f5d55f9ef5aa503ebfce652eacefcfcb1f)), closes [#329](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/329)


### Features

* Add support to add 'Display name' to namespace during creation. ([4ac011e](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/4ac011eb466890d0d9186ad0294d17b6e8f699aa)), closes [#294](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/294)
* Display namespace "display name" in namespace list and module provider page breadcrumbs. ([8406815](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/8406815a47b393000ed53d74bac90bc4df9dbcb6)), closes [#294](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/294)

## [2.54.1](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.54.0...v2.54.1) (2023-01-26)


### Bug Fixes

* Always show example cost to 2 decimal places ([9b73998](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/9b739982defcd9e85bbb9aab87e11fa77d452634)), closes [#324](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/324)
* Fix bug with resource cost calculations where each iteration of a module in a for_each will overwrite the previous cost. ([a258775](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/a2587757a8f6dfe3bc47efad67882d07bfbf746b)), closes [#323](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/323)

# [2.54.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.53.2...v2.54.0) (2023-01-23)


### Bug Fixes

* Update example usage to show pinned version for non-latest module versions ([c0b0a2f](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/c0b0a2f80cb99763cf89aae643156e82a18354aa)), closes [#301](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/301)


### Features

* Add comment above example file Terraform version string when using a beta/unpublished/non-latest version of a module ([51dde84](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/51dde844fd2ff3df4b38ae37b97eb1908e9c59f9)), closes [#317](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/317)
* Update usage example to include Terraform comment above version if using a beta/unpublished/non-latest version ([8dc51b0](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/8dc51b0801f67a38aba7c792f4a6ba6803317a57)), closes [#317](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/317)

## [2.53.2](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.53.1...v2.53.2) (2023-01-23)


### Bug Fixes

* Update links on module search results, module list and homepage module to direct to module provider page without a specific version ([9e2675d](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/9e2675d4f1e8bd996a149a6fb1e281bec7d3b676)), closes [#318](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/318)

## [2.53.1](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.53.0...v2.53.1) (2023-01-22)


### Bug Fixes

* Add warning to module provider page when new features have been added that require a module version to be re-indexed ([416998c](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/416998cf0ec663ae3cdf04e865d541430c96453b)), closes [#314](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/314)

# [2.53.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.52.1...v2.53.0) (2023-01-22)


### Features

* Add ability to configure mode when re-uploaded a previously uploaded module version. ([e7c9dc9](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/e7c9dc90779eff18e9509e39423471adc3961809)), closes [#299](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/299)
* Retain publish state when re-indexing module version ([3fe6daa](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/3fe6daa7568b353fbefbe0819585b31d45f24dab)), closes [#299](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/299)

## [2.52.1](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.52.0...v2.52.1) (2023-01-22)


### Bug Fixes

* Fix display of heredocs in example file content. ([a6358ba](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/a6358bae2130b245440fb271e6a75a1326f777ec)), closes [#310](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/310)

# [2.52.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.51.0...v2.52.0) (2023-01-21)


### Features

* Add Terraform auth tokens that ignore the analytics token checks. ([1ba2dbc](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/1ba2dbc79750c9882cbc5595a59e9cf2e79ede6c)), closes [#244](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/244)

# [2.51.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.50.0...v2.51.0) (2023-01-21)


### Features

* Add alternative themes with user preferences option to select theme. ([1be14ee](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/1be14ee7f9a0472b4c6621a4869ea3ac5cc1c6ce)), closes [#309](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/309)

# [2.50.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.49.1...v2.50.0) (2023-01-20)


### Bug Fixes

* Add '--raw' and '--clean=false' to inframap to retain nodes that are being removed ([a0ee490](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/a0ee49059d397b6436ad7248e872cfb8b0e108c3)), closes [#286](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/286)
* Add draft Terraform init command run before running inframap to make submodules available ([375f6ea](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/375f6ea7cf5c194f53a735df316fd25a7d2625be)), closes [#286](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/286)
* Add red border to resources with costs in resource graph. ([7d1f3e9](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/7d1f3e9a56c14ea543c39172b40e8dae43e2df67)), closes [#286](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/286)
* Implement management of terraformrc file to disable analytics generated from terraform during module extraction ([72e5970](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/72e597087edb93134aa0e5200980cdb99a40d918)), closes [#286](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/286)
* Improve layout of graph by spreading out modules based on number of child resources. This is helped by disabling tiling ([f2c7036](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/f2c7036bb848490df23349e4b191a17ac6b82851)), closes [#286](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/286)
* Update initial graph pan based on size of cy div height/width ([bab8f93](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/bab8f93e1ddaf32eabef8838a5f413a7bd9d3cc8)), closes [#286](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/286)
* Update ModuleExtractor to use terraform graph and terraform-graph-beautifier to generate graph data and JSON ([afa4f65](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/afa4f65e6866956348a73645ac1bbda973f9b8e3)), closes [#286](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/286)


### Features

* Add graph options to show full resource and module names ([796db76](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/796db7682b1d9c3f5e14a16368d92b7f45fa4d70)), closes [#286](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/286)
* Add page to view resource graph ([fe4129c](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/fe4129cacff8fffd65a6e1697a1cfe395880afbc)), closes [#286](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/286)
* Add yearly costs to resources in graph view, where available ([f7689ba](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/f7689ba606d8bd06977a568ee6f7e39a861752dd)), closes [#286](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/286)


### Reverts

* Revert "ci: Update build_wheel step to use test docker image" ([93540a2](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/93540a2fb646baabf7ecf2173d06a4dbd5cbd689))
* Revert "ci: Update wheel build to use python3/pip3, since the test container is based Ubuntu" ([c21b497](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/c21b49791af13722f9fad57270977a6ba13aaf3e))
* Revert "ci: Add installation wheel package for building wheel inside test container" ([066ad26](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/066ad26c349037b91a897cde8e4169de6dcf318f))

## [2.49.1](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.49.0...v2.49.1) (2023-01-14)


### Bug Fixes

* Add 'location' argument to all request parsers that expect arguments via query string ([36220b1](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/36220b129a869b742ca85b7359384dfba3ac43dc)), closes [#283](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/283)

# [2.49.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.48.0...v2.49.0) (2023-01-13)


### Bug Fixes

* Do not show Terraform example usage panel in module provider page if module version has not been published. ([1fdc0ae](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/1fdc0ae749bb79f4f8ade7e30910ae25ccda7a55)), closes [#272](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/272)


### Features

* Add Terraform version constraint to module provider page. ([5d1e382](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/5d1e382de8b8778a757f2587e10ec3fc1e356769)), closes [#272](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/272)

# [2.48.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.47.3...v2.48.0) (2022-12-31)


### Features

* Add ability to provide custom links on module provider pages ([3cfdd2b](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/3cfdd2be03a2d56d4918a8f4ddacc4ba6413576e)), closes [#275](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/275)

## [2.47.3](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.47.2...v2.47.3) (2022-12-31)


### Bug Fixes

* Add warning to login page when no authentication methods are available ([170020d](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/170020db3d2f928baa6b1a3e0915be2a5fbdfc0a)), closes [#259](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/259)
* Hide admin authentication login when admin token has not been configured. ([153def6](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/153def61679de1976ec04683e24a8e8ccf169bc1)), closes [#259](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/259)
* Update initial setup process to automatically redirect user back to initial setup after creating a module ([7ae2608](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/7ae2608ba9914819f3c45c9b689f80ce55bd5088)), closes [#258](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/258)
* Update initial setup process to automatically redirect user to initial-setup after creating a namespace ([05711ee](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/05711eeac344b97cce1271a3377b2f96930d0fd0)), closes [#258](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/258)
* Update initial setup to check for any forms of authentication configured, rather than requiring admin token ([66c43f2](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/66c43f28c0e137d49cf589bf5c2f08f73090f46a)), closes [#258](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/258)

## [2.47.2](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.47.1...v2.47.2) (2022-12-31)


### Bug Fixes

* Hide version text on module provider, by default and increase padding of labels ([de9e477](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/de9e477d5403334af85b95f9fe9df03938048879)), closes [#282](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/282)

## [2.47.1](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.47.0...v2.47.1) (2022-12-16)


### Bug Fixes

* Fix display of < and > symbols in provider requirement version fields ([b10e9a3](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/b10e9a3cfc9e1ca7a7cd8cd0ec30600881077e66)), closes [#275](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/275)

# [2.47.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.46.2...v2.47.0) (2022-12-16)


### Features

* Add audit event logging and audit history page ([9ed7222](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/9ed72226833a4546c98886cb3355ce356f8b53fc)), closes [#253](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/253)
* Add audit events for create/delete namespaces, module providers and module versions ([97ed1a0](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/97ed1a0ca6675f268ef9da9926c00b4ebd9da276)), closes [#253](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/253)
* Add audit events for logins ([b077ca5](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/b077ca5fee06a71d3e7f4aff2fbc1fed927e59ab)), closes [#253](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/253)
* Add audit events for modifications to module provider configuration. ([67826e8](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/67826e8cda81330b9a127141fa8ad45610ced9b5)), closes [#253](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/253)
* Create audit history page and API endpoint to return audit history ([0ce7156](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/0ce71562cf98d199abd19f8cec3fa7a5b67024bc)), closes [#253](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/253)
* Implement audit log querying, limit+offset, sorting and ordering in API and datatable ([5132e46](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/5132e4615a1d272c345348d2f2ccc9bd3332cd78)), closes [#253](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/253)

## [2.46.2](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.46.1...v2.46.2) (2022-12-12)


### Bug Fixes

* Update internal analaytics token to a config value, set by environment variable. ([3013b1b](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/3013b1b5190bf132c032f8507a1629e35e7801b4)), closes [#267](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/267)


### Reverts

* Revert "db: Add DB migration to handle deletion of erroneous analytics rows" ([6f44314](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/6f44314fd29963847f648e3c7b6488c89192d371))

## [2.46.1](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.46.0...v2.46.1) (2022-12-11)


### Bug Fixes

* Move title for current submodule/example above name of module ([f026850](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/f02685065e07595d7969a65d75545bb13038ac6d)), closes [#224](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/224)

# [2.46.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.45.3...v2.46.0) (2022-12-09)


### Bug Fixes

* Fix call to _module_provider_404 in example view ([14149f1](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/14149f1ba0e9ff4f1a849cee47151dc89d1ad612)), closes [#255](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/255)


### Features

* Add user groups and user groups namespace permissions, allowing SSO users to be delegated permissionsed per namespace
* Add ability to delete user group permissions and user groups from UI ([1674817](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/1674817a59fb252fbe3b81d6d14c41fabc3fe6af)), closes [#255](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/255)
* Add configurations to output debug for SAML2 and OpenID connect ([8248929](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/8248929bd89b894f97eb4dfcb1d713d09cb6d71a)), closes [#255](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/255)
* Update authenticated endpoint to return list of namespace permissions ([4144c1b](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/4144c1bb45aa1585ac961a5f6a36026aaa54cf42)), closes [#255](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/255)

## [2.45.3](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.45.2...v2.45.3) (2022-11-06)


### Bug Fixes

* Remove 'Description' from start of description in module provider cards in namespace list and search results ([8eba7c3](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/8eba7c31ca50aa132f6e59f7904143bfcf7b55f8)), closes [#263](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/263)

## [2.45.2](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.45.1...v2.45.2) (2022-11-06)


### Bug Fixes

* Fix ZIP command in initial-setup page ([317f8c1](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/317f8c1c2a018a2227854faa6c5afbce1b8e58d3)), closes [#218](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/218)

## [2.45.1](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.45.0...v2.45.1) (2022-11-05)


### Bug Fixes

* Handle multi-levels of indentation in README and other markdown tabs ([9b6bc26](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/9b6bc2694b9758cafe392bf309106305db964fba)), closes [#222](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/222)

# [2.45.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.44.0...v2.45.0) (2022-10-29)


### Features

* Add support 'number' variable types in usage builder ([482d2aa](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/482d2aa321eefd0d1709ab7674221bbce33014cd))
* Re-design usage builder ([228820e](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/228820e363b81e7ad32288f691756925d8ac3757)), closes [#264](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/264)
* Update usage builder to provide multiple inputs when populating a list variable ([68859c5](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/68859c58d766a3d4083a7038409834101bbda20b))

# [2.44.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.43.1...v2.44.0) (2022-10-07)


### Bug Fixes

* Update tfsec absolute path replacement to replace paths from the extraction directory. ([360b40a](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/360b40afae959d9331597cf81e5c7fc1119828e6)), closes [#228](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/228)


### Features

* Add tab listing security issues in module ([b62d8d7](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/b62d8d727a37f17e32332e29f3653dc9cfe55dfa)), closes [#228](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/228)
* Add tab listing security issues in module ([a0bac97](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/a0bac9721761dd9190928c1d55f2dea08e7f43b6))
* Add tab listing security issues in module ([b728c82](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/b728c8255663785b421ad9604bc11886fe715b57))

## [2.43.1](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.43.0...v2.43.1) (2022-10-04)


### Bug Fixes

* Remove unused import, erroneously added in [#178](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/178) ([e5a00ad](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/e5a00adf8595ad5d00e5300734ac20a5628c9ed6)), closes [#262](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/262)

# [2.43.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.42.0...v2.43.0) (2022-10-01)


### Features

* Add UI to create namespaces and API endpoint to create namespaces ([6e81668](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/6e816681771d45e89261f5e9440308922e9bcabb)), closes [#256](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/256)
* Move namespaces into new table ([e01c820](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/e01c820316d45e9c962a9f738f166cfa388a9f6e)), closes [#178](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/178)
* Update initial setup page to include step for creation of initial namespace ([dd40f13](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/dd40f13fc549cf682355b66c33895a5494ad3c15)), closes [#178](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/178)
* Update module provider creation page to select namespace from pre-existing namespaces ([0e6bdac](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/0e6bdac34c88cf48fd1a9286f5fbea094f909c4f)), closes [#178](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/178)

# [2.42.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.41.0...v2.42.0) (2022-10-01)


### Bug Fixes

* Fix logout icon ([cd343cb](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/cd343cbd58de011a9d67e182934512a1e8cc3d9f)), closes [#7](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/7)


### Features

* Implement OpenID and SAML authentication ([c611a7c](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/c611a7cb7c506c9744dd30c7acec8ea8ce0cad9a))

# [2.41.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.40.4...v2.41.0) (2022-10-01)


### Features

* Provide ability to show unpublished and beta versions in UI. ([c844390](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/c844390742bd92eb77bc1a3dfed088e30b55dddb)), closes [#161](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/161)

## [2.40.4](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.40.3...v2.40.4) (2022-09-29)


### Bug Fixes

* Add timeout for git clone command execution. ([4ee467c](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/4ee467c81e9e088ae4e29ddcef39a6e1352cf438)), closes [#256](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/256)
* Fix bitbucket hook to return error when one or more tags fail to import ([8ac85ff](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/8ac85ffbece1cdfc987081dbe9572b83f2bb9cff)), closes [#256](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/256)
* Stop auto-creation of module providers with Github/Bitbucket hook. ([1cf0a3b](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/1cf0a3b142b27e3800af4c5f23e0ac4fef33c47a)), closes [#256](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/256)
* Update Bitbucket and Github hook endpoints to rollback database transactions when a single version fails to upload. ([3b63ecd](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/3b63ecd5483861c7f42c5bcfc065df24a448f9b7)), closes [#256](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/256)

## [2.40.3](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.40.2...v2.40.3) (2022-09-27)


### Bug Fixes

* Update HTML title in pages to use customised application name, rather than 'Terrareg' ([5597c7c](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/5597c7c8e9958db0553a794dfd5411306ec74657)), closes [#240](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/240)

## [2.40.2](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.40.1...v2.40.2) (2022-09-26)


### Bug Fixes

* Protect against Zip file contents that contain members outside of cwd. ([ba09762](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/ba09762a21415aadd8d62027d4edef006515e89c)), closes [#252](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/252)

## [2.40.1](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.40.0...v2.40.1) (2022-09-22)


### Bug Fixes

* **ui:** Remove icon from yearly cost tag ([f7c3b0d](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/f7c3b0d3d3ae013722407260253327a4e33607c7)), closes [#235](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/235)

# [2.40.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.39.2...v2.40.0) (2022-09-22)


### Features

* Order module search results by relevancy. ([fd9572e](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/fd9572edc4033c6ddf18948b07ef36dbc45a9a55)), closes [#241](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/241)

## [2.39.2](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.39.1...v2.39.2) (2022-09-22)


### Bug Fixes

* **ui:** Update icon used for monthly example cost to money bill, rather than dollar ([2f15f37](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/2f15f378de43f655e1589fe7a557a40dee04617c)), closes [#235](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/235)

## [2.39.1](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.39.0...v2.39.1) (2022-09-20)


### Bug Fixes

* **ui:** Hide 'Terrareg Exclusive' tag from Usage builder tab in module provider when this tag is disabled ([1e5b396](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/1e5b39609cbeaf9b2f0d8443c62c5daf9a30e402))

# [2.39.0](https://gitlab.com/mrmattyboy/terrareg/compare/v2.38.0...v2.39.0) (2022-09-14)


### Features

* Add total major, minor and patch releases in prometheus metrics endpoint. ([b20a9ce](https://gitlab.com/mrmattyboy/terrareg/commit/b20a9ce7f372453cc402a27bd9543c98cd926abf))

# [2.38.0](https://gitlab.com/mrmattyboy/terrareg/compare/v2.37.1...v2.38.0) (2022-09-14)


### Features

* Display custom module provider tabs in UI ([2034de0](https://gitlab.com/mrmattyboy/terrareg/commit/2034de0c727a6ed3dc8f450d966e3aacf58e4be5)), closes [#225](https://gitlab.com/mrmattyboy/terrareg/issues/225)
* Update module extractor to extract tab files into new module version files table ([dd42187](https://gitlab.com/mrmattyboy/terrareg/commit/dd42187915bd631b3a1dcc733d495be28911138e)), closes [#225](https://gitlab.com/mrmattyboy/terrareg/issues/225)
* Update terrareg module version API to return list of module version files. ([950a831](https://gitlab.com/mrmattyboy/terrareg/commit/950a831c277c6437e8596ad11ec181c5285683cc)), closes [#225](https://gitlab.com/mrmattyboy/terrareg/issues/225)

## [2.37.1](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.37.0...v2.37.1) (2022-09-04)


### Bug Fixes

* **ui:** Add current example/submodule to breadcrumb in module provider page ([cb1d0fb](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/cb1d0fbe9cea3f346a1fd4e030919258798eb453)), closes [#233](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/233)

# [2.37.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.36.1...v2.37.0) (2022-09-03)


### Features

* Sanitise data provided from module provider to ensure data is not injected into module provider page ([67d9a11](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/67d9a11146874409ae2ebc1f7add9186c57fe188)), closes [#242](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/242)

## [2.36.1](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.36.0...v2.36.1) (2022-09-01)


### Bug Fixes

* Show cost of example when cost of example is 0 ([51e4a0b](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/51e4a0be67a073d08b762f0f4ab119d18e27692e)), closes [#233](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/233)

# [2.36.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.35.0...v2.36.0) (2022-08-31)


### Features

* Update license of project to GNU GPL 3 ([1d963e6](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/1d963e68955fa8545b1b899ea277b65053a4a590)), closes [#239](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/239)

# [2.35.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.34.1...v2.35.0) (2022-08-22)


### Features

* Display result count and current offsets in module search ([06fb35e](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/06fb35e28af30731dafd009078edd70718db2b74)), closes [#210](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/210)

## [2.34.1](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.34.0...v2.34.1) (2022-08-22)


### Bug Fixes

* Remove incorrect colon without port number in integration URLs when using standard http/https ports ([987321e](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/987321e3fc5837e8ba0bd62ea5b9dd1ed35e4038)), closes [#232](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/232)

# [2.34.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.33.0...v2.34.0) (2022-08-20)


### Bug Fixes

* Fix ID match for CSS for yearly-cost text in module provider page ([249263a](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/249263a4503d91424b03793852b80e816398c348)), closes [#226](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/226)
* Remove module description, published date and owner from example page ([4e04eb1](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/4e04eb14e79dc2731695bfc40afc283f84c73e0a)), closes [#226](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/226)
* Support examples obtaing example cost for modules that use other modules hosted in Terrareg. ([446587a](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/446587ac896daa3f74d932c1da91d80156155781)), closes [#226](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/226)
* **ui:** Move module labels in module provider page to be aligned above the name of the module ([7aad2f1](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/7aad2f1b525472e8319888d157f0c25a490c4b42)), closes [#226](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/226)


### Features

* Add cost label to top of module provider, showing yearly cost ([11dc075](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/11dc075452d4a120c90c8c3c0241d3e48ca8c1f4)), closes [#226](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/226)
* Display estimated monthly cost for module examples in UI ([dad7588](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/dad7588b0744cae7dc937690154e19a8b3d4aea8)), closes [#226](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/226)
* Implement infracost scanning of examples during extraction ([bca620a](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/bca620a6fc168c28f7638dee343dbbf06a7188bc)), closes [#226](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/226)

# [2.33.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.32.1...v2.33.0) (2022-08-17)


### Features

* Implement Github hook support ([32ff7c5](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/32ff7c50b83f04ca31005ff20ad98776e6322436)), closes [#58](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/58)

### Special Thanks

 * David Soff <david@soff.nl>

## [2.32.1](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.32.0...v2.32.1) (2022-08-17)


### Reverts

* Revert "ci: Add CI step to test dry-run release on non-main branches" ([c7a7833](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/c7a78331f78a262fe862cf686102d080919b90fc))

# [2.32.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.31.0...v2.32.0) (2022-08-05)


### Bug Fixes

* Update check_subdirectory_within_base_directory to not add trailing slash if base_dir is root ([19ac2fa](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/19ac2fa84ab595d2679f3d00238ba6787c0b2644)), closes [#69](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/69)
* Update safe_join_paths to handle sub_path argument that has a leading slash, which is now converted to a relative path ([4933a32](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/4933a327ceee0f200ef2082c271f4f6f8a71018b)), closes [#69](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/69)
* Update safe_join_paths to optionally take argument to allow sub directory to be the same as the base directory. ([112cca3](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/112cca3402352eed70272a0c686a4229b9c4d113)), closes [#69](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/69)


### Features

* Add methods to return git_path for module provider, version and submodules. ([257a2f9](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/257a2f99500843c2bd3fb8b7b38350211f2ec9a2)), closes [#69](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/69)
* Allow updating git_path attribute of ModuleProvider from settings tab of module provider page. ([254c48c](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/254c48cc194363ea9ded740307bd158d59e5e1b9)), closes [#69](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/69)
* Update module extractor to use git_path directory when extracting module ([2d33739](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/2d337399667cdee56aa7a19aef9fdb8d093494fc)), closes [#69](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/69)

# [2.31.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.30.1...v2.31.0) (2022-08-05)


### Features

* Add titles to all HTML pages. ([d2d5168](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/d2d51687c810149e231197570a750b497e30f645)), closes [#147](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/147)

## [2.30.1](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.30.0...v2.30.1) (2022-08-05)


### Bug Fixes

* Correct source code URL displayed in submodules and examples to include path to the current code ([222ba9f](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/222ba9ff4caacbfc63434f1c8fe46b984107b80b)), closes [#183](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/183)
* Update source_url in example/submodules to revert to module version browse URL when there is no templated source URL available ([63a424b](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/63a424b5db01178dbf1c0cc51aefd3bdf9eeed99)), closes [#183](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/183)

# [2.30.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.29.3...v2.30.0) (2022-07-28)


### Features

* Add analytics token usage per module provider in prometheus metrics ([0aa1abb](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/0aa1abbc88b74530fc359b4cac0a5d39e53cec83)), closes [#39](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/39)
* Add basic endpoint for prometheus metrics. ([06519e1](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/06519e18fef71b0e186b0b52845f22f1925de20e)), closes [#39](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/39)

## [2.29.3](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.29.2...v2.29.3) (2022-07-22)


### Bug Fixes

* Fix values in string arguments in usage builder. ([2c3e838](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/2c3e8383c59f77c967c45a20b8b5c34c6e0504bd)), closes [#202](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/202)

## [2.29.2](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.29.1...v2.29.2) (2022-07-21)


### Bug Fixes

* Use database transaction when performing module provider creation in server endpoint. ([faef873](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/faef8732a8976c8377031e43e3074615ef0fb695)), closes [#191](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/191)

## [2.29.1](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.29.0...v2.29.1) (2022-07-21)


### Bug Fixes

* Update module provider card widths on homepage to increase in size when window is too small to avoid text overflowing ([7bd77fc](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/7bd77fcbe16f014077d3ded2508f9fd0f72ef778)), closes [#209](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/209)

# [2.29.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.28.2...v2.29.0) (2022-07-20)


### Features

* Add warning to module provider page when viewing the non-latest version of the module. ([84d444a](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/84d444a2e863b58f60a98bdb04843ef7d5cc5d5f)), closes [#194](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/194)

## [2.28.2](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.28.1...v2.28.2) (2022-07-20)


### Bug Fixes

* Add connection pool refresh and pool pre-ping settings to avoid errors after database disconnect. ([5c677a8](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/5c677a8ec326e2183f64ec695678b3e575291867)), closes [#204](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/204)

## [2.28.1](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.28.0...v2.28.1) (2022-07-20)


### Bug Fixes

* Catch errors when deleting module provider and module version in module provider settings. ([f834842](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/f8348424e78e40d587152fff82859f61910d15a7)), closes [#198](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/198)
* Display user-readable error when an authentication error occurs when creating a module ([c1f7708](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/c1f7708861e43ddc4a3dbb446b5de26162234aa8)), closes [#198](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/198)
* Provider user-friendly error when attempting to update settings of module provider when not authenticated ([b95f90b](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/b95f90b62b1dc9bf285267a875af1f6480399dca)), closes [#198](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/198)
* Scroll to error message on create module page when an error occurs ([bbfd60d](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/bbfd60df53d8e26f2316cd7f529b7f169026651b)), closes [#198](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/198)

# [2.28.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.27.2...v2.28.0) (2022-07-20)


### Features

* Increase default admin session timeout to 60 minutes from the previous default of 5. ([38ef7df](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/38ef7df8431066d4c0cdc17fbf8c5113ca86ea24)), closes [#200](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/200) [#199](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/199)
* Store sessions in session table in database to hold active sessions. ([6cdeb8b](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/6cdeb8bff60d6e6c3da62c83ab909fb9847ee22e)), closes [#200](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/200)

## [2.27.2](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.27.1...v2.27.2) (2022-07-19)


### Bug Fixes

* Add interactive and pseudo-terminal flags to docker run command in README ([2ff26e9](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/2ff26e9a36034411be944cc87998ea80f121f51c)), closes [#201](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/201) [#196](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/196)

## [2.27.1](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.27.0...v2.27.1) (2022-07-08)


### Bug Fixes

* Fix module provider page when a module provider only has a single unpublished or beta version. ([aee4ac1](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/aee4ac1b2cb064475ee6f345a6f9406d59594adf)), closes [#192](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/192)

# [2.27.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.26.0...v2.27.0) (2022-07-08)


### Features

* Replace example terraform module calls in READMEs for root module and sub-modules. ([b1550b3](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/b1550b37fe8a61e5748bb86413ce6b5f282cfacb)), closes [#128](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/128)

# [2.26.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.25.0...v2.26.0) (2022-07-08)


### Features

* Improve confirmation when deleting a module version or module provider. ([09c6c91](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/09c6c91f55df92c84535604579c88b4092fa9508)), closes [#184](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/184) [#93](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/93)

# [2.25.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.24.0...v2.25.0) (2022-07-08)


### Features

* Add in-progress message when indexing/publishing a version on the module provider page. ([6e2a77d](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/6e2a77d3ea056ba8353fe65526e71a1762f74931)), closes [#182](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/182)

# [2.24.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.23.0...v2.24.0) (2022-07-08)


### Features

* Add API endpoint to return data on total module provider usage, based on unique analytics tokens. ([e705c71](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/e705c71d95311e27ea7a99733c989ffa37aaa8ec)), closes [#39](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/39)

# [2.23.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.22.0...v2.23.0) (2022-07-07)


### Bug Fixes

* Display all module providers in namespace page, including those without versions or those with only beta/unpublished versions. ([803fbd5](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/803fbd5a4c5dec6d80d53e01847f921ebd6d15ff)), closes [#189](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/189)
* Update module search API endpoint to return terrareg API details for results. ([0a39f35](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/0a39f35104920d0fba4bb0b8df8efc6d48b2cb7b)), closes [#189](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/189)
* Update most downloaded and most recent module API endpoints to return terrareg data for module/version. ([511bb35](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/511bb35fc70e3ffa49d6cc254f29f9ee95656042)), closes [#189](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/189)


### Features

* Update module card listing to exclude published time for unpublished/beta modules and to show warning in description about module being unpublished. ([a3bdc72](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/a3bdc725757261e151dd581c01920b8784a0afe1)), closes [#189](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/189)


### Reverts

* Revert "fix: Update most downloaded and most recent module API endpoints to return terrareg data for module/version." ([e4ca722](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/e4ca722abd1959f9fa1cdd5518d30eb8c82e8f4d))
* Revert "fix: Update module search API endpoint to return terrareg API details for results." ([1e65e13](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/1e65e133c529bb5866b362faea512fdf99faec17))

# [2.22.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.21.0...v2.22.0) (2022-07-06)


### Features

* Add config to allow the auto-generated variable template variables for the usage builder to be disabled ([93a10d0](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/93a10d0771100d88677f4c96c5997b053faa43e2)), closes [#135](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/135)
* Add type 'list' to usage builder, which will generate a list with containing the input variable provided. ([25dc66e](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/25dc66ed1757910f7a3f5e32fb5e39e49bccb046)), closes [#135](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/135)
* Use required terraform input variables for module to auto-populate missing variables from 'variable_template' in terrareg metadata ([1d8a974](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/1d8a974a2c29dc5de8d990d8828281a50fddeec7)), closes [#135](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/135)

# [2.21.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.20.3...v2.21.0) (2022-07-06)


### Features

* Add security scanning of modules during upload/indexing. ([90d6901](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/90d69014ca25362937d5aa4ccb6c36facb484762)), closes [#150](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/150)
* **db:** Move common DB columns of module version and submodule tables into new module_details table ([94c54f1](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/94c54f11a81192b26451d2015e395b22bb4e2338)), closes [#150](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/150)
* Display number of security issues found in module version as part of labels attached to module ([853c1d1](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/853c1d159f7c72942c43cc065d5977ac85608978)), closes [#150](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/150)
* Show security issues for root module, submodules and all examples. ([16cd632](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/16cd6320d86b9da5321ad29fb9464b86b90465de)), closes [#150](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/150)

## [2.20.3](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.20.2...v2.20.3) (2022-07-04)


### Bug Fixes

* Correctly return protocol in all integration URLs ([8a82673](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/8a82673523cefd342c675b225e1bcce1fc23a5b5)), closes [#179](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/179)

## [2.20.2](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.20.1...v2.20.2) (2022-07-04)


### Bug Fixes

* Fix display of custom input fields on module provider page when custom module provider Git URLs have been disable. ([09c268d](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/09c268dcd34d7cce843f0615ad42ef277a04fa41)), closes [#171](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/171)
* Fix display of custom repository in git provider select on module provider page when custom module provider Git URLs have been disable. ([c068916](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/c068916c459d665ee2df09f62ee751f1b62e3112)), closes [#171](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/171)
* Fix error thrown from settings endpoint when setting custom git provider URLs to an empty string when custom git URLs is disabled. ([6bda8d6](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/6bda8d6a0fd0d60711d02d525672cbcd99f0d8d4)), closes [#171](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/171)

## [2.20.1](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.20.0...v2.20.1) (2022-07-04)


### Bug Fixes

* Fix link in initial setup to upload version tab from the git upload tab when an unpublished version has been uploaded ([57918ea](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/57918ead32156a5dbf5b3b1309db0f70267760ae))

# [2.20.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.19.0...v2.20.0) (2022-07-03)


### Bug Fixes

* Remove default 'No description provided' from module blocks and from module page ([44db234](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/44db2341c7354e2aef80afb2ca01fbe11f6a2019)), closes [#154](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/154)


### Features

* Automatically generate module descriptions from README contents ([61f8fcc](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/61f8fcc84b943f9376c125376d187f8a484e5f8a)), closes [#154](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/154)

# [2.19.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.18.0...v2.19.0) (2022-07-03)


### Features

* Add initial-setup page, providing a step-by-step guide on setting up Terrareg ([b493dde](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/b493dde6a1d72c223800926fdcc09840238c8322)), closes [#176](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/176)


### Reverts

* Revert "test: Attempt to move selenium setup to session-wide fixture to ensure display/selenium is only setup once to speed up tests" ([e10e543](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/e10e5437e303b1d40277a8af5f484d1e5b145aba))

# [2.18.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.17.7...v2.18.0) (2022-07-03)


### Features

* Display name of provider in module provider result cards. ([862dbbd](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/862dbbdb3682f2a386fae67bf649b42756ee2951)), closes [#162](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/162)

## [2.17.7](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.17.6...v2.17.7) (2022-07-02)


### Bug Fixes

* Fix SQL warning thrown during query in ExampleFile.get_by_path ([4bb50fb](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/4bb50fbf5f55f310cec60b8e4fc718be5d3bbd5c)), closes [#174](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/174)

## [2.17.6](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.17.5...v2.17.6) (2022-07-02)


### Bug Fixes

* Correctly display success message when indexing a module version on module provider page. ([3528284](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/35282845ab37a20c2551ef74f90336471bcce826)), closes [#170](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/170)

## [2.17.5](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.17.4...v2.17.5) (2022-07-02)


### Bug Fixes

* Fix get_all method of Namespace class to return all namespaces when only_published is False. ([7b529c3](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/7b529c3a3d44f802bd302f4af6ccd4ac5f7b3f41)), closes [#167](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/167)
* Update namespace list page to return all namespaces, including those with only module providers with non-published/beta versions of modules ([67798cc](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/67798cc9d65493551f927b1db99ef850143b6992)), closes [#161](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/161) [#167](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/167)

## [2.17.4](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.17.3...v2.17.4) (2022-07-01)


### Bug Fixes

* **db:** Fix DB migration error catching when index does not exist before attempting to delete it ([3df8214](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/3df82145c84f0a9b4e6965839375f0abaf213026)), closes [#169](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/169)
* **db:** Fix SQL parameter binding in DB migration ef71db86c2a1 ([8fd378d](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/8fd378dafef61945671836cd707dd5a838a824b2)), closes [#168](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/168)

## [2.17.3](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.17.2...v2.17.3) (2022-07-01)


### Bug Fixes

* Fix SQL error thrown in namespace list page (and Terrareg namespace list API endpoint), when using MySQL. ([e318ef0](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/e318ef068aa0b922f17c6106050df855707e2814)), closes [#166](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/166)

## [2.17.2](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.17.1...v2.17.2) (2022-07-01)


### Bug Fixes

* Fix warning around passing a query object into value list of _IN in analytics filter query. ([992e526](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/992e526461112535f8fdece056ce1cac36980f5d)), closes [#156](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/156)

## [2.17.1](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.17.0...v2.17.1) (2022-07-01)


### Bug Fixes

* **docs:** Correct/update information in README about getting starting: ([c8a5879](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/c8a58795921e3dca1754f4ccc0ce750f992fa7e4)), closes [#164](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/164)
* **docs:** Update example command to upload module in README, as the namespace was too short ([f53408d](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/f53408d8f061b331494b92bc4372e88530bf03fe)), closes [#164](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/164)

# [2.17.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.16.0...v2.17.0) (2022-07-01)


### Features

* Add Datadog logo to available provider logos. ([556daa7](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/556daa7a6a6971a52238b369b20c607ece469c05)), closes [#148](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/148)

# [2.16.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.15.0...v2.16.0) (2022-06-30)


### Bug Fixes

* Handle None value for published date in functions that return formatted versions of published date ([b53a447](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/b53a4475f18c0986aa05f72bfd1b580d7ac5e4a0)), closes [#12](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/12)


### Features

* Convert module provider, version, submodule and example page to static page using javascript and APIs. ([598fc25](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/598fc251b08972c106599e7544d97c12981da213)), closes [#12](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/12)

# [2.15.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.14.0...v2.15.0) (2022-06-25)


### Bug Fixes

* Correct name of environment variable that is used to set EXAMPLES_DIRECTORY config ([8186f40](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/8186f408f72039d4e7c8cd89ca27e14320e016ea)), closes [#141](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/141)
* Fix ALLOW_UNIDENTIFIED_DOWNLOADS configuration to respect environment variable, rather than always returning False ([c8b895e](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/c8b895e4ddf4a4db8b74059fdd38a4296c71499f)), closes [#141](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/141)
* Fix environment variable used to obtain API keys to publishing module versions ([d5d8f79](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/d5d8f796f61fef9b3d5d7324d32aaecee562ff2b)), closes [#141](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/141)
* Hide 'publish' button from integrations tab of modules page if publish API keys have been enabled. ([8757f73](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/8757f730c298f27574b5cd7e61c4f34002a5b560)), closes [#141](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/141)


### Features

* Use common method for obtaining boolean environment variables configs. ([f4d1d81](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/f4d1d81cc97cb1e6ecaada237b5552acba94df3f)), closes [#141](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/141)

# [2.14.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.13.1...v2.14.0) (2022-06-25)

WARNING: This database change is non-downgradable, due to incorrect naming of previous foreign keys.

If a downgrade from this version is required, the database must be restored from a pre-upgrade backup.

### Bug Fixes

* Update analytics foreign key to module version to perform no action when module version is removed. ([757b500](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/757b500efa9f3b2732b7f0c71ca2aea28c1a157f)), closes [#153](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/153)


### Features

* Update module version deletion to remove any analytics and update re-creation of module version to migrate analytics to new module version ID ([2ee18e9](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/2ee18e9665750bd845cfee1e5346e786ea7a1b47)), closes [#153](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/153)

## [2.13.1](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.13.0...v2.13.1) (2022-06-25)


### Bug Fixes

* Update basic usage and usage builder rendered terraform to use 2-space indentation, rather than 4, as per terraform best practices ([2c93eab](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/2c93eab105bbbd907c79fc3fc3510a073e4c7369)), closes [#151](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/151)

# [2.13.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.12.1...v2.13.0) (2022-06-25)


### Bug Fixes

* Require a value for the 'Git tag format' field and reword the description in the module creation page. ([9afb2b4](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/9afb2b43bc24f114b5f6b04b4bd3f3f233ae1784)), closes [#152](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/152)


### Features

* Add ability to delete module version in UI from settings tab of module provider ([c88c24f](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/c88c24f6a9f1cf9df64360462f89c19892458c83)), closes [#145](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/145)
* Add API endpoint to delete module version ([9cc5256](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/9cc5256f7b6cecd52820fa408af6d075269e3fd7)), closes [#145](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/145)
* Automatically hide the custom git/browse URL inputs in create module page when selecting a none-custom git provider ([0d5a148](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/0d5a148a5be0d93bcd7b90df1e53c54b911a8294)), closes [#152](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/152)
* Move main.tf to top of list of example files in UI. ([5a4faa2](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/5a4faa2b9e6b9e28965f86983164e8e38c189687))
* Update example UI page to default to show example files ([ffe1566](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/ffe1566821ebfb1a465fabc1e8f88dc5d4394db6)), closes [#146](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/146)

## [2.12.1](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.12.0...v2.12.1) (2022-06-15)


### Bug Fixes

* Update module/submodule/example pages to no longer redirect to tab anchor when loading default tab ([48a25d4](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/48a25d489745e874e37e3309ac3231890d9df59d)), closes [#144](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/144)

# [2.12.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.11.0...v2.12.0) (2022-06-14)


### Bug Fixes

* **test:** Use assert_equal method in module search selenium tests to allow the counting of result cards to be retries, as the cards may not yet be populated on the page ([a4b4716](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/a4b47164e6b8ea9eec1eadcb847f714fd5a3df68)), closes [#125](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/125)
* Update namespace page to include internal modules. ([245693b](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/245693b6569e7956b217279c9f3b2741b5c98b9b)), closes [#125](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/125)
* Update namespace page to show error when no modules exist or namespace does not exist. ([4470af6](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/4470af665fab5375104a145606ea0bbb0b7a99aa)), closes [#125](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/125)


### Features

* Convert namespace list page to use API calls to obtain list of namespaces. ([9d5308a](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/9d5308a18f62c51a77182605d65f8f0be8dbbb1c)), closes [#125](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/125)

# [2.11.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.10.1...v2.11.0) (2022-06-12)


### Bug Fixes

* Update most downloaded and most recently uploaded module functions to exclude internal modules ([411209e](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/411209e4971b8d4ae540c4d15ec8adefc1b4421e)), closes [#137](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/137)
* Update mostdownloaded endpoint to not return download statistics for beta versions ([ac98f6c](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/ac98f6c491cc6acf1ee9d5116f48e945a06f04ec)), closes [#137](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/137)


### Features

* Add 'internal' tag to internal modules ui module provider UI view. ([59f96ee](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/59f96eea0fc7b8b9fc7f74002d10f87eaf344dda)), closes [#137](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/137)
* Update module search and query filter data to exclude internal modules. ([266344f](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/266344fd0c0847f3c4ade1a64b654e0dd782c959)), closes [#137](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/137)

## [2.10.1](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.10.0...v2.10.1) (2022-06-10)


### Bug Fixes

* Update example file module path replacement to add 'version' argument. ([05c5cc3](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/05c5cc3091c7e8cd312afde034c80047fb042d18)), closes [#127](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/127)

# [2.10.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.9.0...v2.10.0) (2022-06-01)


### Bug Fixes

* Fix method call when checking if currently in transaction during creation of transaction ([e96c08f](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/e96c08fc9643cbde31edf27bd4af3eb83686c8bf)), closes [#101](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/101)
* Fix name of __init__ method, which was incorrectly called __enter__ in transaction wrapper. ([5ead4f8](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/5ead4f832c91a84ce2d155d5e8d3c8c363af865b)), closes [#101](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/101)


### Features

* Update bitbucket hook to use transaction whilst handling request, meaning any database changes are rolled back if any issue occurs ([64f8e39](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/64f8e3972cb0cc4dc0b9fb9b054442754c2b2f4d)), closes [#101](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/101)

# [2.9.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.8.0...v2.9.0) (2022-05-27)


### Features

* Add list of provider requirements to module page. ([4bf1a12](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/4bf1a12210edaa3bd3e19944a5f542ca471577c8)), closes [#140](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/140)

# [2.8.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.7.4...v2.8.0) (2022-05-26)


### Bug Fixes

* Fix methods for obtaining variable_template and module details when database columns are empty ([ec22e39](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/ec22e39fe4a69e1bbca9988fe2194b07b74398e5))
* Handle None database value when decoding blob, which previously threw an exception ([9799cc8](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/9799cc82c6d5b57a87df117ca729d506e242683f))
* Revert change to get_readme_content to handle None value before decoding, since this is now handled by the blob decode method ([a4aff2d](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/a4aff2d30f6e3b16dc733bfb330671aa85756780))


### Features

* Add module labels to module provider/version page. ([dcd63e1](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/dcd63e17a7f30d61fe03265205326cdbfff8f6df)), closes [#138](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/138) [#12](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/12)

## [2.7.4](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.7.3...v2.7.4) (2022-05-25)


### Bug Fixes

* **test:** Add missing 'published_at' field in database rows for inserted test data ([986fe2e](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/986fe2ebc0290ee9c257c8b7a2ee166e2c6a93c7)), closes [#136](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/136)

## [2.7.3](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.7.2...v2.7.3) (2022-05-24)


### Bug Fixes

* Handle error when README is empty (when None is stored in DB row) ([b49968d](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/b49968dfab786f0c4d58a99fcc5ff3c9d42d3ad9)), closes [#129](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/129)

## [2.7.2](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.7.1...v2.7.2) (2022-05-24)


### Bug Fixes

* Add 'content' tag to README div, which allows for markup of various types of tags, such as lists ([f3cea4a](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/f3cea4ad084ac7b8045b416cc7eaff3d6b6e8361)), closes [#132](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/132)
* Force version/submodule/example dropdowns from keeping previous value on page load. ([2faa911](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/2faa9116aca6c663bc93b1562dbf4cdb5f6c86af)), closes [#131](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/131)

## [2.7.1](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.7.0...v2.7.1) (2022-05-24)


### Bug Fixes

* Allow empty search, which returns all module providers. ([bdd329a](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/bdd329a97710cc3ed225e23e56779f047d8b140d)), closes [#133](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/133)
* Ensure that first file in examples is left selected, rather than showing the content of the last file. ([c57ec87](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/c57ec87d94c3346cf58a1d24fd0f2e132e923fe7)), closes [#133](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/133)
* Fix name of usage builder in UI in Usage example ([1898ba0](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/1898ba0d8a533675745a75f4d12cf605904451e2)), closes [#133](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/133)
* Fix tiles on homepage occasionally not loading when config request has not completed in time. ([ddeb549](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/ddeb54963fb2abbb5a9539c5e92bde51423ba31b)), closes [#133](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/133)
* Remove trailing/leading new lines in preformatted text for usage builder and example file content. ([51c418f](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/51c418fbd8bd74929e6cfe4202dc0148ae0c35a6)), closes [#133](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/133)
* Show currently selected example file in file list panel ([e2b04c5](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/e2b04c58a8b9049d3166c8e5e8bab62fccd95efe)), closes [#133](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/133)

# [2.7.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.6.0...v2.7.0) (2022-05-23)


### Features

* Update examples in UI to display converted relative module source paths. ([4399886](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/4399886a65ed8d436efcdd491d36ea28d5026f65)), closes [#91](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/91)

# [2.6.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.5.4...v2.6.0) (2022-05-22)


### Bug Fixes

* Update DB migration to use batch operation to handle errors caused by sqlite database ([bf10420](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/bf10420bd75e7c2e6de73834373abbe9051a544a)), closes [#115](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/115)
* Update example terraform version string for beta versions to return exact version match. ([6d4a677](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/6d4a67793e9e64ad97a60c35447390262a40949d)), closes [#115](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/115)
* Update initial module-search page search to occur once document has finished loading. ([5b7099d](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/5b7099d49a5e8030ebaa831e7ff40bbfb8e9dbc7)), closes [#122](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/122)


### Features

* Add column to module version table to handle 'beta' flag ([999cbc6](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/999cbc6faa06eeeb5349346afc4324a0cc3af02e)), closes [#115](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/115)
* Add config for custom labels for trusted namespaces and verified modules in the search filters. ([f533beb](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/f533bebd0be790da7291ad94770890a0d4983687)), closes [#122](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/122)
* Add configuration to disable auto-creation of module providers on version creation/import ([5654b5b](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/5654b5bc13926de2d2c6c0612438f7001a9cadd1)), closes [#123](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/123)
* Add configuration to use custom name for contributed modules in search UI ([19cfa6c](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/19cfa6c7c89c301ba3ef5be17220698515377a92)), closes [#122](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/122)
* Add environment variables for setting SSL public and private key. ([acadd38](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/acadd38bee3ce2a7d64937dd236715412d1529ce)), closes [#99](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/99)
* Add labels to search results with tags for trusted, verified and contributed. ([72b2aa3](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/72b2aa34c25dabb4e3da95f142abd68f6583fc2c)), closes [#122](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/122)
* Add tab to example page to view list of example files and view file contents ([6aef439](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/6aef439ae56501ed5e213f2b7962a34ed3cf6df9)), closes [#91](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/91)
* Extract terraform files from examples during upload and store in  example file table. ([d1a8d76](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/d1a8d76a9a26bdadb1e163c0038306c565773dd7)), closes [#91](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/91)
* Remove beta versions from web interface. ([21a12a5](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/21a12a5d317e25c3c462ba414c82259c191df050)), closes [#115](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/115)
* Support version suffixes when creating module version. ([efcc72f](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/efcc72f3599aeee5ee13e5e237c59225c41e002b)), closes [#115](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/115)


### Reverts

* Revert "chore: Update safe_iglob to support returning relative results." ([538c616](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/538c61699f5a14640a94d250bb016597dbbc8075))

## [2.5.4](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.5.3...v2.5.4) (2022-05-19)


### Bug Fixes

* Add logo for null provider ([613f0b6](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/613f0b6a9a79a80af6ad9a1179bd7af615f2b891)), closes [#120](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/120)

## [2.5.3](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.5.2...v2.5.3) (2022-05-19)


### Bug Fixes

* Update analytics recording to record every module download. ([993bc46](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/993bc46533bc990b9b76ddd88355c2fb32f713ed)), closes [#119](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/119)

## [2.5.2](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.5.1...v2.5.2) (2022-05-12)


### Bug Fixes

* Issue [#118](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/118) Fix endpoint for module version upload in UI ([f50b7e6](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/f50b7e646433c593ea169b693c39c7cae82c25ee))
* Issue [#118](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/118) Fix module version import url in UI to contain module provider ID ([18e41b8](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/18e41b8fd87fc8ac05a00640be448d96f08676c9))
* Issue [#118](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/118) Fix publish endpoint in UI to contain module provider ID ([cecfff5](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/cecfff53d07e940608ad8a44bc0d59a58c6305c4))

## [2.5.1](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.5.0...v2.5.1) (2022-05-11)


### Bug Fixes

* Issue [#111](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/111) Fix validation errors thrown when validation git provider config with provided tag_uri_encoded placeholder ([3113ccd](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/3113ccd319c6c662b14148e2fceb05d2dc681a83))

# [2.5.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.4.1...v2.5.0) (2022-05-11)


### Bug Fixes

* Issue [#113](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/113) Refuse the use of the example analytics token. ([1c9e17c](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/1c9e17cf1c5445246576a8271626c3394fff7bb7))


### Features

* Issue [#109](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/109) Add namespace filtering to module search interface. ([f3798af](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/f3798af1d2958728f9cdadeed725966eb5660bab))
* Issue [#111](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/111) Add URI encoded tag placeholder to browse URL template. ([c467bde](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/c467bdee6dd5010efede6102054e445f0f0a20e9))
* Issue [#113](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/113) Add input for analytics token to 'usage builder' to add to module URL in built terraform. ([e599366](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/e59936644f9bdd9831d241511da7288eab07583e))
* Issue [#113](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/113) Add placeholder analytics token to 'usage' terraform with note about having to replace it. ([240764d](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/240764d71db65bc671e9f5624571474559284a46))
* Issue [#114](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/114) Add custom terraform module version template for use in terraform examples in UI. ([07bfe72](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/07bfe7247a3414267853f23a9cae42981e5b0a13))

## [2.4.1](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.4.0...v2.4.1) (2022-05-10)


### Bug Fixes

* **module-extractor:** Issue [#117](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/117) Remove any terraform-docs configuration files before running terraform docs ([9d46f26](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/9d46f26c88c33b05befe91b6f54395c0c5fe26fb))

# [2.4.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.3.1...v2.4.0) (2022-05-09)


### Features

* Issue [#116](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/116) Display latest environment in analytics token table in module version. ([335177e](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/335177efa26496273cb5ef665b95c27081b18c60))

## [2.3.1](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.3.0...v2.3.1) (2022-05-09)


### Bug Fixes

* Remove stripping of ssh:// from git clone URL when extracting module ([649c281](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/649c281cdb652ea61db9cc748b61aa6cfeb2c1f1))

# [2.3.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.2.0...v2.3.0) (2022-05-09)


### Bug Fixes

* **module-search:** Issue [#96](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/96) Fix error during initial search on module_search page load. ([2e80af3](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/2e80af3630cedb8084ae9076bbbcb8580974868f))


### Features

* Issue [#53](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/53) Add provider filter to module search interface. ([e137c8e](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/e137c8e22f50834b1ada8ae5a9ef89cbe0867e7c))
* Issue [#53](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/53) Update search endpoints to accept multiple namespace, provider and module name arguments. ([9c3e14c](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/9c3e14c8d5868a46e3087fc5ebb5f3bac8d2c951))

# [2.2.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.1.1...v2.2.0) (2022-05-08)


### Features

* Issue [#90](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/90) Add API authentication to ModuleVersion create endpoint. ([76395f6](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/76395f6256ff3a6cf061046dd115cef709336207))
* Issue [#90](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/90) Add API authentication to ModuleVersion publish endpoint. ([90d2c0a](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/90d2c0aa0c4101c30a13852fa0a5fad962f861af))
* Issue [#90](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/90) Add API key authentication for uploading module. ([19f901b](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/19f901b8dbd900165e80e5bf3619c4e2ba39de4f))
* Issue [#90](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/90) Add handling of upload API key as 'secret' from bitbucket web hook. ([59ab633](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/59ab633b1d064fe7256948dc14c5913a3d60a840))

## [2.1.1](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.1.0...v2.1.1) (2022-05-07)


### Bug Fixes

* **module-extractor:** Issue [#108](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/108) Catch error raised when performing git clone. ([956ed9d](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/956ed9d67d5257226d0b6cc86d859be172e141cd))

# [2.1.0](https://gitlab.dockstudios.co.uk/pub/terrareg/compare/v2.0.0...v2.1.0) (2022-05-07)


### Bug Fixes

* **ci:** Issue [#107](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/107) Move mysql service within mysql test stage ([6050426](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/605042660b3727635b69265c1483fad2e6c1dca6))
* **ci:** Issue [#107](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/107) Update mysql service for tests to use internal docker image ([264676d](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/264676de370dd2903ccdbbed2fa27fdecc23cced))
* **ci:** Issue [#107](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/107) Use local repository for python image to avoid rate limits from dockerhub ([bd7e9dc](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/bd7e9dc420e274c65430c8e3a7e2b9a1095b0cd0))
* **db:** Issue [#107](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/107) Convert BLOB columns to MEDIUMBLOBS ([907982a](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/907982aaba181441adf05cc3e9541980d1b1b56e))
* **db:** Issue [#107](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/107) Fix migration script to convert columns to mediumblob ([d832044](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/d832044b581b9513d9f7078a75469ce7c3be9474))
* **db:** Issue [#107](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/107) Fix queries for global stats that attempt to retrieve ID for grouped rows. ([8cb2676](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/8cb2676f589617cfe5e7286af48640249aa19fb5))
* **db:** Issue [#107](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/107) Remove unused database migration ([7ab53fe](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/7ab53fee6770b6bc5914e59af3ba8eb872805f18))
* **db:** Issue [#107](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/107) Update select_module_version_joined_module_provider to take a list of columns/tables to select from. ([1fc42af](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/1fc42af441078cd5f5516cdd8da3696e7f1ec24b))
* **db:** Issue [#107](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/107) Use medium blob dialect for medium blob columns to stop LargeString from reverting to MySQL blob ([0b72a2f](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/0b72a2ffcd32b3fc0f0c20772219bdd57ab31e12))
* **models:** Issue [#107](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/107) Fix queries for MySQL, which perform grouping and were obtaining non-grouped fields (such as ID). ([5b91fa9](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/5b91fa99adc3ceeafba83d383eace78423cf8fbc))
* **module-search:** Issue [#107](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/107) Fix base query used for search results to only include columns from module_provider, as module_version is being aggregated by module_provider.id. ([0801f27](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/0801f27ceb55bad331fa861c36a8ad96ac6c55b9))


### Features

* **ci:** Issue [#107](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/107) Split unit/integration test runs in CI and add execution of integration tests against MySQL ([7a70ad7](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/7a70ad7e56eb871e71cdfcd5ed558b5b2161107d))
* **tests:** Issue [#107](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/107) Add ability to test against custom database URL in integration tests. ([5522690](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/5522690e58c19168799abacac808b33dbf7fc375))

### v2.0.1

 * Reduce size of columns that are too large

### v2.0.0

WARNING: This version does not support migration from previous versions.

 * Update all database columns to use MySQL-compatible types.
 * Convert large data values to blobs

### v1.1.0

 * Add MySQL connector and document URL format to connect to MySQL
 * Fix SQL schema to work with mysql
 * Provide ability to pass SSH private key through environment variable

### v1.0.3

 * Update base image to python:10
 * Remove update of system packages in Dockerfile

### v1.0.2

 * Add exception to be thrown when upload fails to analyse terraform
 * Add tests for module extraction

### v1.0.1

 * Add package updates to Dockerfile

### v1.0.0

 * Initial release
