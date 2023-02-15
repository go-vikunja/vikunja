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
		:style="{
			'--button-white-space': wrap ? 'break-spaces' : 'nowrap',
		}"
	>
		<template v-if="icon">
			<icon 
				v-if="showIconOnly"
				:icon="icon"
				:style="{'color': iconColor !== '' ? iconColor : undefined}"
			/>
			<span class="icon is-small" v-else>
				<icon 
					:icon="icon"
					:style="{'color': iconColor !== '' ? iconColor : undefined}"
				/>
			</span>
		</template>
		<slot />
	</BaseButton>
</template>

<script lang="ts">
const BUTTON_TYPES_MAP = {
  primary: 'is-primary',
  secondary: 'is-outlined',
  tertiary: 'is-text is-inverted underline-none',
} as const

export type ButtonTypes = keyof typeof BUTTON_TYPES_MAP

export default { name: 'x-button' }
</script>

<script setup lang="ts">
import {computed, useSlots} from 'vue'
import BaseButton, {type BaseButtonProps} from '@/components/base/BaseButton.vue'
import type { IconProp } from '@fortawesome/fontawesome-svg-core'

// extending the props of the BaseButton
export interface ButtonProps extends BaseButtonProps {
	variant?: ButtonTypes
	icon?: IconProp
	iconColor?: string
	loading?: boolean
	shadow?: boolean
	wrap?: boolean
}

const {
	variant = 'primary',
	icon = '',
	iconColor = '',
	loading = false,
	shadow = true,
	wrap = true,
} = defineProps<ButtonProps>()

const variantClass = computed(() => BUTTON_TYPES_MAP[variant])

const slots = useSlots()
const showIconOnly = computed(() => icon !== '' && typeof slots.default === 'undefined')
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