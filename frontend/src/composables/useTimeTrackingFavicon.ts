import {watch} from 'vue'
import {createSharedComposable, tryOnMounted} from '@vueuse/core'
import {storeToRefs} from 'pinia'

import {useTimeTrackingStore} from '@/stores/timeTracking'

const FAVICON_SIZE = 32
// Drawn from a PNG rather than the .ico because ICO decoding into a canvas is
// unreliable across browsers.
const BASE_FAVICON = '/images/icons/favicon-32x32.png'
const DOT_COLOR = '#ff4136'

function getFaviconLink(): HTMLLinkElement | null {
	return document.querySelector<HTMLLinkElement>('link[rel="icon"]')
}

// Marks the favicon with a small red dot in the lower left corner while a timer
// is running, so an active time tracking session is visible even when the tab
// isn't focused.
export const useTimeTrackingFavicon = createSharedComposable(() => {
	const {hasActiveTimer} = storeToRefs(useTimeTrackingStore())

	const link = getFaviconLink()
	const originalHref = link?.getAttribute('href') ?? '/favicon.ico'

	let baseImage: HTMLImageElement | null = null

	function loadBaseImage(): Promise<HTMLImageElement> {
		if (baseImage?.complete) {
			return Promise.resolve(baseImage)
		}
		return new Promise((resolve, reject) => {
			const img = new Image()
			img.addEventListener('load', () => {
				baseImage = img
				resolve(img)
			})
			img.addEventListener('error', reject)
			img.src = BASE_FAVICON
		})
	}

	async function drawBadgedFavicon() {
		const targetLink = getFaviconLink()
		if (targetLink === null) {
			return
		}

		const img = await loadBaseImage()
		const canvas = document.createElement('canvas')
		canvas.width = FAVICON_SIZE
		canvas.height = FAVICON_SIZE
		const ctx = canvas.getContext('2d')
		if (ctx === null) {
			return
		}

		ctx.drawImage(img, 0, 0, FAVICON_SIZE, FAVICON_SIZE)

		const radius = FAVICON_SIZE * 0.28
		const cx = radius
		const cy = FAVICON_SIZE - radius
		ctx.beginPath()
		ctx.arc(cx, cy, radius, 0, Math.PI * 2)
		ctx.fillStyle = DOT_COLOR
		ctx.fill()

		targetLink.href = canvas.toDataURL('image/png')
	}

	function restoreFavicon() {
		const targetLink = getFaviconLink()
		if (targetLink !== null) {
			targetLink.href = originalHref
		}
	}

	function update(active: boolean) {
		if (active) {
			void drawBadgedFavicon()
			return
		}
		restoreFavicon()
	}

	watch(hasActiveTimer, update, {flush: 'post'})
	tryOnMounted(() => update(hasActiveTimer.value))
})
