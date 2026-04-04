import {describe, it, expect} from 'vitest'
import UserSettingsModel from './userSettings'
import {PrefixMode} from '@/modules/quickAddMagic'
import {DATE_DISPLAY} from '@/constants/dateDisplay'
import {TIME_FORMAT} from '@/constants/timeFormat'
import {RELATION_KIND} from '@/types/IRelationKind'

describe('UserSettingsModel', () => {
	it('should have correct defaults when constructed without arguments', () => {
		const settings = new UserSettingsModel()
		expect(settings.frontendSettings.showLastViewed).toBe(true)
		expect(settings.frontendSettings.playSoundWhenDone).toBe(true)
		expect(settings.frontendSettings.quickAddMagicMode).toBe(PrefixMode.Default)
		expect(settings.frontendSettings.colorSchema).toBe('auto')
		expect(settings.frontendSettings.alwaysShowBucketTaskCount).toBe(false)
		expect(settings.frontendSettings.commentSortOrder).toBe('asc')
	})

	it('should default showLastViewed to true', () => {
		const settings = new UserSettingsModel()
		expect(settings.frontendSettings.showLastViewed).toBe(true)
	})

	it('should allow showLastViewed to be overridden via frontendSettings', () => {
		const settings = new UserSettingsModel({
			frontendSettings: {
				playSoundWhenDone: true,
				quickAddMagicMode: PrefixMode.Default,
				colorSchema: 'auto',
				allowIconChanges: true,
				filterIdUsedOnOverview: null,
				dateDisplay: DATE_DISPLAY.RELATIVE,
				timeFormat: TIME_FORMAT.HOURS_24,
				defaultTaskRelationType: RELATION_KIND.RELATED,
				backgroundBrightness: null,
				alwaysShowBucketTaskCount: false,
				showLastViewed: false,
				sidebarWidth: null,
				commentSortOrder: 'asc',
				desktopQuickEntryShortcut: 'CmdOrCtrl+Shift+A',
			},
		})
		expect(settings.frontendSettings.showLastViewed).toBe(false)
	})

	it('should replace entire frontendSettings when provided via constructor', () => {
		const settings = new UserSettingsModel({
			frontendSettings: {
				playSoundWhenDone: false,
				quickAddMagicMode: PrefixMode.Default,
				colorSchema: 'dark',
				allowIconChanges: false,
				filterIdUsedOnOverview: null,
				dateDisplay: DATE_DISPLAY.RELATIVE,
				timeFormat: TIME_FORMAT.HOURS_12,
				defaultTaskRelationType: RELATION_KIND.RELATED,
				backgroundBrightness: 50,
				alwaysShowBucketTaskCount: true,
				showLastViewed: true,
				sidebarWidth: 300,
				commentSortOrder: 'desc',
				desktopQuickEntryShortcut: 'CmdOrCtrl+Shift+B',
			},
		})
		expect(settings.frontendSettings.showLastViewed).toBe(true)
		expect(settings.frontendSettings.colorSchema).toBe('dark')
		expect(settings.frontendSettings.sidebarWidth).toBe(300)
	})
})
