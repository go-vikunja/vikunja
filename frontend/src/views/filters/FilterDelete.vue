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

async function remove() {
	await deleteSavedFilter(getSavedFilterIdFromProjectId(props.projectId))
	success({message: t('filters.delete.success')})
	await router.push({name: 'projects.index'})
}
</script>
