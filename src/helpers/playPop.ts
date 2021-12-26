import popSoundFile from '@/assets/audio/pop.mp3'

export const playSoundWhenDoneKey = 'playSoundWhenTaskDone'

export function playPop() {
	const enabled = Boolean(localStorage.getItem(playSoundWhenDoneKey))
	if(!enabled) {
		return
	}

	const popSound = new Audio(popSoundFile)
	popSound.play()
}