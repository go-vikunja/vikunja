import {useAuthStore} from '@/stores/auth'

import popSoundFile from '@/assets/audio/pop.mp3'

export function playPopSound() {
	const play_sound_when_done = useAuthStore().settings.frontend_settings.play_sound_when_done

	if (!play_sound_when_done)
		return

	try {
		const popSound = new Audio(popSoundFile)
		popSound.play()
	} catch (e) {
		console.error('Could not play pop sound:', e)
	}
}
