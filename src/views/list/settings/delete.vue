<template>
	<modal
		@close="$router.back()"
		@submit="deleteList()"
	>
		<template #header><span>{{ $t('list.delete.header') }}</span></template>

		<template #text>
			<p>
				{{ $t('list.delete.text1') }}
			</p>

			<p>
				<strong v-if="totalTasks !== null" class="has-text-white">
					{{ totalTasks > 0 ? $t('list.delete.tasksToDelete', {count: totalTasks}) : $t('list.delete.noTasksToDelete') }}
				</strong>
				<Loading v-else class="is-loading-small"/>
			</p>

			<p>
				{{ $t('misc.cannotBeUndone') }}
			</p>
		</template>
	</modal>
</template>

<script setup lang="ts">
import {computed, ref, watchEffect} from 'vue'
import {useTitle} from '@/composables/useTitle'
import {useI18n} from 'vue-i18n'
import {useStore} from 'vuex'
import {useRoute, useRouter} from 'vue-router'
import {success} from '@/message'
import TaskCollectionService from '@/services/taskCollection'
import Loading from '@/components/misc/loading.vue'

const {t} = useI18n()
const store = useStore()
const route = useRoute()
const router = useRouter()

const totalTasks = ref<number | null>(null)

const list = computed(() => store.getters['lists/getListById'](route.params.listId))

watchEffect(
	() => {
		const taskCollectionService = new TaskCollectionService()
		taskCollectionService.getAll({listId: route.params.listId}).then(() => {
			totalTasks.value = taskCollectionService.totalPages * taskCollectionService.resultCount
		})
	},
)

useTitle(() => t('list.delete.title', {list: list.value.title}))

async function deleteList() {
	await store.dispatch('lists/deleteList', list)
	success({message: t('list.delete.success')})
	router.push({name: 'home'})
}
</script>
