<script setup lang="ts">
import CreateEdit from '@/components/misc/CreateEdit.vue'
import {watch, ref, computed} from 'vue'
import {useProjectStore} from '@/stores/projects'
import ProjectViewModel from '@/models/projectView'
import type {IProjectView} from '@/modelTypes/IProjectView'
import ViewEditForm from '@/components/project/views/ViewEditForm.vue'
import ProjectViewService from '@/services/projectViews'
import XButton from '@/components/input/Button.vue'
import {error, success} from '@/message'
import {useI18n} from 'vue-i18n'
import ProjectService from '@/services/project'
import {RIGHTS} from '@/constants/rights'
import ProjectModel from '@/models/project'
import Message from '@/components/misc/Message.vue'

const {
	projectId,
} = defineProps<{
	projectId: number
}>()

const projectStore = useProjectStore()
const {t} = useI18n()

const views = computed(() => projectStore.projects[projectId]?.views)
const showCreateForm = ref(false)

const projectViewService = ref(new ProjectViewService())
const newView = ref<IProjectView>(new ProjectViewModel({}))
const viewIdToDelete = ref<number | null>(null)
const showDeleteModal = ref(false)
const viewToEdit = ref<IProjectView | null>(null)

const isAdmin = ref<boolean>(false)
watch(
	() => projectId,
	async () => {
		const projectService = new ProjectService()
		const project = await projectService.get(new ProjectModel({id: projectId}))
		isAdmin.value = project.maxRight === RIGHTS.ADMIN
	},
	{immediate: true},
)

async function createView() {
	if (!showCreateForm.value) {
		showCreateForm.value = true
		return
	}

	if (newView.value.title === '') {
		return
	}

	try {
		newView.value.bucketConfigurationMode = newView.value.viewKind === 'kanban'
			? newView.value.bucketConfigurationMode
			: 'none'
		newView.value.projectId = projectId

		const result: IProjectView = await projectViewService.value.create(newView.value)
		success({message: t('project.views.createSuccess')})
		showCreateForm.value = false
		projectStore.setProjectView(result)
		newView.value = new ProjectViewModel({})
	} catch (e) {
		error(e)
	}
}

async function deleteView() {
	if (!viewIdToDelete.value) {
		return
	}

	await projectViewService.value.delete(new ProjectViewModel({
		id: viewIdToDelete.value,
		projectId,
	}))

	projectStore.removeProjectView(projectId, viewIdToDelete.value)

	showDeleteModal.value = false
}

async function saveView() {
	if (viewToEdit.value?.viewKind !== 'kanban') {
		viewToEdit.value.bucketConfigurationMode = 'none'
	}
	const result = await projectViewService.value.update(viewToEdit.value)
	projectStore.setProjectView(result)
	viewToEdit.value = null
}
</script>

<template>
	<CreateEdit
		:title="$t('project.views.header')"
		:primary-label="$t('misc.save')"
		:has-primary-action="false"
	>
		<ViewEditForm
			v-if="showCreateForm"
			v-model="newView"
			class="mb-4"
		/>
		<div
			v-if="isAdmin"
			class="is-flex is-justify-content-end mb-4"
		>
			<XButton
				:loading="projectViewService.loading"
				:disabled="showCreateForm && newView.title === ''"
				@click="createView"
			>
				{{ $t('project.views.create') }}
			</XButton>
		</div>
		
		<Message v-if="!isAdmin">
			{{ $t('project.views.onlyAdminsCanEdit') }}
		</Message>

		<table
			v-if="views?.length > 0"
			class="table has-actions is-striped is-hoverable is-fullwidth"
		>
			<thead>
				<tr>
					<th>{{ $t('project.views.title') }}</th>
					<th>{{ $t('project.views.kind') }}</th>
					<th class="has-text-right">
						{{ $t('project.views.actions') }}
					</th>
				</tr>
			</thead>
			<tbody>
				<tr
					v-for="v in views"
					:key="v.id"
				>
					<template v-if="viewToEdit !== null && viewToEdit.id === v.id">
						<td colspan="3">
							<ViewEditForm
								v-model="viewToEdit"
								class="mb-4"
								:loading="projectViewService.loading"
								:show-save-buttons="true"
								@cancel="viewToEdit = null"
								@update:modelValue="saveView"
							/>
						</td>
					</template>
					<template v-else>
						<td>{{ v.title }}</td>
						<td>{{ v.viewKind }}</td>
						<td class="has-text-right">
							<XButton
								v-if="isAdmin"
								class="is-danger mr-2"
								icon="trash-alt"
								@click="() => {
									viewIdToDelete = v.id
									showDeleteModal = true
								}"
							/>
							<XButton
								v-if="isAdmin"
								icon="pen"
								@click="viewToEdit = {...v}"
							/>
						</td>
					</template>
				</tr>
			</tbody>
		</table>
	</CreateEdit>

	<modal
		:enabled="showDeleteModal"
		@close="showDeleteModal = false"
		@submit="deleteView"
	>
		<template #header>
			<span>{{ $t('project.views.delete') }}</span>
		</template>

		<template #text>
			<p>{{ $t('project.views.deleteText') }}</p>
		</template>
	</modal>
</template>
