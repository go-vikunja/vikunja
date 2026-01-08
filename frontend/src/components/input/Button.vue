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
import BaseButton from '@/components/base/BaseButton.vue'
import type {IconProp} from '@fortawesome/fontawesome-svg-core'

const props = defineProps<ButtonProps>()

const VARIANT_CLASS_MAP = {
	primary: 'is-primary',
	secondary: 'is-outlined',
	tertiary: 'is-text is-inverted underline-none',
} as const

export type ButtonTypes = keyof typeof VARIANT_CLASS_MAP

export interface ButtonProps {
	variant?: ButtonTypes
	icon?: IconProp
	iconColor?: string
	loading?: boolean
	disabled?: boolean
	shadow?: boolean
	wrap?: boolean
}

defineOptions({name: 'XButton'})

// @ts-expect-error - Complex union type from IconProp causes TS2590, but the code is correct
const variant = computed(() => (props.variant ?? 'primary') as ButtonTypes)
const shadow = computed(() => (props.shadow ?? true) as boolean)
const wrap = computed(() => (props.wrap ?? true) as boolean)
const variantClass = computed<string>(() => VARIANT_CLASS_MAP[variant.value])
</script>

<style lang="scss" scoped>
.button {
	// Base structure (replaces Bulma's .button)
	display: inline-flex;
	align-items: center;
	justify-content: center;
	vertical-align: top;
	cursor: pointer;
	text-align: center;
	white-space: var(--button-white-space);

	// Custom styles
	transition: all $transition;
	border: 0;
	text-transform: uppercase;
	font-size: 0.85rem;
	font-weight: bold;
	block-size: auto;
	min-block-size: $button-height;
	box-shadow: var(--shadow-sm);
	line-height: 1;
	padding-inline: .5rem;
	gap: .25rem;

	// Default/Primary variant colors
	background-color: var(--primary);
	color: var(--white);
	border-radius: $radius;

	[dir="rtl"] & {
		flex-direction: row-reverse;
	}

	&:hover {
		box-shadow: var(--shadow-md);
		background-color: var(--primary-dark, color-mix(in srgb, var(--primary) 85%, black));
	}

	&:focus,
	&:focus-visible {
		outline: 2px solid var(--primary);
		outline-offset: 2px;
	}

	&.is-active,
	&.is-focused,
	&:active,
	&:focus,
	&:focus:not(:active) {
		box-shadow: var(--shadow-xs) !important;
	}

	&[disabled] {
		opacity: 0.5;
		cursor: not-allowed;
		pointer-events: none;
	}

	.icon {
		margin: 0 !important;
	}

	// Primary variant (default, explicit)
	&.is-primary {
		background-color: var(--primary);
		color: var(--white);

		&:hover {
			background-color: var(--primary-dark, color-mix(in srgb, var(--primary) 85%, black));
		}
	}

	// Secondary/Outlined variant
	&.is-outlined {
		background-color: var(--scheme-main);
		color: var(--grey-900);

		&:hover {
			color: var(--grey-600);
		}
	}

	// Tertiary/Text variant
	&.is-text {
		background-color: transparent;
		color: var(--text);
		box-shadow: none;

		&:hover {
			background-color: var(--grey-100);
			box-shadow: none;
		}
	}

	&.is-inverted {
		// Used with is-text for tertiary buttons
		color: inherit;
	}

	// Loading state
	&.is-loading {
		color: transparent !important;
		pointer-events: none;
		position: relative;

		&::after {
			content: "";
			position: absolute;
			display: block;
			block-size: 1em;
			inline-size: 1em;
			border: 2px solid currentColor;
			border-radius: 50%;
			border-inline-end-color: transparent;
			border-block-start-color: transparent;
			animation: spin-around 500ms infinite linear;

			// Center the spinner
			inset-inline-start: calc(50% - 0.5em);
			inset-block-start: calc(50% - 0.5em);
		}
	}
}

@keyframes spin-around {
	from {
		transform: rotate(0deg);
	}
	to {
		transform: rotate(360deg);
	}
}

.is-small {
	border-radius: $radius;
}

.underline-none {
	text-decoration: none !important;
}
</style>
