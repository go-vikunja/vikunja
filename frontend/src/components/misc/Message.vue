<template>
	<div class="message-wrapper">
		<div
			class="message"
			:class="[variant, textAlignClass]"
		>
			<slot />
		</div>
	</div>
</template>

<script lang="ts" setup>
import {computed} from 'vue'

const props = withDefaults(defineProps<{
	variant?: 'info' | 'danger' | 'warning' | 'success',
	textAlign?: TextAlignVariant
}>(), {
	variant: 'info',
	textAlign: 'left',
})

const TEXT_ALIGN_MAP = {
	left: '',
	center: 'has-text-centered',
	right: 'has-text-end',
} as const

export type TextAlignVariant = keyof typeof TEXT_ALIGN_MAP

const textAlignClass = computed(() => TEXT_ALIGN_MAP[props.textAlign])
</script>

<style lang="scss" scoped>
.message-wrapper {
	border-radius: $radius;
	background: var(--white);
}

.message {
	padding: .75rem 1rem;
	border-radius: $radius;
}

.info {
	border: 1px solid var(--primary);
	background: hsla(var(--primary-hsl), .05);
}

.danger {
	border: 1px solid var(--danger);
	background: hsla(var(--danger-h), var(--danger-s), var(--danger-l), .05);
}

.warning {
	border: 1px solid var(--warning);
	background: hsla(var(--warning-h), var(--warning-s), var(--warning-l), .05);
}

.success {
	border: 1px solid var(--success);
	background: hsla(var(--success-h), var(--success-s), var(--success-l), .05);
}
</style>
