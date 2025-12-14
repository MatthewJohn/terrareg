-- Initial schema migration for Terrareg
-- This creates all tables matching the Python SQLAlchemy schema

-- Session table
CREATE TABLE IF NOT EXISTS session (
    id VARCHAR(128) PRIMARY KEY,
    expiry DATETIME NOT NULL,
    provider_source_auth MEDIUMBLOB
);

-- Terraform IDP OAuth tables
CREATE TABLE IF NOT EXISTS terraform_idp_authorization_code (
    id INTEGER PRIMARY KEY AUTO_INCREMENT,
    `key` VARCHAR(128) NOT NULL UNIQUE,
    `data` MEDIUMBLOB,
    expiry DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS terraform_idp_access_token (
    id INTEGER PRIMARY KEY AUTO_INCREMENT,
    `key` VARCHAR(128) NOT NULL UNIQUE,
    `data` MEDIUMBLOB,
    expiry DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS terraform_idp_subject_identifier (
    id INTEGER PRIMARY KEY AUTO_INCREMENT,
    `key` VARCHAR(128) NOT NULL UNIQUE,
    `data` MEDIUMBLOB,
    expiry DATETIME NOT NULL
);

-- User group tables
CREATE TABLE IF NOT EXISTS user_group (
    id INTEGER PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(128) NOT NULL UNIQUE,
    site_admin BOOLEAN DEFAULT FALSE NOT NULL
);

-- Namespace table (must be created before user_group_namespace_permission)
CREATE TABLE IF NOT EXISTS namespace (
    id INTEGER PRIMARY KEY AUTO_INCREMENT,
    namespace VARCHAR(128) NOT NULL,
    display_name VARCHAR(128),
    namespace_type VARCHAR(50) NOT NULL DEFAULT 'NONE'
);

CREATE TABLE IF NOT EXISTS user_group_namespace_permission (
    user_group_id INTEGER NOT NULL,
    namespace_id INTEGER NOT NULL,
    permission_type VARCHAR(50),
    PRIMARY KEY (user_group_id, namespace_id),
    FOREIGN KEY (user_group_id) REFERENCES user_group(id) ON UPDATE CASCADE ON DELETE CASCADE,
    FOREIGN KEY (namespace_id) REFERENCES namespace(id) ON UPDATE CASCADE ON DELETE CASCADE
);

-- Git provider table
CREATE TABLE IF NOT EXISTS git_provider (
    id INTEGER PRIMARY KEY,
    name VARCHAR(128) UNIQUE,
    base_url_template VARCHAR(1024),
    clone_url_template VARCHAR(1024),
    browse_url_template VARCHAR(1024),
    git_path_template VARCHAR(1024)
);

-- Namespace redirect table
CREATE TABLE IF NOT EXISTS namespace_redirect (
    id INTEGER PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(128) NOT NULL,
    namespace_id INTEGER NOT NULL,
    FOREIGN KEY (namespace_id) REFERENCES namespace(id) ON UPDATE CASCADE ON DELETE CASCADE
);

-- Module details table (must be created before module_version)
CREATE TABLE IF NOT EXISTS module_details (
    id INTEGER PRIMARY KEY AUTO_INCREMENT,
    readme_content MEDIUMBLOB,
    terraform_docs MEDIUMBLOB,
    tfsec MEDIUMBLOB,
    infracost MEDIUMBLOB,
    terraform_graph MEDIUMBLOB,
    terraform_modules MEDIUMBLOB,
    terraform_version MEDIUMBLOB
);

-- Module provider table
CREATE TABLE IF NOT EXISTS module_provider (
    id INTEGER PRIMARY KEY AUTO_INCREMENT,
    namespace_id INTEGER NOT NULL,
    module VARCHAR(128),
    provider VARCHAR(128),
    repo_base_url_template VARCHAR(1024),
    repo_clone_url_template VARCHAR(1024),
    repo_browse_url_template VARCHAR(1024),
    git_tag_format VARCHAR(128),
    git_path VARCHAR(1024),
    archive_git_path BOOLEAN DEFAULT FALSE,
    verified BOOLEAN DEFAULT NULL,
    git_provider_id INTEGER DEFAULT NULL,
    latest_version_id INTEGER DEFAULT NULL,
    FOREIGN KEY (namespace_id) REFERENCES namespace(id) ON UPDATE CASCADE ON DELETE CASCADE,
    FOREIGN KEY (git_provider_id) REFERENCES git_provider(id) ON UPDATE CASCADE ON DELETE SET NULL
);

-- Module version table
CREATE TABLE IF NOT EXISTS module_version (
    id INTEGER PRIMARY KEY AUTO_INCREMENT,
    module_provider_id INTEGER NOT NULL,
    version VARCHAR(128),
    git_sha VARCHAR(128),
    git_path VARCHAR(1024),
    archive_git_path BOOLEAN DEFAULT FALSE,
    module_details_id INTEGER DEFAULT NULL,
    beta BOOLEAN NOT NULL,
    owner VARCHAR(128),
    description VARCHAR(1024),
    repo_base_url_template VARCHAR(1024),
    repo_clone_url_template VARCHAR(1024),
    repo_browse_url_template VARCHAR(1024),
    published_at DATETIME,
    variable_template MEDIUMBLOB,
    internal BOOLEAN NOT NULL,
    published BOOLEAN DEFAULT NULL,
    extraction_version INTEGER DEFAULT NULL,
    FOREIGN KEY (module_provider_id) REFERENCES module_provider(id) ON UPDATE CASCADE ON DELETE CASCADE,
    FOREIGN KEY (module_details_id) REFERENCES module_details(id) ON UPDATE CASCADE ON DELETE CASCADE
);

-- Add foreign key constraint for module_provider.latest_version_id
ALTER TABLE module_provider
ADD FOREIGN KEY (latest_version_id) REFERENCES module_version(id) ON UPDATE CASCADE ON DELETE SET NULL;

-- Module provider redirect table
CREATE TABLE IF NOT EXISTS module_provider_redirect (
    id INTEGER PRIMARY KEY AUTO_INCREMENT,
    module VARCHAR(128) NOT NULL,
    provider VARCHAR(128) NOT NULL,
    namespace_id INTEGER NOT NULL,
    module_provider_id INTEGER NOT NULL,
    FOREIGN KEY (namespace_id) REFERENCES namespace(id) ON UPDATE CASCADE ON DELETE CASCADE,
    FOREIGN KEY (module_provider_id) REFERENCES module_provider(id) ON UPDATE CASCADE ON DELETE CASCADE
);

-- Submodule table
CREATE TABLE IF NOT EXISTS submodule (
    id INTEGER PRIMARY KEY AUTO_INCREMENT,
    parent_module_version INTEGER NOT NULL,
    module_details_id INTEGER DEFAULT NULL,
    type VARCHAR(128),
    path VARCHAR(1024),
    name VARCHAR(128),
    FOREIGN KEY (parent_module_version) REFERENCES module_version(id) ON UPDATE CASCADE ON DELETE CASCADE,
    FOREIGN KEY (module_details_id) REFERENCES module_details(id) ON UPDATE CASCADE ON DELETE CASCADE
);

-- Analytics tables
CREATE TABLE IF NOT EXISTS analytics (
    id INTEGER PRIMARY KEY AUTO_INCREMENT,
    parent_module_version INTEGER NOT NULL,
    timestamp DATETIME,
    terraform_version VARCHAR(128),
    analytics_token VARCHAR(128),
    auth_token VARCHAR(128),
    environment VARCHAR(128),
    namespace_name VARCHAR(128),
    module_name VARCHAR(128),
    provider_name VARCHAR(128),
    INDEX idx_parent_module_version (parent_module_version)
);

CREATE TABLE IF NOT EXISTS provider_analytics (
    id INTEGER PRIMARY KEY AUTO_INCREMENT,
    provider_version_id INTEGER NOT NULL,
    timestamp DATETIME,
    terraform_version VARCHAR(128),
    namespace_name VARCHAR(128),
    provider_name VARCHAR(128),
    INDEX idx_provider_version_id (provider_version_id)
);

-- Example file table
CREATE TABLE IF NOT EXISTS example_file (
    id INTEGER PRIMARY KEY AUTO_INCREMENT,
    submodule_id INTEGER NOT NULL,
    path VARCHAR(128) NOT NULL,
    content MEDIUMBLOB,
    FOREIGN KEY (submodule_id) REFERENCES submodule(id) ON UPDATE CASCADE ON DELETE CASCADE
);

-- Module version file table
CREATE TABLE IF NOT EXISTS module_version_file (
    id INTEGER PRIMARY KEY AUTO_INCREMENT,
    module_version_id INTEGER NOT NULL,
    path VARCHAR(128) NOT NULL,
    content MEDIUMBLOB,
    FOREIGN KEY (module_version_id) REFERENCES module_version(id) ON UPDATE CASCADE ON DELETE CASCADE
);

-- GPG key table
CREATE TABLE IF NOT EXISTS gpg_key (
    id INTEGER PRIMARY KEY AUTO_INCREMENT,
    namespace_id INTEGER NOT NULL,
    ascii_armor LONGTEXT NOT NULL,
    key_id VARCHAR(1024) NOT NULL,
    fingerprint VARCHAR(1024) NOT NULL UNIQUE,
    source VARCHAR(1024) DEFAULT '',
    source_url VARCHAR(1024) NULL,
    trust_signature TEXT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (namespace_id) REFERENCES namespace(id) ON UPDATE CASCADE ON DELETE CASCADE
);

-- Create indexes for efficient GPG key queries
CREATE INDEX IF NOT EXISTS idx_gpg_key_namespace_id ON gpg_key(namespace_id);
CREATE INDEX IF NOT EXISTS idx_gpg_key_key_id ON gpg_key(key_id);
CREATE INDEX IF NOT EXISTS idx_gpg_key_fingerprint ON gpg_key(fingerprint);

-- Create trigger to update updated_at timestamp for GPG keys
CREATE TRIGGER IF NOT EXISTS update_gpg_key_updated_at
    AFTER UPDATE ON gpg_key
    FOR EACH ROW
BEGIN
    UPDATE gpg_key SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- Provider source table
CREATE TABLE IF NOT EXISTS provider_source (
    name VARCHAR(128) PRIMARY KEY,
    api_name VARCHAR(128),
    provider_source_type VARCHAR(50),
    config MEDIUMBLOB
);

-- Provider category table
CREATE TABLE IF NOT EXISTS provider_category (
    id INTEGER PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(128),
    slug VARCHAR(128) UNIQUE,
    user_selectable BOOLEAN DEFAULT TRUE
);

-- Repository table
CREATE TABLE IF NOT EXISTS repository (
    id INTEGER PRIMARY KEY AUTO_INCREMENT,
    provider_id VARCHAR(128),
    owner VARCHAR(128),
    name VARCHAR(128),
    description MEDIUMBLOB,
    clone_url VARCHAR(1024),
    logo_url VARCHAR(1024),
    provider_source_name VARCHAR(128) NOT NULL,
    FOREIGN KEY (provider_source_name) REFERENCES provider_source(name) ON UPDATE CASCADE ON DELETE CASCADE
);

-- Provider table
CREATE TABLE IF NOT EXISTS provider (
    id INTEGER PRIMARY KEY AUTO_INCREMENT,
    namespace_id INTEGER NOT NULL,
    name VARCHAR(128),
    description VARCHAR(1024),
    tier VARCHAR(50),
    default_provider_source_auth BOOLEAN DEFAULT FALSE,
    provider_category_id INTEGER DEFAULT NULL,
    repository_id INTEGER DEFAULT NULL,
    latest_version_id INTEGER DEFAULT NULL,
    FOREIGN KEY (namespace_id) REFERENCES namespace(id) ON UPDATE CASCADE ON DELETE CASCADE,
    FOREIGN KEY (provider_category_id) REFERENCES provider_category(id) ON UPDATE CASCADE ON DELETE SET NULL,
    FOREIGN KEY (repository_id) REFERENCES repository(id) ON UPDATE CASCADE ON DELETE CASCADE
);

-- Provider version table
CREATE TABLE IF NOT EXISTS provider_version (
    id INTEGER PRIMARY KEY AUTO_INCREMENT,
    provider_id INTEGER NOT NULL,
    gpg_key_id INTEGER NOT NULL,
    version VARCHAR(128),
    git_tag VARCHAR(128),
    beta BOOLEAN NOT NULL,
    published_at DATETIME,
    extraction_version INTEGER DEFAULT NULL,
    protocol_versions MEDIUMBLOB,
    FOREIGN KEY (provider_id) REFERENCES provider(id) ON UPDATE CASCADE ON DELETE CASCADE,
    FOREIGN KEY (gpg_key_id) REFERENCES gpg_key(id) ON UPDATE CASCADE ON DELETE CASCADE
);

-- Add foreign key constraint for provider.latest_version_id
ALTER TABLE provider
ADD FOREIGN KEY (latest_version_id) REFERENCES provider_version(id) ON UPDATE CASCADE ON DELETE SET NULL;

-- Provider version documentation table
CREATE TABLE IF NOT EXISTS provider_version_documentation (
    id INTEGER PRIMARY KEY AUTO_INCREMENT,
    provider_version_id INTEGER NOT NULL,
    name VARCHAR(128) NOT NULL,
    slug VARCHAR(128) NOT NULL,
    title VARCHAR(128),
    description MEDIUMBLOB,
    language VARCHAR(128) NOT NULL,
    subcategory VARCHAR(128),
    filename VARCHAR(128) NOT NULL,
    documentation_type VARCHAR(50) NOT NULL,
    content MEDIUMBLOB,
    FOREIGN KEY (provider_version_id) REFERENCES provider_version(id) ON UPDATE CASCADE ON DELETE CASCADE
);

-- Provider version binary table
CREATE TABLE IF NOT EXISTS provider_version_binary (
    id INTEGER PRIMARY KEY AUTO_INCREMENT,
    provider_version_id INTEGER NOT NULL,
    name VARCHAR(128) NOT NULL,
    operating_system VARCHAR(50) NOT NULL,
    architecture VARCHAR(50) NOT NULL,
    checksum VARCHAR(128) NOT NULL,
    FOREIGN KEY (provider_version_id) REFERENCES provider_version(id) ON UPDATE CASCADE ON DELETE CASCADE
);

-- Audit history table
CREATE TABLE IF NOT EXISTS audit_history (
    id INTEGER PRIMARY KEY AUTO_INCREMENT,
    timestamp DATETIME,
    username VARCHAR(128),
    action VARCHAR(50),
    object_type VARCHAR(128),
    object_id VARCHAR(128),
    old_value VARCHAR(128),
    new_value VARCHAR(128)
);
