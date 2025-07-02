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
                })
                cy.request({
                        method: 'PUT',
                        url: `${Cypress.env('API_URL')}/tasks/${parentB.id}/relations`,
                        headers: {
                                'Authorization': `Bearer ${window.localStorage.getItem('token')}`,
                        },
                        body: {
                                other_task_id: subtask.id,
                                relation_kind: 'subtask',
                        },
                })
        })

        it('shows subtask only once in each project list', () => {
                cy.visit(`/projects/${projectA.id}/1`)
                cy.get('.subtask-nested .task-link').contains(subtask.title).should('exist')
                cy.get('.tasks .task-link').contains(subtask.title).should('have.length', 1)

                cy.visit(`/projects/${projectB.id}/1`)
                cy.get('.subtask-nested .task-link').contains(subtask.title).should('exist')
                cy.get('.tasks .task-link').contains(subtask.title).should('have.length', 1)
        })
})
