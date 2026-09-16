// Native <dialog> elements opened with showModal() render in the browser's
// top-layer, so popups appended to document.body end up visually behind them
// regardless of z-index. Appending into the open dialog lifts them into it.
export function getTopLayerContainer(el?: Element | null): HTMLElement {
	return (el?.closest('dialog[open]') as HTMLElement | null) ?? document.body
}
