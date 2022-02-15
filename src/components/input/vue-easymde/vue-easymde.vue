<template>
  <div class="vue-easymde">
    <textarea
      class="vue-simplemde-textarea"
      :name="name"
      :value="modelValue"
      @input="handleInput($event.target.value)"
    />
  </div>
</template>

<script lang="ts">
import EasyMDE from 'easymde'
import {marked} from 'marked'

export default {
  name: 'vue-easymde',
  props: {
    modelValue: String,
    name: String,
    previewClass: String,
    autoinit: {
      type: Boolean,
      default: true,
    },
    highlight: {
      type: Boolean,
      default: false,
    },
    sanitize: {
      type: Boolean,
      default: false,
    },
    configs: {
      type: Object,
      default() {
        return {}
      },
    },
    previewRender: {
      type: Function,
    },
  },
	emits: ['update:modelValue', 'blur', 'initialized'],
  data() {
    return {
      isValueUpdateFromInner: false,
      easymde: null,
    }
  },
  mounted() {
    if (this.autoinit) this.initialize()
  },
  deactivated() {
    const editor = this.easymde
    if (!editor) return
    const isFullScreen = editor.codemirror.getOption('fullScreen')
    if (isFullScreen) editor.toggleFullScreen()
  },
	beforeUnmount() {
		if (this.easymde) {
			this.easymde.toTextArea()
			this.easymde.cleanup()
			this.easymde = null
		}
	},
  methods: {
    initialize() {
      const configs = Object.assign({
        element: this.$el.firstElementChild,
        initialValue: this.modelValue,
        previewRender: this.previewRender,
        renderingConfig: {},
      }, this.configs)

      // Synchronize the values of value and initialValue
      if (configs.initialValue) {
        this.$emit('update:modelValue', configs.initialValue)
      }

      // Determine whether to enable code highlighting
      if (this.highlight) {
        configs.renderingConfig.codeSyntaxHighlighting = true
      }

      // Set whether to render the input html
      marked.setOptions({ sanitize: this.sanitize })

      // Instantiated editor
      this.easymde = new EasyMDE(configs)

      // Add a custom previewClass
      const className = this.previewClass || ''
      this.addPreviewClass(className)

      // Binding event
      this.bindingEvents()

      this.$nextTick(() => {
        this.$emit('initialized', this.easymde)
      })
    },

    addPreviewClass(className) {
      const wrapper = this.easymde.codemirror.getWrapperElement()
      const preview = document.createElement('div')
      wrapper.nextSibling.className += ` ${className}`
      preview.className = `editor-preview ${className}`
      wrapper.appendChild(preview)
    },

		bindingEvents() {
      this.easymde.codemirror.on('change', this.handleCodemirrorInput)
      this.easymde.codemirror.on('blur', this.handleCodemirrorBlur)
    },

    handleCodemirrorInput(instance, changeObj)  {
			if (changeObj.origin === 'setValue') {
				return
			}
			const val = this.easymde.value()
			this.handleInput(val)
		},

    handleCodemirrorBlur() {
			const val = this.easymde.value()
      this.isValueUpdateFromInner = true
      this.$emit('blur', val)
    },

    handleInput(val) {
      this.isValueUpdateFromInner = true
      this.$emit('update:modelValue', val)
    },
  },

  watch: {
    modelValue(val) {
      if (this.isValueUpdateFromInner) {
        this.isValueUpdateFromInner = false
      } else {
        this.easymde.value(val)
      }
    },
  },
}
</script>

<style lang="scss" scoped>
.vue-easymde .markdown-body {
padding: 0.5em
}

.vue-easymde .editor-preview-active, .vue-easymde .editor-preview-active-side {
display: block;
}
</style>
