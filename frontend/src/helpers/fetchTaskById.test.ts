import {describe, it, expect, beforeEach, vi} from 'vitest'
import {queryClient} from '@/client/queryClient'
import {fetchTaskById} from './fetchTaskById'
const sdk = vi.hoisted(() => ({tasksRead: vi.fn()}))
vi.mock('@/client/generated', () => sdk)

describe('task lookups', () => {
 beforeEach(() => { queryClient.clear(); queryClient.setDefaultOptions({queries: {retry: false, staleTime: 60000}}); sdk.tasksRead.mockReset() })
 it('deduplicates concurrent lookups through the query cache', async () => {
  sdk.tasksRead.mockResolvedValue({data: {id: 1, title: 'one'}})
  const [first, second] = await Promise.all([fetchTaskById(1), fetchTaskById(1)])
  expect(first).toEqual(second)
  expect(sdk.tasksRead).toHaveBeenCalledTimes(1)
 })
 it('retries a later lookup after a transient failure', async () => {
  sdk.tasksRead.mockRejectedValueOnce(new Error('offline')).mockResolvedValueOnce({data: {id: 1}})
  await expect(fetchTaskById(1)).rejects.toThrow('offline')
  await expect(fetchTaskById(1)).resolves.toEqual({id: 1})
 })
})
