export function useQuickAddMode() {
	const isQuickAddMode = new URLSearchParams(window.location.search).get('mode') === 'quick-add'
	return { isQuickAddMode }
}
