<template>
	<component
		:is="componentNodeName"
		class="base-button"
		:class="{ 'base-button--type-button': isButton }"
		v-bind="elementBindings"
		:disabled="disabled || undefined"
	>
		<slot />
	</component>
</template>

<script lang="ts">
import {defineComponent} from 'vue'

// see https://v3.vuejs.org/api/sfc-script-setup.html#usage-alongside-normal-script
export default defineComponent({
  inheritAttrs: false,
})
</script>

<script lang="ts" setup>
// this component removes styling differences between links / vue-router links and button elements
// by doing so we make it easy abstract the functionality from style and enable easier and semantic
// correct button and link usage. Also see: https://css-tricks.com/a-complete-guide-to-links-and-buttons/#accessibility-considerations

// the component tries to heuristically determine what it should be checking the props (see the
// componentNodeName and elementBindings ref for this).

// NOTE: Do NOT use buttons with @click to push routes. => Use router-links instead!

import { ref, watchEffect, computed, useAttrs, PropType } from 'vue'

const BASE_BUTTON_TYPES_MAP =  Object.freeze({
  button: 'button',
  submit: 'submit',
})

type BaseButtonTypes = keyof typeof BASE_BUTTON_TYPES_MAP

const props = defineProps({
	type: {
		type: String as PropType<BaseButtonTypes>,
		default: 'button',
	},
	disabled: {
		type: Boolean,
		default: false,
	},
})


const componentNodeName = ref<Node['nodeName']>('button')
interface ElementBindings {
	type?: string;
	rel?: string,
}

const elementBindings = ref({})

const attrs = useAttrs()
watchEffect(() => {
	// by default this component is a button element with the attribute of the type "button" (default prop value)
	let nodeName = 'button'
	let bindings: ElementBindings = {type: props.type}

	// if we find a "to" prop we set it as router-link
	if ('to' in attrs) {
		nodeName = 'router-link'
		bindings = {}
	}

	// if there is a href we assume the user wants an external link via a link element
	// we also set a predefined value for the attribute rel, but make it possible to overwrite this by the user.
	if ('href' in attrs) {
		nodeName = 'a'
		bindings = {rel: 'noreferrer noopener nofollow'}
	}

	componentNodeName.value = nodeName
	elementBindings.value = {
		...bindings,
		...attrs,
	}
})

const isButton = computed(() => componentNodeName.value === 'button')
</script>

<style lang="scss">
// NOTE: we do not use scoped styles to reduce specifity and make it easy to overwrite

// We reset the default styles of a button element to enable easier styling
:where(.base-button--type-button) {
	border: 0;
	margin: 0;
	padding: 0;
	text-decoration: none;
	background-color: transparent;
	text-align: center;
	appearance: none;
}

:where(.base-button) {
	cursor: pointer;
	display: block;
	color: inherit;
	font: inherit;
	user-select: none;
	pointer-events: auto; // disable possible resets

	&:focus {
		outline: transparent;
	}

	&[disabled] {
		cursor: default;
	}
}
</style>
