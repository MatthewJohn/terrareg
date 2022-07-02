# Changelog

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


### Bug Fixes

* Update analaytics foreign key to module version to perform no action when module version is removed. ([757b500](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/757b500efa9f3b2732b7f0c71ca2aea28c1a157f)), closes [#153](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/153)


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

* Issue [#113](https://gitlab.dockstudios.co.uk/pub/terrareg/issues/113) Refuse the use of the example analaytics token. ([1c9e17c](https://gitlab.dockstudios.co.uk/pub/terrareg/commit/1c9e17cf1c5445246576a8271626c3394fff7bb7))


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
