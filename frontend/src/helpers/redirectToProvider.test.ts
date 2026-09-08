import {describe, it, expect, vi, beforeEach, afterEach} from 'vitest'

import {createCodeChallenge} from './pkce'
import {getAutoRedirectProvider, redirectToProvider} from './redirectToProvider'
import type {IProvider} from '@/types/IProvider'

const provider = {key: 'authentik', name: 'Authentik'} as IProvider

const soleProviderContext = {
	localAuthEnabled: false,
	ldapAuthEnabled: false,
	openIdEnabled: true,
	providers: [provider],
	isDesktopApp: false,
	justLoggedOut: false,
	hasCopyableRedirect: false,
}

describe('getAutoRedirectProvider', () => {
	it('returns the provider when it is the only way to log in', () => {
		expect(getAutoRedirectProvider(soleProviderContext)).toBe(provider)
	})

	it('does not redirect when the login url carries a copyable oauth destination', () => {
		expect(getAutoRedirectProvider({...soleProviderContext, hasCopyableRedirect: true})).toBeUndefined()
	})

	it('does not redirect inside the desktop app', () => {
		expect(getAutoRedirectProvider({...soleProviderContext, isDesktopApp: true})).toBeUndefined()
	})

	it('does not redirect right after an explicit logout', () => {
		expect(getAutoRedirectProvider({...soleProviderContext, justLoggedOut: true})).toBeUndefined()
	})

	it('does not redirect when local or ldap auth is available', () => {
		expect(getAutoRedirectProvider({...soleProviderContext, localAuthEnabled: true})).toBeUndefined()
		expect(getAutoRedirectProvider({...soleProviderContext, ldapAuthEnabled: true})).toBeUndefined()
	})

	it('does not redirect when there is a choice of providers', () => {
		expect(getAutoRedirectProvider({
			...soleProviderContext,
			providers: [provider, {key: 'other', name: 'Other'} as IProvider],
		})).toBeUndefined()
	})

	it('does not redirect when openid is disabled or has no providers', () => {
		expect(getAutoRedirectProvider({...soleProviderContext, openIdEnabled: false})).toBeUndefined()
		expect(getAutoRedirectProvider({...soleProviderContext, providers: []})).toBeUndefined()
	})
})

describe('redirectToProvider', () => {
	const oidcProvider = {
		key: 'authentik',
		name: 'Authentik',
		authUrl: 'https://id.example.com/application/o/authorize/',
		clientId: 'vikunja',
		scope: null,
		logoutUrl: '',
	} as unknown as IProvider

	beforeEach(() => {
		localStorage.clear()
		vi.stubGlobal('location', {
			href: 'https://vikunja.example.com/login',
			protocol: 'https:',
			host: 'vikunja.example.com',
		})
	})

	afterEach(() => {
		vi.unstubAllGlobals()
	})

	async function authorizeRequest(): Promise<URLSearchParams> {
		await redirectToProvider(oidcProvider)
		return new URL(window.location.href).searchParams
	}

	it('keeps sending the parameters it always did', async () => {
		const params = await authorizeRequest()

		expect(params.get('client_id')).toBe('vikunja')
		expect(params.get('response_type')).toBe('code')
		expect(params.get('scope')).toBe('openid email profile')
		expect(params.get('redirect_uri')).toBe('https://vikunja.example.com/auth/openid/authentik')
		expect(params.get('state')).toBe(localStorage.getItem('state'))
	})

	it('draws state from the csprng, not Math.random', async () => {
		const random = vi.spyOn(Math, 'random')

		await authorizeRequest()

		expect(random).not.toHaveBeenCalled()
	})

	it('sends a nonce and keeps it for the callback to check the id token against', async () => {
		const params = await authorizeRequest()

		expect(params.get('nonce')).toBeTruthy()
		expect(params.get('nonce')).toBe(localStorage.getItem('nonce'))
	})

	it('sends an S256 challenge for the verifier it stored', async () => {
		const params = await authorizeRequest()
		const verifier = localStorage.getItem('codeVerifier')

		expect(verifier).toBeTruthy()
		expect(params.get('code_challenge_method')).toBe('S256')
		expect(params.get('code_challenge')).toBe(await createCodeChallenge(verifier as string))
	})

	it('leaves pkce out of the request when the browser cannot hash the verifier', async () => {
		localStorage.setItem('codeVerifier', 'stale-from-an-earlier-flow')
		vi.stubGlobal('crypto', {getRandomValues: crypto.getRandomValues.bind(crypto)})

		const params = await authorizeRequest()

		expect(params.get('code_challenge')).toBeNull()
		expect(params.get('code_challenge_method')).toBeNull()
		expect(localStorage.getItem('codeVerifier')).toBeNull()
		expect(params.get('nonce')).toBeTruthy()
	})
})
