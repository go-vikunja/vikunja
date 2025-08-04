import {UserFactory} from '../../factories/user'
import {ProjectFactory} from '../../factories/project'
import {TaskFactory} from '../../factories/task'
import {login} from '../../support/authenticateUser'
import {DATE_DISPLAY} from '../../../src/constants/dateDisplay'
import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'

dayjs.extend(relativeTime)

const createdDate = new Date(Date.UTC(2022, 6, 25, 12))
const now = new Date(Date.UTC(2022, 6, 30, 12))

const expectedFormats = {
	[DATE_DISPLAY.RELATIVE]: dayjs(createdDate).from(now),
	[DATE_DISPLAY.MM_DD_YYYY]: dayjs(createdDate).format('MM-DD-YYYY'),
	[DATE_DISPLAY.DD_MM_YYYY]: dayjs(createdDate).format('DD-MM-YYYY'),
	[DATE_DISPLAY.YYYY_MM_DD]: dayjs(createdDate).format('YYYY-MM-DD'),
	[DATE_DISPLAY.MM_SLASH_DD_YYYY]: dayjs(createdDate).format('MM/DD/YYYY'),
	[DATE_DISPLAY.DD_SLASH_MM_YYYY]: dayjs(createdDate).format('DD/MM/YYYY'),
	[DATE_DISPLAY.YYYY_SLASH_MM_DD]: dayjs(createdDate).format('YYYY/MM/DD'),
	[DATE_DISPLAY.DAY_MONTH_YEAR]: new Intl.DateTimeFormat('en', {
		day: 'numeric',
		month: 'long',
		year: 'numeric',
	}).format(createdDate),
	[DATE_DISPLAY.WEEKDAY_DAY_MONTH_YEAR]: new Intl.DateTimeFormat('en', {
		weekday: 'long',
		day: 'numeric',
		month: 'long',
		year: 'numeric',
	}).format(createdDate),
}

describe('Date display setting', () => {
	Object.entries(expectedFormats).forEach(([format, expected]) => {
		it(`shows ${format}`, () => {
			const user = UserFactory.create(1, {
				frontend_settings: JSON.stringify({dateDisplay: format}),
			})[0]
			const project = ProjectFactory.create(1, {owner_id: user.id})[0]
			TaskFactory.truncate()
			const task = TaskFactory.create(1, {
				id: 1,
				project_id: project.id,
				created_by_id: user.id,
				created: createdDate.toISOString(),
				updated: createdDate.toISOString(),
			})[0]

			cy.clock(now, ['Date'])
			login(user)
			cy.visit(`/tasks/${task.id}`)
			cy.get('.task-view .created time span').should('contain', expected)
		})
	})
})
