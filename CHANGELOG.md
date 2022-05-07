# Changelog

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
