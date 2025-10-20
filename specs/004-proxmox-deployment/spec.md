# Feature Specification: Proxmox LXC Automated Deployment

**Feature Branch**: `004-proxmox-deployment`  
**Created**: 2025-10-19  
**Status**: Draft  
**Input**: User description: "Create an automated deployment strategy for Vikunja on Proxmox LXC containers with interactive setup, easy updates, and high uptime during development"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Initial Deployment with Interactive Setup (Priority: P1)

A system administrator wants to deploy Vikunja to their Proxmox cluster for the first time. They run a single command that guides them through selecting database type (SQLite, PostgreSQL, or MySQL), configuring the public domain/URL, setting resource allocations, and completing the installation. Within 5-10 minutes, they have a fully functional Vikunja instance accessible via their configured domain.

**Why this priority**: This is the foundational capability - without easy initial deployment, the system cannot be used. This delivers immediate value by reducing deployment time from hours to minutes.

**Independent Test**: Can be fully tested by running the deployment script on a fresh Proxmox node and accessing the Vikunja web interface at the configured domain. Success is verified when a user can log in and create their first task.

**Acceptance Scenarios**:

1. **Given** a Proxmox node with internet access, **When** administrator runs the deployment script and selects SQLite as database, **Then** a complete Vikunja instance (backend, frontend, MCP server) is deployed and accessible
2. **Given** a Proxmox node with existing PostgreSQL server, **When** administrator runs deployment script and provides PostgreSQL connection details, **Then** Vikunja connects to external database and initializes schema
3. **Given** administrator enters an invalid domain name during setup, **When** validation occurs, **Then** system prompts for correction before proceeding
4. **Given** insufficient resources on Proxmox node, **When** deployment attempts to create container, **Then** system warns administrator and suggests minimum requirements
5. **Given** deployment completes successfully, **When** administrator navigates to configured domain, **Then** Vikunja login page loads within 3 seconds

---

### User Story 2 - Seamless Updates from Main Branch (Priority: P1)

A system administrator receives notification that new features have been merged to main branch. They run a single update command that automatically pulls the latest changes, runs database migrations, and performs a rolling restart with health checks to ensure zero downtime. The update completes in under 5 minutes with automatic rollback if any component fails health checks.

**Why this priority**: Easy updates are critical for maintaining security and accessing new features. Without this, administrators may defer updates, leading to technical debt and security vulnerabilities. This is P1 because continuous deployment is core to the development workflow requirement.

**Independent Test**: Can be tested by deploying a known older version, running the update command, and verifying that the new version is running without service interruption. Monitor active user sessions during update to confirm zero dropped connections.

**Acceptance Scenarios**:

1. **Given** an existing Vikunja deployment on version N, **When** administrator runs update command, **Then** system updates to latest version N+1 while maintaining active user sessions
2. **Given** an update includes database migrations, **When** update process runs, **Then** migrations execute successfully and new schema is applied without data loss
3. **Given** a backend update fails health checks, **When** automatic rollback triggers, **Then** system reverts to previous working version within 2 minutes
4. **Given** frontend and backend versions become incompatible, **When** update detects version mismatch, **Then** system prevents partial deployment and shows clear error message
5. **Given** an update is in progress, **When** another administrator attempts to run update, **Then** system prevents concurrent updates and shows lock status

---

### User Story 3 - Configuration Management (Priority: P2)

An administrator needs to modify Vikunja configuration after initial deployment (e.g., change public URL, update database connection, adjust resource limits). They run a reconfiguration command that presents current settings, allows selective updates, and applies changes with automatic service restart and validation.

**Why this priority**: Configuration changes are common during system lifecycle but less frequent than deployments and updates. This is P2 because initial deployment (P1) can include all necessary configuration, making this enhancement rather than core functionality.

**Independent Test**: Can be tested by changing a configuration value (e.g., public URL), running reconfigure command, and verifying new setting takes effect without requiring manual file edits or complete redeployment.

**Acceptance Scenarios**:

1. **Given** an existing deployment, **When** administrator runs reconfigure command and changes public URL, **Then** system updates configuration and restarts affected services
2. **Given** administrator changes database from SQLite to PostgreSQL, **When** reconfigure command runs with migration option, **Then** system exports data, updates configuration, imports to new database
3. **Given** invalid configuration value is entered, **When** system validates, **Then** error is shown before any changes are applied
4. **Given** configuration change requires service restart, **When** changes are saved, **Then** system performs graceful restart with health check validation

---

### User Story 4 - Health Monitoring and Status Checks (Priority: P2)

An administrator wants to verify the deployment is functioning correctly. They run a status command that checks all three components (backend, frontend, MCP server), database connectivity, disk space, memory usage, and reports any issues with suggested remediation steps.

**Why this priority**: Health monitoring is important for production operations but not required for basic deployment functionality. This is P2 because systems can operate without automated health checks using manual inspection.

**Independent Test**: Can be tested by running status command on a healthy deployment (shows all green), then stopping one component and re-running (shows specific failure with remediation advice).

**Acceptance Scenarios**:

1. **Given** all services are running normally, **When** administrator runs status check, **Then** system reports healthy status for all components with uptime and resource usage
2. **Given** backend service has crashed, **When** status check runs, **Then** system identifies failed service and provides restart command
3. **Given** database connection is lost, **When** status check runs, **Then** system reports database connectivity issue with troubleshooting steps
4. **Given** disk space is below 10% free, **When** status check runs, **Then** system warns about low disk space and suggests cleanup actions

---

### User Story 5 - Backup and Restore (Priority: P3)

An administrator wants to create backups before major updates or for disaster recovery. They run a backup command that exports database, configuration files, and uploaded files to a designated backup location with timestamp. The restore command can recreate the exact system state from any backup.

**Why this priority**: Backup capability is valuable for production systems but not essential for initial deployment and updates. This is P3 because it's a safety feature that enhances but doesn't block core functionality.

**Independent Test**: Can be tested by creating a backup, making changes to the system (create tasks, upload files), restoring from backup, and verifying system returns to exact backup state.

**Acceptance Scenarios**:

1. **Given** a running Vikunja instance, **When** administrator runs backup command, **Then** system creates timestamped backup with database dump, config files, and uploaded files
2. **Given** a backup file exists, **When** administrator runs restore command, **Then** system recreates exact system state including all tasks, users, and files
3. **Given** backup operation encounters an error, **When** failure occurs, **Then** partial backup is cleaned up and clear error message is displayed
4. **Given** multiple backups exist, **When** administrator lists backups, **Then** system shows all available backups with timestamps and sizes

---

### Edge Cases

- What happens when Proxmox node loses internet connectivity during deployment? System should pause and resume when connectivity returns, or fail gracefully with clear next steps.
- How does system handle deployment to a node with conflicting port allocations? System checks for port conflicts before starting and suggests alternative ports or asks to stop conflicting services.
- What happens if database migrations fail midway during update? System maintains pre-migration backup and provides rollback command to restore previous state.
- How does system handle updates when disk space is insufficient? Pre-flight check ensures adequate space (minimum 2GB free) before starting update process.
- What happens if administrator cancels deployment midway (Ctrl+C)? System detects interruption, performs cleanup of partial installation, and provides resume or clean retry option.
- How does system handle multiple Vikunja instances on same Proxmox cluster? Each deployment uses unique container ID, ports, and configuration to prevent conflicts.
- What happens during update if a user is actively uploading a large file? System waits for active uploads to complete (with 5-minute timeout) before proceeding with backend restart.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide a single-command deployment script that can be executed directly from a URL (curl/wget pattern like tteck helper scripts)
- **FR-002**: Deployment script MUST interactively prompt for: database type (SQLite/PostgreSQL/MySQL), database connection details (if external), public domain/URL, admin email, container resource allocation (CPU cores, RAM, disk size)
- **FR-003**: System MUST automatically install all dependencies including: Go runtime for backend, Node.js runtime for MCP server, web server for frontend static files, selected database client libraries
- **FR-004**: System MUST create a Debian-based LXC container with appropriate templates for Proxmox integration
- **FR-005**: System MUST configure all three components (backend API, frontend static files, MCP server) to start automatically on container boot using systemd services
- **FR-006**: Update mechanism MUST check main branch for new commits and determine if update is needed before proceeding
- **FR-007**: Update process MUST execute database migrations in correct order with automatic backup before migration starts
- **FR-008**: Update process MUST perform health checks on each component after update and automatically rollback if any health check fails
- **FR-009**: System MUST support blue-green deployment pattern where new version starts on alternate ports, passes health checks, then traffic switches with zero downtime
- **FR-010**: System MUST provide a configuration file (YAML format) that persists deployment settings and can be modified for reconfiguration
- **FR-011**: System MUST validate all configuration inputs (domain format, database connectivity, port availability) before applying changes
- **FR-012**: Status check command MUST verify: service running status, HTTP endpoint responsiveness, database connectivity, disk space (warn at <20% free), memory usage (warn at >80% used)
- **FR-013**: System MUST log all deployment, update, and reconfiguration operations to persistent log files with timestamps
- **FR-014**: System MUST provide rollback capability to revert to previous version within 5 minutes of failed update
- **FR-015**: System MUST support both fresh deployment and upgrade from manually installed Vikunja instances by detecting existing installations
- **FR-016**: Backup mechanism MUST export: complete database dump, configuration files, uploaded task attachments, and create single compressed archive
- **FR-017**: Restore mechanism MUST validate backup integrity before proceeding with restoration
- **FR-018**: System MUST assign static IP to container and configure reverse proxy (Nginx) for SSL termination and domain routing
- **FR-019**: System MUST provide uninstall command that removes container, cleans up resources, and optionally preserves data for future reinstallation
- **FR-020**: System MUST support running multiple independent Vikunja instances on same Proxmox cluster with automatic conflict avoidance

### Key Entities

- **LXC Container**: Represents the isolated Linux container running on Proxmox, contains all Vikunja components, has allocated resources (CPU, RAM, disk), has unique container ID and IP address
- **Deployment Configuration**: Represents the persistent configuration including database type and connection, public URL/domain, resource allocations, service ports, and admin email, stored as YAML file in container
- **Service Component**: Represents each of the three services (Backend API, Frontend Static Files, MCP Server), each with systemd unit file, health check endpoint, versioning information, and restart policies
- **Update Package**: Represents a deployable version from main branch including source code for all components, database migration scripts, dependency specifications, and version metadata
- **Backup Archive**: Represents a point-in-time system snapshot including database dump, configuration files, uploaded files, timestamp and version information, stored as compressed tar.gz

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Administrator can complete fresh Vikunja deployment from script execution to accessible web interface in under 10 minutes on a standard Proxmox node
- **SC-002**: Updates from main branch complete in under 5 minutes with zero dropped user connections for active sessions
- **SC-003**: System maintains 99.9% uptime during update operations (maximum 5 seconds of inaccessibility during traffic switchover)
- **SC-004**: Failed updates automatically rollback to previous working version in under 2 minutes
- **SC-005**: Health check command executes in under 10 seconds and provides clear status for all components
- **SC-006**: Backup creation completes in under 5 minutes for database with up to 10,000 tasks and 1GB of attachments
- **SC-007**: System supports up to 5 concurrent Vikunja instances on same Proxmox cluster without manual conflict resolution
- **SC-008**: 95% of deployments complete successfully on first attempt without requiring manual intervention
- **SC-009**: Configuration changes requiring service restart complete with less than 10 seconds of service interruption
- **SC-010**: All deployment operations provide real-time progress feedback with completion percentage

## Scope *(mandatory)*

### In Scope

- Automated LXC container creation and provisioning on Proxmox
- Interactive setup wizard for initial deployment configuration
- Installation and configuration of backend (Go), frontend (Vue.js static), and MCP server (Node.js)
- Support for SQLite, PostgreSQL, and MySQL database backends
- Automatic database migration execution during updates
- Blue-green deployment for zero-downtime updates
- Health checking and automatic rollback on failed updates
- Configuration management and reconfiguration capabilities
- Status monitoring and health checks
- Backup and restore functionality
- SSL/TLS setup with reverse proxy (Nginx)
- Multiple instance support on same Proxmox cluster
- Systemd service management for auto-start on boot
- Logging of all operations for troubleshooting

### Out of Scope

- Kubernetes or Docker Compose deployment strategies (Proxmox LXC only)
- Clustering/high-availability across multiple Proxmox nodes
- Automated SSL certificate management via Let's Encrypt (manual SSL setup only)
- Database replication or high-availability database configurations
- Custom backup scheduling or retention policies (manual backup invocation only)
- Integration with external monitoring systems (Prometheus, Grafana, etc.)
- Automated security patching of underlying OS
- Multi-tenant deployments with tenant isolation
- Load balancing across multiple Vikunja instances

## Assumptions *(mandatory)*

- Proxmox VE 7.0 or higher is installed and accessible
- Administrator has root access to Proxmox host
- Proxmox node has at least 2 CPU cores and 4GB RAM available for Vikunja container
- Proxmox node has reliable internet connectivity for downloading dependencies
- Administrator has basic Linux command-line knowledge
- DNS is configured externally to point domain to Proxmox node IP
- If using external database, database server is accessible from container network
- SSL certificates are provided by administrator (not auto-generated)
- Container storage is on local or shared storage accessible to Proxmox node
- Default container template (Debian 12) is available in Proxmox

## Dependencies *(mandatory)*

- Proxmox VE API access for container creation and management
- Debian 12 LXC template availability in Proxmox
- Internet connectivity for downloading: Go compiler/runtime, Node.js runtime, npm packages, Vikunja source code from Git repository
- External database server if PostgreSQL or MySQL is selected (for non-SQLite deployments)
- DNS infrastructure for domain name resolution
- SSL/TLS certificates for HTTPS termination

## Constraints *(mandatory)*

- Deployment script must be compatible with Bash 4.0+ (Proxmox host default shell)
- All operations must work with standard Proxmox permissions (no custom kernel modules)
- Container must use unprivileged LXC mode for security (no privileged containers)
- Update process must complete within 10 minutes to minimize risk window
- Maximum rollback time of 5 minutes to restore service after failed update
- System must work with Proxmox default networking (bridge mode)
- All services must run without requiring root privileges inside container
- Configuration file must be human-readable and editable (YAML format)
- Backup archives must be portable between different Proxmox nodes

## Risks *(mandatory)*

1. **Database migration failures during updates**: Mitigation - Always create automatic backup before migrations, provide manual rollback procedure in documentation
2. **Resource exhaustion on Proxmox node during deployment**: Mitigation - Pre-flight checks for minimum available resources, clear warning messages before proceeding
3. **Network connectivity loss during critical update phases**: Mitigation - Implement resumable operations with state tracking, document manual recovery procedures
4. **Port conflicts with existing services**: Mitigation - Check port availability before deployment, provide port customization during setup
5. **Incompatible Proxmox versions or configurations**: Mitigation - Check Proxmox version during deployment, maintain compatibility with LTS versions
6. **SSL certificate expiration affecting updates**: Mitigation - Separate certificate management from update process, document certificate renewal procedures
7. **Disk space exhaustion during backup or update**: Mitigation - Check available space before operations, clean up old backups automatically
8. **Concurrent update attempts causing state corruption**: Mitigation - Implement file-based locking mechanism, detect and prevent concurrent operations

## Open Questions *(optional)*

[No open questions at this time - all critical design decisions have reasonable defaults based on Proxmox and Vikunja architecture]
