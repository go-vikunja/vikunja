import {PrefixMode} from '@/modules/parseTaskText'

const key = 'quickAddMagicMode'

export const setQuickAddMagicMode = (mode: PrefixMode) => {
	localStorage.setItem(key, mode)
}

export const getQuickAddMagicMode = (): PrefixMode => {
	const mode = localStorage.getItem(key)

	// @ts-ignore
	return PrefixMode[mode] || PrefixMode.Disabled
}
