import {useAuthStore} from '@/stores/auth'

import popSoundFile from '@/assets/audio/pop.mp3'

export function playPopSound() {
	const playSoundWhenDone = useAuthStore().settings.frontendSettings.playSoundWhenDone

	if (!playSoundWhenDone)
		return

	try {
		const popSound = new Audio(popSoundFile)
		popSound.play()
	} catch (e) {
		console.error('Could not play pop sound:', e)
	}
}
