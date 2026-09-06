// Per-user entitlements resolved by the server on GET /user. Flags are 0/1,
// limits are the maximum and only present when the user is limited.
export const ENTITLEMENT = {
	ADMIN_PANEL: 'admin_panel',
	AUDIT_LOGS: 'audit_logs',
	TIME_TRACKING: 'time_tracking',
	TEAM_CREATION: 'team_creation',
	MAX_PROJECTS: 'max_projects',
	MAX_STORAGE_BYTES: 'max_storage_bytes',
} as const

export type Entitlement = typeof ENTITLEMENT[keyof typeof ENTITLEMENT]

export type EntitlementFlag =
	| typeof ENTITLEMENT.ADMIN_PANEL
	| typeof ENTITLEMENT.AUDIT_LOGS
	| typeof ENTITLEMENT.TIME_TRACKING
	| typeof ENTITLEMENT.TEAM_CREATION

export type EntitlementLimit =
	| typeof ENTITLEMENT.MAX_PROJECTS
	| typeof ENTITLEMENT.MAX_STORAGE_BYTES

export const ERROR_CODE_LIMIT_REACHED = 20001
export const ERROR_CODE_FEATURE_DISABLED_FOR_USER = 20002
