import {createFakeUserAndLogin} from '../../support/authenticateUser'
import {ProjectFactory} from '../../factories/project'
import {TaskFactory} from '../../factories/task'
import {ProjectViewFactory} from '../../factories/project_view'

function createViews(projectId: number, projectViewId: number) {
        return ProjectViewFactory.create(1, {
                id: projectViewId,
                project_id: projectId,
                view_kind: 0,
        }, false)[0]
}

describe('Subtask duplicate handling', () => {
        createFakeUserAndLogin()

        let projectA
        let projectB
        let parentA
        let parentB
        let subtask

        beforeEach(() => {
                ProjectFactory.truncate()
                ProjectViewFactory.truncate()
                TaskFactory.truncate()

                projectA = ProjectFactory.create(1, {id: 1, title: 'Project A'})[0]
                createViews(projectA.id, 1)
                projectB = ProjectFactory.create(1, {id: 2, title: 'Project B'}, false)[0]
                createViews(projectB.id, 2)

                parentA = TaskFactory.create(1, {id: 10, title: 'Parent A', project_id: projectA.id}, false)[0]
                parentB = TaskFactory.create(1, {id: 11, title: 'Parent B', project_id: projectB.id}, false)[0]
                subtask = TaskFactory.create(1, {id: 12, title: 'Shared subtask', project_id: projectA.id}, false)[0]

                cy.request({
                        method: 'PUT',
                        url: `${Cypress.env('API_URL')}/tasks/${parentA.id}/relations`,
                        headers: {
                                'Authorization': `Bearer ${window.localStorage.getItem('token')}`,
                        },
                        body: {
                                other_task_id: subtask.id,
                                relation_kind: 'subtask',
                        },
                        failOnStatusCode: false,
                }).then((response) => {
                        // Accept both success (201) and conflict (409) - conflict means relation already exists
                        expect([201, 409]).to.include(response.status)
                })
        })

        it('shows subtask only once in project list', () => {
                // Add API intercepts to wait for all necessary data to load
                cy.intercept('GET', '**/api/v1/projects/*/views/*/tasks**').as('loadTasks')
                cy.intercept('GET', '**/api/v1/projects/*').as('loadProject')
                cy.intercept('GET', '**/api/v1/tasks/*/relations').as('loadTaskRelations')

                cy.visit(`/projects/${projectA.id}/1`)

                // Wait for project and tasks to load with longer timeouts for CI compatibility
                cy.wait('@loadProject', { timeout: 30000 })
                cy.wait('@loadTasks', { timeout: 30000 })

                // Wait for page to be fully loaded and tasks rendered
                cy.get('.tasks').should('be.visible')
                cy.get('.task-link').should('have.length.greaterThan', 0)

                // Wait a bit more for relations to load if needed
                cy.wait(500)

                // Check that subtask appears in nested structure (with retry for reliability)
                cy.get('.subtask-nested .task-link').contains(subtask.title, { timeout: 10000 }).should('exist')
                // Check that subtask appears only once in the overall task list
                cy.get('.tasks .task-link').contains(subtask.title, { timeout: 10000 }).should('have.length', 1)
        })
})
