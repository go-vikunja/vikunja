import {ref, toValue, watch, type MaybeRefOrGetter} from 'vue'
import {debouncedWatch, tryOnMounted, useWindowSize} from '@vueuse/core'

// TODO: also add related styles
// OR: replace with vueuse function
export function useAutoHeightTextarea(value: MaybeRefOrGetter<string>) {
	const textarea = ref<HTMLTextAreaElement | null>(null)
	const minHeight = ref(0)
	const height = ref('')

	// adapted from https://github.com/LeaVerou/stretchy/blob/47f5f065c733029acccb755cae793009645809e2/src/stretchy.js#L34
	function resize(textareaEl: HTMLTextAreaElement | null) {
		if (!textareaEl) return

		let empty

		// the value here is the attribute value
		if (!textareaEl.value && textareaEl.placeholder) {
			empty = true
			textareaEl.value = textareaEl.placeholder
		}

		// const cs = getComputedStyle(textareaEl)

		textareaEl.style.minHeight = ''
		textareaEl.style.height = '0'
		height.value = textareaEl.scrollHeight + 'px'

		textareaEl.style.height = height.value

		// calculate min-height for the first time
		if (!minHeight.value) {
			minHeight.value = parseFloat(height.value)
		}

		textareaEl.style.minHeight = minHeight.value.toString()


		if (empty) {
			textareaEl.value = ''
		}

	}

	tryOnMounted(() => {
		if (textarea.value) {
			// we don't want scrollbars
			textarea.value.style.overflowY = 'hidden'
		}
	})

	const {width: windowWidth} = useWindowSize()

	debouncedWatch(
		windowWidth,
		() => resize(textarea.value),
		{debounce: 200},
	)

	// It is not possible to get notified of a change of the value attribute of a textarea without workarounds (setTimeout) 
	// So instead we watch the value that we bound to it.
	watch(
		() => [textarea.value, toValue(value)],
		() => resize(textarea.value),
		{
			immediate: true, // calculate initial size
			flush: 'post', // resize after value change is rendered to DOM
		},
	)

	return {
		textarea,
		height,
	}
}
