import {test, expect} from '../../support/fixtures'
import {UserFactory} from '../../factories/user'
import {ProjectFactory} from '../../factories/project'
import {TaskFactory} from '../../factories/task'
import {login} from '../../support/authenticateUser'
import {DATE_DISPLAY} from '../../../src/constants/dateDisplay'
import {TIME_FORMAT} from '../../../src/constants/timeFormat'
import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime.js'

dayjs.extend(relativeTime)

const createdDate = new Date(Date.UTC(2022, 6, 25, 12))
const now = new Date(Date.UTC(2022, 6, 30, 12))

const expectedFormats12h = {
	[DATE_DISPLAY.RELATIVE]: dayjs(createdDate).from(now),
	[DATE_DISPLAY.MM_DD_YYYY]: dayjs(createdDate).format('MM-DD-YYYY hh:mm A'),
	[DATE_DISPLAY.DD_MM_YYYY]: dayjs(createdDate).format('DD-MM-YYYY hh:mm A'),
	[DATE_DISPLAY.YYYY_MM_DD]: dayjs(createdDate).format('YYYY-MM-DD hh:mm A'),
	[DATE_DISPLAY.MM_SLASH_DD_YYYY]: dayjs(createdDate).format('MM/DD/YYYY hh:mm A'),
	[DATE_DISPLAY.DD_SLASH_MM_YYYY]: dayjs(createdDate).format('DD/MM/YYYY hh:mm A'),
	[DATE_DISPLAY.YYYY_SLASH_MM_DD]: dayjs(createdDate).format('YYYY/MM/DD hh:mm A'),
	[DATE_DISPLAY.DAY_MONTH_YEAR]: new Intl.DateTimeFormat('en', {
		day: 'numeric',
		month: 'long',
		year: 'numeric',
		hour: 'numeric',
		minute: 'numeric',
		hour12: true,
	}).format(createdDate),
	[DATE_DISPLAY.WEEKDAY_DAY_MONTH_YEAR]: new Intl.DateTimeFormat('en', {
		weekday: 'long',
		day: 'numeric',
		month: 'long',
		year: 'numeric',
		hour: 'numeric',
		minute: 'numeric',
		hour12: true,
	}).format(createdDate),
}

const expectedFormats24h = {
	[DATE_DISPLAY.RELATIVE]: dayjs(createdDate).from(now),
	[DATE_DISPLAY.MM_DD_YYYY]: dayjs(createdDate).format('MM-DD-YYYY HH:mm'),
	[DATE_DISPLAY.DD_MM_YYYY]: dayjs(createdDate).format('DD-MM-YYYY HH:mm'),
	[DATE_DISPLAY.YYYY_MM_DD]: dayjs(createdDate).format('YYYY-MM-DD HH:mm'),
	[DATE_DISPLAY.MM_SLASH_DD_YYYY]: dayjs(createdDate).format('MM/DD/YYYY HH:mm'),
	[DATE_DISPLAY.DD_SLASH_MM_YYYY]: dayjs(createdDate).format('DD/MM/YYYY HH:mm'),
	[DATE_DISPLAY.YYYY_SLASH_MM_DD]: dayjs(createdDate).format('YYYY/MM/DD HH:mm'),
	[DATE_DISPLAY.DAY_MONTH_YEAR]: new Intl.DateTimeFormat('en', {
		day: 'numeric',
		month: 'long',
		year: 'numeric',
		hour: 'numeric',
		minute: 'numeric',
		hour12: false,
	}).format(createdDate),
	[DATE_DISPLAY.WEEKDAY_DAY_MONTH_YEAR]: new Intl.DateTimeFormat('en', {
		weekday: 'long',
		day: 'numeric',
		month: 'long',
		year: 'numeric',
		hour: 'numeric',
		minute: 'numeric',
		hour12: false,
	}).format(createdDate),
}

test.describe('Date display setting', () => {
	Object.entries(expectedFormats12h).forEach(([format, expected]) => {
		test(`shows ${format} with 12h time format`, async ({page, apiContext}) => {
			const user = (await UserFactory.create(1, {
				frontend_settings: JSON.stringify({dateDisplay: format, timeFormat: TIME_FORMAT.HOURS_12}),
			}))[0]
			const project = (await ProjectFactory.create(1, {owner_id: user.id}))[0]
			TaskFactory.truncate()
			const task = (await TaskFactory.create(1, {
				id: 1,
				project_id: project.id,
				created_by_id: user.id,
				created: createdDate.toISOString(),
				updated: createdDate.toISOString(),
			}))[0]

			await page.clock.install({time: now})
			await login(page, apiContext, user)
			await page.goto(`/tasks/${task.id}`)
			await expect(page.locator('.task-view .created time span')).toContainText(expected)
		})
	})

	Object.entries(expectedFormats24h).forEach(([format, expected]) => {
		test(`shows ${format} with 24h time format`, async ({page, apiContext}) => {
			const user = (await UserFactory.create(1, {
				frontend_settings: JSON.stringify({dateDisplay: format, timeFormat: TIME_FORMAT.HOURS_24}),
			}))[0]
			const project = (await ProjectFactory.create(1, {owner_id: user.id}))[0]
			TaskFactory.truncate()
			const task = (await TaskFactory.create(1, {
				id: 1,
				project_id: project.id,
				created_by_id: user.id,
				created: createdDate.toISOString(),
				updated: createdDate.toISOString(),
			}))[0]

			await page.clock.install({time: now})
			await login(page, apiContext, user)
			await page.goto(`/tasks/${task.id}`)
			await expect(page.locator('.task-view .created time span')).toContainText(expected)
		})
	})
})
