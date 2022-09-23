import type { IBucket } from '@/modelTypes/IBucket'
import type { IUserSettings } from '@/modelTypes/IUserSettings'
import type { IList } from '@/modelTypes/IList'
import type { IAttachment } from '@/modelTypes/IAttachment'
import type { ILabel } from '@/modelTypes/ILabel'
import type { INamespace } from '@/modelTypes/INamespace'
import type { IUser } from '@/modelTypes/IUser'

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
	logoVisible: boolean,
}

export interface AttachmentState {
	attachments: IAttachment[],
}

export interface AuthState {
	authenticated: boolean,
	isLinkShareAuth: boolean,
	info: IUser | null,
	needsTotpPasscode: boolean,
	avatarUrl: string,
	lastUserInfoRefresh: Date | null,
	settings: IUserSettings,
	isLoading: boolean,
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
	isLoading: boolean,
}

export interface ListState {
	lists: { [id: IList['id']]: IList },
	isLoading: boolean,
}

export interface NamespaceState {
	namespaces: INamespace[]
	isLoading: boolean,
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