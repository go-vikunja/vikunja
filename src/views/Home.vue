<template>
	<div class="content has-text-centered">
		<h2 v-if="userInfo">
			{{ $t(welcome, {username: userInfo.name !== '' ? userInfo.name : userInfo.username}) }}!
		</h2>
		<message variant="danger" v-if="deletionScheduledAt !== null" class="mb-4">
			{{
				$t('user.deletion.scheduled', {
					date: formatDateShort(deletionScheduledAt),
					dateSince: formatDateSince(deletionScheduledAt),
				})
			}}
			<router-link :to="{name: 'user.settings', hash: '#deletion'}">
				{{ $t('user.deletion.scheduledCancel') }}
			</router-link>
		</message>
		<add-task
			:listId="defaultListId"
			@taskAdded="updateTaskList"
			class="is-max-width-desktop"
		/>
		<template v-if="!hasTasks && !loading">
			<template v-if="defaultNamespaceId > 0">
				<p class="mt-4">{{ $t('home.list.newText') }}</p>
				<x-button
					:to="{ name: 'list.create', params: { id: defaultNamespaceId } }"
					:shadow="false"
					class="ml-2"
				>
					{{ $t('home.list.new') }}
				</x-button>
			</template>
			<p class="mt-4" v-if="migratorsEnabled">
				{{ $t('home.list.importText') }}
			</p>
			<x-button
				v-if="migratorsEnabled"
				:to="{ name: 'migrate.start' }"
				:shadow="false">
				{{ $t('home.list.import') }}
			</x-button>
		</template>
		<div v-if="listHistory.length > 0" class="is-max-width-desktop has-text-left mt-4">
			<h3>{{ $t('home.lastViewed') }}</h3>
			<div class="is-flex list-cards-wrapper-2-rows">
				<list-card
					v-for="(l, k) in listHistory"
					:key="`l${k}`"
					:list="l"
				/>
			</div>
		</div>
		<ShowTasks class="mt-4" :show-all="true" v-if="hasLists" :key="showTasksKey"/>

		<transition name="modal">
			<task-detail-view-modal v-if="showTaskDetail" />
		</transition>
	</div>
</template>

<script lang="ts" setup>
import {ref, computed} from 'vue'
import {useStore} from 'vuex'

import Message from '@/components/misc/message.vue'
import ShowTasks from '@/views/tasks/ShowTasks.vue'
import ListCard from '@/components/list/partials/list-card.vue'
import AddTask from '@/components/tasks/add-task.vue'

import {getHistory} from '@/modules/listHistory'
import {parseDateOrNull} from '@/helpers/parseDateOrNull'
import {formatDateShort, formatDateSince} from '@/helpers/time/formatDate'
import {useDateTimeSalutation} from '@/composables/useDateTimeSalutation'
import TaskDetailViewModal, { useShowModal } from '@/views/tasks/TaskDetailViewModal.vue'

const showTaskDetail = useShowModal()

const welcome = useDateTimeSalutation()

const store = useStore()
const listHistory = computed(() => {
	return getHistory()
		.map(l => store.getters['lists/getListById'](l.id))
		.filter(l => l !== null)
})


const migratorsEnabled = computed(() => store.state.config.availableMigrators?.length > 0)
const userInfo = computed(() => store.state.auth.info)
const hasTasks = computed(() => store.state.hasTasks)
const defaultListId = computed(() => store.state.auth.defaultListId)
const defaultNamespaceId = computed(() => store.state.namespaces.namespaces?.[0]?.id || 0)
const hasLists = computed (() => store.state.namespaces.namespaces?.[0]?.lists.length > 0)
const loading = computed(() => store.state.loading && store.state.loadingModule === 'tasks')
const deletionScheduledAt = computed(() => parseDateOrNull(store.state.auth.info?.deletionScheduledAt))

// This is to reload the tasks list after adding a new task through the global task add.
// FIXME: Should use vuex (somehow?)
const showTasksKey = ref(0)
function updateTaskList() {
	showTasksKey.value++
}
</script>

<style lang="scss" scoped>
.list-cards-wrapper-2-rows {
	flex-wrap: wrap;
	max-height: calc(#{$list-height * 2} + #{$list-spacing * 2} - 4px);
	overflow: hidden;

	@media screen and (max-width: $mobile) {
		max-height: calc(#{$list-height * 4} + #{$list-spacing * 4} - 4px);
	}
}
</style>