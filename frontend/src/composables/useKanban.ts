import {computed, ref, toValue, watch, type MaybeRefOrGetter} from 'vue'
import {useQuery} from '@tanstack/vue-query'
import {klona} from 'klona/lite'
import {kanbanQuery, type BucketResponse} from '@/client/queries/kanban'
import type {TaskFilterParams} from '@/client/queries/tasks'

export function useKanban(
	project: MaybeRefOrGetter<number>,
	view: MaybeRefOrGetter<number>,
	params: MaybeRefOrGetter<TaskFilterParams>,
) {
	const query = useQuery(computed(() => kanbanQuery(toValue(project), toValue(view), toValue(params))))
	const dragBuckets = ref<BucketResponse[] | null>(null)
	const buckets = computed(() => dragBuckets.value ?? query.data.value?.buckets ?? [])
	function startDrag() { dragBuckets.value = klona(query.data.value?.buckets ?? []) }
	function endDrag() { dragBuckets.value = null }
	watch([() => toValue(project), () => toValue(view), () => toValue(params)], endDrag)
	return {...query, buckets, dragBuckets, startDrag, endDrag}
}
