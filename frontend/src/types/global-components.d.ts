import type { FunctionalComponent } from 'vue'
import type { Notifications } from '@kyvg/vue3-notification'
// import FontAwesomeIcon from '@/components/misc/Icon'
import type { FontAwesomeIcon as FontAwesomeIconFixedTypes } from './vue-fontawesome'
import type XButton from '@/components/input/Button.vue'
import type Modal from '@/components/misc/Modal.vue'
import type Card from '@/components/misc/Card.vue'

// Here we define globally imported components
// See: https://github.com/vuejs/language-tools/wiki/Global-Component-Types
declare module 'vue' {
	export interface GlobalComponents {
		Icon: FontAwesomeIconFixedTypes
		Notifications: FunctionalComponent<Notifications>
		XButton: typeof XButton,
		Modal: typeof Modal,
		Card: typeof Card,
	}
}

export {}
