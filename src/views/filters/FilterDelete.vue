<template>
	<modal
		@close="$router.back()"
		@submit="deleteSavedFilter()"
	>
		<template #header><span>{{ $t('filters.delete.header') }}</span></template>
		
		<template #text>
			<p>{{ $t('filters.delete.text') }}</p>
		</template>
	</modal>
</template>

<script setup lang="ts">
import { store } from '@/store'
import { useI18n } from 'vue-i18n'
import { useRouter, useRoute } from 'vue-router'
import {success} from '@/message'

import SavedFilterModel from '@/models/savedFilter'
import SavedFilterService from '@/services/savedFilter'
import {getSavedFilterIdFromListId} from '@/helpers/savedFilter'

const { t } = useI18n()
const router = useRouter()
const route = useRoute()

async function deleteSavedFilter() {
	// We assume the listId in the route is the pseudolist
	const savedFilterId = getSavedFilterIdFromListId(route.params.listId)

	const filterService = new SavedFilterService()
	const filter = new SavedFilterModel({id: savedFilterId})

	await filterService.delete(filter)
	await store.dispatch('namespaces/loadNamespaces')
	success({message: t('filters.delete.success')})
	router.push({name: 'namespaces.index'})
}
</script>
