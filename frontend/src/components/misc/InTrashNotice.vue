<template>
	<Message
		variant="warning"
		class="in-trash"
	>
		<p>{{ label }}</p>
		<p class="has-text-grey">
			{{ $t('trash.daysLeft', trashDaysLeft(deletedAt)) }}
		</p>
		<!-- The trash only lists what the user can restore -->
		<div
			v-if="canRestore"
			class="in-trash__actions"
		>
			<XButton
				v-cy="'restoreTrashed'"
				icon="undo"
				:loading="restoring"
				@click="emit('restore')"
			>
				{{ $t('trash.restore') }}
			</XButton>
			<XButton
				v-if="!authStore.isLinkShareAuth"
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
