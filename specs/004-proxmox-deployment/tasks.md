# Tasks: Proxmox LXC Automated Deployment

**Input**: Design documents from `/specs/004-proxmox-deployment/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/cli-interface.md

**Tests**: Integration tests are included as per Constitution requirement for infrastructure automation.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions
- Deployment scripts: `deploy/proxmox/`
- Library functions: `deploy/proxmox/lib/`
- Templates: `deploy/proxmox/templates/`
- Integration tests: `deploy/proxmox/tests/integration/`
- Documentation: `deploy/proxmox/docs/`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic directory structure

- [X] T001 Create deploy/proxmox/ directory structure as defined in plan.md
- [X] T002 [P] Create deploy/proxmox/lib/ directory for shared library functions
- [X] T003 [P] Create deploy/proxmox/templates/ directory for configuration templates
- [X] T004 [P] Create deploy/proxmox/tests/integration/ directory for integration tests
- [X] T005 [P] Create deploy/proxmox/tests/fixtures/ directory for test fixtures
- [X] T006 [P] Create deploy/proxmox/docs/ directory for documentation
- [X] T007 Create deploy/proxmox/.shellcheckrc for Bash linting configuration
- [X] T008 Create deploy/proxmox/README.md with project overview and quickstart link

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core library functions that ALL user stories depend on

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T009 Implement logging functions (log_info, log_success, log_warn, log_error, log_debug) in deploy/proxmox/lib/common.sh
- [X] T010 [P] Implement progress indicators (progress_start, progress_update, progress_complete, progress_fail) in deploy/proxmox/lib/common.sh
- [X] T011 [P] Implement input validation functions (validate_ip, validate_domain, validate_port, validate_email) in deploy/proxmox/lib/common.sh
- [X] T012 [P] Implement privilege check functions (check_root, check_proxmox) in deploy/proxmox/lib/common.sh
- [X] T013 Implement lock management functions (acquire_lock, release_lock, check_lock) in deploy/proxmox/lib/common.sh
- [X] T014 [P] Implement configuration management functions (load_config, save_config, update_config) in deploy/proxmox/lib/common.sh
- [X] T015 [P] Implement state management functions (get_state, set_state, update_deployed_version) in deploy/proxmox/lib/common.sh
- [X] T016 [P] Implement error handling functions (error, trap_errors, cleanup_on_error) in deploy/proxmox/lib/common.sh
- [X] T017 Create deployment configuration YAML template in deploy/proxmox/templates/deployment-config.yaml
- [X] T018 Create color codes and output formatting constants in deploy/proxmox/lib/common.sh

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Initial Deployment with Interactive Setup (Priority: P1) 🎯 MVP

**Goal**: Enable single-command deployment of Vikunja to Proxmox LXC with interactive prompts for database, network, and resource configuration. Complete deployment in <10 minutes.

**Independent Test**: Run vikunja-install.sh on fresh Proxmox node, complete interactive setup, verify Vikunja web interface accessible and functional.

### Integration Tests for User Story 1

- [X] T019 [P] [US1] Create test-fresh-install.sh integration test in deploy/proxmox/tests/integration/ for SQLite deployment scenario
- [X] T020 [P] [US1] Create test-postgresql-install.sh integration test in deploy/proxmox/tests/integration/ for PostgreSQL deployment scenario
- [X] T021 [P] [US1] Create test-validation-errors.sh integration test in deploy/proxmox/tests/integration/ for input validation scenarios
- [X] T022 [P] [US1] Create mock-proxmox-api.sh fixture in deploy/proxmox/tests/fixtures/ for CI testing

### Library Functions for User Story 1

- [X] T023 [P] [US1] Implement Proxmox CLI wrapper functions (pct_create, pct_start, pct_exec, pvesh_get) in deploy/proxmox/lib/proxmox-api.sh
- [X] T024 [P] [US1] Implement LXC container creation functions (create_container, configure_network, allocate_resources) in deploy/proxmox/lib/lxc-setup.sh
- [X] T025 [P] [US1] Implement container provisioning functions (install_dependencies, setup_go, setup_nodejs, clone_repository) in deploy/proxmox/lib/lxc-setup.sh
- [X] T026 [P] [US1] Implement database setup functions (setup_sqlite, setup_postgresql, setup_mysql, test_db_connection) in deploy/proxmox/lib/lxc-setup.sh
- [X] T027 [US1] Implement build functions (build_backend, build_frontend, build_mcp) in deploy/proxmox/lib/lxc-setup.sh
- [X] T028 [P] [US1] Implement systemd service creation functions (generate_systemd_unit, enable_service, start_service) in deploy/proxmox/lib/service-setup.sh
- [X] T029 [P] [US1] Implement nginx configuration functions (generate_nginx_config, enable_site, reload_nginx) in deploy/proxmox/lib/nginx-setup.sh
- [X] T030 [P] [US1] Implement health check functions (check_component_health, check_backend_health, check_frontend_health, check_mcp_health, check_database_connection) in deploy/proxmox/lib/health-check.sh

### Templates for User Story 1

- [X] T031 [P] [US1] Create vikunja-backend.service systemd unit template in deploy/proxmox/templates/
- [X] T032 [P] [US1] Create vikunja-mcp.service systemd unit template in deploy/proxmox/templates/
- [X] T033 [P] [US1] Create nginx-vikunja.conf nginx site configuration template in deploy/proxmox/templates/
- [X] T034 [P] [US1] Create health-check.sh script template (deployed to container) in deploy/proxmox/templates/

### Main Installation Script for User Story 1

- [X] T035 [US1] Create vikunja-install-main.sh main script with argument parsing and help text in deploy/proxmox/
- [X] T036 [US1] Implement interactive prompts section (instance ID, container ID, database type, network config, resources) in deploy/proxmox/vikunja-install-main.sh
- [X] T037 [US1] Implement configuration validation section in deploy/proxmox/vikunja-install-main.sh
- [X] T038 [US1] Implement pre-flight checks (resources, ports, DNS) in deploy/proxmox/vikunja-install-main.sh
- [X] T039 [US1] Implement deployment orchestration (container creation → provisioning → build → services → health check) in deploy/proxmox/vikunja-install-main.sh
- [X] T040 [US1] Implement post-deployment summary and next steps output in deploy/proxmox/vikunja-install-main.sh
- [X] T041 [US1] Implement cleanup on failure logic in deploy/proxmox/vikunja-install-main.sh
- [X] T042 [US1] Add exit code handling and error messages with remediation steps in deploy/proxmox/vikunja-install-main.sh

### Bootstrap Architecture Implementation

- [X] T042E1 [US1] Create vikunja-install-bootstrap.sh that downloads full installer package in deploy/proxmox/
- [X] T042E2 [US1] Rename vikunja-install.sh to vikunja-install-main.sh in deploy/proxmox/
- [X] T042E3 [US1] Copy bootstrap to vikunja-install.sh as curl-able entry point in deploy/proxmox/
- [X] T042E4 [US1] Update quickstart.md Option 2 to use git clone pattern in specs/004-proxmox-deployment/
- [X] T042E5 [US1] Update README.md Architecture section with bootstrap explanation in deploy/proxmox/

**Checkpoint**: User Story 1 complete - can deploy Vikunja to Proxmox with single curl command. Run integration tests to verify.

---

## Phase 3.5: Documentation Updates (Post-Bootstrap)

**Goal**: Update all documentation to reflect bootstrap architecture and validate curl-based installation works.

**Independent Test**: Execute curl command from quickstart.md on fresh Proxmox node, verify bootstrap downloads all files and launches installer.

- [X] T042E6 [DOC] Update ARCHITECTURE.md (when created) with bootstrap downloader pattern explanation in deploy/proxmox/docs/
- [X] T042E7 [DOC] Add troubleshooting section for bootstrap download failures in TROUBLESHOOTING.md in deploy/proxmox/docs/
- [X] T042E8 [TEST] Test curl-based installation from branch URL: bash <(curl -fsSL https://raw.githubusercontent.com/aroige/vikunja/004-proxmox-deployment/deploy/proxmox/vikunja-install.sh)
- [X] T042E9 [TEST] Validate all library files and templates are downloaded correctly in /tmp/vikunja-installer-*
- [X] T042E10 [DOC] Update contracts/cli-interface.md with bootstrap architecture notes if needed in specs/004-proxmox-deployment/

**Checkpoint**: Bootstrap architecture validated and documented. Ready for Phase 4 implementation.

---

## Phase 3.9: Regression Testing for Phase 1-3 Fixes

**Goal**: Validate all fixes from initial deployment testing before proceeding to Phase 4.

**Context**: During manual testing of Phase 1-3, nine critical issues were discovered and fixed:
1. Node.js version incompatibility (18.20.8 → 22.18.0 for Vite 7.1.10)
2. Repository URL using upstream instead of aroige fork (missing mcp-server/)
3. MCP build using pnpm instead of npm (package-lock.json vs pnpm-lock.yaml mismatch)
4. Service generation double-prefixing service_type ("vikunja-backend-blue" → "backend")
5. Frontend URL auto-detection not configured (backend advertising 127.0.0.1:3456)
6. Nginx configuration issues (API proxy double-slash, WebSocket path, upgrade mapping)
7. User registration disabled (missing VIKUNJA_SERVICE_ENABLEREGISTRATION)
8. Database configuration hardcoded to SQLite (PostgreSQL/MySQL configuration not passed to service generation)
9. Service enable/start functions failing despite systemctl success (tee pipe error handling)

**Independent Test**: Execute clean deployment from scratch using fixed scripts, verify all components start correctly with proper service names and API auto-detection working.

### Database Configuration Review (Pre-Testing)

- [ ] T042R0 [CODE] Review and fix database configuration environment variables for all database types (sqlite/postgresql/mysql)
  - Verify `generate_systemd_unit()` function signature accepts database parameters
  - Verify main script passes DATABASE_* variables to service generation
  - Verify SQLite configuration sets VIKUNJA_DATABASE_PATH correctly
  - Verify PostgreSQL/MySQL configuration sets HOST, PORT, DATABASE, USER, PASSWORD correctly
  - Verify default ports set correctly (PostgreSQL: 5432, MySQL: 3306)
  - Add database configuration validation in config validation section
  - Document expected environment variables for each database type

### Regression Test Tasks

- [ ] T042R1 [TEST] Execute clean deployment test on fresh LXC container using fixed vikunja-install-main.sh
- [ ] T042R2 [TEST] Verify Node.js 22 installation and Vite frontend build completes successfully
- [ ] T042R3 [TEST] Verify mcp-server/ directory exists and MCP server builds with `npm ci`
- [ ] T042R4 [TEST] Verify systemd services created with correct names (vikunja-backend-blue.service, not vikunja-backend-blue-blue.service)
- [ ] T042R5 [TEST] Verify backend VIKUNJA_SERVICE_FRONTENDURL set to http://<IP>:80 or http://<domain>:80
- [ ] T042R6 [TEST] Verify nginx configuration has correct API proxy (`location /api` without trailing slash)
- [ ] T042R7 [TEST] Verify WebSocket path is `/api/v1/ws` and upgrade mapping directive present
- [ ] T042R8 [TEST] Verify frontend loads and connects to API without manual URL configuration (window.API_URL should auto-detect)
- [ ] T042R9 [TEST] Verify user registration is enabled (VIKUNJA_SERVICE_ENABLEREGISTRATION=true in backend service)
- [ ] T042R10 [TEST] Verify health checks pass for all components (backend, frontend, MCP server)
- [ ] T042R11 [TEST] Verify user can successfully register a new account via /register endpoint
- [ ] T042R12 [TEST] Verify VIKUNJA_SERVICE_PUBLICURL is set (not the non-existent VIKUNJA_SERVICE_FRONTENDURL)
- [ ] T042R13 [TEST] Verify VIKUNJA_SERVICE_ROOTPATH set to /opt/vikunja in backend service
- [ ] T042R14 [TEST] Verify VIKUNJA_DATABASE_TYPE=sqlite and VIKUNJA_DATABASE_PATH=/opt/vikunja/vikunja.db
- [ ] T042R15 [TEST] Verify database file created at correct location /opt/vikunja/vikunja.db
- [ ] T042R16 [TEST] Verify /api/v1/info returns correct frontendUrl matching deployment domain/IP
- [ ] T042R17 [TEST] Verify database configuration passed correctly to systemd service (all DB types)
- [ ] T042R18 [TEST] Test SQLite deployment: verify VIKUNJA_DATABASE_TYPE=sqlite and VIKUNJA_DATABASE_PATH set
- [ ] T042R19 [TEST] Test PostgreSQL deployment: verify all VIKUNJA_DATABASE_* environment variables set correctly
- [ ] T042R20 [TEST] Test MySQL deployment: verify all VIKUNJA_DATABASE_* environment variables set correctly with port 3306
- [ ] T042R21 [TEST] Verify default database ports applied correctly (PostgreSQL: 5432, MySQL: 3306) when not explicitly provided
- [ ] T042R22 [TEST] Document any remaining issues in specs/004-proxmox-deployment/research.md Section 6.x

**Checkpoint**: All Phase 1-3 fixes validated via clean deployment. No regression issues found. Ready for Phase 4 implementation.

---

## Phase 4: User Story 2 - Seamless Updates from Main Branch (Priority: P1)

**Goal**: Enable zero-downtime updates via blue-green deployment with automatic rollback on failures. Updates complete in <5 minutes with <5 seconds downtime.

**Independent Test**: Deploy older version using vikunja-install.sh, run vikunja-update.sh, verify new version running with zero dropped connections and <5s downtime.

### Integration Tests for User Story 2

- [ ] T043 [P] [US2] Create test-update-cycle.sh integration test in deploy/proxmox/tests/integration/ for successful update scenario
- [ ] T044 [P] [US2] Create test-rollback.sh integration test in deploy/proxmox/tests/integration/ for automatic rollback on failure
- [ ] T045 [P] [US2] Create test-migration.sh integration test in deploy/proxmox/tests/integration/ for database migration execution
- [ ] T046 [P] [US2] Create test-concurrent-update.sh integration test in deploy/proxmox/tests/integration/ for lock mechanism validation

### Library Functions for User Story 2

- [ ] T047 [P] [US2] Implement blue-green deployment functions (determine_inactive_color, start_on_inactive_port, switch_traffic, stop_old_deployment) in deploy/proxmox/lib/blue-green.sh
- [ ] T048 [P] [US2] Implement git operations functions (check_for_updates, pull_latest, get_commit_hash, checkout_commit) in deploy/proxmox/lib/lxc-setup.sh
- [ ] T049 [P] [US2] Implement backup functions (create_pre_migration_backup, verify_backup_integrity, cleanup_old_backups) in deploy/proxmox/lib/backup-restore.sh
- [ ] T050 [US2] Implement migration execution functions (run_migrations, check_migration_status, restore_from_backup) in deploy/proxmox/lib/lxc-setup.sh
- [ ] T051 [US2] Implement rollback functions (rollback_to_blue, rollback_to_green, cleanup_failed_deployment) in deploy/proxmox/lib/blue-green.sh
- [ ] T052 [P] [US2] Implement nginx upstream switching functions (update_nginx_upstream, test_nginx_config) in deploy/proxmox/lib/nginx-setup.sh
- [ ] T053 [P] [US2] Implement health check validation functions (wait_for_healthy, retry_with_timeout, full_health_check) in deploy/proxmox/lib/health-check.sh

### Main Update Script for User Story 2

- [ ] T054 [US2] Create vikunja-update.sh main script with argument parsing in deploy/proxmox/
- [ ] T055 [US2] Implement update detection logic (check git for new commits, show changes) in deploy/proxmox/vikunja-update.sh
- [ ] T056 [US2] Implement pre-update backup creation in deploy/proxmox/vikunja-update.sh
- [ ] T057 [US2] Implement blue-green update orchestration (determine color → pull → build → migrate → start inactive → health check → switch → stop old) in deploy/proxmox/vikunja-update.sh
- [ ] T058 [US2] Implement automatic rollback on health check failure in deploy/proxmox/vikunja-update.sh
- [ ] T059 [US2] Implement version mismatch detection and prevention in deploy/proxmox/vikunja-update.sh
- [ ] T060 [US2] Implement concurrent update prevention (lock checking) in deploy/proxmox/vikunja-update.sh
- [ ] T061 [US2] Implement update summary output (time, downtime, versions, rollback availability) in deploy/proxmox/vikunja-update.sh

**Checkpoint**: User Story 2 complete - can update deployments with zero downtime and automatic rollback. Run integration tests to verify.

---

## Phase 5: User Story 3 - Configuration Management (Priority: P2)

**Goal**: Enable post-deployment configuration changes via interactive reconfigure command with validation and graceful service restarts.

**Independent Test**: Deploy Vikunja, run reconfigure command to change domain, verify new configuration applied and services restarted successfully.

### Integration Tests for User Story 3

- [ ] T062 [P] [US3] Create test-reconfigure.sh integration test in deploy/proxmox/tests/integration/ for domain change scenario
- [ ] T063 [P] [US3] Create test-database-migration-config.sh integration test in deploy/proxmox/tests/integration/ for database type change with data migration
- [ ] T064 [P] [US3] Create test-config-validation.sh integration test in deploy/proxmox/tests/integration/ for invalid configuration rejection

### Library Functions for User Story 3

- [ ] T065 [P] [US3] Implement configuration display functions (show_current_config, format_config_table) in deploy/proxmox/lib/common.sh
- [ ] T066 [P] [US3] Implement configuration prompts (prompt_for_domain, prompt_for_database, prompt_for_resources) in deploy/proxmox/lib/common.sh
- [ ] T067 [US3] Implement database migration functions (export_database, import_database, migrate_database_type) in deploy/proxmox/lib/backup-restore.sh
- [ ] T068 [US3] Implement graceful restart functions (graceful_restart_backend, graceful_restart_mcp, reload_nginx_config) in deploy/proxmox/lib/service-setup.sh

### Management Command Implementation for User Story 3

- [ ] T069 [US3] Create vikunja-manage.sh main script with subcommand routing in deploy/proxmox/
- [ ] T070 [US3] Implement reconfigure subcommand with interactive mode in deploy/proxmox/vikunja-manage.sh
- [ ] T071 [US3] Implement configuration validation before applying changes in deploy/proxmox/vikunja-manage.sh
- [ ] T072 [US3] Implement selective configuration update logic in deploy/proxmox/vikunja-manage.sh
- [ ] T073 [US3] Implement configuration apply with service restart orchestration in deploy/proxmox/vikunja-manage.sh

**Checkpoint**: User Story 3 complete - can reconfigure deployments without redeployment. Run integration tests to verify.

---

## Phase 6: User Story 4 - Health Monitoring and Status Checks (Priority: P2)

**Goal**: Provide comprehensive status monitoring with component health checks, resource usage, and remediation advice.

**Independent Test**: Deploy Vikunja, run status command (shows healthy), stop backend service, re-run status (shows failure with restart command).

### Integration Tests for User Story 4

- [ ] T074 [P] [US4] Create test-health-checks.sh integration test in deploy/proxmox/tests/integration/ for all-healthy scenario
- [ ] T075 [P] [US4] Create test-component-failure.sh integration test in deploy/proxmox/tests/integration/ for component failure detection
- [ ] T076 [P] [US4] Create test-resource-warnings.sh integration test in deploy/proxmox/tests/integration/ for disk/memory warnings

### Library Functions for User Story 4

- [ ] T077 [P] [US4] Implement resource monitoring functions (check_cpu_usage, check_memory_usage, check_disk_usage) in deploy/proxmox/lib/health-check.sh
- [ ] T078 [P] [US4] Implement process monitoring functions (check_systemd_status, check_process_uptime) in deploy/proxmox/lib/health-check.sh
- [ ] T079 [P] [US4] Implement remediation advice functions (suggest_restart, suggest_cleanup, suggest_troubleshooting) in deploy/proxmox/lib/health-check.sh
- [ ] T080 [US4] Implement status caching functions (cache_health_results, read_cached_health, is_cache_stale) in deploy/proxmox/lib/health-check.sh
- [ ] T081 [P] [US4] Implement status formatting functions (format_status_table, format_json_output, color_status_output) in deploy/proxmox/lib/common.sh

### Management Command Implementation for User Story 4

- [ ] T082 [US4] Implement status subcommand in deploy/proxmox/vikunja-manage.sh
- [ ] T083 [US4] Implement watch mode for continuous status monitoring in deploy/proxmox/vikunja-manage.sh
- [ ] T084 [US4] Implement JSON output format option for programmatic use in deploy/proxmox/vikunja-manage.sh
- [ ] T085 [US4] Implement last operations history display in deploy/proxmox/vikunja-manage.sh
- [ ] T086 [US4] Implement remediation suggestions based on detected issues in deploy/proxmox/vikunja-manage.sh

**Checkpoint**: User Story 4 complete - can monitor deployment health with detailed status. Run integration tests to verify.

---

## Phase 7: User Story 5 - Backup and Restore (Priority: P3)

**Goal**: Enable backup creation with encryption and restore capability for disaster recovery.

**Independent Test**: Deploy Vikunja, create tasks and files, create backup, make more changes, restore from backup, verify original state restored.

### Integration Tests for User Story 5

- [ ] T087 [P] [US5] Create test-backup-restore.sh integration test in deploy/proxmox/tests/integration/ for full backup/restore cycle
- [ ] T088 [P] [US5] Create test-backup-encryption.sh integration test in deploy/proxmox/tests/integration/ for encrypted backup
- [ ] T089 [P] [US5] Create test-backup-failure.sh integration test in deploy/proxmox/tests/integration/ for backup failure cleanup

### Library Functions for User Story 5

- [ ] T090 [P] [US5] Implement backup creation functions (backup_database, backup_files, backup_configuration, create_archive) in deploy/proxmox/lib/backup-restore.sh
- [ ] T091 [P] [US5] Implement backup verification functions (verify_checksum, validate_backup_structure, check_archive_integrity) in deploy/proxmox/lib/backup-restore.sh
- [ ] T092 [P] [US5] Implement backup encryption functions (encrypt_archive, decrypt_archive, prompt_for_password) in deploy/proxmox/lib/backup-restore.sh
- [ ] T093 [US5] Implement restore functions (extract_archive, restore_database, restore_files, restore_configuration) in deploy/proxmox/lib/backup-restore.sh
- [ ] T094 [US5] Implement backup management functions (list_backups, cleanup_old_backups, create_safety_backup) in deploy/proxmox/lib/backup-restore.sh
- [ ] T095 [P] [US5] Implement compression functions (compress_backup, calculate_compression_ratio) in deploy/proxmox/lib/backup-restore.sh

### Management Command Implementation for User Story 5

- [ ] T096 [US5] Implement backup subcommand in deploy/proxmox/vikunja-manage.sh
- [ ] T097 [US5] Implement restore subcommand with confirmation prompts in deploy/proxmox/vikunja-manage.sh
- [ ] T098 [US5] Implement backup list subcommand showing available backups in deploy/proxmox/vikunja-manage.sh
- [ ] T099 [US5] Implement automatic backup rotation (retention policy) in deploy/proxmox/vikunja-manage.sh
- [ ] T100 [US5] Implement safety backup before restore in deploy/proxmox/vikunja-manage.sh

**Checkpoint**: User Story 5 complete - can backup and restore deployments. Run integration tests to verify.

---

## Phase 8: Additional Management Commands

**Goal**: Complete the management command suite with remaining subcommands.

- [ ] T101 [P] Implement logs subcommand (view, follow, filter by component) in deploy/proxmox/vikunja-manage.sh
- [ ] T102 [P] Implement restart subcommand (graceful restart with health checks) in deploy/proxmox/vikunja-manage.sh
- [ ] T103 [P] Implement stop subcommand in deploy/proxmox/vikunja-manage.sh
- [ ] T104 [P] Implement start subcommand in deploy/proxmox/vikunja-manage.sh
- [ ] T105 [P] Implement uninstall subcommand with data preservation option in deploy/proxmox/vikunja-manage.sh
- [ ] T106 [P] Implement list subcommand to show all instances in deploy/proxmox/vikunja-manage.sh

---

## Phase 9: Edge Case Handling

**Goal**: Handle edge cases and error scenarios identified in spec.md.

- [ ] T107 [P] Implement network connectivity loss detection and recovery in deploy/proxmox/vikunja-install-main.sh
- [ ] T108 [P] Implement port conflict detection and resolution in deploy/proxmox/lib/proxmox-api.sh
- [ ] T109 [P] Implement deployment cancellation handling (Ctrl+C cleanup) in deploy/proxmox/vikunja-install-main.sh
- [ ] T110 [P] Implement disk space monitoring during operations in deploy/proxmox/lib/common.sh
- [ ] T111 [P] Implement multiple instance conflict avoidance in deploy/proxmox/lib/proxmox-api.sh
- [ ] T112 [P] Implement active upload detection before service restart in deploy/proxmox/vikunja-update.sh

---

## Phase 10: Documentation & Polish

**Goal**: Complete documentation and final polish.

- [ ] T113 [P] Create deploy/proxmox/docs/README.md comprehensive user guide (expand from quickstart.md)
- [ ] T114 [P] Create deploy/proxmox/docs/ARCHITECTURE.md documenting blue-green deployment pattern
- [ ] T115 [P] Create deploy/proxmox/docs/TROUBLESHOOTING.md with common issues and solutions
- [ ] T116 [P] Create deploy/proxmox/docs/DEVELOPMENT.md with testing and contribution guidelines
- [ ] T117 [P] Add inline documentation and comments to all library functions
- [ ] T118 [P] Add script version constants and --version flag to all main scripts
- [ ] T119 [P] Add shellcheck validation to all Bash scripts
- [ ] T120 Run integration test suite end-to-end on actual Proxmox node
- [ ] T121 Validate against quickstart.md step-by-step instructions
- [ ] T122 [P] Create deploy/proxmox/CHANGELOG.md documenting initial release
- [ ] T123 Security review: credentials handling, permissions, input validation
- [ ] T124 Performance testing: verify <10min deploy, <5min update, 99.9% uptime targets

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - **BLOCKS all user stories**
- **User Story 1 (Phase 3)**: Depends on Foundational phase - Core deployment capability
- **User Story 2 (Phase 4)**: Depends on Foundational phase - Can run parallel with US1 BUT requires existing deployment to test
- **User Story 3 (Phase 5)**: Depends on US1 completion (needs deployment to reconfigure)
- **User Story 4 (Phase 6)**: Depends on US1 completion (needs deployment to monitor)
- **User Story 5 (Phase 7)**: Depends on US1 completion (needs deployment to backup)
- **Phase 8**: Depends on Phase 2 (uses foundational functions)
- **Phase 9**: Can integrate throughout phases (add to relevant scripts)
- **Documentation (Phase 10)**: Can proceed in parallel with implementation

### User Story Dependencies

```
Foundation (Phase 2) ─┬─→ User Story 1 (P1) ──┬─→ User Story 3 (P2)
                      │                        ├─→ User Story 4 (P2)
                      │                        └─→ User Story 5 (P3)
                      └─→ User Story 2 (P1)*

* US2 can be implemented in parallel with US1, but requires a deployment (from US1) to test updates
```

### Recommended Execution Order

**For MVP (Minimum Viable Product)**:
1. Phase 1: Setup
2. Phase 2: Foundational (CRITICAL - blocks everything)
3. Phase 3: User Story 1 (Initial Deployment)
4. **STOP HERE** - You now have a working deployment system
5. Test, demo, validate with users

**For Full Feature Set**:
1. Complete MVP (Phases 1-3)
2. Phase 4: User Story 2 (Updates) - P1 priority
3. Phase 5: User Story 3 (Configuration) - P2 priority
4. Phase 6: User Story 4 (Monitoring) - P2 priority
5. Phase 7: User Story 5 (Backup) - P3 priority
6. Phase 8-10: Management commands, edge cases, documentation

### Parallel Opportunities

**Within Phases**:
- All tasks marked [P] within a phase can run in parallel
- Setup tasks (T002-T006): All parallel
- Foundational library functions (T010-T016, T018): Can run in parallel
- Templates (T031-T034): Can run in parallel
- Integration tests: Can run in parallel after implementation

**Across User Stories** (if team capacity allows):
- After Foundation complete, US1, US2, US3, US4, US5 can be implemented by different developers
- However, US2-US5 need US1 for testing (requires actual deployment)
- Recommended: Complete US1 first, then parallelize US2-US5

**Documentation**:
- Phase 10 documentation (T113-T117, T122) can run in parallel with implementation

---

## Parallel Example: Foundation Phase

```bash
# After T009 (logging functions) complete, launch in parallel:
- T010: Progress indicators
- T011: Validation functions  
- T012: Privilege checks
- T014: Configuration management
- T015: State management
- T016: Error handling
- T018: Color codes

# After T017 (config template) complete:
- T031-T034: All templates in parallel
```

---

## Parallel Example: User Story 1

```bash
# Integration tests (can run in parallel, but may need deployment setup):
- T019: SQLite test
- T020: PostgreSQL test
- T021: Validation test

# Library functions (can run in parallel):
- T023: Proxmox CLI wrappers
- T024: LXC creation
- T025: Provisioning
- T026: Database setup
- T028: Systemd service creation
- T029: Nginx configuration
- T030: Health checks

# Templates (can run in parallel):
- T031: Backend systemd unit
- T032: MCP systemd unit
- T033: Nginx config
- T034: Health check script

# Main script (sequential - depends on library functions):
- T035 → T036 → T037 → T038 → T039 → T040 → T041 → T042
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

**Goal**: Working deployment system in minimum time

1. ✅ Complete Phase 1: Setup (T001-T008)
2. ✅ Complete Phase 2: Foundational (T009-T018) - **CRITICAL PATH**
3. ✅ Complete Phase 3: User Story 1 (T019-T042)
4. **VALIDATE**: Run T019-T022 integration tests
5. **DEMO**: Deploy Vikunja to Proxmox in <10 minutes
6. **FEEDBACK**: Gather user feedback before proceeding

**Estimated Time**: 4-5 days

### Incremental Delivery

**After MVP**:
1. Add User Story 2 (Updates) → **Value**: Zero-downtime updates
2. Add User Story 3 (Reconfigure) → **Value**: Post-deployment flexibility
3. Add User Story 4 (Monitoring) → **Value**: Operational visibility
4. Add User Story 5 (Backup) → **Value**: Disaster recovery
5. Polish & Documentation → **Value**: Production readiness

**Each increment**:
- Is independently testable
- Adds specific user value
- Can be deployed/demoed
- Builds on previous work

**Total Estimated Time**: 8-12 days

### Parallel Team Strategy

With 3 developers after Foundation (Phase 2) complete:

**Developer A**: User Story 1 (deployment) - 2-3 days
**Developer B**: Start documentation (Phase 10 docs) - ongoing
**Developer C**: Start User Story 2 (updates) - 2-3 days (needs US1 for testing)

Then continue with US3, US4, US5 in priority order.

---

## Testing Strategy

### Integration Testing

**When**: After each user story phase completes
**How**: Run integration tests for that story
**Validate**: Independent test criteria from spec.md

**Test Execution**:
```bash
# User Story 1
cd deploy/proxmox/tests/integration
./test-fresh-install.sh
./test-postgresql-install.sh
./test-validation-errors.sh

# User Story 2
./test-update-cycle.sh
./test-rollback.sh
./test-migration.sh

# Continue for all user stories...
```

### Manual Testing

**When**: Before marking user story complete
**How**: Follow quickstart.md guide manually
**Validate**: All acceptance scenarios from spec.md pass

### CI Integration

**Setup**: Mock Proxmox API (deploy/proxmox/tests/fixtures/mock-proxmox-api.sh)
**Run**: Subset of integration tests that don't require real Proxmox
**Validate**: Script syntax (shellcheck), basic functionality

---

## Task Summary

**Total Tasks**: 124
- Phase 1 (Setup): 8 tasks
- Phase 2 (Foundational): 10 tasks  
- Phase 3 (US1 - Deploy): 24 tasks
- Phase 4 (US2 - Update): 19 tasks
- Phase 5 (US3 - Reconfigure): 13 tasks
- Phase 6 (US4 - Monitor): 15 tasks
- Phase 7 (US5 - Backup): 14 tasks
- Phase 8 (Management): 6 tasks
- Phase 9 (Edge Cases): 6 tasks
- Phase 10 (Documentation): 12 tasks

**Parallel Tasks**: 76 tasks marked [P] can run in parallel (61%)
**Critical Path**: Phase 2 (Foundational) → Phase 3 (US1) → Phases 4-7 can parallelize

**MVP Scope**: Phases 1-3 (42 tasks, ~4-5 days)
**Full Feature**: All phases (124 tasks, ~8-12 days)

---

## Success Criteria Validation

Each user story maps to success criteria from spec.md:

- **US1 (Deploy)** → SC-001: <10 min deployment, SC-008: 95% success rate
- **US2 (Update)** → SC-002: <5 min update, SC-003: 99.9% uptime, SC-004: <2 min rollback
- **US3 (Reconfigure)** → SC-009: <10 sec service interruption
- **US4 (Monitor)** → SC-005: <10 sec health check, SC-010: Real-time feedback
- **US5 (Backup)** → SC-006: <5 min backup, SC-007: 5 instances support

All tasks contribute to achieving these measurable outcomes.

---

## Notes

- **[P] tasks**: Different files, no shared state, can run in parallel
- **[Story] labels**: Enable tracking which tasks belong to which user story
- **File paths**: Included in every task description for clarity
- **Integration tests**: Match the 5 user stories from spec.md
- **Checkpoints**: After each phase, validate story works independently
- **Dependencies**: Foundation blocks all stories; US1 required for US2-US5 testing
- **MVP strategy**: Phases 1-3 deliver working deployment system
- **Incremental**: Each story adds value without breaking previous work
- **Exit codes**: Follow contract in contracts/cli-interface.md (0-20 range)
- **Constitution**: Integration testing appropriate for infrastructure automation
