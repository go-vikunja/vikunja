import {parseDateOrNull} from '@/helpers/parseDateOrNull'

export function toISOStringOrNull(date: Date | string | null | undefined): string | null {
	return parseDateOrNull(date)?.toISOString() ?? null
}
