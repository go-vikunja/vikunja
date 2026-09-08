// Per-user entitlements resolved by the server on GET /user. Flags are 0/1,
// limits are the maximum and only present when the user is limited.
export const ENTITLEMENT_FLAGS = ['admin_panel', 'audit_logs', 'time_tracking', 'team_creation'] as const
export const ENTITLEMENT_LIMITS = ['max_projects', 'max_storage_bytes'] as const

export type EntitlementFlag = typeof ENTITLEMENT_FLAGS[number]
export type EntitlementLimit = typeof ENTITLEMENT_LIMITS[number]
export type Entitlement = EntitlementFlag | EntitlementLimit

export const ENTITLEMENT = {
	ADMIN_PANEL: 'admin_panel',
	AUDIT_LOGS: 'audit_logs',
	TIME_TRACKING: 'time_tracking',
	TEAM_CREATION: 'team_creation',
	MAX_PROJECTS: 'max_projects',
	MAX_STORAGE_BYTES: 'max_storage_bytes',
} as const satisfies Record<string, Entitlement>
