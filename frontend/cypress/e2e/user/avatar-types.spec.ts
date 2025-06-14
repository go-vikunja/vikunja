import {UserFactory, type UserAttributes} from '../../factories/user'
import {login} from '../../support/authenticateUser'

const avatarProviders = ['initials', 'gravatar', 'marble', 'upload', 'ldap'] as const

describe('User avatars', () => {
  avatarProviders.forEach(provider => {
    describe(`Avatar provider ${provider}`, () => {
      let user: UserAttributes

      beforeEach(() => {
        const overrides: Partial<UserAttributes & {avatar_provider?: string; avatar_file_id?: number; email?: string}> = {
          username: `user_${provider}`,
          avatar_provider: provider,
        }

        if (provider === 'gravatar') {
          overrides.email = `user_${provider}@example.com`
        }
        if (provider === 'upload' || provider === 'ldap') {
          overrides.avatar_file_id = 1
        }

        user = UserFactory.create(1, overrides)[0] as UserAttributes
        login(user)
      })

      it('Shows the avatar image', () => {
        cy.visit('/')
        cy.get('.username-dropdown-trigger img.avatar')
          .should('be.visible')
          .and(($img) => {
            expect($img[0].naturalWidth).to.be.greaterThan(0)
          })
      })
    })
  })
})
