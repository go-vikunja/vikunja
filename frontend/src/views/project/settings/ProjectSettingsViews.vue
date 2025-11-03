<script setup lang="ts">
import CreateEdit from '@/components/misc/CreateEdit.vue'
import {watch, ref, shallowReactive} from 'vue'
import {useProjectStore} from '@/stores/projects'
import ProjectViewModel from '@/models/projectView'
import type {IProjectView} from '@/modelTypes/IProjectView'
import ViewEditForm from '@/components/project/views/ViewEditForm.vue'
import ProjectViewService from '@/services/projectViews'
import XButton from '@/components/input/Button.vue'
import {error, success} from '@/message'
import {useI18n} from 'vue-i18n'
import ProjectService from '@/services/project'
import {PERMISSIONS} from '@/constants/permissions'
import ProjectModel from '@/models/project'
import Message from '@/components/misc/Message.vue'
import draggable from 'zhyswan-vuedraggable'
import {calculateItemPosition} from '@/helpers/calculateItemPosition'

const props = defineProps<{
	projectId: number
}>()

const projectStore = useProjectStore()
const {t} = useI18n()

const views = ref<IProjectView[]>([])
watch(
	() => projectStore.projects[props.projectId]?.views || [],
	allViews => {
		views.value = [...allViews]
	},
	{
		deep: true,
		immediate: true,
	},
)

const showCreateForm = ref(false)

const projectViewService = shallowReactive(new ProjectViewService())
const newView = ref<IProjectView>(ProjectViewModel.createWithDefaultFilter())
const viewIdToDelete = ref<number | null>(null)
const showDeleteModal = ref(false)
const viewToEdit = ref<IProjectView | null>(null)

const isAdmin = ref<boolean>(false)
watch(
	() => props.projectId,
	async () => {
		const projectService = new ProjectService()
		const project = await projectService.get(new ProjectModel({id: props.projectId}))
		isAdmin.value = project.maxPermission === PERMISSIONS.ADMIN
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
		newView.value.projectId = props.projectId

		const result: IProjectView = await projectViewService.create(newView.value)
		success({message: t('project.views.createSuccess')})
		showCreateForm.value = false
		projectStore.setProjectView(result)
		newView.value = new ProjectViewModel({})
	} catch (e) {
		error(e)
	}
}

async function deleteView(viewId: number) {
	if (!viewId) {
		return
	}

	await projectViewService.delete(new ProjectViewModel({
		id: viewId,
		projectId: props.projectId,
	}))

	projectStore.removeProjectView(props.projectId, viewId)

	showDeleteModal.value = false
}

async function saveView(view: IProjectView) {
	if (view?.viewKind !== 'kanban') {
		view.bucketConfigurationMode = 'none'
	}
	const result = await projectViewService.update(view)
	projectStore.setProjectView(result)
	viewToEdit.value = null
	success({message: t('project.views.updateSuccess')})
}

async function saveViewPosition(e) {
	const view = views.value[e.newIndex]
	const viewBefore = views.value[e.newIndex - 1]
	const viewAfter = views.value[e.newIndex + 1]
	
	const position = calculateItemPosition(
		viewBefore?.position,
		viewAfter?.position,
	)
	const result = await projectViewService.update({
		...view,
		position,
	})
	projectStore.setProjectView(result)
	success({message: t('project.views.updateSuccess')})
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
			class="mbe-4"
		/>
		<div
			v-if="isAdmin"
			class="is-flex is-justify-content-end mbe-4"
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
					<th class="has-text-end">
						{{ $t('project.views.actions') }}
					</th>
				</tr>
			</thead>
			<draggable
				v-model="views"
				tag="tbody"
				item-key="id"
				handle=".handle"
				:animation="100"
				@end="saveViewPosition"
			>
				<template #item="{element: v}">
					<tr>
						<template v-if="viewToEdit !== null && viewToEdit.id === v.id">
							<td colspan="3">
								<ViewEditForm
									v-model="viewToEdit"
									class="mbe-4"
									:loading="projectViewService.loading"
									:show-save-buttons="true"
									@cancel="viewToEdit = null"
									@update:modelValue="saveView(viewToEdit)"
								/>
							</td>
						</template>
						<template v-else>
							<td>{{ v.title }}</td>
							<td>{{ v.viewKind }}</td>
							<td class="has-text-end actions">
								<XButton
									v-if="isAdmin"
									class="is-danger mie-2"
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
								<span class="icon handle">
									<Icon icon="grip-lines" />
								</span>
							</td>
						</template>
					</tr>
				</template>
			</draggable>
		</table>
	</CreateEdit>

	<Modal
		:enabled="showDeleteModal"
		@close="showDeleteModal = false"
		@submit="deleteView(viewIdToDelete)"
	>
		<template #header>
			<span>{{ $t('project.views.delete') }}</span>
		</template>

		<template #text>
			<p>{{ $t('project.views.deleteText') }}</p>
		</template>
	</Modal>
</template>

<style scoped>
.handle {
	cursor: grab;
	margin-inline-start: .25rem;
}

.actions {
	display: flex;
	align-items: center;
	justify-content: flex-end;
}
</style>
