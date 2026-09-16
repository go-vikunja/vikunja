import {getFullBaseUrl} from '@/helpers/getFullBaseUrl'
import {useConfigStore} from '@/stores/config'

interface AppBase {
	origin: string
	pathname: string // with trailing slash
}

function getAcceptedBases(): AppBase[] {
	const bases: AppBase[] = [{
		origin: window.location.origin,
		pathname: getFullBaseUrl(),
	}]

	// Permalinks the app generates (comment links, mails) use the configured frontend url,
	// which can differ from the open origin.
	const {frontendUrl} = useConfigStore()
	if (!frontendUrl) {
		return bases
	}

	let base: URL
	try {
		base = new URL(frontendUrl)
	} catch {
		return bases
	}

	if (base.protocol !== 'http:' && base.protocol !== 'https:') {
		return bases
	}

	bases.push({
		origin: base.origin,
		pathname: base.pathname.endsWith('/') ? base.pathname : base.pathname + '/',
	})

	return bases
}

// Task id when `href` is this app's task detail route (current origin or configured
// frontend url, base path aware), else null.
export function parseTaskIdFromUrl(href: string): number | null {
	if (!href) {
		return null
	}

	let url: URL
	try {
		url = new URL(href, window.location.origin)
	} catch {
		return null
	}

	for (const base of getAcceptedBases()) {
		if (url.origin !== base.origin) {
			continue
		}

		const prefix = base.pathname + 'tasks/'
		if (!url.pathname.startsWith(prefix)) {
			continue
		}

		const rest = url.pathname.slice(prefix.length)
		if (!/^\d+$/.test(rest)) {
			continue
		}

		const id = Number(rest)
		if (id > 0) {
			return id
		}
	}

	return null
}
