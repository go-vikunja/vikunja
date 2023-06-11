import popSoundFile from '@/assets/audio/pop.mp3'

export const playSoundWhenDoneKey = 'playSoundWhenTaskDone'

export function playPopSound() {
	const popSound = new Audio(popSoundFile)
	popSound.play()
}
