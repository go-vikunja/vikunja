import type {Editor} from '@tiptap/core'

// Native <dialog> elements opened with showModal() render in the browser's
// top-layer, so popups appended to document.body end up visually behind them
// regardless of z-index. Appending to the open dialog itself lifts the popup
// into the same top-layer stacking context.
export function getPopupContainer(editor?: Editor): HTMLElement {
	const editorEl = editor?.view?.dom as HTMLElement | undefined
	const dialog = editorEl?.closest('dialog[open]') as HTMLElement | null
	return dialog ?? document.body
}
