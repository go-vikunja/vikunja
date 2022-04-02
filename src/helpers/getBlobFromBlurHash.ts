import {decode} from 'blurhash'

export async function getBlobFromBlurHash(blurHash: string): Promise<Blob | null> {
	if (blurHash === '') {
		return null
	}

	const pixels = decode(blurHash, 32, 32)
	const canvas = document.createElement('canvas')
	canvas.width = 32
	canvas.height = 32
	const ctx = canvas.getContext('2d')
	if (ctx === null) {
		return null
	}
	
	const imageData = ctx.createImageData(32, 32)
	imageData.data.set(pixels)
	ctx.putImageData(imageData, 0, 0)

	return new Promise<Blob>((resolve, reject) => {
		canvas.toBlob(b => {
			if (b === null) {
				reject(b)
				return
			}

			resolve(b)
		})
	})
}
