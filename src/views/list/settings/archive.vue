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
import {defineComponent} from 'vue'
export default defineComponent({name: 'list-setting-archive'})
</script>

<script setup lang="ts">
import {computed} from 'vue'
import {useStore} from 'vuex'
import {useRouter, useRoute} from 'vue-router'
import {useI18n} from 'vue-i18n'

import { success } from '@/message'
import { useTitle } from '@/composables/useTitle'

const {t} = useI18n({useScope: 'global'})
const store = useStore()
const router = useRouter()
const route = useRoute()

const list = computed(() => store.getters['lists/getListById'](route.params.listId))
useTitle(() => t('list.archive.title', {list: list.value.title}))

async function archiveList() {
	try {
		const newList = await store.dispatch('lists/updateList', {
			...list.value,
			isArchived: !list.value.isArchived,
		})
		store.commit('currentList', newList)
		success({message: t('list.archive.success')})
	} finally {
		router.back()
	}
}
</script>
