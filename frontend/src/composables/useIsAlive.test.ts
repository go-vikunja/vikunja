import {effectScope} from 'vue'
import {describe, expect, it} from 'vitest'

import {useIsAlive} from './useIsAlive'

describe('useIsAlive', () => {
	it('flips to false once the owning scope is disposed', () => {
		const scope = effectScope()
		const alive = scope.run(() => useIsAlive())!

		expect(alive.value).toBe(true)
		scope.stop()
		expect(alive.value).toBe(false)
	})
})
