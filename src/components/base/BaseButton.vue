<!-- a disabled link of any kind is not a link -->
<!-- we have a router link -->
<!-- just a normal link -->
<!-- a button it shall be -->
<!-- note that we only pass the click listener here -->
<template>
	<div
		v-if="disabled === true && (to !== undefined || href !== undefined)"
		class="base-button"
		:aria-disabled="disabled || undefined"
		ref="button"
	>
		<slot/>
	</div>
	<router-link
		v-else-if="to !== undefined"
		:to="to"
		class="base-button"
		ref="button"
	>
		<slot/>
	</router-link>
	<a v-else-if="href !== undefined"
		class="base-button"
		:href="href"
		rel="noreferrer noopener nofollow"
		target="_blank"
		ref="button"
	>
		<slot/>
	</a>
	<button
		v-else
		:type="type"
		class="base-button base-button--type-button"
		:disabled="disabled || undefined"
		ref="button"
		@click="(event: MouseEvent) => emit('click', event)"
	>
		<slot/>
	</button>
</template>

<script lang="ts">
const BASE_BUTTON_TYPES_MAP = {
	BUTTON: 'button',
	SUBMIT: 'submit',
} as const

export type BaseButtonTypes = typeof BASE_BUTTON_TYPES_MAP[keyof typeof BASE_BUTTON_TYPES_MAP] | undefined
</script>

<script lang="ts" setup>
// this component removes styling differences between links / vue-router links and button elements
// by doing so we make it easy abstract the functionality from style and enable easier and semantic
// correct button and link usage. Also see: https://css-tricks.com/a-complete-guide-to-links-and-buttons/#accessibility-considerations

// the component tries to heuristically determine what it should be checking the props 

// NOTE: Do NOT use buttons with @click to push routes. => Use router-links instead!

import {unrefElement} from '@vueuse/core'
import {ref, type HTMLAttributes} from 'vue'
import type {RouteLocationRaw} from 'vue-router'

export interface BaseButtonProps extends /* @vue-ignore */ HTMLAttributes {
	type?: BaseButtonTypes
	disabled?: boolean
	to?: RouteLocationRaw
	href?: string
}

export interface BaseButtonEmits {
	(e: 'click', payload: MouseEvent): void
}

const {
	type = BASE_BUTTON_TYPES_MAP.BUTTON,
	disabled = false,
} = defineProps<BaseButtonProps>()

const emit = defineEmits<BaseButtonEmits>()

const button = ref<HTMLElement | null>(null)

function focus() {
	unrefElement(button)?.focus()
}

defineExpose({
	focus,
})
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
	display: inline-block;
	color: inherit;
	font: inherit;
	user-select: none;
	pointer-events: auto; // disable possible resets

	&:focus, &.is-focused {
		outline: transparent;
	}

	&[disabled] {
		cursor: default;
	}
}
</style>
