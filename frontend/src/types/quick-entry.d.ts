interface Window {
	quickEntry?: {
		close: () => void
		resize: (width: number, height: number) => void
		showMainWindow: () => void
	}
}
