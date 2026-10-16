import {computed, toValue, type MaybeRefOrGetter} from 'vue'
import {useI18n} from 'vue-i18n'
import type {Label, ProjectView, User} from '@/client/generated'
import {ensureLabels, getLabelById, refreshLabels} from '@/client/queries/labels'
import type {FilterBucketPlacement} from '@/client/queries/taskCache'
import type {FilterBucketMoveInput} from '@/client/queries/taskMutations'
import {refreshTask, type TaskFilterParams, type TaskResponse} from '@/client/queries/tasks'
import {searchProjectUsers} from '@/client/queries/userSearch'
import {
	findFilterValue,
	parseBucketFilter,
	planBucketFilterMove,
	type BucketFilterEdit,
} from '@/helpers/filterBuckets'
import {error} from '@/message'

type WithId<T> = T & {id: number}

type Drop = {
	task: TaskResponse
	from: number
	to: number
	index: number
	position: number
}

const MEMBERSHIP_FIELDS = /\b(assignees|labels)\b/

export function useFilterBucketDrop(
	project: MaybeRefOrGetter<number>,
	view: MaybeRefOrGetter<ProjectView | null>,
	params: MaybeRefOrGetter<TaskFilterParams>,
) {
	const {t} = useI18n({useScope: 'global'})

	const filters = computed(() => (toValue(view)?.bucket_configuration ?? [])
		.map(bucket => parseBucketFilter(bucket.filter)))
	const filterOf = (bucketId: number) => filters.value[bucketId] ?? null
	const canDrop = (bucketId: number) => filterOf(bucketId) !== null

	async function findLabel(value: string) {
		const id = Number(value)
		return getLabelById(await ensureLabels(), id) ?? getLabelById(await refreshLabels(), id)
	}

	async function resolve(task: TaskResponse, {field, value, add}: BucketFilterEdit) {
		const found = add
			? field === 'assignees'
				? (await searchProjectUsers(task.project_id, value)).find(user => user.username === value)
				: await findLabel(value)
			: findFilterValue(task, field, value)
		if (found?.id === undefined) throw new Error(t('project.kanban.filterBucketValueNotFound', {value}))
		return {
			field,
			add,
			item: found as WithId<User | Label>,
		}
	}

	async function planWrites(task: TaskResponse, edits: BucketFilterEdit[]) {
		const resolved = await Promise.all(edits.map(edit => resolve(task, edit)))
		const labels = resolved
			.filter(edit => edit.field === 'labels')
			.map(({add, item}) => ({
				add,
				label: item as WithId<Label>,
			}))
		const assigneeEdits = resolved.filter(edit => edit.field === 'assignees')
		const assignees = assigneeEdits.length === 0
			? null
			: assigneeEdits.reduce((current, {add, item}) => [
				...current.filter(user => user.id !== item.id),
				...(add ? [item as WithId<User>] : []),
			], task.assignees as WithId<User>[])
		return {
			labels,
			assignees,
		}
	}

	function placement(target: number, index: number): FilterBucketPlacement | null {
		const current = toValue(view)
		const boardParams = toValue(params)
		const unevaluated = [
			current?.filter?.filter,
			boardParams.filter,
			...(current?.bucket_configuration ?? [])
				.filter((_, id) => filters.value[id] === null)
				.map(bucket => bucket.filter?.filter),
		]
		if (toValue(project) < 0 || MEMBERSHIP_FIELDS.test(unevaluated.join(' '))) return null
		return {
			filters: filters.value,
			includeNulls: Boolean(current?.filter?.filter_include_nulls || boardParams.filter_include_nulls),
			target,
			index,
		}
	}

	async function prepare({task, from, to, index, position}: Drop): Promise<FilterBucketMoveInput | null> {
		const target = filterOf(to)
		const viewId = toValue(view)?.id
		if (!target || viewId === undefined) return null
		const {labels, assignees} = await refreshTask(task.id)
			.then(fresh => planWrites(fresh, planBucketFilterMove(fresh, filterOf(from), target)))
			.catch(cause => {
				error(cause)
				throw cause
			})
		return {
			project: toValue(project),
			view: viewId,
			params: toValue(params),
			taskId: task.id,
			labels,
			assignees,
			position,
			placement: placement(to, index),
		}
	}

	return {
		canDrop,
		prepare,
	}
}
