import popSoundFile from '@/assets/audio/pop.mp3'

export const playSoundWhenDoneKey = 'playSoundWhenTaskDone'

export function playPop() {
	const enabled = localStorage.getItem(playSoundWhenDoneKey) === 'true'
	if (!enabled) {
		return
	}

	playPopSound()
}

export function playPopSound() {
	const popSound = new Audio(popSoundFile)
	popSound.play()
}
