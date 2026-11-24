// A Safari "Add to Dock" web app on macOS draws its window toolbar over the page and reports no safe-area
// inset, so the height is measured: 40px, windowed and full screen (Safari 26.6, macOS 26.6).
const DOCK_TOOLBAR_HEIGHT = 40

const TOP_INSET_PROPERTY = '--app-top-inset'

// navigator.standalone is Safari-only, and iOS is the only other platform that sets it; those have touch points.
export function isMacDockApp(nav: Navigator = navigator): boolean {
	return (nav as Navigator & {standalone?: boolean}).standalone === true && nav.maxTouchPoints === 0
}

export function applyDockAppOffset(nav: Navigator = navigator, root: HTMLElement = document.documentElement) {
	if (isMacDockApp(nav)) {
		root.style.setProperty(TOP_INSET_PROPERTY, `${DOCK_TOOLBAR_HEIGHT}px`)
	}
}

export function appTopInset(root: HTMLElement = document.documentElement): number {
	return parseFloat(getComputedStyle(root).getPropertyValue(TOP_INSET_PROPERTY)) || 0
}
