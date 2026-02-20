import type {FunctionalComponent} from 'vue'
import type {Notifications} from '@kyvg/vue3-notification'
import type XButton from '@/components/input/Button.vue'
import type Modal from '@/components/misc/Modal.vue'
import type Card from '@/components/misc/Card.vue'

declare module 'vue' {
	export interface GlobalComponents {
		Notifications: FunctionalComponent<Notifications>
		XButton: typeof XButton,
		Modal: typeof Modal,
		Card: typeof Card,
	}
}

export {}
