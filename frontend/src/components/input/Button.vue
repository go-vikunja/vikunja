<template>
	<BaseButton
		class="button"
		:class="[
			variantClass,
			{
				'is-loading': getLoading(),
				'has-no-shadow': hasNoShadow,
			}
		]"
		:disabled="getDisabled() || getLoading()"
		:style="{
			'--button-white-space': buttonWhiteSpace,
		}"
		:type="getType()"
		:to="props.to"
		:href="props.href"
		:open-external-in-new-tab="getOpenExternalInNewTab()"
	>
		<template v-if="props.icon">
			<Icon
				v-if="!$slots.default"
				:icon="props.icon as IconProp"
				:style="{color: props.iconColor}"
			/>
			<span
				v-else
				class="icon is-small"
			>
				<Icon
					:icon="props.icon as IconProp"
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
import type {RouteLocationRaw} from 'vue-router'
import type {IconProp} from '@fortawesome/fontawesome-svg-core'

type ButtonTypes = 'primary' | 'secondary' | 'tertiary'
// Simpler icon type to avoid complex union type issues
type SimpleIconType = IconProp | string | string[] | Record<string, unknown>

interface ButtonProps {
	variant?: ButtonTypes
	icon?: SimpleIconType
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

const props = defineProps<ButtonProps>()

// Handle defaults
const getVariant = () => props.variant ?? 'primary'
const getLoading = () => props.loading ?? false
const getDisabled = () => props.disabled ?? false
const getShadow = () => props.shadow ?? true
const getWrap = () => props.wrap ?? true
const getType = () => props.type ?? 'button'
const getOpenExternalInNewTab = () => props.openExternalInNewTab ?? true

const VARIANT_CLASS_MAP = {
	primary: 'is-primary',
	secondary: 'is-outlined',
	tertiary: 'is-text is-inverted underline-none',
} as const

defineOptions({name: 'XButton'})

const variantClass = computed(() => VARIANT_CLASS_MAP[getVariant()])
const hasNoShadow = computed(() => !getShadow() || getVariant() === 'tertiary')
const buttonWhiteSpace = computed(() => getWrap() ? 'break-spaces' : 'nowrap')
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
