<template>
	<BaseCheckbox
		class="fancycheckbox"
		:class="{
			'is-disabled': disabled,
			'is-block': isBlock,
		}"
		:disabled="disabled"
		:model-value="modelValue"
		@update:model-value="value => emit('update:modelValue', value)"
	>
		<CheckboxIcon class="fancycheckbox__icon" />
		<span v-if="$slots.default" class="fancycheckbox__content">
			<slot/>
		</span>
	</BaseCheckbox>
</template>

<script setup lang="ts">
import CheckboxIcon from '@/assets/checkbox.svg?component'

import BaseCheckbox from '@/components/base/BaseCheckbox.vue'

defineProps({
	modelValue: {
		type: Boolean,
	},
	disabled: {
		type: Boolean,
	},
	isBlock: {
		type: Boolean,
		default: false,
	},
})

const emit = defineEmits<{
  (event: 'update:modelValue', value: boolean): void
}>()
</script>


<style lang="scss" scoped>
.fancycheckbox {
  display: inline-block;
  padding-right: 5px;
  padding-top: 3px;

	&.is-block {
		display: block;
		margin: .5rem .2rem;
	}
}

.fancycheckbox__content {
	font-size: 0.8rem;
	vertical-align: top;
	padding-left: .5rem;
}

.fancycheckbox__icon:deep() {
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

.fancycheckbox:not(:has(input:disabled)):hover .fancycheckbox__icon,
.fancycheckbox:has(input:checked) .fancycheckbox__icon {
	--stroke-color: var(--primary);
}
</style>

<style lang="scss">
// Since css-has-pseudo doesn't work with deep classes,
// the following rules can't be scoped

.fancycheckbox:has(:not(input:checked)) .fancycheckbox__icon {
	path {
		transition-delay: 0.05s;
	}
}

.fancycheckbox:has(input:checked) .fancycheckbox__icon {
	path {
		stroke-dashoffset: 60;
	}

	polyline {
		stroke-dashoffset: 42;
		transition-delay: 0.15s;
	}
}
</style>