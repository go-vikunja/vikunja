import {afterEach, beforeEach, describe, it, expect} from 'vitest'

import {endLinkFlow, getAutoRedirectProvider, isLinkingIdentity, redirectToProvider} from './redirectToProvider'
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
	const originalLocation = window.location
	const fullProvider = {
		...provider,
		auth_url: 'https://idp.example.com/auth',
		client_id: 'vikunja',
		scope: 'openid',
		logout_url: '',
	} as IProvider

	beforeEach(() => {
		localStorage.clear()
		Object.defineProperty(window, 'location', {
			configurable: true,
			value: {
				...originalLocation,
				href: 'http://localhost/user/settings/linked-accounts',
			},
		})
	})

	afterEach(() => {
		Object.defineProperty(window, 'location', {
			configurable: true,
			value: originalLocation,
		})
	})

	it('uses the login redirect uri when linking', () => {
		redirectToProvider(fullProvider, {linking: true})

		expect(isLinkingIdentity(localStorage.getItem('state')!)).toBe(true)
		const url = new URL(window.location.href)
		expect(url.searchParams.get('prompt')).toBe('select_account')
		expect(url.searchParams.get('redirect_uri')).toBe('http://localhost/auth/openid/authentik')
		expect(url.searchParams.get('state')).toBe(localStorage.getItem('state'))
	})

	it('drops a stale linking marker on a login redirect', () => {
		redirectToProvider(fullProvider, {linking: true})
		const linkState = localStorage.getItem('state')!
		redirectToProvider(fullProvider)

		expect(isLinkingIdentity(linkState)).toBe(false)
		expect(new URL(window.location.href).searchParams.has('prompt')).toBe(false)
	})

	it('is not in link mode for a different state', () => {
		redirectToProvider(fullProvider, {linking: true})

		expect(isLinkingIdentity('other-state')).toBe(false)
		expect(isLinkingIdentity('')).toBe(false)
	})

	it('is not in link mode for a missing state without a marker', () => {
		expect(isLinkingIdentity(null)).toBe(false)
		expect(isLinkingIdentity(undefined)).toBe(false)
	})

	it('ignores a foreign, empty or missing state when ending a flow', () => {
		redirectToProvider(fullProvider, {linking: true})
		const linkState = localStorage.getItem('state')!

		endLinkFlow('other-state')
		endLinkFlow('')
		endLinkFlow(null)

		expect(isLinkingIdentity(linkState)).toBe(true)
		expect(localStorage.getItem('state')).toBe(linkState)
	})

	it('consumes the marker and state of its own flow', () => {
		redirectToProvider(fullProvider, {linking: true})
		const linkState = localStorage.getItem('state')!

		endLinkFlow(linkState)

		expect(isLinkingIdentity(linkState)).toBe(false)
		expect(localStorage.getItem('state')).toBeNull()
	})
})
