import { parseValue, decodeKey, encodeKey } from './helpers'

let mapFunctions = {}
let objAvoided = []
let elementAvoided = []
let keyPressed = false

function dispatchShortkeyEvent(pKey) {
  const e = new CustomEvent('shortkey', { bubbles: false })

  if (mapFunctions[pKey].key) {
    e.srcKey = mapFunctions[pKey].key
  }

  const elm = mapFunctions[pKey].el

  if (!mapFunctions[pKey].propagte) {
    elm[elm.length - 1].dispatchEvent(e)
  } else {
    elm.forEach((elmItem) => elmItem.dispatchEvent(e))
  }
}

function keyDown(pKey) {
  if (
    (!mapFunctions[pKey].once && !mapFunctions[pKey].push) ||
    (mapFunctions[pKey].push && !keyPressed)
  ) {
    dispatchShortkeyEvent(pKey)
  }
}

function fillMappingFunctions(
  mappingFunctions,
  { b, push, once, focus, propagte, el },
) {
  for (let key in b) {
    const k = encodeKey(b[key])
    const propagated = mappingFunctions[k] && mappingFunctions[k].propagte
    const elm =
      mappingFunctions[k] && mappingFunctions[k].el
        ? mappingFunctions[k].el
        : []

    elm.push(el)

    mappingFunctions[k] = {
      push,
      once,
      focus,
      key,
      propagte: propagated || propagte,
      el: elm,
    }
  }
}

function bindValue(value, el, binding, vnode) {
  const { modifiers } = binding
  const push = !!modifiers.push
  const avoid = !!modifiers.avoid
  const focus = !modifiers.focus
  const once = !!modifiers.once
  const propagte = !!modifiers.propagte

  if (avoid) {
    objAvoided = objAvoided.filter((itm) => !itm === el)
    objAvoided.push(el)
  } else {
    fillMappingFunctions(mapFunctions, {
      b: value,
      push,
      once,
      focus,
      propagte,
      el: vnode.el,
    })
  }
}

function unbindValue(value, el) {
  for (let key in value) {
    const k = encodeKey(value[key])
    const idxElm = mapFunctions[k].el.indexOf(el)

    if (mapFunctions[k].el.length > 1 && idxElm > -1) {
      mapFunctions[k].el.splice(idxElm, 1)
    } else {
      delete mapFunctions[k]
    }
  }
}

function availableElement(decodedKey) {
  const objectIsAvoided = !!objAvoided.find(
    (r) => r === document.activeElement,
  )
  const filterAvoided = !!elementAvoided.find(
    (selector) =>
      document.activeElement && document.activeElement.matches(selector),
  )
  return !!mapFunctions[decodedKey] && !(objectIsAvoided || filterAvoided)
}

function keyDownListener(pKey) {
  const decodedKey = decodeKey(pKey)

  // Check avoidable elements
  if (!availableElement(decodedKey)) {
    return
  }

  if (!mapFunctions[decodedKey].propagte) {
    pKey.preventDefault()
    pKey.stopPropagation()
  }

  if (mapFunctions[decodedKey].focus) {
    keyDown(decodedKey)
    keyPressed = true
  } else if (!keyPressed) {
    const elm = mapFunctions[decodedKey].el
    elm[elm.length - 1].focus()
    keyPressed = true
  }
}

function keyUpListener(pKey) {
  const decodedKey = decodeKey(pKey)

  if (!availableElement(decodedKey)) {
    keyPressed = false
    return
  }

  if (!mapFunctions[decodedKey].propagte) {
    pKey.preventDefault()
    pKey.stopPropagation()
  }

  if (mapFunctions[decodedKey].once || mapFunctions[decodedKey].push) {
    dispatchShortkeyEvent(decodedKey)
  }

  keyPressed = false
}

// register key presses that happen before mounting of directive
// if (process?.env?.NODE_ENV !== 'test') {
//   (() => {
    document.addEventListener('keydown', keyDownListener, true)
    document.addEventListener('keyup', keyUpListener, true)
  // })()
// }

function install(app, options) {
  elementAvoided = [...(options && options.prevent ? options.prevent : [])]

  app.directive('shortkey', {
    beforeMount(el, binding, vnode) {
      // Mapping the commands
      const value = parseValue(binding.value)
      bindValue(value, el, binding, vnode)
    },

    updated(el, binding, vnode) {
      const oldValue = parseValue(binding.oldValue)
      unbindValue(oldValue, el)

      const newValue = parseValue(binding.value)
      bindValue(newValue, el, binding, vnode)
    },

    unmounted(el, binding) {
      const value = parseValue(binding.value)
      unbindValue(value, el)
    },
  })
}

export default {
  install,
  encodeKey,
  decodeKey,
  keyDown,
}
