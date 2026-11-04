import {afterEach, beforeEach, describe, it, expect} from 'vitest'

import {getAutoRedirectProvider, isLinkingIdentity, redirectToProvider} from './redirectToProvider'
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
			value: {...originalLocation, href: 'http://localhost/user/settings/linked-accounts'},
		})
	})

	afterEach(() => {
		Object.defineProperty(window, 'location', {configurable: true, value: originalLocation})
	})

	it('uses the login redirect uri when linking', () => {
		redirectToProvider(fullProvider, {linking: true})

		expect(isLinkingIdentity()).toBe(true)
		const url = new URL(window.location.href)
		expect(url.searchParams.get('prompt')).toBe('select_account')
		expect(url.searchParams.get('redirect_uri')).toBe('http://localhost/auth/openid/authentik')
		expect(url.searchParams.get('state')).toBe(localStorage.getItem('state'))
	})

	it('drops a stale linking marker on a login redirect', () => {
		redirectToProvider(fullProvider, {linking: true})
		redirectToProvider(fullProvider)

		expect(isLinkingIdentity()).toBe(false)
		expect(new URL(window.location.href).searchParams.has('prompt')).toBe(false)
	})
})
