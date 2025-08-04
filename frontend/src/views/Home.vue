<template>
	<div class="content has-text-centered">
		<h2 v-if="salutation">
			{{ salutation }}
		</h2>

		<Message
			v-if="deletionScheduledAt !== null"
			variant="danger"
			class="mbe-4"
		>
			{{
				$t('user.deletion.scheduled', {
					date: formatDisplayDate(deletionScheduledAt),
					dateSince: formatDateSince(deletionScheduledAt),
				})
			}}
			<RouterLink :to="{name: 'user.settings.deletion'}">
				{{ $t('user.deletion.scheduledCancel') }}
			</RouterLink>
		</Message>
		<AddTask
			class="is-max-width-desktop"
			@taskAdded="updateTaskKey"
		/>
		<ImportHint v-if="tasksLoaded" />
		<div
			v-if="projectHistory.length > 0"
			class="is-max-width-desktop has-text-start mbs-4"
		>
			<h3>{{ $t('home.lastViewed') }}</h3>
			<ProjectCardGrid
				v-cy="'projectCardGrid'"
				:projects="projectHistory"
				:show-even-number-of-projects="true"
			/>
		</div>
		<ShowTasks
			v-if="projectStore.hasProjects"
			:key="showTasksKey"
			class="show-tasks"
			@tasksLoaded="tasksLoaded = true"
		/>
	</div>
</template>

<script lang="ts" setup>
import {ref, computed} from 'vue'

import Message from '@/components/misc/Message.vue'
import ShowTasks from '@/views/tasks/ShowTasks.vue'
import ProjectCardGrid from '@/components/project/partials/ProjectCardGrid.vue'
import AddTask from '@/components/tasks/AddTask.vue'
import ImportHint from '@/components/home/ImportHint.vue'

import {getHistory} from '@/modules/projectHistory'
import {parseDateOrNull} from '@/helpers/parseDateOrNull'
import {formatDateSince, formatDisplayDate} from '@/helpers/time/formatDate'
import {useDaytimeSalutation} from '@/composables/useDaytimeSalutation'

import {useProjectStore} from '@/stores/projects'
import {useAuthStore} from '@/stores/auth'

const salutation = useDaytimeSalutation()

const authStore = useAuthStore()
const projectStore = useProjectStore()

const projectHistory = computed(() => {
	// If we don't check this, it tries to load the project background right after logging out	
	if(!authStore.authenticated) {
		return []
	}
	
	return getHistory()
		.map(l => projectStore.projects[l.id])
		.filter(l => Boolean(l))
})

const tasksLoaded = ref(false)

const deletionScheduledAt = computed(() => parseDateOrNull(authStore.info?.deletionScheduledAt))

// This is to reload the tasks list after adding a new task through the global task add.
// FIXME: Should use pinia (somehow?)
const showTasksKey = ref(0)

function updateTaskKey() {
	showTasksKey.value++
}
</script>

<style scoped lang="scss">
.show-tasks {
	margin-block-start: 2rem;
}
</style>
