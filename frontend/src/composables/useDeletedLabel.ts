import {useI18n} from 'vue-i18n'

import type {User} from '@/client/generated'
import {useGlobalNow} from '@/composables/useGlobalNow'
import {getDisplayName} from '@/helpers/user'
import {formatDateSince} from '@/helpers/time/formatDate'
import {useAuthStore} from '@/stores/auth'

type Trashed = {
	deleted_at?: string
	deleted_by?: User
}

// keyPrefix names three strings: keyPrefix (deleter unknown), keyPrefix + 'By' and keyPrefix + 'ByYou'.
export function useDeletedLabel(keyPrefix: string) {
	const {t} = useI18n({useScope: 'global'})
	const authStore = useAuthStore()
	// formatDateSince would otherwise start the shared clock during render, outside setup
	useGlobalNow()

	return (item: Trashed) => {
		const since = formatDateSince(item.deleted_at)
		if (!item.deleted_by) {
			return t(keyPrefix, {since})
		}
		if (item.deleted_by.id === authStore.info?.id) {
			return t(`${keyPrefix}ByYou`, {since})
		}
		return t(`${keyPrefix}By`, {user: getDisplayName(item.deleted_by), since})
	}
}
