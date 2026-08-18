import {describe, it, expect} from 'vitest'

import {getAutoRedirectProvider} from './redirectToProvider'
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
