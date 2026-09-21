import type {BasicColorSchema} from '@vueuse/core'
import type {SupportedLocale} from '@/i18n'
import type {DefaultProjectViewKind} from '@/constants/projectView'
import type {Priority} from '@/constants/priorities'
import type {DateDisplay} from '@/constants/dateDisplay'
import type {TimeFormat} from '@/constants/timeFormat'
import type {IRelationKind} from '@/types/IRelationKind'
import type {TaskReminder} from '@/client/generated'

export interface FrontendSettings {
	play_sound_when_done: boolean
	quick_add_magic_mode: PrefixMode
	color_schema: BasicColorSchema
	allow_icon_changes: boolean
	filter_id_used_on_overview: number | null
	default_view?: DefaultProjectViewKind
	minimum_priority?: Priority
	date_display: DateDisplay
	time_format: TimeFormat
	default_task_relation_type: IRelationKind
	background_brightness: number | null
	always_show_bucket_task_count: boolean
	show_last_viewed: boolean
	sidebar_width: number | null
	comment_sort_order: 'asc' | 'desc'
	desktop_quick_entry_shortcut: string
	quick_add_default_reminders: {relative_period?: number}[]
	time_tracking_default_start?: string
	default_due_time?: string
}

export function taskRemindersFromSettings(
	reminders: Readonly<FrontendSettings['quick_add_default_reminders']> | undefined,
): TaskReminder[] {
	return (reminders ?? []).map(reminder => ({relative_period: reminder.relative_period}))
}

export interface ExtraSettingsLink {
	text: string
	url: string
}

export interface ExtraSettingsLinks {
	[key: string]: ExtraSettingsLink
}

import type {UserGeneralSettings} from '@/client/generated'

export type UserSettingsResponse = Required<Omit<UserGeneralSettings, '$schema' | 'frontend_settings' | 'extra_settings_links' | 'language'>> & {
	frontend_settings: FrontendSettings
	extra_settings_links: ExtraSettingsLinks
	language: SupportedLocale
}

import {getBrowserLanguage} from '@/i18n'
import {PrefixMode} from '@/modules/quickAddMagic'
import {DEFAULT_PROJECT_VIEW_SETTINGS} from '@/constants/projectView'
import {PRIORITIES} from '@/constants/priorities'
import {DATE_DISPLAY} from '@/constants/dateDisplay'
import {TIME_FORMAT} from '@/constants/timeFormat'
import {RELATION_KIND} from '@/types/IRelationKind'

export function createUserSettingsDraft(data: UserGeneralSettings = {}): UserSettingsResponse {
	const frontend = typeof data.frontend_settings === 'object' && data.frontend_settings !== null
		? data.frontend_settings as Partial<FrontendSettings> : {}
	return {
		name: data.name ?? '',
		email_reminders_enabled: data.email_reminders_enabled ?? true,
		discoverable_by_name: data.discoverable_by_name ?? false,
		discoverable_by_email: data.discoverable_by_email ?? false,
		overdue_tasks_reminders_enabled: data.overdue_tasks_reminders_enabled ?? true,
		overdue_tasks_reminders_time: data.overdue_tasks_reminders_time ?? '09:00',
		default_project_id: data.default_project_id ?? 0,
		week_start: data.week_start ?? 0,
		timezone: data.timezone ?? '',
		language: (data.language || getBrowserLanguage()) as SupportedLocale,
		extra_settings_links: Object.fromEntries(Object.entries(data.extra_settings_links ?? {}).filter((entry): entry is [string, ExtraSettingsLink] => {
			const value = entry[1]
			return !!value && typeof value === 'object' && 'text' in value && typeof value.text === 'string' && 'url' in value && typeof value.url === 'string'
		})),
		frontend_settings: {
			play_sound_when_done: true,
			quick_add_magic_mode: PrefixMode.Default,
			color_schema: 'auto',
			allow_icon_changes: true,
			filter_id_used_on_overview: null,
			default_view: DEFAULT_PROJECT_VIEW_SETTINGS.FIRST,
			minimum_priority: PRIORITIES.MEDIUM,
			date_display: DATE_DISPLAY.RELATIVE,
			time_format: TIME_FORMAT.HOURS_24,
			default_task_relation_type: RELATION_KIND.RELATED,
			background_brightness: null,
			always_show_bucket_task_count: false,
			show_last_viewed: true,
			sidebar_width: null,
			comment_sort_order: 'asc',
			desktop_quick_entry_shortcut: 'CmdOrCtrl+Shift+A',
			quick_add_default_reminders: [],
			default_due_time: undefined,
			time_tracking_default_start: '09:00',
			...frontend,
			quick_add_default_reminders: (frontend.quick_add_default_reminders ?? []).map(reminder => ({...reminder})),
		},
	}
}
