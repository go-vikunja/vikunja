<template>
	<Modal
		@close="$router.back()"
		@submit="remove()"
	>
		<template #header>
			<span>{{ $t('filters.delete.header') }}</span>
		</template>

		<template #text>
			<p>{{ $t('filters.delete.text') }}</p>
		</template>
	</Modal>
</template>

<script setup lang="ts">
import {onUnmounted, ref} from 'vue'
import {useI18n} from 'vue-i18n'
import {useRouter} from 'vue-router'

import {getSavedFilterIdFromProjectId} from '@/client/queries/projects'
import {deleteSavedFilter} from '@/client/queries/savedFilters'
import {success} from '@/message'

const props = defineProps<{
	projectId: number,
}>()

const {t} = useI18n({useScope: 'global'})
const router = useRouter()

// useMounted() never flips back on unmount, so track liveness ourselves.
const alive = ref(true)
onUnmounted(() => {
	alive.value = false
})

async function remove() {
	const id = props.projectId
	const filterId = getSavedFilterIdFromProjectId(id)
	if (filterId <= 0) {
		return
	}

	await deleteSavedFilter(filterId)
	// The route param can change on this same instance, so a stale delete must not navigate.
	if (!alive.value || props.projectId !== id) {
		return
	}
	success({message: t('filters.delete.success')})
	await router.push({name: 'projects.index'})
}
</script>
