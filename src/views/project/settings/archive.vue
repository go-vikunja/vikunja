<template>
	<modal
		@close="$router.back()"
		@submit="archiveList()"
	>
		<template #header><span>{{ list.isArchived ? $t('list.archive.unarchive') : $t('list.archive.archive') }}</span></template>
		
		<template #text>
			<p>{{ list.isArchived ? $t('list.archive.unarchiveText') : $t('list.archive.archiveText') }}</p>
		</template>
	</modal>
</template>

<script lang="ts">
export default {name: 'list-setting-archive'}
</script>

<script setup lang="ts">
import {computed} from 'vue'
import {useRouter, useRoute} from 'vue-router'
import {useI18n} from 'vue-i18n'

import {success} from '@/message'
import {useTitle} from '@/composables/useTitle'

import {useBaseStore} from '@/stores/base'
import {useListStore} from '@/stores/lists'

const {t} = useI18n({useScope: 'global'})
const listStore = useListStore()
const router = useRouter()
const route = useRoute()

const list = computed(() => listStore.getListById(route.params.listId))
useTitle(() => t('list.archive.title', {list: list.value.title}))

async function archiveList() {
	try {
		const newList = await listStore.updateList({
			...list.value,
			isArchived: !list.value.isArchived,
		})
		useBaseStore().setCurrentList(newList)
		success({message: t('list.archive.success')})
	} finally {
		router.back()
	}
}
</script>
