<template>
	<div class="message" :class="variant">
		<slot/>
	</div>
</template>

<script lang="ts" setup>
defineProps({
	variant: {
		type: String,
		default: 'info',
	},
})
</script>

<style lang="scss" scoped>
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
