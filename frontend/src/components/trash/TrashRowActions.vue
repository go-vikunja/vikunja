<template>
	<span class="trash-row-actions">
		<BaseButton
			v-tooltip="$t('trash.restore')"
			v-cy="restoreCy"
			data-trash-action="restore"
			class="trash-action"
			:aria-label="$t('trash.restore')"
			:aria-disabled="pending || undefined"
			@click="emit('restore')"
		>
			<Icon icon="undo" />
		</BaseButton>
		<BaseButton
			v-tooltip="$t('trash.deleteForever')"
			v-cy="deleteForeverCy"
			data-trash-action="delete-forever"
			class="trash-action trash-action--danger"
			:aria-label="$t('trash.deleteForever')"
			:aria-disabled="pending || undefined"
			@click="emit('deleteForever')"
		>
			<Icon icon="trash-alt" />
		</BaseButton>
	</span>
</template>

<script setup lang="ts">
import BaseButton from '@/components/base/BaseButton.vue'

defineProps<{
	pending: boolean,
	restoreCy: string,
	deleteForeverCy: string,
}>()

const emit = defineEmits<{
	restore: [],
	deleteForever: [],
}>()
</script>

<style lang="scss" scoped>
.trash-row-actions {
	display: inline-flex;
	gap: .25rem;
}

.trash-action {
	display: inline-flex;
	align-items: center;
	justify-content: center;
	inline-size: 2rem;
	block-size: 2rem;
	border-radius: $radius;
	color: var(--grey-500);
	transition: color $transition, background-color $transition;

	&:hover {
		color: var(--primary);
		background-color: var(--grey-200);
	}

	&--danger:hover {
		color: var(--danger);
	}
}
</style>
