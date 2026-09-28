import {
	describe,
	it,
	expect,
	beforeEach,
	afterEach,
	vi,
} from 'vitest'

const {settings, playMock, createdSources} = vi.hoisted(() => ({
	settings: {playSoundWhenDone: true},
	playMock: vi.fn(),
	createdSources: [] as string[],
}))

vi.mock('@/stores/auth', () => ({
	useAuthStore: () => ({settings: {frontendSettings: settings}}),
}))

class FakeAudio {
	constructor(src: string) {
		createdSources.push(src)
	}

	play = playMock
}

// Stands in for the promise play() returns and records whether anyone subscribed to its rejection.
function watchedPlayPromise() {
	const state = {rejectionHandled: false}
	const promise: PromiseLike<void> & {catch: unknown, finally: unknown} = {
		then(_onFulfilled?: unknown, onRejected?: unknown) {
			state.rejectionHandled ||= typeof onRejected === 'function'
			return promise
		},
		catch(onRejected?: unknown) {
			state.rejectionHandled ||= typeof onRejected === 'function'
			return promise
		},
		finally() {
			return promise
		},
	} as PromiseLike<void> & {catch: unknown, finally: unknown}

	return {promise, state}
}

describe('playPopSound', () => {
	beforeEach(() => {
		settings.playSoundWhenDone = true
		createdSources.length = 0
		playMock.mockReset()
		playMock.mockResolvedValue(undefined)
		vi.stubGlobal('Audio', FakeAudio)
	})

	afterEach(() => {
		vi.unstubAllGlobals()
	})

	it('plays the sound when the setting is enabled', async () => {
		const {playPopSound} = await import('./playPop')

		playPopSound()

		expect(createdSources).toHaveLength(1)
		expect(playMock).toHaveBeenCalledOnce()
	})

	it('does not play the sound when the setting is disabled', async () => {
		settings.playSoundWhenDone = false
		const {playPopSound} = await import('./playPop')

		playPopSound()

		expect(createdSources).toHaveLength(0)
		expect(playMock).not.toHaveBeenCalled()
	})

	// An unhandled AbortError or NotAllowedError from play() would be reported to Sentry.
	it('subscribes to the play() rejection', async () => {
		const {promise, state} = watchedPlayPromise()
		playMock.mockReturnValue(promise)
		const {playPopSound} = await import('./playPop')

		playPopSound()

		expect(state.rejectionHandled).toBe(true)
	})
})
