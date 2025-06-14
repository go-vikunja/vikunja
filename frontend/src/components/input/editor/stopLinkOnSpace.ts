import {Extension} from '@tiptap/core'

export default Extension.create({
    name: 'stopLinkOnSpace',

    addKeyboardShortcuts() {
        return {
            Space: ({editor}) => {
                if (editor.isActive('link')) {
                    editor.commands.unsetLink()
                }
                return false
            },
        }
    },
})
