import {useAuthStore} from '@/stores/auth'

import popSoundFile from '@/assets/audio/pop.mp3'

export function playPopSound() {
	const playSoundWhenDone = useAuthStore().settings.frontendSettings.playSoundWhenDone

	if (!playSoundWhenDone)
		return

	try {
		const popSound = new Audio(popSoundFile)
		// play() rejects when the browser blocks playback without a user gesture or aborts the load.
		// Nothing the user can act on, and an unhandled rejection would be reported to Sentry.
		popSound.play().catch(() => {})
	} catch (e) {
		console.error('Could not play pop sound:', e)
	}
}
