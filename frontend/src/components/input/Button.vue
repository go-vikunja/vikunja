<template>
	<BaseButton
		class="button"
		:class="[
			variantClass,
			{
				'is-loading': loading,
				'has-no-shadow': !shadow || variant === 'tertiary',
			}
		]"
		:disabled="disabled || loading"
		:style="{
			'--button-white-space': wrap ? 'break-spaces' : 'nowrap',
		}"
		:type="buttonType"
		:to="props.to"
		:href="props.href"
		:open-external-in-new-tab="openExternalInNewTab"
	>
		<template v-if="props.icon">
			<Icon
				v-if="!$slots.default"
				:icon="props.icon"
				:style="{color: props.iconColor}"
			/>
			<span
				v-else
				class="icon is-small"
			>
				<Icon
					:icon="props.icon"
					:style="{color: props.iconColor}"
				/>
			</span>
		</template>
		<span>
			<slot />
		</span>
	</BaseButton>
</template>

<script setup lang="ts">
import {computed} from 'vue'
import BaseButton from '@/components/base/BaseButton.vue'
import type {IconProp} from '@fortawesome/fontawesome-svg-core'
import type {RouteLocationRaw} from 'vue-router'

const VARIANT_CLASS_MAP = {
	primary: 'is-primary',
	secondary: 'is-outlined',
	tertiary: 'is-text is-inverted underline-none',
} as const

type ButtonTypes = 'primary' | 'secondary' | 'tertiary'

interface ButtonProps {
	variant?: ButtonTypes
	icon?: IconProp
	iconColor?: string
	loading?: boolean
	disabled?: boolean
	shadow?: boolean
	wrap?: boolean
	type?: 'button' | 'submit'
	to?: RouteLocationRaw
	href?: string
	openExternalInNewTab?: boolean
}

const props = defineProps<ButtonProps>() as ButtonProps

// Provide defaults with explicit typing
const variant = computed((): ButtonTypes => (props.variant ?? 'primary') as ButtonTypes)
const loading = computed((): boolean => props.loading ?? false)
const disabled = computed((): boolean => props.disabled ?? false)
const shadow = computed((): boolean => props.shadow ?? true)
const wrap = computed((): boolean => props.wrap ?? true)
const openExternalInNewTab = computed((): boolean => props.openExternalInNewTab ?? true)
const buttonType = computed((): 'button' | 'submit' => props.type ?? 'button')

defineOptions({name: 'XButton'})

const variantClass = computed(() => VARIANT_CLASS_MAP[variant.value])
</script>

<style lang="scss" scoped>
.button {
	transition: all $transition;
	border: 0;
	text-transform: uppercase;
	font-size: 0.85rem;
	font-weight: bold;
	block-size: auto;
	min-block-size: $button-height;
	box-shadow: var(--shadow-sm);
	white-space: var(--button-white-space);
	line-height: 1;
	display: inline-flex;
	padding-inline: 0; // override bulma style // override bulma style
	padding-inline: .5rem;
	gap: .25rem;

	[dir="rtl"] & {
		flex-direction: row-reverse;
	}

	&:hover {
		box-shadow: var(--shadow-md);
	}

	&.fullheight {
		padding-inline-end: 7px;
		block-size: 100%;
	}

	&.is-active,
	&.is-focused,
	&:active,
	&:focus,
	&:focus:not(:active) {
		box-shadow: var(--shadow-xs) !important;
	}

	&.is-primary.is-outlined:hover {
		color: var(--white);
	}

	.icon {
		margin: 0 !important;
	}
}

.is-small {
	border-radius: $radius;
}

.underline-none {
	text-decoration: none !important;
}
</style>
