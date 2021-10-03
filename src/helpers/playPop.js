export const playSoundWhenDoneKey = 'playSoundWhenTaskDone'

export const playPop = () => {
	const enabled = localStorage.getItem(playSoundWhenDoneKey) === 'true' || localStorage.getItem(playSoundWhenDoneKey) === null
	if(!enabled) {
		return
	}

	const popSound = new Audio('/audio/pop.mp3')
	popSound.play()
}