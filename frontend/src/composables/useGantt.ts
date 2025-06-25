import {ref, shallowRef, watch, toValue, type Ref} from 'vue'
import {tryOnUnmounted} from '@vueuse/core'

export interface GanttBarInput { id: string; start: Date; end: Date; meta?: Record<string, unknown> }
export interface TimeScale { range: Date[]; unit: 'hour' | 'day' }

export interface UseGanttOptions {
  startDate: Date
  endDate: Date
  unit?: 'hour' | 'day'
  bars?: GanttBarInput[]
  window?: Window | undefined
  watchOptions?: { immediate?: boolean; flush?: 'pre'|'post'|'sync' }
}

export interface UseGanttReturn {
  timeScale: Ref<TimeScale>
  bars: Ref<GanttBarInput[]>
  isDragSupported: Ref<boolean>
  moveBar(id: string, newStart: Date, newEnd: Date): void
  zoom(newStart: Date, newEnd: Date): void
}

function createRange(start: Date, end: Date, unit: 'hour' | 'day'): Date[] {
	const range: Date[] = []
	const current = new Date(start)
	while (current <= end) {
		range.push(new Date(current))
		if (unit === 'hour') {
			current.setHours(current.getHours() + 1)
		} else {
			current.setDate(current.getDate() + 1)
		}
	}
	return range
}

export function useGantt(options: UseGanttOptions): UseGanttReturn {
	const bars = shallowRef(options.bars ?? [])
	const timeScale = ref<TimeScale>({
		range: createRange(options.startDate, options.endDate, options.unit ?? 'day'),
		unit: options.unit ?? 'day',
	})
	const isDragSupported = ref(typeof window !== 'undefined' && 'PointerEvent' in window)

	function moveBar(id: string, newStart: Date, newEnd: Date) {
		const idx = bars.value.findIndex(b => b.id === id)
		if (idx !== -1) {
			bars.value[idx] = { ...bars.value[idx], start: newStart, end: newEnd }
		}
	}

	function zoom(newStart: Date, newEnd: Date) {
		timeScale.value = {
			range: createRange(newStart, newEnd, timeScale.value.unit),
			unit: timeScale.value.unit,
		}
	}

	watch(
		() => [toValue(options.startDate), toValue(options.endDate), options.unit],
		([start, end, unit]) => {
			timeScale.value.unit = unit ?? 'day'
			zoom(start, end)
		},
		options.watchOptions,
	)

	tryOnUnmounted(() => {
		// nothing to cleanup yet
	})

	return { timeScale, bars, isDragSupported, moveBar, zoom }
}
