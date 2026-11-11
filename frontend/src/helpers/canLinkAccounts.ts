import type {ConfigState} from '@/stores/config'

export function canLinkAccounts(
	openidConnect: ConfigState['auth']['openid_connect'],
	info: {is_local_user?: boolean} | null,
): boolean {
	return openidConnect.providers.length > 0
		&& info?.is_local_user === true
}
