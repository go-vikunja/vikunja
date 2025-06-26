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
		:type="type"
		:to="to"
		:href="href"
		:open-external-in-new-tab="openExternalInNewTab"
	>
		<template v-if="icon">
			<Icon
				v-if="!$slots.default"
				:icon="icon"
				:style="{color: iconColor}"
			/>
			<span
				v-else
				class="icon is-small"
			>
				<Icon
					:icon="icon"
					:style="{color: iconColor}"
				/>
			</span>
		</template>
		<slot />
	</BaseButton>
</template>

<script setup lang="ts">
import {computed} from 'vue'
import BaseButton, {type BaseButtonProps} from '@/components/base/BaseButton.vue'
import type {IconProp} from '@fortawesome/fontawesome-svg-core'
import type {RouteLocationRaw} from 'vue-router'

export type ButtonTypes = 'primary' | 'secondary' | 'tertiary'

export interface ButtonProps {
	variant?: ButtonTypes
	icon?: IconProp
	iconColor?: string
	loading?: boolean
	disabled?: boolean
	shadow?: boolean
	wrap?: boolean
	// BaseButton props (explicit types to avoid union complexity)
	type?: 'button' | 'submit' | undefined
	to?: RouteLocationRaw
	href?: string
	openExternalInNewTab?: boolean
}

const props = withDefaults(defineProps<ButtonProps>(), {
	variant: 'primary',
	icon: undefined,
	iconColor: undefined,
	loading: false,
	disabled: false,
	shadow: true,
	wrap: true,
	type: undefined,
	to: undefined,
	href: undefined,
	openExternalInNewTab: true,
})

defineOptions({name: 'XButton'})

const VARIANT_CLASS_MAP = {
	primary: 'is-primary',
	secondary: 'is-outlined',
	tertiary: 'is-text is-inverted underline-none',
} as const

const variantClass = computed(() => VARIANT_CLASS_MAP[props.variant])
</script>

<style lang="scss" scoped>
.button {
	transition: all $transition;
	border: 0;
	text-transform: uppercase;
	font-size: 0.85rem;
	font-weight: bold;
	height: auto;
	min-height: $button-height;
	box-shadow: var(--shadow-sm);
	display: inline-flex;
	white-space: var(--button-white-space);
	line-height: 1;

	&:hover {
		box-shadow: var(--shadow-md);
	}

	&.fullheight {
		padding-right: 7px;
		height: 100%;
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
}

.is-small {
	border-radius: $radius;
}

.underline-none {
	text-decoration: none !important;
}
</style>
