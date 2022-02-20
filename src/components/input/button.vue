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
	>
		<icon :icon="icon" v-if="showIconOnly"/>
		<span class="icon is-small" v-else-if="icon !== ''">
			<icon :icon="icon"/>
		</span>
		<slot />
	</BaseButton>
</template>

<script lang="ts">
export default {
	name: 'x-button',
}
</script>

<script setup lang="ts">
import {computed, useSlots, PropType} from 'vue'
import BaseButton from '@/components/base/BaseButton.vue'

const BUTTON_TYPES_MAP =  Object.freeze({
  primary: 'is-primary',
  secondary: 'is-outlined',
  tertiary: 'is-text is-inverted underline-none',
})

type ButtonTypes = keyof typeof BUTTON_TYPES_MAP

const props = defineProps({
	variant: {
		type: String as PropType<ButtonTypes>,
		default: 'primary',
	},
	icon: {
		default: '',
	},
	loading: {
		type: Boolean,
		default: false,
	},
	shadow: {
		type: Boolean,
		default: true,
	},
})

const variantClass = computed(() => BUTTON_TYPES_MAP[props.variant])

const slots = useSlots()
const showIconOnly = computed(() => props.icon !== '' && typeof slots.default === 'undefined')
</script>

<style lang="scss" scoped>
.button {
  transition: all $transition;
  border: 0;
  text-transform: uppercase;
  font-size: 0.85rem;
  font-weight: bold;
  min-height: $button-height;
  box-shadow: var(--shadow-sm);
  display: inline-flex;

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