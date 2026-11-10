<template>
	<Message
		variant="warning"
		class="in-trash"
	>
		<p>{{ label }}</p>
		<p class="has-text-grey">
			{{ $t('trash.daysLeft', trashDaysLeft(deletedAt)) }}
		</p>
		<!-- Restore stays user-only; the trash lists only what the user can restore -->
		<div
			v-if="canRestore && !authStore.isLinkShareAuth"
			class="in-trash__actions"
		>
			<XButton
				v-cy="'restoreTrashed'"
				icon="undo"
				:loading="restoring"
				@click="emit('restore')"
			>
				{{ restoreLabel ?? $t('trash.restore') }}
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
import {useAuthStore} from '@/stores/auth'

defineProps<{
	label: string,
	deletedAt: string,
	canRestore: boolean,
	restoreLabel?: string,
	restoring: boolean,
	trashRoute: RouteLocationRaw,
}>()

const emit = defineEmits<{
	restore: [],
}>()

const authStore = useAuthStore()
</script>

<style lang="scss" scoped>
.in-trash__actions {
	display: flex;
	gap: .5rem;
	margin-block-start: .75rem;
}
</style>
