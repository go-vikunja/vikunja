import type {ILabel} from '@/modelTypes/ILabel'

export function useLabelStyles() {
	function getLabelStyles(label: ILabel) {
		return {
			'background': label.hexColor || 'var(--grey-200)',
			'color': label.textColor || 'var(--grey-800)',
		}
	}

	return {
		getLabelStyles,
	}
}