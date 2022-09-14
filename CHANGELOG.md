# Changelog

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
