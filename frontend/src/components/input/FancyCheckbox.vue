<template>
	<BaseCheckbox
		class="fancy-checkbox"
		:class="{
			'is-disabled': disabled,
			'is-block': isBlock,
		}"
		:disabled="disabled"
		:model-value="modelValue"
		@update:modelValue="value => emit('update:modelValue', value)"
	>
		<CheckboxIcon class="fancy-checkbox__icon" />
		<span
			v-if="$slots.default"
			class="fancy-checkbox__content"
		>
			<slot />
		</span>
	</BaseCheckbox>
</template>

<script setup lang="ts">
import CheckboxIcon from '@/assets/checkbox.svg?component'
import BaseCheckbox from '@/components/base/BaseCheckbox.vue'

withDefaults(defineProps<{
	modelValue: boolean,
	disabled?: boolean,
	isBlock?: boolean
}>(), {
	disabled: false,
	isBlock: false,
})

const emit = defineEmits<{
	'update:modelValue': [value: boolean]
}>()
</script>

<style lang="scss" scoped>
.fancy-checkbox {
  display: inline-block;
  padding-inline-end: 5px;
  padding-block-start: 3px;

	&.is-block {
		display: block;
		margin: .5rem .2rem;
	}
}

.fancy-checkbox__content {
	font-size: 0.8rem;
	vertical-align: top;
	padding-inline-start: .5rem;
}

.fancy-checkbox__icon:deep() {
	position: relative;
	z-index: 1;
	stroke: var(--stroke-color, #c8ccd4);
	transform: translate3d(0, 0, 0);
	transition: all 0.2s ease;

	path,
	polyline {
		transition: all 0.2s linear, color 0.2s ease;
	}
}

.fancy-checkbox:hover input:not(:disabled) + .fancy-checkbox__icon,
.fancy-checkbox input:checked + .fancy-checkbox__icon {
	--stroke-color: var(--primary);
}
</style>

<style lang="scss">
// Since css-has-pseudo doesn't work with deep classes,
// the following rules can't be scoped

.fancy-checkbox :not(input:checked) + .fancy-checkbox__icon {
	path {
		transition-delay: 0.05s;
	}
}

.fancy-checkbox input:checked + .fancy-checkbox__icon {
	path {
		stroke-dashoffset: 60;
	}

	polyline {
		stroke-dashoffset: 42;
		transition-delay: 0.15s;
	}
}
</style>
