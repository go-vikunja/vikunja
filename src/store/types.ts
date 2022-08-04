import type { IBucket } from '@/modelTypes/IBucket'
import type { IUserSettings } from '@/modelTypes/IUserSettings'
import type { IList } from '@/modelTypes/IList'
import type { IAttachment } from '@/modelTypes/IAttachment'
import type { ILabel } from '@/modelTypes/ILabel'
import type { INamespace } from '@/modelTypes/INamespace'

export interface RootStoreState {
	loading: boolean,
	loadingModule: null,
	currentList: IList,
	background: string,
	blurHash: string,
	hasTasks: boolean,
	menuActive: boolean,
	keyboardShortcutsActive: boolean,
	quickActionsActive: boolean,
}

export interface AttachmentState {
	attachments: IAttachment[],
}

export const AUTH_TYPES = {
	'UNKNOWN': 0,
	'USER': 1,
	'LINK_SHARE': 2,
} as const

export interface Info {
	id: number // what kind of id is this?
	type: typeof AUTH_TYPES[keyof typeof AUTH_TYPES],
	getAvatarUrl: () => string
	settings: IUserSettings
	name: string
	email: string
	exp: any
}
export interface AuthState {
	authenticated: boolean,
	isLinkShareAuth: boolean,
	info: Info | null,
	needsTotpPasscode: boolean,
	avatarUrl: string,
	lastUserInfoRefresh: Date | null,
	settings: IUserSettings,
}

export interface ConfigState {
	version: string,
	frontendUrl: string,
	motd: string,
	linkSharingEnabled: boolean,
	maxFileSize: '20MB',
	registrationEnabled: boolean,
	availableMigrators: [],
	taskAttachmentsEnabled: boolean,
	totpEnabled: boolean,
	enabledBackgroundProviders: [],
	legal: {
		imprintUrl: string,
		privacyPolicyUrl: string,
	},
	caldavEnabled: boolean,
	userDeletionEnabled: boolean,
	taskCommentsEnabled: boolean,
	auth: {
		local: {
			enabled: boolean,
		},
		openidConnect: {
			enabled: boolean,
			redirectUrl: string,
			providers: [],
		},
	},
}

export interface KanbanState {
	buckets: IBucket[],
	listId: IList['id'],
	bucketLoading: {},
	taskPagesPerBucket: {
		[id: IBucket['id']]: number
	},
	allTasksLoadedForBucket: {
		[id: IBucket['id']]: boolean
	},
}

export interface LabelState {
	labels: {
		[id: ILabel['id']]: ILabel
	},
	loaded: boolean,
}

export interface ListState {
	[id: IList['id']]: IList
}

export interface NamespaceState {
	namespaces: INamespace[]
}

export interface TaskState {}


export type StoreState = RootStoreState & {
	config: ConfigState,
	auth: AuthState,
	namespaces: NamespaceState,
	kanban: KanbanState,
	tasks: TaskState,
	lists: ListState,
	attachments: AttachmentState,
	labels: LabelState,
}