import {ref, computed, onMounted, onUnmounted, watch} from 'vue'
import {useMediaQuery} from '@vueuse/core'
import {useAuthStore} from '@/stores/auth'

const BULMA_MOBILE_BREAKPOINT = 768
const DEFAULT_SIDEBAR_WIDTH = 300
const MIN_SIDEBAR_WIDTH = 200
const MAX_SIDEBAR_WIDTH = 500

function clampWidth(width: number): number {
	return Math.max(MIN_SIDEBAR_WIDTH, Math.min(MAX_SIDEBAR_WIDTH, width))
}

// Shared state across all component instances
const isResizing = ref(false)
const currentWidth = ref(DEFAULT_SIDEBAR_WIDTH)
let initialized = false
let watcherInitialized = false

// Captured body styles to restore after resize
let previousUserSelect: string | undefined
let previousCursor: string | undefined

function setupWatcher(authStore: ReturnType<typeof useAuthStore>) {
	if (watcherInitialized) return
	watcherInitialized = true

	watch(
		() => authStore.settings?.frontendSettings?.sidebarWidth,
		(newWidth) => {
			if (isResizing.value) return

			if (newWidth !== null && newWidth !== undefined) {
				currentWidth.value = clampWidth(newWidth)
			} else {
				currentWidth.value = DEFAULT_SIDEBAR_WIDTH
			}
		},
	)
}

export function useSidebarResize() {
	const authStore = useAuthStore()
	const isMobile = useMediaQuery(`(max-width: ${BULMA_MOBILE_BREAKPOINT}px)`)

	// Initialize width from settings only once
	onMounted(() => {
		if (initialized) return
		initialized = true

		const savedWidth = authStore.settings?.frontendSettings?.sidebarWidth
		if (savedWidth !== null && savedWidth !== undefined) {
			currentWidth.value = clampWidth(savedWidth)
		}
	})

	// Register settings watcher only once
	setupWatcher(authStore)

	const sidebarWidth = computed(() => {
		if (isMobile.value) {
			return '70vw'
		}
		return `${currentWidth.value}px`
	})

	function startResize(event: MouseEvent | TouchEvent) {
		if (isMobile.value) return

		event.preventDefault()
		isResizing.value = true

		document.addEventListener('mousemove', handleResize)
		document.addEventListener('mouseup', stopResize)
		document.addEventListener('touchmove', handleResize)
		document.addEventListener('touchend', stopResize)

		// Capture current styles and set new values
		previousUserSelect = document.body.style.userSelect
		previousCursor = document.body.style.cursor
		document.body.style.userSelect = 'none'
		document.body.style.cursor = 'ew-resize'
	}

	function handleResize(event: MouseEvent | TouchEvent) {
		if (!isResizing.value) return

		const clientX = 'touches' in event ? event.touches[0].clientX : event.clientX

		// Handle RTL direction
		const isRtl = document.dir === 'rtl'
		let newWidth: number

		if (isRtl) {
			newWidth = window.innerWidth - clientX
		} else {
			newWidth = clientX
		}

		currentWidth.value = clampWidth(newWidth)
	}

	function stopResize() {
		if (!isResizing.value) return

		isResizing.value = false

		document.removeEventListener('mousemove', handleResize)
		document.removeEventListener('mouseup', stopResize)
		document.removeEventListener('touchmove', handleResize)
		document.removeEventListener('touchend', stopResize)

		// Restore previous styles
		if (previousUserSelect !== undefined) {
			document.body.style.userSelect = previousUserSelect
			previousUserSelect = undefined
		}
		if (previousCursor !== undefined) {
			document.body.style.cursor = previousCursor
			previousCursor = undefined
		}

		// Save width to user settings
		saveWidth()
	}

	async function saveWidth() {
		const savedWidth = authStore.settings?.frontendSettings?.sidebarWidth
		// Only save if width actually changed
		if (savedWidth === currentWidth.value) return

		const newSettings = {
			...authStore.settings,
			frontendSettings: {
				...authStore.settings.frontendSettings,
				sidebarWidth: currentWidth.value,
			},
		}
		await authStore.saveUserSettings({
			settings: newSettings,
			showMessage: false,
		})
	}

	// Cleanup on unmount
	onUnmounted(stopResize)

	return {
		sidebarWidth,
		currentWidth,
		isResizing,
		startResize,
		isMobile,
		DEFAULT_SIDEBAR_WIDTH,
	}
}
