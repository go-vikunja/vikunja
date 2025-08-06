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
		<span>
			<slot />
		</span>
	</BaseButton>
</template>

<script setup lang="ts">
import {computed} from 'vue'
import BaseButton, {type BaseButtonProps} from '@/components/base/BaseButton.vue'
import type {IconProp} from '@fortawesome/fontawesome-svg-core'

export type ButtonTypes = keyof typeof VARIANT_CLASS_MAP

export interface ButtonProps extends /* @vue-ignore */ BaseButtonProps {
	variant?: ButtonTypes
	icon?: IconProp
	iconColor?: string
	loading?: boolean
	disabled?: boolean
	shadow?: boolean
	wrap?: boolean
}

const props = withDefaults(defineProps<ButtonProps>(), {
	variant: 'primary',
	icon: undefined,
	iconColor: undefined,
	loading: false,
	disabled: false,
	shadow: true,
	wrap: true,
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
