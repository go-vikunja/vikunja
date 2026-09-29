import {defineComponent, h} from 'vue'
import {mount, type VueWrapper} from '@vue/test-utils'
import {QueryClient, VueQueryPlugin} from '@tanstack/vue-query'
import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest'

import {useSecretMutation} from './secretMutation'

const mutationFn = vi.fn<(password: string) => Promise<string>>()

let client: QueryClient
let wrapper: VueWrapper
let mutation: ReturnType<typeof useSecretMutation<string, Error, string>>

beforeEach(() => {
	mutationFn.mockReset()
	client = new QueryClient()
	wrapper = mount(defineComponent({
		setup() {
			mutation = useSecretMutation({mutationFn})
			return () => h('div')
		},
	}), {
		global: {
			plugins: [[VueQueryPlugin, {queryClient: client}]],
		},
	})
})

afterEach(() => wrapper.unmount())

function settleGc() {
	return new Promise(resolve => setTimeout(resolve))
}

describe('useSecretMutation', () => {
	it('holds the variables only while the mutation is pending', async () => {
		let resolve!: (token: string) => void
		mutationFn.mockReturnValue(new Promise(done => { resolve = done }))

		const pending = mutation.mutateAsync('password')
		await vi.waitFor(() => expect(mutation.variables.value).toBe('password'))
		expect(client.getMutationCache().getAll().map(entry => entry.options.gcTime)).toEqual([0])

		resolve('token')
		expect(await pending).toBe('token')
		expect(mutation.variables.value).toBeUndefined()
		expect(mutation.data.value).toBeUndefined()
		await settleGc()
		expect(client.getMutationCache().getAll()).toEqual([])
	})

	it('drops the variables once mutateAsync rejects', async () => {
		const cause = new Error('wrong password')
		mutationFn.mockRejectedValue(cause)

		await expect(mutation.mutateAsync('password')).rejects.toBe(cause)

		expect(mutation.variables.value).toBeUndefined()
		expect(mutation.error.value).toBeNull()
		await settleGc()
		expect(client.getMutationCache().getAll()).toEqual([])
	})

	it('drops the variables once mutate settles', async () => {
		mutationFn.mockResolvedValue('token')
		const onSettled = vi.fn()

		mutation.mutate('password', {onSettled})
		await vi.waitFor(() => expect(onSettled).toHaveBeenCalledTimes(1))
		await settleGc()

		expect(onSettled.mock.calls[0]?.slice(0, 3)).toEqual([
			'token',
			null,
			'password',
		])
		expect(mutation.variables.value).toBeUndefined()
		expect(client.getMutationCache().getAll()).toEqual([])
	})

	it('drops the variables once mutate fails', async () => {
		mutationFn.mockRejectedValue(new Error('wrong password'))
		const onError = vi.fn()

		mutation.mutate('password', {onError})
		await vi.waitFor(() => expect(onError).toHaveBeenCalledTimes(1))
		await settleGc()

		expect(mutation.variables.value).toBeUndefined()
		expect(client.getMutationCache().getAll()).toEqual([])
	})
})
