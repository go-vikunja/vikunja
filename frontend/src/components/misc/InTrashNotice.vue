<template>
	<Message
		variant="warning"
		class="in-trash"
	>
		<p>{{ label }}</p>
		<p class="has-text-grey">
			{{ $t('trash.daysLeft', trashDaysLeft(deletedAt)) }}
		</p>
		<div class="in-trash__actions">
			<XButton
				v-if="canRestore"
				v-cy="restoreCy"
				icon="undo"
				:loading="restoring"
				@click="emit('restore')"
			>
				{{ $t('trash.restore') }}
			</XButton>
			<XButton
				variant="secondary"
				:to="trashRoute"
			>
				{{ $t('trash.open') }}
			</XButton>
		</div>
	</Message>
</template>

<script setup lang="ts">
import type {RouteLocationRaw} from 'vue-router'

import Message from '@/components/misc/Message.vue'
import {trashDaysLeft} from '@/client/queries/trash'

defineProps<{
	label: string,
	deletedAt: string,
	canRestore: boolean,
	restoring: boolean,
	trashRoute: RouteLocationRaw,
	restoreCy: string,
}>()

const emit = defineEmits<{
	restore: [],
}>()
</script>

<style lang="scss" scoped>
.in-trash__actions {
	display: flex;
	gap: .5rem;
	margin-block-start: .75rem;
}
</style>
