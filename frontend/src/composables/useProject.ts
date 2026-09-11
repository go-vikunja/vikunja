import {computed, ref, toValue, watch, type MaybeRefOrGetter} from 'vue'
import {useQuery} from '@tanstack/vue-query'
import {useI18n} from 'vue-i18n'
import {useRouter} from 'vue-router'

import {
	duplicateProject,
	normalizeProject,
	projectQuery,
	updateProject,
} from '@/client/queries/projects'
import type {ProjectResponse} from '@/client/queries/projects'
import {success} from '@/message'

export function useProject(projectId: MaybeRefOrGetter<number>) {
	const router = useRouter()
	const {t} = useI18n({useScope: 'global'})
	const query = useQuery(computed(() => projectQuery(Number(toValue(projectId)))))
	const project = ref<ProjectResponse>(normalizeProject({id: Number(toValue(projectId))}))
	const loadedProjectId = ref(0)
	const isSaving = ref(false)

	watch([
		() => Number(toValue(projectId)),
		query.data,
		query.isFetching,
		query.isError,
	], ([id, value, isFetching]) => {
		if (loadedProjectId.value !== id) {
			project.value = normalizeProject({id})
		}
		if (!isFetching && value && loadedProjectId.value !== id) {
			project.value = normalizeProject(value)
			loadedProjectId.value = id
		}
	}, {immediate: true})

	async function save() {
		if (loadedProjectId.value !== Number(toValue(projectId))) {
			throw new Error('Project details are not loaded')
		}
		isSaving.value = true
		try {
			project.value = await updateProject({
				id: project.value.id,
				title: project.value.title,
				description: project.value.description,
				hex_color: project.value.hex_color,
				identifier: project.value.identifier,
				is_archived: project.value.is_archived,
				is_favorite: project.value.is_favorite,
				parent_project_id: project.value.parent_project_id,
				position: project.value.position,
			})
			success({message: t('project.edit.success')})
		} finally {
			isSaving.value = false
		}
	}

	async function duplicate(parentProjectId: number, duplicateShares = false) {
		const duplicated = await duplicateProject({
			projectId: Number(toValue(projectId)),
			parentProjectId,
			duplicateShares,
		})
		success({message: t('project.duplicate.success')})
		await router.push({name: 'project.index', params: {projectId: duplicated.id}})
	}

	return {
		project,
		isLoading: computed(() =>
			query.isFetching.value ||
			(!query.isError.value && loadedProjectId.value !== Number(toValue(projectId))) ||
			isSaving.value,
		),
		error: query.error,
		save,
		duplicateProject: duplicate,
	}
}
