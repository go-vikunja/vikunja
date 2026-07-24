import {computed} from 'vue'

import {formatDate} from '@/helpers/time/formatDate'
import {useTimeFormat} from '@/composables/useTimeFormat'
import {TIME_FORMAT} from '@/constants/timeFormat'

// Locale-aware clock-time label honouring the user's 12/24h preference, using
// the same format strings as formatDisplayDateFormat so the planner matches
// the rest of the app.
export function usePlannerTimeFormatter() {
	const {store: timeFormat} = useTimeFormat()
	return computed(() => (date: Date | string) =>
		formatDate(date, timeFormat.value === TIME_FORMAT.HOURS_24 ? 'HH:mm' : 'hh:mm A'),
	)
}
