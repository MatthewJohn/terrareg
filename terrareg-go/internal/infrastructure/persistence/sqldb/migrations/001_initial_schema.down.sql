-- Rollback initial schema migration

-- Drop tables in reverse order of dependencies

DROP TABLE IF EXISTS audit_history;
DROP TABLE IF EXISTS provider_version_binary;
DROP TABLE IF EXISTS provider_version_documentation;
DROP TABLE IF EXISTS provider_version;
DROP TABLE IF EXISTS provider;
DROP TABLE IF EXISTS repository;
DROP TABLE IF EXISTS provider_category;
DROP TABLE IF EXISTS provider_source;
DROP TABLE IF EXISTS gpg_key;
DROP TABLE IF EXISTS module_version_file;
DROP TABLE IF EXISTS example_file;
DROP TABLE IF EXISTS provider_analytics;
DROP TABLE IF EXISTS analytics;
DROP TABLE IF EXISTS submodule;
DROP TABLE IF EXISTS module_provider_redirect;
DROP TABLE IF EXISTS module_version;
DROP TABLE IF EXISTS module_provider;
DROP TABLE IF EXISTS module_details;
DROP TABLE IF EXISTS namespace_redirect;
DROP TABLE IF EXISTS git_provider;
DROP TABLE IF EXISTS user_group_namespace_permission;
DROP TABLE IF EXISTS namespace;
DROP TABLE IF EXISTS user_group;
DROP TABLE IF EXISTS terraform_idp_subject_identifier;
DROP TABLE IF EXISTS terraform_idp_access_token;
DROP TABLE IF EXISTS terraform_idp_authorization_code;
DROP TABLE IF EXISTS session;
