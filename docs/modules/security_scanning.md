
# Security Scanning

Security scanning of modules is performed automatically, without any additional setup.

To disable security scanning results, set the [ENABLE_SECURITY_SCANNING](../CONFIG.md#enable_security_scanning) configuration.

This configuration does not change whether security scans are performing during module indexing, instead, it disables the display of security vulnerabilities in the UI. This means that if the configuration is reverted in future, the security issues are immediately displayed without having to re-index modules.
