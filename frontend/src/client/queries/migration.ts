import {queryOptions, useMutation} from '@tanstack/vue-query'
import {migrationCsvStatus, migrationMicrosoftTodoStatus, migrationPlankaStatus, migrationTicktickStatus, migrationTodoistStatus, migrationTrelloStatus, migrationVikunjaFileStatus, migrationWekanStatus, migrationTodoistAuth, migrationTrelloAuth, migrationMicrosoftTodoAuth, migrationCsvMigrate, migrationMicrosoftTodoMigrate, migrationPlankaMigrate, migrationTicktickMigrate, migrationTodoistMigrate, migrationTrelloMigrate, migrationVikunjaFileMigrate, migrationWekanMigrate, migrationCsvDetect, migrationCsvPreview, type MigrationCredentialsBodyWritable, type MigrationMigrateBodyWritable, type MigrationCsvMigrateData} from '@/client/generated'
import {contextMutationOptions} from './contextMutation'
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
export const migrationKeys = {status: (provider: MigrationProvider) => ['migration', provider, 'status'] as const}
export function migrationStatusQuery(provider: MigrationProvider) {
	return queryOptions({queryKey: migrationKeys.status(provider), queryFn: async ({signal}) => (await statusOperations[provider]({signal})).data, staleTime: 0, retry: false})
}
const authOperations = {
	'todoist': migrationTodoistAuth,
	'trello': migrationTrelloAuth,
	'microsoft-todo': migrationMicrosoftTodoAuth,
}
export function migrationAuthMutationOptions() {
	return contextMutationOptions({mutationFn: async (provider: keyof typeof authOperations) => (await authOperations[provider]()).data, toastError: () => false})
}
const fileOperations = {
	'ticktick': migrationTicktickMigrate,
	'wekan': migrationWekanMigrate,
	'vikunja-file': migrationVikunjaFileMigrate,
}
const oauthOperations = {
	'todoist': migrationTodoistMigrate,
	'trello': migrationTrelloMigrate,
	'microsoft-todo': migrationMicrosoftTodoMigrate,
}
export type StartMigrationInput =
	| {kind: 'file', provider: keyof typeof fileOperations, file: File}
	| {kind: 'oauth', provider: keyof typeof oauthOperations, body: MigrationMigrateBodyWritable}
	| {kind: 'credentials', provider: 'planka', body: MigrationCredentialsBodyWritable}
	| {kind: 'csv', provider: 'csv', body: MigrationCsvMigrateData['body']}
export function startMigrationMutationOptions() {
	return {
		...contextMutationOptions({
			mutationFn: async (input: StartMigrationInput) => {
				switch (input.kind) {
					case 'file': return (await fileOperations[input.provider]({body: {import: input.file}})).data
					case 'oauth': return (await oauthOperations[input.provider]({body: input.body})).data
					case 'credentials': return (await migrationPlankaMigrate({body: input.body})).data
					case 'csv': return (await migrationCsvMigrate({body: input.body})).data
				}
			},
			onSettled: ({provider}, client) => client.invalidateQueries({queryKey: migrationKeys.status(provider)}),
			toastError: () => false,
		}),
		// Input holds the plaintext Planka password.
		gcTime: 0,
	}
}
export function migrationCompletedMutationOptions() {
	return contextMutationOptions({
		mutationFn: async () => undefined,
		onSettled: (_input, client) => Promise.all([client.invalidateQueries({queryKey: projectKeys.all}), client.invalidateQueries({queryKey: taskKeys.all})]),
	})
}
export function detectCsvMutationOptions() {
	return contextMutationOptions({mutationFn: async (file: File) => (await migrationCsvDetect({body: {import: file}})).data, toastError: () => false})
}
export function previewCsvMutationOptions() {
	return contextMutationOptions({mutationFn: async (body: MigrationCsvMigrateData['body']) => (await migrationCsvPreview({body})).data, toastError: () => false})
}

export const useMigrationAuthMutation = () => useMutation(migrationAuthMutationOptions())
export const useStartMigrationMutation = () => useMutation(startMigrationMutationOptions())
export const useDetectCsvMutation = () => useMutation(detectCsvMutationOptions())
export const usePreviewCsvMutation = () => useMutation(previewCsvMutationOptions())
