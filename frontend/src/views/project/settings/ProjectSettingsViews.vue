<script setup lang="ts">
import CreateEdit from '@/components/misc/CreateEdit.vue'
import {computed, watch, ref} from 'vue'
import {useQuery} from '@tanstack/vue-query'
import type {ProjectView} from '@/client/generated'
import {
	createProjectView,
	createProjectViewDraft,
	deleteProjectView,
	projectViewsQuery,
	updateProjectView,
	type ProjectViewDraft,
} from '@/client/queries/projectViews'
import ViewEditForm from '@/components/project/views/ViewEditForm.vue'
import XButton from '@/components/input/Button.vue'
import {error, success} from '@/message'
import {useI18n} from 'vue-i18n'
import {PERMISSIONS} from '@/constants/permissions'
import Message from '@/components/misc/Message.vue'
import draggable from 'zhyswan-vuedraggable'
import {calculateItemPosition} from '@/helpers/calculateItemPosition'
import {useProject} from '@/composables/useProject'
import ErrorMessage from '@/components/misc/Error.vue'

const props = defineProps<{
	projectId: number
}>()

const {t} = useI18n()
const query = useQuery(computed(() => projectViewsQuery(props.projectId)))
const {project, error: projectError} = useProject(() => props.projectId)
const loadError = computed(() => projectError.value ?? query.error.value)

const views = ref<ProjectView[]>([])
watch(
	query.data,
	allViews => {
		views.value = [...(allViews ?? [])]
	},
	{
		deep: true,
		immediate: true,
	},
)

const showCreateForm = ref(false)

type EditableProjectView = ProjectViewDraft & Pick<ProjectView, 'id' | 'project_id'>
const createNewView = (): EditableProjectView => ({
	...createProjectViewDraft(),
	project_id: props.projectId,
})
const newView = ref<EditableProjectView>(createNewView())
const viewIdToDelete = ref<number | null>(null)
const showDeleteModal = ref(false)
const viewToEdit = ref<ProjectView | null>(null)
const isMutating = ref(false)
const isLoading = computed(() => query.isPending.value || isMutating.value)

const isAdmin = computed(() => project.value?.max_permission === PERMISSIONS.ADMIN)

async function createView() {
	if (!showCreateForm.value) {
		showCreateForm.value = true
		return
	}

	if (newView.value.title === '') {
		return
	}

	try {
		isMutating.value = true
		newView.value.bucket_configuration_mode = newView.value.view_kind === 'kanban'
			? newView.value.bucket_configuration_mode
			: 'none'

		await createProjectView({projectId: props.projectId, view: newView.value})
		success({message: t('project.views.createSuccess')})
		showCreateForm.value = false
		newView.value = createNewView()
	} catch (e) {
		error(e)
	} finally {
		isMutating.value = false
	}
}

async function deleteView(viewId: number | null) {
	if (!viewId) {
		return
	}

	isMutating.value = true
	try {
		await deleteProjectView({projectId: props.projectId, viewId})
		showDeleteModal.value = false
	} finally {
		isMutating.value = false
	}
}

async function saveView(view: ProjectView) {
	if (!view.id) {
		return
	}
	const updated = createProjectViewDraft(view)
	if (updated.view_kind !== 'kanban') {
		updated.bucket_configuration_mode = 'none'
	}
	isMutating.value = true
	try {
		await updateProjectView({
			projectId: props.projectId,
			viewId: view.id,
			view: updated,
		})
		viewToEdit.value = null
		success({message: t('project.views.updateSuccess')})
	} finally {
		isMutating.value = false
	}
}

async function saveViewPosition(e: {newIndex: number}) {
	const view = views.value[e.newIndex]
	if (!view?.id) {
		return
	}
	const viewBefore = views.value[e.newIndex - 1]
	const viewAfter = views.value[e.newIndex + 1]
	
	const position = calculateItemPosition(
		viewBefore?.position,
		viewAfter?.position,
	)
	isMutating.value = true
	try {
		await updateProjectView({
			projectId: props.projectId,
			viewId: view.id,
			view: createProjectViewDraft({...view, position}),
		})
		success({message: t('project.views.updateSuccess')})
	} finally {
		isMutating.value = false
	}
}
</script>

<template>
	<CreateEdit
		:title="$t('project.views.header')"
		:primary-label="$t('misc.save')"
		:has-primary-action="false"
	>
		<ErrorMessage v-if="loadError" />
		<template v-else>
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
					:loading="isLoading"
					:disabled="showCreateForm && newView.title === ''"
					@click="createView"
				>
					{{ $t('project.views.create') }}
				</XButton>
			</div>

			<Message v-if="!isAdmin">
				{{ $t('project.views.onlyAdminsCanEdit') }}
			</Message>

			<div
				v-if="views?.length > 0"
				class="has-horizontal-overflow"
			>
				<table class="table has-actions is-striped is-hoverable is-fullwidth">
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
											:loading="isLoading"
											:show-save-buttons="true"
											@cancel="viewToEdit = null"
											@update:modelValue="saveView"
										/>
									</td>
								</template>
								<template v-else>
									<td>{{ v.title }}</td>
									<td>{{ v.view_kind }}</td>
									<td class="has-text-end actions">
										<XButton
											v-if="isAdmin"
											class="is-danger mie-2"
											:aria-label="$t('project.views.delete')"
											icon="trash-alt"
											@click="() => {
												viewIdToDelete = v.id ?? null
												showDeleteModal = true
											}"
										/>
										<XButton
											v-if="isAdmin"
											icon="pen"
											:aria-label="$t('project.views.edit')"
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
			</div>
		</template>
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
