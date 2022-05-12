# Changelog

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