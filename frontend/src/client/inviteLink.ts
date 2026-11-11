import type {RegisterUserRequestWritable} from './generated'

// Keep static imports type-only so this runs before router and telemetry.
let token: string | null = null
const replaceState = window.history.replaceState.bind(window.history)
const listeners = new Set<() => void>()
function consumeFragment() {
	if (!/\/register\/?$/i.test(window.location.pathname) || !window.location.hash.startsWith('#invite-link=')) return
	token = new URLSearchParams(window.location.hash.slice(1)).get('invite-link')
	replaceState(window.history.state, '', window.location.pathname + window.location.search)
	listeners.forEach(listener => listener())
}
consumeFragment()
window.addEventListener('popstate', consumeFragment, true)
window.addEventListener('hashchange', consumeFragment, true)

export function onInviteLinkChange(listener: () => void): () => void {
	listeners.add(listener)
	return () => { listeners.delete(listener) }
}

export function hasInviteLink(): boolean {
	return token !== null
}

export async function checkInviteLink() {
	if (token === null) throw new Error('No invitation')
	const currentToken = token
	const {inviteLinksCheck} = await import('./generated')
	return inviteLinksCheck({body: {token: currentToken}})
}

export async function registerViaInviteLink(credentials: RegisterUserRequestWritable) {
	if (token === null) throw new Error('No invitation')
	const currentToken = token
	const {authRegister} = await import('./generated')
	const response = await authRegister({body: {...credentials, invite_token: currentToken}})
	if (token === currentToken) token = null
	return response
}
