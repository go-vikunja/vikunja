import {queryOptions, useMutation} from '@tanstack/vue-query'
import {
	migrationCsvStatus,
	migrationMicrosoftTodoStatus,
	migrationPlankaStatus,
	migrationTicktickStatus,
	migrationTodoistStatus,
	migrationTrelloStatus,
	migrationVikunjaFileStatus,
	migrationWekanStatus,
	migrationTodoistAuth,
	migrationTrelloAuth,
	migrationMicrosoftTodoAuth,
	migrationCsvMigrate,
	migrationMicrosoftTodoMigrate,
	migrationPlankaMigrate,
	migrationTicktickMigrate,
	migrationTodoistMigrate,
	migrationTrelloMigrate,
	migrationVikunjaFileMigrate,
	migrationWekanMigrate,
	migrationCsvDetect,
	migrationCsvPreview,
} from '@/client/generated'
import type {
	MigrationCredentialsBodyWritable,
	MigrationMigrateBodyWritable,
	MigrationCsvMigrateData,
	MigrationCsvPreviewData,
} from '@/client/generated'
import type {MigrationProviderOfKind} from '@/views/migrate/migrators'
import {contextMutationOptions} from './contextMutation'
import {useSecretMutation} from './secretMutation'
import {projectKeys} from './projects'
import {taskKeys} from './tasks'
const statusOperations = {
	'csv': migrationCsvStatus,
	'microsoft-todo': migrationMicrosoftTodoStatus,
	'planka': migrationPlankaStatus,
	'ticktick': migrationTicktickStatus,
	'todoist': migrationTodoistStatus,
	'trello': migrationTrelloStatus,
	'vikunja-file': migrationVikunjaFileStatus,
	'wekan': migrationWekanStatus,
}
export type MigrationProvider = keyof typeof statusOperations
export type MigrationKind = 'file' | 'oauth' | 'credentials' | 'csv'
export const migrationKeys = {
	status: (provider: MigrationProvider) => ['migration', provider, 'status'] as const,
}
export async function fetchMigrationStatus(provider: MigrationProvider, signal?: AbortSignal) {
	return (await statusOperations[provider]({signal})).data
}
export function migrationStatusQuery(provider: MigrationProvider) {
	return queryOptions({
		queryKey: migrationKeys.status(provider),
		queryFn: ({signal}) => fetchMigrationStatus(provider, signal),
		staleTime: 0,
		retry: false,
	})
}
const authOperations = {
	'todoist': migrationTodoistAuth,
	'trello': migrationTrelloAuth,
	'microsoft-todo': migrationMicrosoftTodoAuth,
} satisfies Record<MigrationProviderOfKind<'oauth'>, unknown>
export function migrationAuthMutationOptions() {
	return contextMutationOptions({
		mutationFn: async (provider: MigrationProviderOfKind<'oauth'>) => (await authOperations[provider]()).data,
		toastError: () => false,
	})
}
const fileOperations = {
	'ticktick': migrationTicktickMigrate,
	'wekan': migrationWekanMigrate,
	'vikunja-file': migrationVikunjaFileMigrate,
} satisfies Record<MigrationProviderOfKind<'file'>, unknown>
const oauthOperations = {
	'todoist': migrationTodoistMigrate,
	'trello': migrationTrelloMigrate,
	'microsoft-todo': migrationMicrosoftTodoMigrate,
} satisfies Record<MigrationProviderOfKind<'oauth'>, unknown>
const credentialsOperations = {
	'planka': migrationPlankaMigrate,
} satisfies Record<MigrationProviderOfKind<'credentials'>, unknown>
const csvOperations = {
	'csv': migrationCsvMigrate,
} satisfies Record<MigrationProviderOfKind<'csv'>, unknown>
export type StartMigrationInput =
	| {
		kind: 'file',
		provider: MigrationProviderOfKind<'file'>,
		file: File,
	}
	| {
		kind: 'oauth',
		provider: MigrationProviderOfKind<'oauth'>,
		body: MigrationMigrateBodyWritable,
	}
	| {
		kind: 'credentials',
		provider: MigrationProviderOfKind<'credentials'>,
		body: MigrationCredentialsBodyWritable,
	}
	| {
		kind: 'csv',
		provider: MigrationProviderOfKind<'csv'>,
		body: MigrationCsvMigrateData['body'],
	}
export function startMigrationMutationOptions() {
	return contextMutationOptions({
		mutationFn: async (input: StartMigrationInput) => {
			switch (input.kind) {
				case 'file': return (await fileOperations[input.provider]({body: {import: input.file}})).data
				case 'oauth': return (await oauthOperations[input.provider]({body: input.body})).data
				case 'credentials': return (await credentialsOperations[input.provider]({body: input.body})).data
				case 'csv': return (await csvOperations[input.provider]({body: input.body})).data
			}
		},
		onSettled: ({provider}, client) => client.invalidateQueries({queryKey: migrationKeys.status(provider)}),
		toastError: () => false,
	})
}
export function migrationCompletedMutationOptions() {
	return contextMutationOptions({
		mutationFn: async () => undefined,
		onSettled: (_input, client) => Promise.all([
			client.invalidateQueries({queryKey: projectKeys.all}),
			client.invalidateQueries({queryKey: taskKeys.all}),
		]),
	})
}
export function detectCsvMutationOptions() {
	return contextMutationOptions({
		mutationFn: async (file: File) => (await migrationCsvDetect({body: {import: file}})).data,
		toastError: () => false,
	})
}
export function previewCsvMutationOptions() {
	return contextMutationOptions({
		mutationFn: async (body: MigrationCsvPreviewData['body']) => (await migrationCsvPreview({body})).data,
		toastError: () => false,
	})
}

export const useMigrationAuthMutation = () => useMutation(migrationAuthMutationOptions())
export const useStartMigrationMutation = () => useSecretMutation(startMigrationMutationOptions())
export const useDetectCsvMutation = () => useMutation(detectCsvMutationOptions())
export const usePreviewCsvMutation = () => useMutation(previewCsvMutationOptions())
