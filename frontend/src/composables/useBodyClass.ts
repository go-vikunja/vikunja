import {ref, watchEffect} from 'vue'
import {tryOnBeforeUnmount} from '@vueuse/core'

export function useBodyClass(className: string, defaultValue = false) {
	const isActive = ref(defaultValue)

	watchEffect(() => {
		if(isActive.value) {
			document.body.classList.add(className)
			return
		}
		
		document.body.classList.remove(className)
	})

	tryOnBeforeUnmount(() => isActive.value && document.body.classList.remove(className))

	return isActive
}
