<script setup lang="ts">
import CreateEdit from '@/components/misc/create-edit.vue'
import {computed, ref} from 'vue'
import {useProjectStore} from '@/stores/projects'
import ProjectViewModel from '@/models/projectView'
import type {IProjectView} from '@/modelTypes/IProjectView'
import ViewEditForm from '@/components/project/views/viewEditForm.vue'
import ProjectViewService from '@/services/projectViews'
import XButton from '@/components/input/button.vue'
import {error, success} from '@/message'
import {useI18n} from 'vue-i18n'

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

async function createView() {
	if (!showCreateForm.value) {
		showCreateForm.value = true
		return
	}

	if (newView.value.title === '') {
		return
	}

	try {
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
	const result = await projectViewService.value.update(viewToEdit.value)
	projectStore.setProjectView(result)
	viewToEdit.value = null
}
</script>

<template>
	<CreateEdit
		:title="$t('project.views.header')"
		:primary-label="$t('misc.save')"
	>
		<ViewEditForm
			v-if="showCreateForm"
			v-model="newView"
			class="mb-4"
		/>
		<div class="is-flex is-justify-content-end">
			<x-button
				@click="createView"
				:loading="projectViewService.loading"
			>
				{{ $t('project.views.create') }}
			</x-button>
		</div>

		<table
			v-if="views?.length > 0"
			class="table has-actions is-striped is-hoverable is-fullwidth"
		>
			<thead>
			<tr>
				<th>{{ $t('project.views.title') }}</th>
				<th>{{ $t('project.views.kind') }}</th>
				<th class="has-text-right">{{ $t('project.views.actions') }}</th>
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
						/>
						<div class="is-flex is-justify-content-end">
							<x-button
								variant="tertiary"
								@click="viewToEdit = null"
								class="mr-2"
							>
								{{ $t('misc.cancel') }}
							</x-button>
							<x-button
								@click="saveView"
								:loading="projectViewService.loading"
							>
								{{ $t('misc.save') }}
							</x-button>
						</div>
					</td>
				</template>
				<template v-else>
					<td>{{ v.title }}</td>
					<td>{{ v.viewKind }}</td>
					<td class="has-text-right">
						<x-button
							class="is-danger mr-2"
							icon="trash-alt"
							@click="() => {
							viewIdToDelete = v.id
							showDeleteModal = true
						}"
						/>
						<x-button
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

<style scoped lang="scss">

</style>