import {createFakeUserAndLogin} from '../../support/authenticateUser'

describe('Duplicate Notifications', () => {
        createFakeUserAndLogin()

        it('Merges duplicate notifications and shows count', () => {
                cy.visit('/')

                cy.window().then(win => {
                        const app = win.document.getElementById('app') as any
                        // Access the vue app instance to trigger notifications
                        const vueApp = (app as any).__vue_app__
                        vueApp.config.globalProperties.$message.success({message: 'Duplicate Test'})
                        vueApp.config.globalProperties.$message.success({message: 'Duplicate Test'})
                })

                cy.get('.global-notification .vue-notification.success')
                        .should('have.length', 1)
                        .find('.notification-content')
                        .should('contain', 'Duplicate Test')
                        .find('span')
                        .should('contain', '×2')
        })
})
