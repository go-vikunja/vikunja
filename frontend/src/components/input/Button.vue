<template>
	<BaseButton
		class="button"
		:class="[
			variantClass,
			{
				'is-loading': props.loading,
				'has-no-shadow': !props.shadow || props.variant === 'tertiary',
			}
		]"
		:disabled="props.disabled || props.loading"
		:style="{
			'--button-white-space': props.wrap ? 'break-spaces' : 'nowrap',
		}"
		:type="props.type"
		:to="props.to"
		:href="props.href"
		:open-external-in-new-tab="props.openExternalInNewTab"
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
		<slot />
	</BaseButton>
</template>

<script setup lang="ts">
import {computed, type PropType} from 'vue'
import BaseButton from '@/components/base/BaseButton.vue'
import type {IconProp} from '@fortawesome/fontawesome-svg-core'
import type {RouteLocationRaw} from 'vue-router'

const props = defineProps({
	variant: {
		type: String as PropType<'primary' | 'secondary' | 'tertiary'>,
		default: 'primary'
	},
	icon: {
		type: Object as PropType<IconProp>,
		default: undefined
	},
	iconColor: {
		type: String,
		default: undefined
	},
	loading: {
		type: Boolean,
		default: false
	},
	disabled: {
		type: Boolean,
		default: false
	},
	shadow: {
		type: Boolean,
		default: true
	},
	wrap: {
		type: Boolean,
		default: true
	},
	type: {
		type: String as PropType<'button' | 'submit' | undefined>,
		default: undefined
	},
	to: {
		type: Object as PropType<RouteLocationRaw>,
		default: undefined
	},
	href: {
		type: String,
		default: undefined
	},
	openExternalInNewTab: {
		type: Boolean,
		default: true
	},
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
