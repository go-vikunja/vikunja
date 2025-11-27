export interface ICustomShortcut {
	actionId: string        // e.g., "task.markDone"
	keys: string[]          // e.g., ["t"] or ["Control", "s"]
	isCustomized: boolean   // true if user changed from default
}

export interface ICustomShortcutsMap {
	[actionId: string]: string[]  // Maps "task.markDone" -> ["t"]
}

export interface ValidationResult {
	valid: boolean
	error?: string  // i18n key
	conflicts?: ShortcutAction[]
}

// Re-export from shortcuts.ts to avoid circular dependencies
export interface ShortcutAction {
	actionId: string           // Unique ID like "general.toggleMenu"
	title: string             // i18n key for display
	keys: string[]            // Default keys
	customizable: boolean     // Can user customize this?
	contexts?: string[]       // Which routes/contexts apply
	category: ShortcutCategory
}

export enum ShortcutCategory {
	GENERAL = 'general',
	NAVIGATION = 'navigation',
	TASK_ACTIONS = 'taskActions',
	PROJECT_VIEWS = 'projectViews',
	LIST_VIEW = 'listView',
	GANTT_VIEW = 'ganttView',
}
