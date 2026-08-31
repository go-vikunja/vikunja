<template>
	<Modal
		@close="$router.back()"
		@submit="remove()"
	>
		<template #header>
			<span>{{ $t('filters.delete.header') }}</span>
		</template>

		<template #text>
			<ErrorMessage v-if="!(filterId > 0)" />
			<p v-else>
				{{ $t('filters.delete.text') }}
			</p>
		</template>
	</Modal>
</template>

<script setup lang="ts">
import {computed} from 'vue'
import {useRouter} from 'vue-router'

import {useIsAlive} from '@/composables/useIsAlive'

import ErrorMessage from '@/components/misc/Error.vue'

import {getSavedFilterIdFromProjectId} from '@/client/queries/projects'
import {useDeleteSavedFilterMutation} from '@/client/queries/savedFilters'

const props = defineProps<{
	projectId: number,
}>()

const router = useRouter()

const filterId = computed(() => getSavedFilterIdFromProjectId(props.projectId))

const alive = useIsAlive()
const deleteMutation = useDeleteSavedFilterMutation(id => alive.value && filterId.value === id)

async function remove() {
	const id = props.projectId
	if (!(filterId.value > 0)) {
		return
	}

	try {
		await deleteMutation.mutateAsync(filterId.value)
	} catch {
		return
	}
	// The route param can change on this same instance, so a stale delete must not navigate.
	if (!alive.value || props.projectId !== id) {
		return
	}
	await router.push({name: 'projects.index'})
}
</script>
