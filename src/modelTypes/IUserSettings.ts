
import type {IAbstract} from './IAbstract'
import type {IProject} from './IProject'
import type {PrefixMode} from '@/modules/parseTaskText'
import type {BasicColorSchema} from '@vueuse/core'
import type {SupportedLocale} from '@/i18n'

export interface IFrontendSettings {
	playSoundWhenDone: boolean
	quickAddMagicMode: PrefixMode
	colorSchema: BasicColorSchema
	filterIdUsedOnOverview: IProject['id'] | null
}

export interface IUserSettings extends IAbstract {
	name: string
	emailRemindersEnabled: boolean
	discoverableByName: boolean
	discoverableByEmail: boolean
	overdueTasksRemindersEnabled: boolean
	overdueTasksRemindersTime: string | Date
	defaultProjectId: undefined | IProject['id']
	weekStart: 0 | 1 | 2 | 3 | 4 | 5 | 6
	timezone: string
	language: SupportedLocale
	frontendSettings: IFrontendSettings
}