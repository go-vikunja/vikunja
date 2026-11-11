import {beforeEach, describe, expect, it, vi} from 'vitest'
import {defineComponent, h} from 'vue'
import {mount} from '@vue/test-utils'
import {QueryClient, VueQueryPlugin} from '@tanstack/vue-query'
import {
	detectCsvMutationOptions,
	migrationAuthMutationOptions,
	migrationKeys,
	migrationStatusQuery,
	previewCsvMutationOptions,
	startMigrationMutationOptions,
	useStartMigrationMutation,
	type MigrationProvider,
	type StartMigrationInput,
} from './migration'
import {MIGRATORS, isMigratorOfKind} from '@/views/migrate/migrators'

const sdk = vi.hoisted(() => ({
	migrationCsvStatus: vi.fn(),
	migrationMicrosoftTodoStatus: vi.fn(),
	migrationPlankaStatus: vi.fn(),
	migrationTicktickStatus: vi.fn(),
	migrationTodoistStatus: vi.fn(),
	migrationTrelloStatus: vi.fn(),
	migrationVikunjaFileStatus: vi.fn(),
	migrationWekanStatus: vi.fn(),
	migrationTodoistAuth: vi.fn(),
	migrationTrelloAuth: vi.fn(),
	migrationMicrosoftTodoAuth: vi.fn(),
	migrationCsvMigrate: vi.fn(),
	migrationMicrosoftTodoMigrate: vi.fn(),
	migrationPlankaMigrate: vi.fn(),
	migrationTicktickMigrate: vi.fn(),
	migrationTodoistMigrate: vi.fn(),
	migrationTrelloMigrate: vi.fn(),
	migrationVikunjaFileMigrate: vi.fn(),
	migrationWekanMigrate: vi.fn(),
	migrationCsvDetect: vi.fn(),
	migrationCsvPreview: vi.fn(),
}))
vi.mock('@/client/generated', () => sdk)
vi.mock('@/message', () => ({success: vi.fn(), error: vi.fn()}))

function startInputFor(provider: MigrationProvider): StartMigrationInput {
	const file = new File(['export'], 'export.json')
	if (isMigratorOfKind(provider, 'file')) {
		return {
			kind: 'file',
			provider,
			file,
		}
	}
	if (isMigratorOfKind(provider, 'oauth')) {
		return {
			kind: 'oauth',
			provider,
			body: {code: 'code'},
		}
	}
	if (isMigratorOfKind(provider, 'credentials')) {
		return {
			kind: 'credentials',
			provider,
			body: {url: 'https://planka.example'},
		}
	}
	if (isMigratorOfKind(provider, 'csv')) {
		return {
			kind: 'csv',
			provider,
			body: {
				import: file,
				config: '{}',
			},
		}
	}
	throw new Error(`${provider} has no migration kind`)
}

describe('migration queries', () => {
	beforeEach(() => vi.resetAllMocks())

	it('reads the status of the requested provider only', async () => {
		const client = new QueryClient()
		sdk.migrationTrelloStatus.mockResolvedValue({data: {started_at: '2024-01-15T10:00:00Z'}})

		expect(await client.fetchQuery(migrationStatusQuery('trello'))).toEqual({started_at: '2024-01-15T10:00:00Z'})
		expect(sdk.migrationTrelloStatus).toHaveBeenCalledWith({signal: expect.any(AbortSignal)})
		expect(sdk.migrationCsvStatus).not.toHaveBeenCalled()
	})

	it.each([
		['ticktick', 'migrationTicktickMigrate'],
		['wekan', 'migrationWekanMigrate'],
		['vikunja-file', 'migrationVikunjaFileMigrate'],
	] as const)('uploads a %s export through its own operation', async (provider, operation) => {
		const client = new QueryClient()
		const file = new File(['export'], 'export.json')
		sdk[operation].mockResolvedValue({data: {}})

		await client.getMutationCache().build(client, startMigrationMutationOptions()).execute({
			kind: 'file',
			provider,
			file,
		})

		expect(sdk[operation]).toHaveBeenCalledWith({body: {import: file}})
	})

	it.each([
		['todoist', 'migrationTodoistMigrate'],
		['trello', 'migrationTrelloMigrate'],
		['microsoft-todo', 'migrationMicrosoftTodoMigrate'],
	] as const)('sends the %s oauth code through its own operation', async (provider, operation) => {
		const client = new QueryClient()
		sdk[operation].mockResolvedValue({data: {}})

		await client.getMutationCache().build(client, startMigrationMutationOptions()).execute({
			kind: 'oauth',
			provider,
			body: {code: `${provider}-code`},
		})

		expect(sdk[operation]).toHaveBeenCalledWith({body: {code: `${provider}-code`}})
	})

	it.each([
		['todoist', 'migrationTodoistAuth'],
		['trello', 'migrationTrelloAuth'],
		['microsoft-todo', 'migrationMicrosoftTodoAuth'],
	] as const)('reads the %s auth url from its own operation', async (provider, operation) => {
		const client = new QueryClient()
		sdk[operation].mockResolvedValue({data: {url: `https://${provider}.example/auth`}})

		const result = await client.getMutationCache().build(client, migrationAuthMutationOptions()).execute(provider)

		expect(result).toEqual({url: `https://${provider}.example/auth`})
		expect(sdk[operation]).toHaveBeenCalledWith()
	})

	it('sends planka credentials to the planka operation', async () => {
		const client = new QueryClient()
		const body = {
			url: 'https://planka.example',
			username: 'user',
			password: 'secret',
		}
		sdk.migrationPlankaMigrate.mockResolvedValue({data: {}})

		await client.getMutationCache().build(client, startMigrationMutationOptions()).execute({
			kind: 'credentials',
			provider: 'planka',
			body,
		})

		expect(sdk.migrationPlankaMigrate).toHaveBeenCalledWith({body})
	})

	it('sends the csv file and its config to the csv migrate operation', async () => {
		const client = new QueryClient()
		const body = {
			import: new File(['title\n'], 'tasks.csv'),
			config: '{"delimiter":","}',
		}
		sdk.migrationCsvMigrate.mockResolvedValue({data: {}})

		await client.getMutationCache().build(client, startMigrationMutationOptions()).execute({
			kind: 'csv',
			provider: 'csv',
			body,
		})

		expect(sdk.migrationCsvMigrate).toHaveBeenCalledWith({body})
	})

	it.each(Object.values(MIGRATORS).map(({id}) => id))(
		'starts a %s migration through exactly one operation',
		async provider => {
			const client = new QueryClient()
			const migrateOperations = Object.entries(sdk).filter(([name]) => name.endsWith('Migrate'))
			for (const [, operation] of migrateOperations) operation.mockResolvedValue({data: {}})

			await client.getMutationCache().build(client, startMigrationMutationOptions()).execute(startInputFor(provider))

			expect(migrateOperations.filter(([, operation]) => operation.mock.calls.length > 0)).toHaveLength(1)
		},
	)

	it('invalidates the status of the provider it started and no other', async () => {
		const client = new QueryClient()
		client.setQueryData(migrationKeys.status('todoist'), {started_at: null})
		client.setQueryData(migrationKeys.status('trello'), {started_at: null})
		sdk.migrationTodoistMigrate.mockResolvedValue({data: {}})

		await client.getMutationCache().build(client, startMigrationMutationOptions()).execute({
			kind: 'oauth',
			provider: 'todoist',
			body: {code: 'code'},
		})

		expect(client.getQueryState(migrationKeys.status('todoist'))?.isInvalidated).toBe(true)
		expect(client.getQueryState(migrationKeys.status('trello'))?.isInvalidated).toBe(false)
	})

	it('keeps the started migration input out of the mutation cache', async () => {
		const client = new QueryClient()
		sdk.migrationPlankaMigrate.mockResolvedValue({data: {}})
		let mutation!: ReturnType<typeof useStartMigrationMutation>
		const wrapper = mount(defineComponent({
			setup() {
				mutation = useStartMigrationMutation()
				return () => h('div')
			},
		}), {
			global: {
				plugins: [[VueQueryPlugin, {queryClient: client}]],
			},
		})

		await mutation.mutateAsync({
			kind: 'credentials',
			provider: 'planka',
			body: {
				url: 'https://planka.example',
				username: 'user',
				password: 'secret',
			},
		})
		await new Promise(resolve => setTimeout(resolve))

		expect(client.getMutationCache().getAll()).toEqual([])
		wrapper.unmount()
	})

	it('detects and previews a csv through the csv operations', async () => {
		const client = new QueryClient()
		const file = new File(['title\n'], 'tasks.csv')
		sdk.migrationCsvDetect.mockResolvedValue({data: {delimiter: ','}})
		sdk.migrationCsvPreview.mockResolvedValue({data: {total_rows: 1}})

		await client.getMutationCache().build(client, detectCsvMutationOptions()).execute(file)
		await client.getMutationCache().build(client, previewCsvMutationOptions()).execute({
			import: file,
			config: '{}',
		})

		expect(sdk.migrationCsvDetect).toHaveBeenCalledWith({body: {import: file}})
		expect(sdk.migrationCsvPreview).toHaveBeenCalledWith({body: {import: file, config: '{}'}})
	})
})
