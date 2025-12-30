# Terrareg Go Refactoring and Implementation Plan

## Overview
This plan organizes the comprehensive refactoring tasks for the Terrareg Go implementation into logical phases that build on each other. The focus is on cleaning up architectural issues, removing legacy code, and implementing missing functionality while maintaining DDD principles.

**CRITICAL CONSTRAINT**: All functionality must remain identical to the Python application:
- **API Compatibility**: All APIs must match Python implementation exactly
- **Database Schema**: DB schema must remain identical to Python version
- **Functional Parity**: End-user experience must be indistinguishable from Python
- **Behavioral Consistency**: All business logic must produce same results as Python

This is a refactoring effort, not a feature change. The goal is cleaner architecture while maintaining 100% backward compatibility.

## Phase 1: Foundation and Analysis
**Goal**: Understand current state and create comprehensive documentation

### 1.1 Entity Analysis and Documentation
- **Tasks**:
  - Catalog ALL entities in the codebase (domain, infrastructure, application layers)
  - Document purpose, layer, and DDD compliance for each entity
  - Identify overlaps and separation of concerns issues
  - Create entity relationship map

### 1.2 TODO and Legacy Investigation
- **Tasks**:
  - Find and catalog all TODO comments and legacy_stub references
  - Analyze Python vs Go implementation gaps
  - Create comprehensive development prompt instructions based on migration patterns

### 1.3 Critical Issues Assessment
- **Focus Areas**:
  - Module ingestion and re-import logic
  - Graph data format mismatches
  - Missing analytics and audit implementations

---

## Phase 2: Authentication System Overhaul
**Goal**: Complete redesign of authentication to follow clean DDD principles

### 2.1 Current Auth Analysis and Cleanup
- **Tasks**:
  - Analyze user models: determine necessity and DB storage requirements
  - Audit auth methods for duplication and remove legacy stubs
  - Evaluate AuthFactory: DDD compliance and naming
  - Implement missing `loadUserGroupsForAuthMethod` method
  - Clean up admin auth method unnecessary logic

### 2.2 New Authentication Architecture Design
- **Proposed Structure**:
  ```
  BaseAuthMethod interface:
    - IsEnabled() bool
    - Authenticate(credentials) (BaseAuthenticatedUser, error)
    - GetType() AuthMethodType

  BaseAuthenticatedUser interface:
    - GetUsername() string
    - GetGroups() []UserGroup
    - HasPermission(permission) bool
    - GetTokenType() TokenType
  ```

### 2.3 Context-Based Authentication Implementation
- **Tasks**:
  - Replace mutex locks with context-based user storage
  - Implement different user types: AdminUser, APIKeyUser, UnauthenticatedUser
  - Refactor AuthFactory to use dependency injection properly
  - Remove auth method attribute setting (current anti-pattern)

---

## Phase 3: Configuration and Storage Architecture
**Goal**: Consolidate and clarify core service architectures

### 3.1 Configuration System Cleanup
- **Tasks**:
  - Audit configuration layers: infra/domain, repository, service, DTO
  - Eliminate overlap between configuration components
  - Centralize ALL configuration defaults in domain models
  - Remove service-level hardcoded defaults (e.g., exampleFileExtensions in processor)

### 3.2 Storage Architecture Consolidation
- **Clarify Responsibilities**:
  - **Temporary Storage**: For module/provider analysis (short-lived)
  - **Data Storage**: For archives (local or S3, long-lived)
  - **Storage Service**: Operations on stored data
  - **Storage Factory**: Creation of storage instances

- **Tasks**:
  - Clarify storage service vs storage factory responsibilities
  - Consolidate storage architecture with clear separation
  - Optimize GenerateModuleSourceCommand to use pre-stored archives

---

## Phase 4: Module System and Data Flow
**Goal**: Fix core module functionality and data flow

### 4.1 Module Ingestion Fixes
- **Critical Issues**:
  - Ensure re-import of existing versions creates new records (not updates)
  - Fix module version ID management (prevent ID reuse)
  - Verify module_details_id population during import
  - Test transaction safety during module processing

### 4.2 Data Model Corrections
- **Tasks**:
  - Move ImportModuleVersionRequest to domain model (remove from infrastructure)
  - Fix graph data format to match expected specification
  - Ensure proper DDD layer separation for all module-related models

### 4.3 Performance Optimization
- **Tasks**:
  - Optimize GenerateModuleSourceCommand to use storage service
  - Eliminate on-the-fly archive generation
  - Implement proper caching for frequently accessed modules

---

## Phase 5: Analytics and Audit Implementation
**Goal**: Implement missing business functionality

### 5.1 Analytics System
- **Components**:
  - Module download statistics
  - Usage metrics tracking
  - Performance analytics
  - Reporting interfaces

### 5.2 Audit History System
- **Components**:
  - Change tracking for all entities
  - User action logging
  - Module version history
  - Compliance reporting

---

## Phase 6: Final Integration and Testing
**Goal**: Ensure all components work together properly

### 6.1 Integration Testing
- **Areas**:
  - End-to-end module import workflow
  - Authentication flow with all user types
  - Configuration loading and validation
  - Storage operations across different backends

### 6.2 Documentation Completion
- **Tasks**:
  - Update AI_DEVELOPMENT_GUIDE.md with all architectural decisions
  - Create comprehensive testing guides
  - Document migration patterns for future development

### 6.3 Integration Test Coverage Matrix

The following integration tests exist in Python but need implementation in Go. Status tracking for each test suite:

#### ✅ Completed Test Suites

| Test Suite | Go Test File | Python Reference | Status |
|------------|--------------|------------------|--------|
| Submodule | `test/integration/model/submodule_test.go` | `test/integration/terrareg/models/test_submodule.py` | ✅ 11 tests passing |
| Example File | `test/integration/model/example_file_test.go` | `test/integration/terrareg/models/test_example_file.py` | ✅ 12 tests passing |
| Module Provider Redirect | `test/integration/model/module_provider_redirect_test.go` | `test/integration/terrareg/models/test_module_provider_redirect.py` | ✅ 15+ tests passing |
| Module Details | `test/integration/model/module_details_test.go` | `test/integration/terrareg/models/test_module_details.py` | ✅ 14 tests passing |
| Analytics | `test/integration/analytics/analytics_test.go` | Various in `test/integration/terrareg/` | ✅ 11 tests passing |
| Module Version (Domain) | `internal/domain/module/model/module_version_test.go` | `test/integration/terrareg/models/test_module_version.py` | ✅ 40+ tests passing |
| Module Version (Integration) | `test/integration/model/module_version_integration_test.go` | `test/integration/terrareg/models/test_module_version.py` | ✅ 6 tests passing |
| Module Search | `test/integration/search/module_search_test.go` | `test/integration/terrareg/module_search/test_search_module_providers.py` | ✅ 9 tests (26 subtests) passing |
| Storage Workflow | `test/integration/storage/*.go` | Various in `test/unit/terrareg/test_file_storage.py` | ✅ 8 tests (30+ subtests) passing |
| Module Extractor | `test/integration/module/module_extractor_test.go` | `test/integration/terrareg/module_extractor/test_process_upload.py` | ✅ 35 tests passing |
| Module Provider | `test/integration/model/module_provider_test.go` | `test/integration/terrareg/models/test_module_provider.py` | ✅ 30+ tests passing |
| Enums | `test/integration/model/enums_test.go` | Various enum tests in Python | ✅ 7 tests passing |
| Git Provider | `test/integration/model/git_provider_test.go` | `test/integration/terrareg/models/test_git_provider.py` | ✅ Implemented |
| Namespace | `test/integration/model/namespace_test.go` | `test/integration/terrareg/models/test_namespace.py` | ✅ Implemented |
| Provider Category | `test/integration/model/provider_category_test.go` | `test/integration/terrareg/models/test_provider_category.py` | ✅ Implemented |
| Provider Logo | `test/integration/model/provider_logo_test.go` | `test/integration/terrareg/models/test_provider_logo.py` | ✅ Implemented |
| Session | `test/integration/model/session_test.go` | `test/integration/terrareg/models/test_session.py` | ✅ Implemented |
| User Group | `test/integration/model/user_group_test.go` | `test/integration/terrareg/models/test_user_group.py` | ✅ Implemented |
| User Group NS Permission | `test/integration/model/user_group_namespace_permission_test.go` | `test/integration/terrareg/models/test_user_group_namespace_permission.py` | ✅ Implemented |

#### ❌ Missing Test Suites (High Priority)

| Test Suite | Python Reference | Complexity | Blockers | Notes |
|------------|------------------|------------|----------|-------|
| **Provider Extractor** | `test/integration/terrareg/test_provider_extractor.py` | High | Provider package schema fix + GitHub/GPG mocks | Tests provider release extraction, GPG verification |
| **Provider Model** | `test/integration/terrareg/models/test_provider.py` | High | Provider package schema fix | Basic provider CRUD operations |
| **Provider Version** | `test/integration/terrareg/models/test_provider_version.py` | High | Provider package schema fix | Provider version lifecycle |
| **Provider Version Binary** | `test/integration/terrareg/models/test_provider_version_binary.py` | High | Provider package schema fix | Binary storage/retrieval by OS/arch |
| **Provider Version Documentation** | `test/integration/terrareg/models/test_provider_version_documentation.py` | High | Provider package schema fix | Provider documentation handling |
| **Provider Category Factory** | `test/integration/terrareg/models/test_provider_category_factory.py` | Medium | Provider package schema fix | Provider category creation logic |
| **Provider Source Factory** | `test/integration/terrareg/models/test_provider_source_factory.py` | Medium | Provider package schema fix | Provider source management |
| **Provider Source (GitHub)** | `test/integration/terrareg/models/provider_source/test_github_provider_source.py` | High | GitHub API mocking | GitHub-specific provider source logic |
| **Provider Source (Base)** | `test/integration/terrareg/models/provider_source/test_base_provider_source.py` | Medium | - | Base provider source functionality |

#### ❌ Missing Test Suites (Medium Priority)

| Test Suite | Python Reference | Complexity | Blockers | Notes |
|------------|------------------|------------|----------|-------|
| **Module Version (Integration)** | `test/integration/terrareg/models/test_module_version.py` | Medium | None (domain tests exist) | Integration-level module version tests |
| **Module Provider (Integration)** | `test/integration/terrareg/models/test_module_provider.py` | Medium | None (partial exists) | Full module provider integration tests |
| **Example (Model)** | `test/integration/terrareg/models/test_example.py` | Medium | None | Example model (not just example file) |
| **Base Submodule** | `test/integration/terrareg/models/test_base_submodule.py` | Low | None | Base submodule functionality |

#### ❌ Missing Test Suites (API/Handler Level)

| Test Suite | Python Reference | Complexity | Notes |
|------------|------------------|------------|-------|
| **Module Upload Endpoints** | `test/unit/terrareg/server/test_module_version_upload.py` | High | Tests module upload API handlers |
| **Module Version Import** | `test/unit/terrareg/server/test_api_module_version_import.py` | High | Tests module import from Git |
| **Module Version Download** | `test/unit/terrareg/server/test_api_module_version_download.py` | Medium | Tests module download endpoints |
| **Module Version Details** | `test/unit/terrareg/server/test_api_module_version_details.py` | Medium | Tests version details API |
| **Module Version Create** | `test/unit/terrareg/server/test_api_module_version_create.py` | High | Tests version creation API |
| **Module List** | `test/unit/terrareg/server/test_api_module_list.py` | Medium | Tests module listing |
| **Namespace Providers** | `test/unit/terrareg/server/test_api_namespace_providers.py` | Medium | Tests namespace provider listing |
| **Provider List** | `test/unit/terrareg/server/test_api_provider_list.py` | Medium | Tests provider listing API |
| **Provider Versions** | `test/unit/terrareg/server/test_api_provider_versions.py` | High | Tests provider version listing |
| **Provider V2 API** | `test/unit/terrareg/server/api/terraform/v2/` | High | Full Terraform v2 provider protocol |
| **Provider Categories API** | `test/unit/terrareg/server/api/terraform/v2/test_provider_categories.py` | Medium | Provider category endpoints |
| **GPG Keys API** | `test/unit/terrareg/server/api/terraform/v2/test_gpg_keys.py` | Medium | GPG key management |
| **GitHub Webhook** | `test/unit/terrareg/server/test_api_module_provider_github_hook.py` | High | GitHub webhook handling |
| **Bitbucket Webhook** | `test/integration/terrareg/server/test_api_module_provider_bitbucket_hook_integration.py` | High | Bitbucket webhook handling |
| **Search Filters** | `test/integration/terrareg/module_search/test_get_search_filters.py` | Medium | Search filter validation |

#### ❌ Missing Test Suites (Utility/Feature)

| Test Suite | Python Reference | Complexity | Notes |
|------------|------------------|------------|-------|
| **Provider Documentation Type** | `test/integration/terrareg/test_provider_documentation_type.py` | Low | Provider documentation types enum |
| **Provider Source Type** | `test/integration/terrareg/test_provider_source_type.py` | Low | Provider source types enum |
| **Provider Tier** | `test/integration/terrareg/test_provider_tier.py` | Low | Provider tiers (community, official, partner) |
| **Repository Kind** | `test/integration/terrareg/test_repository_kind.py` | Low | Repository kinds (github, gitlab, etc.) |
| **Repository Release Metadata** | `test/integration/terrareg/test_repository_release_metadata.py` | Medium | Release metadata parsing |
| **Registry Resource Type** | `test/integration/terrareg/test_registry_resource_type.py` | Low | Registry resource types enum |

#### ❌ Missing Test Suites (Selenium/UI)

| Test Suite | Python Reference | Complexity | Notes |
|------------|------------------|------------|-------|
| **Create Namespace** | `test/selenium/test_create_namespace.py` | High | UI-based namespace creation |
| **Create Module Provider** | `test/selenium/test_create_module_provider.py` | High | UI-based module creation |
| **Login** | `test/selenium/test_login.py` | High | UI authentication flow |
| **Namespace List** | `test/selenium/test_namespace_list.py` | Medium | Namespace listing UI |
| **Homepage** | `test/selenium/test_homepage.py` | Low | Homepage rendering |
| **Provider Search** | `test/selenium/test_provider_search.py` | Medium | Provider search UI |

---

## Implementation Priority Order

### **Immediate (High Impact, Low Risk)**
1. Entity analysis and documentation
2. TODO investigation and prompt creation
3. Configuration default centralization
4. Module ingestion fixes

### **Medium Priority (High Impact, Medium Risk)**
5. Authentication system analysis
6. Storage architecture clarification
7. Graph data format fixes
8. GenerateModuleSourceCommand optimization

### **Lower Priority (Critical but Complex)**
9. Complete authentication redesign
10. Analytics system implementation
11. Audit history system implementation

---

## Test Implementation Priority Order

Based on blockers and dependencies, here's the recommended order for implementing missing integration tests:

### **Phase 1: No Blockers (Can Start Immediately)**
1. ~~**Module Extractor Tests**~~ - ✅ **COMPLETED** (35 tests)
2. ~~**Module Version Integration Tests**~~ - ✅ **COMPLETED** (6 tests)
3. ~~**Module Provider Integration Tests**~~ - ✅ **COMPLETED** (30+ tests)
4. ~~**Enum Type Tests**~~ - ✅ **COMPLETED** (7 tests)
5. **Example Model Tests** - Simple model tests (NOT example_file - that's done)
6. **Base Submodule Tests** - Low complexity
7. **Utility/Feature Tests** - Provider tier (partial), source type, documentation type, repository kind, registry resource type, repository release metadata
8. **Search Filter Tests** - Search filter validation

### **Phase 2: After Provider Package Schema Fix**
1. **Provider Model Tests** - All provider-related tests
2. **Provider Version Tests** - Version lifecycle
3. **Provider Version Binary Tests** - Binary handling
4. **Provider Version Documentation Tests** - Documentation handling
5. **Provider Category/Source Factory Tests** - Factory pattern tests
6. **Provider Source Tests** (Base and GitHub)

### **Phase 3: API/Handler Level Tests**
1. **Module Upload Endpoints** - Critical for module publishing
2. **Module Version Import/Download** - Core module operations
3. **Provider V2 API Tests** - Full Terraform provider protocol
4. **Webhook Tests** - GitHub/Bitbucket integration

### **Phase 4: Selenium/UI Tests (Optional)**
1. **UI-based tests** - Lower priority, can be deferred

---

## Critical Files to Examine/Modify

### **Entity and Configuration**
- `/internal/domain/` - All entity models
- `/internal/infrastructure/config/` - Configuration loading
- `/internal/domain/config/` - Configuration models

### **Authentication**
- `/internal/domain/auth/` - Auth domain models
- `/internal/infrastructure/auth/` - Auth implementations
- `/internal/infrastructure/container/container.go` - DI container

### **Module System**
- `/internal/domain/module/` - Module domain layer
- `/internal/application/command/` - Import commands
- `/internal/infrastructure/storage/` - Storage implementations

### **Missing Implementations**
- `/internal/analytics/` - Currently empty
- `/internal/audit/` - Currently empty

---

## Success Metrics

### **Technical**
- All entities follow clear DDD principles
- No architectural violations or layer mixing
- Zero TODO comments or legacy stubs remaining
- Complete test coverage for all refactored components
- **Test Coverage**: Target 80%+ integration test parity with Python implementation

### **Functional**
- Module import/reimport works correctly
- Authentication is clean and extensible
- Configuration is centralized and type-safe
- Storage architecture is clear and efficient
- Analytics and audit systems are operational

### **Test Coverage Metrics** (Current Status - Updated 2025-12-29)

| Category | Python Tests | Go Tests | Coverage |
|----------|--------------|----------|----------|
| Model Tests | 22 test files | 20 test files | ~91% |
| Module Integration | 8 test files | 9 test files | ~100% |
| Provider Integration | 10 test files | 1 test file (blocked) | ~10% |
| API/Handler Tests | 30+ test files | Partial | ~20% |
| Storage Tests | Complete | Complete | 100% |
| Search Tests | Complete | Complete | 100% |
| **Overall** | **70+ test files** | **40+ test files** | **~60%** |

**Recent Additions (Dec 2025)**:
- Module Extractor integration tests: 35 tests covering ZIP/TAR.GZ extraction, metadata processing, external tool mocks, hidden metadata files, non-root directories, .terraformignore parsing, non-root repo with .tfignore, multiple .tfignore files, wildcard patterns
- Module Version integration tests: 6 tests covering version lifecycle, publishing, beta detection
- Module Provider expanded tests: +3 tests for multiple versions, verified status, git configuration
- Enum type tests: 7 tests covering ProviderTier, ProviderSourceType, NamespaceType, ProviderBinaryOperatingSystemType, ProviderBinaryArchitectureType
- **Python test references**: Added Python test references to all Golang integration tests for traceability
- Test infrastructure: Created `archive_helpers.go` and `module_test_helpers.go` with helper functions for creating test modules, archives, and database fixtures
- SystemCommandService: Created abstraction layer for external tool mocking (terraform-docs, tfsec, infracost, terraform, git)
- Bug fix: Fixed `GetPathspecFilter` to use `strings.Split` instead of `filepath.SplitList` for proper .terraformignore parsing

**Remaining blockers**:
- Provider package schema mismatch still blocks provider-related integration tests

---

## Known Blockers and Critical Issues

### Provider Package Schema Mismatch

**Status**: ❌ **HIGH PRIORITY BLOCKER**

**Impact**: Blocks all provider-related integration tests (10+ test suites)

**Issue**: The provider package models (`internal/infrastructure/persistence/sqldb/provider/models.go`) do not match the actual database schema from `/app/modules.db`.

**Affected Files**:
- `/app/terrareg-go/internal/infrastructure/persistence/sqldb/provider/models.go`
- `/app/terrareg-go/internal/infrastructure/persistence/sqldb/provider/repository.go`

**Schema Mismatches**:

| Table | Field | Provider Model | Actual DB | Status |
|-------|-------|----------------|-----------|--------|
| provider | name | `size:255` | `VARCHAR(128)` | ❌ Size mismatch |
| provider | tier | `size:50` | `VARCHAR(9)` | ❌ Size mismatch |
| provider | description | `type:text` | `VARCHAR(1024)` | ❌ Type mismatch |
| provider | created_at/updated_at | Present | Missing | ❌ Extra fields |
| provider_version | gpg_key_id | `*int` (nullable) | `INTEGER` (not null) | ❌ Nullability mismatch |
| provider_version | published_at | `*time.Time` | `DATETIME` (as string) | ❌ Type mismatch |
| provider_version | protocol_versions | `type:text` | `BLOB` | ❌ Type mismatch |
| provider_version_binary | version_id | `VersionID` | `provider_version_id` | ❌ Name mismatch |
| provider_version_binary | checksum | Missing | `checksum` | ❌ Missing field |
| gpg_key | provider_id | Present | Missing | ❌ Wrong foreign key |
| gpg_key | namespace_id | Missing | `INTEGER` (not null) | ❌ Missing field |
| gpg_key | ascii_armor | `type:longtext` | `BLOB` | ❌ Type mismatch |

**Resolution Required**:
1. Update provider models to match actual database schema
2. Update model-to-domain conversion functions
3. Update domain-to-model conversion functions
4. Update repository queries
5. Re-run provider search integration tests

**Estimated Effort**: 4-8 hours

---

### **Maintainability**
- Clear documentation for future developers
- Consistent patterns across all components
- Proper separation of concerns
- Easy to extend and modify