import popSoundFile from '@/assets/audio/pop.mp3'

export const playSoundWhenDoneKey = 'playSoundWhenTaskDone'

export function playPopSound() {
	try {
		const popSound = new Audio(popSoundFile)
		popSound.play()
	} catch (e) {
		console.error('Could not play pop sound:', e)
	}
}
