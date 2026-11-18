import {test as nodeTest, type Page} from '@playwright/test'
import {readFile, readdir} from 'node:fs/promises'
import {test, expect} from '../../support/fixtures'
import {setupApiUrl} from '../../support/authenticateUser'
import {ProjectFactory} from '../../factories/project'
import {TaskFactory} from '../../factories/task'
import {TaskCommentFactory} from '../../factories/task_comment'
import {BucketFactory} from '../../factories/bucket'
import {TaskBucketFactory} from '../../factories/task_buckets'
import {LinkShareFactory} from '../../factories/link_sharing'
import {createDefaultViews} from '../project/prepareProjects'

const DESKTOP = {
	width: 1280,
	height: 720,
}
const PHONE = {
	width: 390,
	height: 844,
}
const NARROW = {
	width: 1080,
	height: 720,
}
const TABLET = {
	width: 900,
	height: 720,
}
const SHORT = {
	width: 1280,
	height: 360,
}
const CENTRED = '.default .modal-content, .hint-modal .modal-content, .top .modal-content'

type State = 'task' | 'sidebar' | 'skip' | 'about' | 'shortcuts' | 'search' | 'kanban'
	| 'scrolling' | 'confirmation' | 'fullscreen' | 'shadowedClose' | 'notifications'
	| 'login' | 'share' | 'dropdown' | 'popup'

interface GuardRow {
	id: string
	file: string
	selector: string
	property: string
	viewport: {
		width: number
		height: number
	}
	state: State
	pre43: number
	dockDelta: number
	// CSSOM selector, where the measured element needs a more specific locator.
	ruleSelector?: string
	declaration?: string
	// Preserve repeated selector/property rules at different breakpoints in G1.
	occurrence?: number
	layout?: 'top' | 'height' | 'scroll' | 'logo' | 'dropdown' | 'popup'
	js?: boolean
}

// Files are relative to frontend/src. Each source declaration has a row, even when
// a later media rule repeats the same selector/property. Repeated viewport rows
// refer to the same occurrence; grouped selectors stay grouped for the CSSOM key.
const CSS_ROWS: GuardRow[] = [
	{
		id: 'header',
		file: 'components/home/AppHeader.vue',
		selector: '.navbar',
		property: 'inset-block-start',
		viewport: DESKTOP,
		state: 'task',
		pre43: 0, // 3f9835b63 AppHeader.vue: inset-block-start: 0
		dockDelta: 40,
		layout: 'top',
	},
	{
		id: 'header-phone',
		file: 'components/home/AppHeader.vue',
		selector: '.navbar',
		property: 'inset-block-start',
		viewport: PHONE,
		state: 'task',
		pre43: 0, // AppHeader.vue: the fixed header uses the same top at both widths.
		dockDelta: 40,
		layout: 'top',
	},
	{
		id: 'sidebar-desktop',
		file: 'components/home/Navigation.vue',
		selector: '.menu-container',
		property: 'inset-block-start',
		viewport: DESKTOP,
		state: 'task',
		pre43: 64, // 3f9835b63 Navigation.vue: inset-block-start: $navbar-height (4rem)
		dockDelta: 40,
		layout: 'top',
	},
	{
		id: 'sidebar-phone',
		file: 'components/home/Navigation.vue',
		selector: '.menu-container',
		property: 'inset-block-start',
		viewport: PHONE,
		state: 'sidebar',
		pre43: 0, // 3f9835b63 Navigation.vue, max-width $tablet: inset-block-start: 0
		dockDelta: 40,
		occurrence: 2,
		layout: 'top',
	},
	{
		id: 'sidebar-close-phone',
		file: 'components/home/ContentAuth.vue',
		selector: '.menu-hide-button',
		property: 'inset-block-start',
		viewport: PHONE,
		state: 'sidebar',
		pre43: 8, // 3f9835b63 ContentAuth.vue: inset-block-start: 0.5rem
		dockDelta: 40,
	},
	{
		id: 'container-min-height',
		file: 'components/home/ContentAuth.vue',
		selector: '.app-container',
		property: 'min-block-size',
		viewport: DESKTOP,
		state: 'task',
		pre43: 655, // 3f9835b63 ContentAuth.vue: calc(100vh - 65px)
		dockDelta: -40,
	},
	{
		id: 'container-min-height-phone',
		file: 'components/home/ContentAuth.vue',
		selector: '.app-container',
		property: 'min-block-size',
		viewport: PHONE,
		state: 'task',
		pre43: 779, // 3f9835b63 ContentAuth.vue: calc(100vh - 65px)
		dockDelta: -40,
	},
	{
		id: 'container-padding-phone',
		file: 'components/home/ContentAuth.vue',
		selector: '.app-container',
		property: 'padding-block-start',
		viewport: PHONE,
		state: 'task',
		pre43: 64, // 3f9835b63 ContentAuth.vue, max-width $tablet: padding-block-start: $navbar-height
		dockDelta: 40,
	},
	{
		id: 'content-min-height-phone',
		file: 'components/home/ContentAuth.vue',
		selector: '.app-content',
		property: 'min-block-size',
		viewport: PHONE,
		state: 'task',
		pre43: 780, // 3f9835b63 ContentAuth.vue: calc(100vh - 4rem)
		dockDelta: -40,
	},
	{
		id: 'content-padding-desktop',
		file: 'components/home/ContentAuth.vue',
		selector: '.app-content',
		property: 'padding-block-start',
		declaration: 'padding',
		viewport: DESKTOP,
		state: 'task',
		pre43: 88, // 3f9835b63 ContentAuth.vue: padding: $navbar-height + 1.5rem 1.5rem 0 1.5rem
		dockDelta: 40,
	},
	{
		id: 'loading-wrapper-min-height',
		file: 'components/home/ContentAuth.vue',
		selector: '.app-content > .loader-container',
		property: 'min-block-size',
		viewport: DESKTOP,
		state: 'task',
		pre43: 616, // 3f9835b63 ContentAuth.vue: calc(100vh - #{$navbar-height + 1.5rem + 1rem})
		dockDelta: -40,
	},
	{
		id: 'loading-wrapper-min-height-phone',
		file: 'components/home/ContentAuth.vue',
		selector: '.app-content > .loader-container',
		property: 'min-block-size',
		viewport: PHONE,
		state: 'task',
		pre43: 740, // 3f9835b63 ContentAuth.vue: calc(100vh - #{$navbar-height + 1.5rem + 1rem})
		dockDelta: -40,
	},
	{
		id: 'skip-link-desktop',
		file: 'styles/global.scss',
		selector: '.skip-to-content:focus',
		property: 'inset-block-start',
		viewport: DESKTOP,
		state: 'skip',
		pre43: 0, // 3f9835b63 global.scss: .skip-to-content:focus { inset-block-start: 0 }
		dockDelta: 40,
		layout: 'top',
	},
	{
		id: 'skip-link-phone',
		file: 'styles/global.scss',
		selector: '.skip-to-content:focus',
		property: 'inset-block-start',
		viewport: PHONE,
		state: 'skip',
		pre43: 0, // 3f9835b63 global.scss: .skip-to-content:focus { inset-block-start: 0 }
		dockDelta: 40,
		layout: 'top',
	},
	{
		id: 'modal-container-desktop',
		file: 'components/misc/Modal.vue',
		selector: 'dialog[open] .modal-container',
		ruleSelector: '.modal-container',
		property: 'padding-block-start',
		viewport: DESKTOP,
		state: 'scrolling',
		pre43: 0, // 3f9835b63 Modal.vue: padding-block-start: env(safe-area-inset-top)
		dockDelta: 40,
	},
	{
		id: 'modal-centred-top',
		file: 'components/misc/Modal.vue',
		selector: 'dialog[open].hint-modal .modal-content',
		ruleSelector: CENTRED,
		property: 'inset-block-start',
		viewport: DESKTOP,
		state: 'about',
		pre43: 360, // 3f9835b63 Modal.vue: inset-block-start: 50% (720px container)
		dockDelta: 20,
	},
	{
		id: 'modal-centred-max-height',
		file: 'components/misc/Modal.vue',
		selector: 'dialog[open].hint-modal .modal-content',
		ruleSelector: CENTRED,
		property: 'max-block-size',
		viewport: DESKTOP,
		state: 'about',
		pre43: 688, // 3f9835b63 Modal.vue: calc(100dvh - 2rem)
		dockDelta: -40,
	},
	{
		id: 'modal-height-bound-top',
		file: 'components/misc/Modal.vue',
		selector: 'dialog[open].default .modal-content',
		ruleSelector: CENTRED,
		property: 'inset-block-start',
		viewport: DESKTOP,
		state: 'shortcuts',
		pre43: 360, // 3f9835b63 Modal.vue: inset-block-start: 50% (720px container)
		dockDelta: 20,
	},
	{
		id: 'modal-height-bound-max-height',
		file: 'components/misc/Modal.vue',
		selector: 'dialog[open].default .modal-content',
		ruleSelector: CENTRED,
		property: 'max-block-size',
		viewport: DESKTOP,
		state: 'shortcuts',
		pre43: 688, // 3f9835b63 Modal.vue: calc(100dvh - 2rem)
		dockDelta: -40,
		layout: 'height',
	},
	{
		id: 'modal-search-top',
		file: 'components/misc/Modal.vue',
		selector: 'dialog[open].top .modal-content',
		ruleSelector: '.top .modal-content',
		property: 'inset-block-start',
		viewport: DESKTOP,
		state: 'search',
		pre43: 48, // 3f9835b63 Modal.vue: .top .modal-content { inset-block-start: 3rem }
		dockDelta: 40,
		layout: 'top',
	},
	{
		id: 'modal-search-max-height',
		file: 'components/misc/Modal.vue',
		selector: 'dialog[open].top .modal-content',
		ruleSelector: '.top .modal-content',
		property: 'max-block-size',
		viewport: DESKTOP,
		state: 'search',
		pre43: 624, // 3f9835b63 Modal.vue: .top .modal-content { max-block-size: calc(100dvh - 6rem) }
		dockDelta: -40,
	},
	{
		id: 'modal-close-wide',
		file: 'components/misc/Modal.vue',
		selector: 'dialog[open] .modal-container > .close',
		ruleSelector: '.close',
		property: 'inset-block-start',
		viewport: DESKTOP,
		state: 'shortcuts',
		pre43: 8, // 3f9835b63 Modal.vue: .close { inset-block-start: .5rem }
		dockDelta: 40,
	},
	{
		id: 'modal-close-shadowed',
		file: 'components/misc/Modal.vue',
		selector: 'dialog[open] .modal-container > .close',
		ruleSelector: '.close',
		property: 'inset-block-start',
		viewport: NARROW,
		state: 'shadowedClose',
		pre43: 69, // 3f9835b63 Modal.vue: calc(5px + $modal-margin), $modal-margin = 4rem
		dockDelta: 40,
		occurrence: 2,
	},
	{
		id: 'modal-close-fullscreen',
		file: 'components/misc/Modal.vue',
		selector: 'dialog[open].fullscreen .modal-container > .close',
		ruleSelector: '.fullscreen .close',
		property: 'inset-block-start',
		viewport: NARROW,
		state: 'fullscreen',
		pre43: 8, // 3f9835b63 Modal.vue: .fullscreen & { inset-block-start: .5rem } in close media rule
		dockDelta: 40,
	},
	{
		id: 'modal-close-tablet',
		file: 'components/misc/Modal.vue',
		selector: 'dialog[open] .modal-container > .close',
		ruleSelector: '.close',
		property: 'inset-block-start',
		viewport: TABLET,
		state: 'shortcuts',
		pre43: 12, // 3f9835b63 Modal.vue: .close, $tablet..($desktop + 84px): inset-block-start: .75rem
		dockDelta: 40,
		occurrence: 3,
	},
	{
		id: 'modal-container-phone',
		file: 'components/misc/Modal.vue',
		selector: 'dialog[open] .modal-container',
		ruleSelector: '.modal-container',
		property: 'padding-block-start',
		viewport: PHONE,
		state: 'shortcuts',
		pre43: 0, // 3f9835b63 Modal.vue, max-width $tablet: padding-block-start: env(safe-area-inset-top)
		dockDelta: 40,
		occurrence: 2,
	},
	{
		id: 'modal-card-phone',
		file: 'components/misc/Modal.vue',
		selector: 'dialog[open] .keyboard-shortcuts.card',
		ruleSelector: '.card',
		property: 'min-block-size',
		viewport: PHONE,
		state: 'shortcuts',
		pre43: 844, // 3f9835b63 Modal.vue: calc(100dvh - env(safe-area-inset-top) - env(safe-area-inset-bottom))
		dockDelta: -40,
	},
	{
		id: 'modal-header-min-height',
		file: 'components/misc/Modal.vue',
		selector: 'dialog[open] .modal-content:has(.modal-header)',
		ruleSelector: '.modal-content:has(.modal-header)',
		property: 'min-block-size',
		viewport: PHONE,
		state: 'confirmation',
		pre43: 844, // 3f9835b63 Modal.vue: calc(100dvh - env(safe-area-inset-top) - env(safe-area-inset-bottom))
		dockDelta: -40,
	},
	{
		id: 'modal-fullscreen-phone',
		file: 'components/misc/Modal.vue',
		selector: 'dialog[open].fullscreen .modal-container',
		ruleSelector: '.modal-container',
		property: 'padding-block-start',
		viewport: PHONE,
		state: 'fullscreen',
		pre43: 0, // 3f9835b63 Modal.vue, max-width $tablet: padding-block-start: env(safe-area-inset-top)
		dockDelta: 40,
		occurrence: 2,
	},
	{
		id: 'modal-fullscreen-close-phone',
		file: 'components/misc/Modal.vue',
		selector: 'dialog[open].fullscreen .modal-container > .close',
		ruleSelector: '.close',
		property: 'inset-block-start',
		viewport: PHONE,
		state: 'fullscreen',
		pre43: 8, // 3f9835b63 Modal.vue: .close { inset-block-start: .5rem }; mobile fullscreen keeps it visible
		dockDelta: 40,
	},
	{
		id: 'kanban-desktop',
		file: 'components/project/views/ProjectKanban.vue',
		selector: '.kanban',
		property: 'block-size',
		viewport: DESKTOP,
		state: 'kanban',
		pre43: 573, // 3f9835b63 ProjectKanban.vue: 100vh - 4.5rem - 1.5rem - 1rem - 1.5rem - 11px
		dockDelta: -40,
		layout: 'height',
	},
	{
		id: 'kanban-phone',
		file: 'components/project/views/ProjectKanban.vue',
		selector: '.kanban',
		property: 'block-size',
		viewport: PHONE,
		state: 'kanban',
		// 3f9835b63 ProjectKanban.vue: calc(#{$crazy-height-calculation} - #{$filter-container-height} + 9px)
		// Interpolation expands to 100vh - 8.5rem - 11px - 1rem - 2.69rem + 9px (no parentheses).
		pre43: 646.96,
		dockDelta: -40,
		occurrence: 2,
		layout: 'height',
	},
	{
		id: 'notifications-phone',
		file: 'components/notifications/Notifications.vue',
		selector: '.notifications .notifications-list',
		property: 'max-block-size',
		viewport: PHONE,
		state: 'notifications',
		pre43: 764, // 3f9835b63 Notifications.vue: calc(100vh - 1rem - #{$navbar-height})
		dockDelta: -40,
	},
	{
		id: 'comments-container',
		file: 'components/tasks/partials/Comments.vue',
		selector: '.comments-container',
		property: 'scroll-margin-block-start',
		viewport: DESKTOP,
		state: 'task',
		pre43: 64, // 3f9835b63 Comments.vue: .comments-container { scroll-margin-block-start: 4rem }
		dockDelta: 40,
		layout: 'scroll',
	},
	{
		id: 'comment-item',
		file: 'components/tasks/partials/Comments.vue',
		selector: '#comment-1',
		ruleSelector: '.media.comment',
		property: 'scroll-margin-block-start',
		viewport: DESKTOP,
		state: 'task',
		pre43: 64, // 3f9835b63 Comments.vue: .media.comment { scroll-margin-block-start: 4rem }
		dockDelta: 40,
		layout: 'scroll',
	},
	{
		id: 'comments-container-phone',
		file: 'components/tasks/partials/Comments.vue',
		selector: '.comments-container',
		property: 'scroll-margin-block-start',
		viewport: PHONE,
		state: 'task',
		pre43: 64, // 3f9835b63 Comments.vue: .comments-container { scroll-margin-block-start: 4rem }
		dockDelta: 40,
		layout: 'scroll',
	},
	{
		id: 'comment-item-phone',
		file: 'components/tasks/partials/Comments.vue',
		selector: '#comment-1',
		ruleSelector: '.media.comment',
		property: 'scroll-margin-block-start',
		viewport: PHONE,
		state: 'task',
		pre43: 64, // 3f9835b63 Comments.vue: .media.comment { scroll-margin-block-start: 4rem }
		dockDelta: 40,
		layout: 'scroll',
	},
	{
		id: 'login-shell',
		file: 'components/misc/NoAuthWrapper.vue',
		selector: '.no-auth-wrapper',
		property: 'padding-block-start',
		viewport: DESKTOP,
		state: 'login',
		pre43: 0, // 3f9835b63 NoAuthWrapper.vue: .no-auth-wrapper has no top padding (initial 0)
		dockDelta: 40,
		layout: 'logo',
	},
	{
		id: 'share-shell',
		file: 'components/home/ContentLinkShare.vue',
		selector: '.link-share-container',
		property: 'padding-block-start',
		viewport: DESKTOP,
		state: 'share',
		pre43: 0, // 3f9835b63 ContentLinkShare.vue: .link-share-container has no top padding (initial 0)
		dockDelta: 40,
		layout: 'logo',
	},
	{
		id: 'login-shell-phone',
		file: 'components/misc/NoAuthWrapper.vue',
		selector: '.no-auth-wrapper',
		property: 'padding-block-start',
		viewport: PHONE,
		state: 'login',
		pre43: 0, // 3f9835b63 NoAuthWrapper.vue: no top padding at either width.
		dockDelta: 40,
		layout: 'logo',
	},
	{
		id: 'share-shell-phone',
		file: 'components/home/ContentLinkShare.vue',
		selector: '.link-share-container',
		property: 'padding-block-start',
		viewport: PHONE,
		state: 'share',
		pre43: 0, // 3f9835b63 ContentLinkShare.vue: no top padding at either width.
		dockDelta: 40,
		layout: 'logo',
	},
]

// Desktop sidebar/content padding is replaced on phones. Mobile modals override
// centering/max-height and hide the close button except in fullscreen (guarded
// above). Their container padding also has its own phone occurrence above.
// These two source consumers call appTopInset() in JS and cannot be found by G1.
// Their pinned anchors and saturated menus make the consumer's own floor bind.
const JS_CONSUMERS: GuardRow[] = [
	// Dropdown.vue's fixedNavbarBottom floor is width-independent and binds at 1280x360.
	// At 390x360, the kanban bucket footer and loading overlay intercept fixture clicks.
	{
		id: 'dropdown-floor',
		file: 'components/misc/Dropdown.vue',
		selector: '.task-actions-menu .dropdown-menu',
		property: 'top',
		viewport: SHORT,
		state: 'dropdown',
		pre43: 8, // 3f9835b63 Dropdown.vue: fixedNavbarBottom() returns 0 in a dialog; padding.top = 0 + 8
		dockDelta: 40,
		layout: 'dropdown',
		js: true,
	},
	{
		id: 'popup-floor',
		file: 'components/misc/Popup.vue',
		selector: '.popup:has(.datepicker-popup)',
		property: 'top',
		viewport: SHORT,
		state: 'popup',
		pre43: 72, // Upstream Popup.vue: VIEWPORT_PADDING.top = 72, used by flip/shift/size
		dockDelta: 40,
		layout: 'popup',
		js: true,
	},
]
// The date picker becomes a bottom sheet on phones; Popup's JS floor is inactive.

async function behaveLikeDockApp(page: Page) {
	await page.addInitScript(() => {
		Object.defineProperty(navigator, 'standalone', {get: () => true})
	})
}

async function gotoApp(page: Page, url: string, dock: boolean) {
	await page.goto(url)
	// App.vue applies the inset before mounting; wait for it before checking either mode.
	await expect(page.locator('#app')).not.toBeEmpty()
	expect(
		await page.evaluate(() => document.documentElement.style.getPropertyValue('--app-top-inset')),
		`${dock ? 'Dock' : 'normal'} document must initialize its own app inset`,
	).toBe(dock ? '40px' : '')
}

async function computed(page: Page, selector: string, property: string) {
	return page.locator(selector).first().evaluate((element, property) => {
		const value = getComputedStyle(element).getPropertyValue(property)
		if (!/^-?[\d.]+px$/.test(value)) {
			throw new Error(`${property} must resolve to px, got ${value}`)
		}
		return parseFloat(value)
	}, property)
}

function rect(page: Page, selector: string) {
	return page.locator(selector).first().evaluate(element => {
		const {top, bottom, height} = element.getBoundingClientRect()
		return {
			top,
			bottom,
			height,
		}
	})
}

async function settled(page: Page, selector: string) {
	await expect(page.locator(selector).first()).toBeVisible()
	await page.evaluate(() => document.fonts.ready)
	await page.waitForFunction(() => document.getAnimations().every(animation => animation.playState !== 'running'))
}

async function prepareFixture(ownerId: number) {
	await ProjectFactory.create(1, {
		title: 'Guard',
		owner_id: ownerId,
	})
	await createDefaultViews(1)
	await TaskFactory.create(1, {
		id: 1,
		title: 'Guard task',
		description: '<p>Deterministic inset guard description.</p>',
		created_by_id: ownerId,
	})
	await BucketFactory.create(1, {
		title: 'Guard bucket',
		project_view_id: 4,
	})
	await TaskBucketFactory.create(1, {
		task_id: 1,
		bucket_id: 1,
		project_view_id: 4,
	})
	// Twelve tall comments leave ample trailing content after both scroll targets.
	await TaskCommentFactory.create(12, {
		author_id: ownerId,
		comment: (id: number) => `<p>Guard comment ${id}</p>${'<p>Trailing guard content.</p>'.repeat(6)}`,
		created: '2026-01-01T12:00:00Z',
		updated: '2026-01-01T12:00:00Z',
	})
	await LinkShareFactory.create(1, {
		hash: 'inset-guard-share',
		shared_by_id: ownerId,
	})
}

async function openTask(page: Page, dock = false) {
	await gotoApp(page, '/tasks/1#comment-1', dock)
	await expect(page.locator('.task-view h1.title')).toContainText('Guard task')
	await expect(page.locator('#comment-1')).toBeVisible()
	await page.evaluate(() => window.scrollTo(0, 0))
}

async function openKanban(page: Page, dock = false) {
	await gotoApp(page, '/projects/1/4', dock)
	await expect(page.locator('.kanban .bucket .task')).toContainText('Guard task')
}

async function openTaskDialog(page: Page, dock: boolean) {
	await openKanban(page, dock)
	await page.locator('.bucket .task').first().click()
	await settled(page, 'dialog[open].scrolling .task-view')
}

async function openShortcuts(page: Page, dock: boolean) {
	await openTask(page, dock)
	// ContentAuth.vue hides the floating button on phones; AppHeader's user menu keeps the same action.
	await page.locator('.navbar .username-dropdown-trigger').click()
	await page.locator('.navbar-end').getByRole('button', {
		name: 'Keyboard Shortcuts',
		exact: true,
	}).click()
	await settled(page, 'dialog[open] .keyboard-shortcuts')
}

async function openState(page: Page, state: State, dock = false) {
	if (state === 'login') {
		await gotoApp(page, '/login', dock)
		await expect(page).toHaveURL('/login')
		await expect(page.locator('#loginform #username')).toBeVisible()
		return
	}
	if (state === 'share') {
		await gotoApp(page, '/share/inset-guard-share/auth?view=1', dock)
		await expect(page).toHaveURL('/projects/1/1#share-auth-token=inset-guard-share')
		await expect(page.locator('.link-share-container h1.title')).toHaveText('Guard')
		await expect(page.locator('.tasks')).toContainText('Guard task')
		return
	}
	if (state === 'kanban') {
		await openKanban(page, dock)
		return
	}
	if (state === 'scrolling' || state === 'dropdown') {
		await openTaskDialog(page, dock)
		if (state === 'dropdown') {
			// Fix the real Dropdown reference at y=200 in BOTH modes. Its long menu
			// must open upward and exhaust the space above it (floor 8 / 48).
			// No modal/header/ancestor movement can supply this 40px difference.
			await page.locator('.task-actions-menu').evaluate((element: HTMLElement) => {
				Object.assign(element.style, {
					position: 'fixed',
					top: '200px',
					left: '700px',
					width: '40px',
					height: '32px',
				})
			})
			await page.locator('.task-actions-menu .toolbar-icon-button').click()
		}
		return
	}
	if (state === 'shortcuts' || state === 'shadowedClose') {
		await openShortcuts(page, dock)
		if (state === 'shadowedClose') {
			// At 1024..1108px the later .75rem rule ALWAYS shadows the 5px + 4rem
			// rule. Isolate it in this document only, without supplying an inset or
			// changing the declaration under test. A reload restores the stylesheet.
			const removed = await page.evaluate(() => {
				const close = document.querySelector('dialog[open] .modal-container > .close')!
				const candidates: CSSStyleRule[] = []
				function visit(rules: CSSRuleList) {
					for (const rule of Array.from(rules)) {
						if (rule instanceof CSSStyleRule
							&& rule.selectorText.replace(/\[data-v-[\w-]+\]/g, '').trim() === '.close'
							// Heading.vue also has a scoped .close rule; only Modal.vue's rules match this button.
							&& close.matches(rule.selectorText)
							&& rule.style.getPropertyValue('inset-block-start')) {
							candidates.push(rule)
						}
						if ('cssRules' in rule) visit((rule as CSSGroupingRule).cssRules)
					}
				}
				for (const sheet of Array.from(document.styleSheets)) visit(sheet.cssRules)
				if (candidates.length !== 3) {
					throw new Error(`Expected three Modal.vue close declarations, got ${candidates.length}`)
				}
				return Boolean(candidates[2].style.removeProperty('inset-block-start'))
			})
			expect(removed, 'isolate only the later tablet close declaration').toBe(true)
		}
		return
	}
	if (state === 'about' || state === 'confirmation') {
		await gotoApp(page, state === 'about' ? '/about' : '/projects/1/settings/delete', dock)
		await settled(page, state === 'about' ? 'dialog[open] .card' : 'dialog[open] .modal-header')
		return
	}
	await openTask(page, dock)
	if (state === 'sidebar') {
		const toggle = page.locator('.navbar .menu-show-button')
		if (await toggle.getAttribute('aria-expanded') !== 'true') await toggle.click()
		await settled(page, '.menu-container.is-active')
	} else if (state === 'skip') {
		await page.locator('.skip-to-content').focus()
	} else if (state === 'notifications') {
		await page.locator('.notifications .trigger-button').click()
	} else if (state === 'search') {
		await page.getByRole('button', {
			name: 'Open the search/quick action bar',
			exact: true,
		}).click()
		await expect(page.locator('.quick-actions .action-input input')).toBeFocused()
	} else if (state === 'fullscreen') {
		await page.locator('.attachments .attachment-open').first().click()
		await settled(page, 'dialog[open].fullscreen .image-lightbox__image.is-loaded')
	} else if (state === 'popup') {
		// Saturate the viewport independently of compact-picker geometry or side preference.
		// Upstream crossAxis shift slides over the anchor; only the Dock floor changes.
		await page.addStyleTag({content: '.datepicker-popup { min-block-size: 600px; }'})
		await page.locator('.property-due-date .datepicker button.show').evaluate((element: HTMLElement) => {
			Object.assign(element.style, {
				position: 'fixed',
				top: '260px',
				left: '700px',
				width: '100px',
				height: '32px',
				// SimpleButton.vue's .1rem block margin moves the border box below the fixed top.
				margin: '0',
			})
		})
		await page.locator('.property-due-date .datepicker button.show').click()
	}
}

function normalizeSelector(selector: string) {
	return selector
		.replace(/\[data-v-[\w-]+\]/g, '')
		.split(',')
		.map(part => part.trim().replace(/\s*([>+~])\s*/g, '$1').replace(/\s+/g, ' '))
		.sort()
		.join(',')
}

function rowKey(row: GuardRow) {
	return `${normalizeSelector(row.ruleSelector ?? row.selector)} :: ${row.declaration ?? row.property}`
		+ ` :: ${row.occurrence ?? 1}`
}

async function insetRules(page: Page) {
	return page.evaluate(() => {
		const found: {
			identity: string
			selector: string
			property: string
		}[] = []
		function visit(rules: CSSRuleList, sheetIdentity: string, parentPath: number[] = []) {
			for (const [index, rule] of Array.from(rules).entries()) {
				const path = [...parentPath, index]
				if (rule instanceof CSSStyleRule) {
					// cssText preserves authored shorthands; style iteration can expand
					// the ContentAuth padding shorthand into unresolved longhands.
					for (const declaration of rule.style.cssText.split(';')) {
						const colon = declaration.indexOf(':')
						if (colon < 0 || !declaration.slice(colon + 1).includes('--app-top-inset')) continue
						const property = declaration.slice(0, colon).trim()
						found.push({
							identity: JSON.stringify([sheetIdentity, path, rule.selectorText, property]),
							selector: rule.selectorText,
							property,
						})
					}
				}
				// Include inactive @media/@supports rules as well as active ones.
				if ('cssRules' in rule) visit((rule as CSSGroupingRule).cssRules, sheetIdentity, path)
			}
		}
		const sheetOccurrences = new Map<string, number>()
		for (const sheet of Array.from(document.styleSheets)) {
			const owner = sheet.ownerNode
			if (!sheet.href && !(owner instanceof Element)) {
				throw new Error('Inset inventory requires a stylesheet href or owner element')
			}
			// Inline owners need a stable identity across documents, independent of
			// other stylesheets being loaded. Preserve identical sheet instances too.
			const source = JSON.stringify(sheet.href ? ['href', sheet.href] : ['owner', (owner as Element).outerHTML])
			const occurrence = (sheetOccurrences.get(source) ?? 0) + 1
			sheetOccurrences.set(source, occurrence)
			visit(sheet.cssRules, JSON.stringify([source, occurrence]))
		}
		return found
	})
}

async function checkRow(page: Page, row: GuardRow, dock: boolean, normalPositions: Map<string, number>) {
	await settled(page, row.selector)
	const expected = row.pre43 + (dock ? row.dockDelta : 0)
	if (!row.js) {
		expect(Math.abs(await computed(page, row.selector, row.property) - expected), row.id).toBeLessThanOrEqual(0.5)
	}
	if (row.layout === 'scroll') {
		await page.locator(row.selector).first().evaluate(element => {
			element.scrollIntoView({
				behavior: 'instant',
				block: 'start',
			})
		})
		await expect.poll(async () => Math.abs((await rect(page, row.selector)).top - expected)).toBeLessThanOrEqual(0.5)
	} else if (row.layout === 'logo') {
		const heading = row.state === 'login' ? '.content h2.title' : 'h1.title'
		for (const selector of [`${row.selector} .logo`, `${row.selector} ${heading}`]) {
			await settled(page, selector)
			const {top} = await rect(page, selector)
			const key = `${row.id} :: ${selector}`
			if (dock) {
				expect(Math.abs(top - normalPositions.get(key)! - row.dockDelta), key).toBeLessThanOrEqual(0.5)
			} else {
				normalPositions.set(key, top)
			}
		}
	} else if (row.layout === 'top' || row.layout === 'height') {
		expect(Math.abs((await rect(page, row.selector))[row.layout] - expected)).toBeLessThanOrEqual(0.5)
	} else if (row.layout === 'dropdown' || row.layout === 'popup') {
		const dropdown = row.layout === 'dropdown'
		const anchor = dropdown ? '.task-actions-menu' : '.property-due-date .datepicker button.show'
		const anchorTop = dropdown ? 200 : 260
		expect(Math.abs((await rect(page, anchor)).top - anchorTop)).toBeLessThanOrEqual(0.5)
		await expect.poll(async () => Math.abs((await rect(page, row.selector)).top - expected)).toBeLessThanOrEqual(0.5)
		const bounds = await rect(page, row.selector)
		expect(bounds.height).toBeGreaterThan(100)
		if (dropdown) {
			expect(bounds.bottom).toBeLessThanOrEqual(anchorTop - 4 + 0.5)
			if (!dock) expect(bounds.top).toBeLessThan(40)
			expect(bounds.top).toBeGreaterThanOrEqual(dock ? 40 : 0)
			const content = page.locator('.task-actions-menu .dropdown-content')
			expect(await content.evaluate(element => element.scrollHeight > element.clientHeight)).toBe(true)
		} else {
			const viewportBottom = row.viewport.height - 8
			expect(Math.abs(bounds.bottom - viewportBottom)).toBeLessThanOrEqual(0.5)
			expect(Math.abs(bounds.height - (viewportBottom - expected))).toBeLessThanOrEqual(0.5)
			expect(await page.locator(row.selector).evaluate(element => element.scrollHeight > element.clientHeight)).toBe(true)
		}
	}
}

const ALL_ROWS = [...CSS_ROWS, ...JS_CONSUMERS]
for (const state of [...new Set(ALL_ROWS.map(row => row.state))]) {
	test.describe(`Inset declarations: ${state}`, () => {
		const stateRows = ALL_ROWS.filter(row => row.state === state)
		for (const viewport of [...new Set(stateRows.map(row => row.viewport))]) {
			test(`normal tab and Dock app at ${viewport.width}x${viewport.height}`, async ({
				page,
				apiContext,
				currentUser,
				userToken,
			}) => {
				await prepareFixture(currentUser.id)
				await setupApiUrl(page)
				await page.setViewportSize(viewport)
				if (state !== 'login' && state !== 'share') {
					await page.addInitScript(token => {
						localStorage.setItem('token', token)
						// Keep the phone install banner from intercepting signed-in interactions.
						localStorage.setItem('hideAddToHomeScreenMessage', 'true')
					}, userToken)
				}
				if (state === 'fullscreen') {
					const response = await apiContext.put('tasks/1/attachments', {
						headers: {Authorization: `Bearer ${userToken}`},
						multipart: {
							files: {
								name: 'guard.png',
								mimeType: 'image/png',
								buffer: await readFile('tests/fixtures/image-blue.png'),
							},
						},
					})
					expect(response.ok()).toBe(true)
				}
				const normalPositions = new Map<string, number>()
				for (const dock of [false, true]) {
					if (dock) {
						if (state === 'share') await page.evaluate(() => localStorage.clear())
						await behaveLikeDockApp(page)
					}
					// Reopening the same hash URL otherwise keeps the normal document, its dialogs,
					// and its CSSOM edits alive, and never runs the Dock init script.
					await page.goto('about:blank')
					await openState(page, state, dock)
					for (const row of stateRows.filter(row => row.viewport === viewport)) {
						await test.step(`${row.id}: ${dock ? 'Dock' : 'normal'}`, async () => {
							await checkRow(page, row, dock, normalPositions)
						})
					}
					if (state === 'task') {
						expect((await rect(page, '.navbar')).height).toBe(64)
					}
					if (state === 'scrolling') {
						// 3f9835b63 Modal.vue: .scrolling .modal-content { margin: $modal-margin auto } (4rem).
						expect(await computed(page, 'dialog[open] .modal-content', 'margin-block-start')).toBe(64)
						expect((await rect(page, 'dialog[open] .modal-content')).top).toBe(dock ? 104 : 64)
					}
				}
			})
		}
	})
}

// Use the base test so this source-only check does not start a browser or the
// automatic API/database fixture from support/fixtures.
nodeTest('every source inset consumer has a guard row', async () => {
	const sourceRoot = new URL('../../../src/', import.meta.url)
	const found = new Set<string>()
	const foundJsCalls = new Map<string, number>()
	async function visit(directory = '') {
		for (const entry of await readdir(new URL(directory, sourceRoot), {withFileTypes: true})) {
			const file = `${directory}${entry.name}`
			if (entry.isDirectory()) {
				await visit(`${file}/`)
				continue
			}
			if (!entry.isFile() || !/\.(vue|ts|scss)$/.test(file)
				|| file.endsWith('.test.ts') || file === 'composables/useDockAppOffset.ts') continue
			const source = await readFile(new URL(file, sourceRoot), 'utf8')
			if (source.includes('app-top-inset') || source.includes('appTopInset(')) found.add(file)
			const jsCalls = source.match(/\bappTopInset\(/g)?.length ?? 0
			if (jsCalls) foundJsCalls.set(file, jsCalls)
		}
	}
	await visit()
	const expectedJsCalls = new Map<string, number>()
	for (const row of JS_CONSUMERS) {
		expectedJsCalls.set(row.file, (expectedJsCalls.get(row.file) ?? 0) + 1)
	}
	const jsCallMismatches = [...new Set([...foundJsCalls.keys(), ...expectedJsCalls.keys()])]
		.sort()
		.map(file => ({
			file,
			expected: expectedJsCalls.get(file) ?? 0,
			found: foundJsCalls.get(file) ?? 0,
		}))
		.filter(({expected, found}) => expected !== found)
	expect(jsCallMismatches, `JS call counts must equal JS_CONSUMERS rows; ${jsCallMismatches
		.map(({file, expected, found}) => `${file}: expected ${expected}, found ${found}`)
		.join('; ')}`).toEqual([])
	const expected = new Set([...CSS_ROWS, ...JS_CONSUMERS].map(row => row.file))
	const missing = [...found].filter(file => !expected.has(file)).sort()
	const extra = [...expected].filter(file => !found.has(file)).sort()
	expect({
		missingGuardFiles: missing,
		extraGuardFiles: extra,
	}, `Source inventory must equal guarded files; missing: ${missing.join(', ') || '(none)'}; extra: ${extra.join(', ') || '(none)'}`).toEqual({
		missingGuardFiles: [],
		extraGuardFiles: [],
	})
})

test('every loaded inset declaration has a guard row, including anonymous shells', async ({
	page,
	browser,
	currentUser,
	userToken,
}) => {
	await prepareFixture(currentUser.id)
	await setupApiUrl(page)
	await page.addInitScript(token => localStorage.setItem('token', token), userToken)
	const declarations = new Map<string, {selector: string, property: string}>()
	for (const state of ['task', 'kanban', 'about'] as const) {
		await openState(page, state)
		for (const declaration of await insetRules(page)) declarations.set(declaration.identity, declaration)
	}
	// Separate storage prevents a signed-in login redirect or share mode leaking
	// into the signed-in inventory. Configure the anonymous API just like WP01.
	const anonymous = await browser.newContext({baseURL: process.env.BASE_URL || 'http://127.0.0.1:4173'})
	try {
		const shell = await anonymous.newPage()
		await setupApiUrl(shell)
		for (const state of ['login', 'share'] as const) {
			await openState(shell, state)
			for (const declaration of await insetRules(shell)) declarations.set(declaration.identity, declaration)
		}
	} finally {
		await anonymous.close()
	}
	// Deduplicate raw declarations across snapshots before stripping Vue scopes.
	// Distinct scoped rules with the same normalized selector must each get a row.
	const occurrences = new Map<string, number>()
	const found = new Set([...declarations.values()].map(({selector, property}) => {
		const key = `${normalizeSelector(selector)} :: ${property}`
		const occurrence = (occurrences.get(key) ?? 0) + 1
		occurrences.set(key, occurrence)
		return `${key} :: ${occurrence}`
	}))
	const expected = new Set(CSS_ROWS.map(rowKey))
	const missing = [...found].filter(key => !expected.has(key)).sort()
	const extra = [...expected].filter(key => !found.has(key)).sort()
	expect({
		missingGuardRows: missing,
		extraGuardRows: extra,
	}, 'CSS inventory must equal guarded declarations; JS floors are listed separately in JS_CONSUMERS').toEqual({
		missingGuardRows: [],
		extraGuardRows: [],
	})
})
