import {reactive, watch} from 'vue'

const STORAGE_KEY = 'gantt-arrow-config'

export interface GanttArrowConfig {
	enabled: boolean
	strokeWidth: number
	dashArray: string
	opacity: number
	pathMode: 'bezier' | 'stepped' | 'stepRounded'
	// Bezier
	cp1X: number
	cp1Y: number
	cp2X: number
	cp2Y: number
	// Stepped
	exitDir: 'right' | 'bottom'
	entryDir: 'left' | 'top'
	exitOffset: number
	entryOffset: number
	exitLength: number
	entryLength: number
	cornerRadius: number
	// General
	arrowSize: number
	showDots: boolean
	dotRadius: number
	showShadow: boolean
	shadowWidth: number
	shadowOpacity: number
	palette: 'multi' | 'mono'
}

const DEFAULTS: GanttArrowConfig = {
	enabled: true,
	strokeWidth: 1.5,
	dashArray: '2,2',
	opacity: 0.5,
	pathMode: 'stepRounded',
	cp1X: 0.5,
	cp1Y: 0,
	cp2X: 0.5,
	cp2Y: 0,
	exitDir: 'bottom',
	entryDir: 'left',
	exitOffset: 0.5,
	entryOffset: 0.5,
	exitLength: 30,
	entryLength: 20,
	cornerRadius: 11,
	arrowSize: 8,
	showDots: false,
	dotRadius: 3,
	showShadow: false,
	shadowWidth: 4,
	shadowOpacity: 0.2,
	palette: 'multi',
}

function loadConfig(): GanttArrowConfig {
	try {
		const stored = localStorage.getItem(STORAGE_KEY)
		if (stored) {
			const parsed = JSON.parse(stored)
			return {...DEFAULTS, ...parsed}
		}
	} catch {
		// ignore
	}
	return {...DEFAULTS}
}

function saveConfig(config: GanttArrowConfig) {
	try {
		localStorage.setItem(STORAGE_KEY, JSON.stringify(config))
	} catch {
		// ignore
	}
}

// Singleton reactive config
const arrowConfig = reactive<GanttArrowConfig>(loadConfig())

// Auto-save on any change
watch(arrowConfig, (val) => {
	saveConfig(val)
}, {deep: true})

export function useGanttArrowConfig() {
	function resetToDefaults() {
		Object.assign(arrowConfig, DEFAULTS)
	}

	function importConfig(json: string) {
		try {
			const parsed = JSON.parse(json)
			Object.assign(arrowConfig, {...DEFAULTS, ...parsed})
		} catch {
			// ignore
		}
	}

	function exportConfig(): string {
		return JSON.stringify(arrowConfig, null, 2)
	}

	return {
		config: arrowConfig,
		defaults: DEFAULTS,
		resetToDefaults,
		importConfig,
		exportConfig,
	}
}
