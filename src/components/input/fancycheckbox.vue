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
		<span class="fancycheckbox__content">
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

.fancycheckbox__icon {
	--stroke-color: #c8ccd4;
	position: relative;
	z-index: 1;
	stroke: var(--stroke-color);
	transform: translate3d(0, 0, 0);
	transition: all 0.2s ease;

	:deep(path) {
		// stroke-dasharray: 60;
		transition: all 0.2s linear, color 0.2s ease;
	}

	:deep(polyline) {
		// stroke-dasharray: 22;
		// stroke-dashoffset: 66;
		transition: all 0.2s linear, color 0.2s ease;
	}
}

.fancycheckbox:not(:has(input:disabled)):hover .fancycheckbox__icon {
	--stroke-color: var(--primary);
}

.fancycheckbox:has(input:checked) .fancycheckbox__icon:deep() {
	--stroke-color: var(--primary);

	path {
		stroke-dashoffset: 60;
	}

	polyline {
		stroke-dashoffset: 42;
		transition-delay: 0.15s;
	}
}
</style>