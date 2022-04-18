export const isAppleDevice = (): boolean => {
	return navigator.userAgent.includes('Mac') || [
		'iPad Simulator',
		'iPhone Simulator',
		'iPod Simulator',
		'iPad',
		'iPhone',
		'iPod',
	].includes(navigator.platform)
}
