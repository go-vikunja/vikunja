// Licensed "pro" features the server may advertise via /info's enabled_pro_features.
// Use these instead of bare strings when calling configStore.isProFeatureEnabled.
export const PRO_FEATURE = {
	ADMIN_PANEL: 'admin_panel',
	TIME_TRACKING: 'time_tracking',
} as const

export type ProFeature = typeof PRO_FEATURE[keyof typeof PRO_FEATURE]
