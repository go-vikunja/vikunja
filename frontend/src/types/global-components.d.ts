// import FontAwesomeIcon from '@/components/misc/Icon'
import type { FontAwesomeIcon as FontAwesomeIconFixedTypes } from './vue-fontawesome'
import type XButton from '@/components/input/button.vue'
import type Modal from '@/components/misc/modal.vue'
import type Card from '@/components/misc/card.vue'

// Here we define globally imported components
// See:
// https://github.com/johnsoncodehk/volar/blob/2ca8fd3434423c7bea1c8e08132df3b9ce84eea7/extensions/vscode-vue-language-features/README.md#usage
// Under the hidden collapsible "Define Global Components"

declare module '@vue/runtime-core' {
	export interface GlobalComponents {
		Icon: FontAwesomeIconFixedTypes
		XButton: typeof XButton,
		Modal: typeof Modal,
		Card: typeof Card,
	}
}

export {}
