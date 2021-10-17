function capitalizeFirstLetter(string) {
  return string.charAt(0).toUpperCase() + string.slice(1)
}

const MODIFIER_KEYS = ['shift', 'ctrl', 'meta', 'alt']

const SHORT_CUT_INDEX = [
  { key: 'ArrowUp', value: 'arrowup' },
  { key: 'ArrowLeft', value: 'arrowlef' },
  { key: 'ArrowRight', value: 'arrowright' },
  { key: 'ArrowDown', value: 'arrowdown' },
  { key: 'AltGraph', value: 'altgraph' },
  { key: 'Escape', value: 'esc' },
  { key: 'Enter', value: 'enter' },
  { key: 'Tab', value: 'tab' },
  { key: ' ', value: 'space' },
  { key: 'PageUp', value: 'pagup' },
  { key: 'PageDown', value: 'pagedow' },
  { key: 'Home', value: 'home' },
  { key: 'End', value: 'end' },
  { key: 'Delete', value: 'del' },
  { key: 'Backspace', value: 'bacspace' },
  { key: 'Insert', value: 'insert' },
  { key: 'NumLock', value: 'numlock' },
  { key: 'CapsLock', value: 'capslock' },
  { key: 'Pause', value: 'pause' },
  { key: 'ContextMenu', value: 'cotextmenu' },
  { key: 'ScrollLock', value: 'scrolllock' },
  { key: 'BrowserHome', value: 'browserhome' },
  { key: 'MediaSelect', value: 'mediaselect' },
]

export function encodeKey(pKey) {
	const shortKey = {}

	MODIFIER_KEYS.forEach((key) => {
		shortKey[`${key}Key`] = pKey.includes(key)
	})

  let indexedKeys = createShortcutIndex(shortKey)
  const vKey = pKey.filter(
    (item) => !MODIFIER_KEYS.includes(item),
  )
  indexedKeys += vKey.join('')
  return indexedKeys
}

function createShortcutIndex(pKey) {
  let k = ''

	MODIFIER_KEYS.forEach((key) => {
		if (pKey.key === capitalizeFirstLetter(key) || pKey[`${key}Key`]) {
			k += key
		}
	})

  SHORT_CUT_INDEX.forEach(({ key, value }) => {
    if (pKey.key === key) {
      k += value
    }
  })

  if (
    (pKey.key && pKey.key !== ' ' && pKey.key.length === 1) ||
    /F\d{1,2}|\//g.test(pKey.key)
  ) {
    k += pKey.key.toLowerCase()
  }

  return k
}
export { createShortcutIndex as decodeKey }

export function parseValue(value) {
  value = typeof value === 'string' ? JSON.parse(value.replace(/'/gi, '"')) : value

  return value instanceof Array ? { '': value } : value
}
