import {PrefixMode} from '@/modules/parseTaskText'

const key = 'quickAddMagicMode'

export const setQuickAddMagicMode = (mode: PrefixMode) => {
	localStorage.setItem(key, mode)
}

export const getQuickAddMagicMode = (): PrefixMode => {
	const mode = localStorage.getItem(key)

	switch (mode) {
		case PrefixMode.Default:
			return PrefixMode.Default
		case PrefixMode.Todoist:
			return PrefixMode.Todoist
	}

	return PrefixMode.Disabled
}
