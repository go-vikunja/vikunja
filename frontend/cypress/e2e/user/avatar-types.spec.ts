import { UserFactory, type UserAttributes } from '../../factories/user'
import { login } from '../../support/authenticateUser'

const avatarProviders = [
	'initials',
	'gravatar',
	'marble',
	'upload',
	'ldap',
] as const

function createFakeUserWithAvatar(provider: string) {
	const overrides: Partial<UserAttributes & {
		avatar_provider?: string;
		avatar_file_id?: number;
		email?: string
	}> = {
		username: `user_${provider}`,
		avatar_provider: provider,
	}

	if (provider === 'gravatar') {
		overrides.email = `user_${provider}@example.com`
	}
	if (provider === 'upload' || provider === 'ldap') {
		overrides.avatar_file_id = 1
	}

	return UserFactory.create(1, overrides)[0] as UserAttributes
}

describe('User avatars', () => {
	avatarProviders.forEach(provider => {
		it(`Shows the avatar image for ${provider}`, () => {

			login(createFakeUserWithAvatar(provider))

			cy.visit('/')
			cy.get('.username-dropdown-trigger img.avatar')
				.should('be.visible')
				.and(($img) => {
					expect($img[0].naturalWidth).to.be.greaterThan(0)
				})
		})
	})
})
