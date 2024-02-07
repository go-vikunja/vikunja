import {ref, unref, watch} from 'vue'
import {debouncedWatch, tryOnMounted, useWindowSize, type MaybeRef} from '@vueuse/core'

// TODO: also add related styles
// OR: replace with vueuse function
export function useAutoHeightTextarea(value: MaybeRef<string>) {
	const textarea = ref<HTMLTextAreaElement | null>(null)
	const minHeight = ref(0)

	// adapted from https://github.com/LeaVerou/stretchy/blob/47f5f065c733029acccb755cae793009645809e2/src/stretchy.js#L34
	function resize(textareaEl: HTMLTextAreaElement | null) {
		if (!textareaEl) return

		let empty

		// the value here is the attribute value
		if (!textareaEl.value && textareaEl.placeholder) {
			empty = true
			textareaEl.value = textareaEl.placeholder
		}

		const cs = getComputedStyle(textareaEl)

		textareaEl.style.minHeight = ''
		textareaEl.style.height = '0'
		const offset = textareaEl.offsetHeight - parseFloat(cs.paddingTop) - parseFloat(cs.paddingBottom)
		const height = textareaEl.scrollHeight + offset + 'px'

		textareaEl.style.height = height

		// calculate min-height for the first time
		if (!minHeight.value) {
			minHeight.value = parseFloat(height)
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
		() => [textarea.value, unref(value)],
		() => resize(textarea.value),
		{
			immediate: true, // calculate initial size
			flush: 'post', // resize after value change is rendered to DOM
		},
	)

	return textarea
}