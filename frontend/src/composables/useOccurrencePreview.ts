import {computed, onScopeDispose, shallowRef, toValue, watch, type MaybeRefOrGetter} from 'vue'
import {hashKey, useQuery, useQueryClient} from '@tanstack/vue-query'

import type {TaskRepeatWritable} from '@/client/generated'
import {
	classifyPreviewError,
	occurrenceKeys,
	occurrencePreviewQuery,
	PREVIEW_COUNT,
	type OccurrencePreviewInput,
} from '@/client/queries/occurrencePreview'
import type {classifyRepeat} from '@/helpers/rrule'
import {error} from '@/message'

export const PREVIEW_DEBOUNCE_MS = 250

export type OccurrencePreviewSource = {
	repeat: TaskRepeatWritable | null,
	repeatClass: ReturnType<typeof classifyRepeat>,
	dueDate: Date | null,
	initialDueTime: string | null,
	taskId: number,
	done: boolean,
	repeatsFromCurrentDate: boolean,
}

export type OccurrencePreviewState = 'disabled' | 'loading' | 'ready' | 'stale' | 'ruleHint' | 'unavailable'

// An undated done task will not seed on save, so previewing a generated date for it would promise one it never gets.
export function previewInputFor(source: OccurrencePreviewSource | null): OccurrencePreviewInput | null {
	if (!source || !source.repeat || source.repeatClass !== 'editable' || source.taskId <= 0) return null
	const common = {
		// A plain copy: in-place edits of the editor's rule must reach the key only through the debounce.
		repeat: JSON.parse(JSON.stringify(source.repeat)) as TaskRepeatWritable,
		task_id: source.taskId,
		repeats_from_current_date: source.repeatsFromCurrentDate,
		count: PREVIEW_COUNT,
	}
	if (source.dueDate) return {...common, due_date: source.dueDate.toISOString()}
	if (source.done || !source.initialDueTime) return null
	return {...common, initial_due_time: source.initialDueTime}
}

function sameRequest(a: OccurrencePreviewInput | null, b: OccurrencePreviewInput | null) {
	return hashKey(occurrenceKeys.preview(a)) === hashKey(occurrenceKeys.preview(b))
}

export function useOccurrencePreview(source: MaybeRefOrGetter<OccurrencePreviewSource | null>) {
	const queryClient = useQueryClient()
	const input = computed(() => previewInputFor(toValue(source)))
	const settled = shallowRef<OccurrencePreviewInput | null>(null)
	let timer: ReturnType<typeof setTimeout> | undefined
	// Only edits to the same task wait: switching task, enabling or disabling settles at once, and later edits wait again.
	watch(input, latest => {
		clearTimeout(timer)
		const current = settled.value
		if (!latest || !current || latest.task_id !== current.task_id) {
			settled.value = latest
			return
		}
		if (sameRequest(latest, current)) return
		timer = setTimeout(() => {
			settled.value = latest
		}, PREVIEW_DEBOUNCE_MS)
	}, {immediate: true})
	onScopeDispose(() => clearTimeout(timer))

	const query = useQuery(computed(() => occurrencePreviewQuery(settled.value)))

	// The query opts out of the global error toast, so a 401 has to be surfaced here as everywhere else.
	watch(query.error, cause => {
		if (cause && classifyPreviewError(cause) === 'auth') error(cause)
	})

	const state = computed<OccurrencePreviewState>(() => {
		if (!input.value) return 'disabled'
		if (query.isError.value) {
			const kind = classifyPreviewError(query.error.value)
			return kind === 'ruleHint' ? 'ruleHint' : 'unavailable'
		}
		if (!query.data.value) return 'loading'
		// A cached answer shown while it is asked again can already be out of date: the server evaluates against its clock.
		if (query.isPlaceholderData.value || query.isFetching.value || !sameRequest(input.value, settled.value)) return 'stale'
		return 'ready'
	})

	const view = computed(() => state.value === 'ready' || state.value === 'stale'
		? query.data.value ?? null
		: null)

	function cancel() {
		const taskId = toValue(source)?.taskId
		if (!taskId) return Promise.resolve()
		return queryClient.cancelQueries({queryKey: occurrenceKeys.task(taskId)})
	}

	return {
		state,
		view,
		isFetching: query.isFetching,
		cancel,
	}
}
