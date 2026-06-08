import {ref, computed} from 'vue'
import {acceptHMRUpdate, defineStore} from 'pinia'

import {useWebSocket} from '@/composables/useWebSocket'
import {useTimeEntryService, parseTimeEntry} from '@/services/timeEntry'
import {useAuthStore} from '@/stores/auth'

import type {ITimeEntry} from '@/modelTypes/ITimeEntry'

export const useTimeTrackingStore = defineStore('timeTracking', () => {
	const activeTimer = ref<ITimeEntry | null>(null)
	const browsedEntries = ref<ITimeEntry[]>([])

	const hasActiveTimer = computed(() => activeTimer.value !== null)

	async function browseEntries(filter: string) {
		const {items} = await useTimeEntryService().getAll({
			filter,
			filterTimezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
			perPage: 250,
		})
		browsedEntries.value = items
	}

	// Drop a deleted entry from the list and clear the active timer if it was it.
	// Shared by the local delete and the cross-tab WebSocket "timer.deleted".
	function applyTimerDeletion(id: number) {
		browsedEntries.value = browsedEntries.value.filter(entry => entry.id !== id)
		if (activeTimer.value?.id === id) {
			activeTimer.value = null
		}
	}

	async function removeEntry(id: number) {
		await useTimeEntryService().remove(id)
		applyTimerDeletion(id)
	}

	// Replace an already-loaded entry in place so a stop (or any update) is
	// reflected without a refetch. Never inserts — an event for an entry that
	// isn't in the current filter shouldn't appear in the list.
	function patchInList(entry: ITimeEntry) {
		const index = browsedEntries.value.findIndex(existing => existing.id === entry.id)
		if (index !== -1) {
			browsedEntries.value.splice(index, 1, entry)
		}
	}

	// Reconcile the active timer from a timer event (WebSocket) or a local
	// action: an entry with an end time is a stop — clear it if it's the one we
	// track; otherwise it is the running timer.
	function applyTimerEvent(entry: ITimeEntry) {
		patchInList(entry)
		if (entry.endTime !== null) {
			if (activeTimer.value?.id === entry.id) {
				activeTimer.value = null
			}
			return
		}
		activeTimer.value = entry
	}

	// Source of truth on (re)connect: the caller's own running timer, if any.
	async function hydrateActiveTimer() {
		const userId = useAuthStore().info?.id
		if (userId === undefined) {
			activeTimer.value = null
			return
		}

		const {items} = await useTimeEntryService().getAll({
			filter: `user_id = ${userId} && end_time = null`,
			perPage: 1,
		})
		activeTimer.value = items[0] ?? null
	}

	// Create any entry (manual, with an end time, or a running timer when end is
	// omitted) and reconcile the active timer from the result.
	async function createEntry(payload: Partial<ITimeEntry>) {
		const entry = await useTimeEntryService().create(payload)
		applyTimerEvent(entry)
		return entry
	}

	async function updateEntry(payload: Partial<ITimeEntry> & {id: number}) {
		const entry = await useTimeEntryService().update(payload)
		applyTimerEvent(entry)
		return entry
	}

	async function stopTimer() {
		const entry = await useTimeEntryService().stopTimer()
		applyTimerEvent(entry)
		return entry
	}

	let unsubscribers: Array<() => void> = []
	function subscribeToTimerEvents() {
		const {subscribe} = useWebSocket()
		// Ignore messages without a payload (e.g. subscribe acknowledgements).
		const onEvent = (msg: {data?: unknown}) => {
			if (msg.data == null) {
				return
			}
			applyTimerEvent(parseTimeEntry(msg.data as Record<string, unknown>))
		}
		const onDelete = (msg: {data?: unknown}) => {
			if (msg.data == null) {
				return
			}
			applyTimerDeletion(parseTimeEntry(msg.data as Record<string, unknown>).id)
		}
		unsubscribers.push(subscribe('timer.created', onEvent))
		unsubscribers.push(subscribe('timer.updated', onEvent))
		unsubscribers.push(subscribe('timer.deleted', onDelete))
	}
	function unsubscribeFromTimerEvents() {
		unsubscribers.forEach(unsubscribe => unsubscribe())
		unsubscribers = []
	}

	return {
		activeTimer,
		browsedEntries,
		hasActiveTimer,
		applyTimerEvent,
		applyTimerDeletion,
		hydrateActiveTimer,
		browseEntries,
		createEntry,
		updateEntry,
		stopTimer,
		removeEntry,
		subscribeToTimerEvents,
		unsubscribeFromTimerEvents,
	}
})

if (import.meta.hot) {
	import.meta.hot.accept(acceptHMRUpdate(useTimeTrackingStore, import.meta.hot))
}
