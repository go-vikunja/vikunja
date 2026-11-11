import type {Editor} from '@tiptap/core'
import {getTopLayerContainer} from '@/helpers/getTopLayerContainer'

export function getPopupContainer(editor?: Editor): HTMLElement {
	return getTopLayerContainer(editor?.view?.dom as HTMLElement | undefined)
}
