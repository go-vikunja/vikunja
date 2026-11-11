import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest'

let eventListeners: ReturnType<typeof vi.spyOn>
const api = vi.hoisted(() => ({check: vi.fn(), register: vi.fn()}))
vi.mock('./generated', () => ({inviteLinksCheck: api.check, authRegister: api.register}))

beforeEach(() => {
	vi.resetModules()
	vi.resetAllMocks()
	eventListeners = vi.spyOn(window, 'addEventListener')
})
afterEach(() => {
	for (const [type, listener, options] of eventListeners.mock.calls) {
		window.removeEventListener(type, listener, options)
	}
	vi.restoreAllMocks()
	window.history.replaceState({}, '', '/')
})

describe('invitation bootstrap', () => {
	it.each(['/register', '/vikunja/register'])('removes the secret before returning control at %s', async path => {
		window.history.replaceState({position: 2}, '', `${path}?lang=en#invite-link=secret%2Btoken`)
		const invitation = await import('./inviteLink')
		expect(window.location.pathname + window.location.search + window.location.hash).toBe(`${path}?lang=en`)
		expect(window.history.state).toEqual({position: 2})
		expect(invitation.hasInviteLink()).toBe(true)
		await invitation.checkInviteLink()
		expect(api.check).toHaveBeenCalledWith({body: {token: 'secret+token'}})
	})

	it.each(['/register#other=value', '/share#link=secret', '/register'])('preserves unrelated navigation %s', async path => {
		window.history.replaceState({}, '', path)
		const invitation = await import('./inviteLink')
		expect(window.location.pathname + window.location.hash).toBe(path)
		expect(invitation.hasInviteLink()).toBe(false)
	})

	it('treats an empty invitation as invalid instead of falling back to public signup', async () => {
		window.history.replaceState({}, '', '/register#invite-link=')
		const invitation = await import('./inviteLink')
		expect(invitation.hasInviteLink()).toBe(true)
		await invitation.checkInviteLink()
		expect(api.check).toHaveBeenCalledWith({body: {token: ''}})
	})

	it('retains the secret on failure and forgets it after account creation', async () => {
		window.history.replaceState({}, '', '/register#invite-link=synthetic-secret')
		const invitation = await import('./inviteLink')
		const credentials = {username: 'guest', email: 'guest@example.com', password: '12345678'}
		api.register.mockRejectedValueOnce(new Error('Duplicate email'))
		await expect(invitation.registerViaInviteLink(credentials)).rejects.toThrow('Duplicate email')
		expect(invitation.hasInviteLink()).toBe(true)
		api.register.mockResolvedValueOnce({data: {id: 42}})
		await invitation.registerViaInviteLink(credentials)
		expect(api.register).toHaveBeenLastCalledWith({body: {...credentials, invite_token: 'synthetic-secret'}})
		expect(invitation.hasInviteLink()).toBe(false)
		await expect(invitation.registerViaInviteLink(credentials)).rejects.toThrow('No invitation')
	})

	it('consumes subsequent fragments before navigation observers and notifies without the token', async () => {
		window.history.replaceState({}, '', '/register')
		const invitation = await import('./inviteLink')
		const changed = vi.fn()
		const unsubscribe = invitation.onInviteLinkChange(changed)
		window.history.replaceState({}, '', '/register#invite-link=new-secret')
		const wrappedHistory = vi.spyOn(window.history, 'replaceState')
		const observed = vi.fn(() => expect(window.location.hash).toBe(''))
		window.addEventListener('popstate', observed)
		window.dispatchEvent(new PopStateEvent('popstate'))
		expect(wrappedHistory).not.toHaveBeenCalled()
		expect(observed).toHaveBeenCalledOnce()
		expect(changed).toHaveBeenCalledWith()
		await invitation.checkInviteLink()
		expect(api.check).toHaveBeenCalledWith({body: {token: 'new-secret'}})
		unsubscribe()
	})

	it('does not discard a new invitation when an earlier registration finishes', async () => {
		window.history.replaceState({}, '', '/register#invite-link=first')
		const invitation = await import('./inviteLink')
		let finish!: () => void
		api.register.mockReturnValue(new Promise<void>(resolve => { finish = resolve }))
		const registration = invitation.registerViaInviteLink({username: 'guest', email: 'guest@example.com', password: '12345678'})
		await vi.waitFor(() => expect(api.register).toHaveBeenCalledOnce())
		window.history.replaceState({}, '', '/register#invite-link=second')
		window.dispatchEvent(new PopStateEvent('popstate'))
		finish()
		await registration
		await invitation.checkInviteLink()
		expect(api.check).toHaveBeenCalledWith({body: {token: 'second'}})
	})

})
