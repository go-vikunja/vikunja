<template>
	<div class="message-wrapper">
		<div class="message" :class="[variant, textAlignClass]">
			<slot/>
		</div>
	</div>
</template>

<script lang="ts" setup>
import {computed} from 'vue'

const props = defineProps({
	variant: {
		type: String,
		default: 'info',
	},
	textAlign: {
		type: String,
		default: 'left',
	},
})

const textAlignClass = computed(() => {
	switch (props.textAlign) {
		case 'left':
			return ''
		case 'right':
			return 'has-text-right'
		case 'center':
			return 'has-text-centered'
	}
})

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
