<template>
    <!-- TODO: Fix the icons -->
    <vue-easymde v-model="text" :configs="config" @change="bubble"/>
</template>

<script>
    import VueEasymde from 'vue-easymde'

    export default {
        name: 'easymde',
        components: {
            VueEasymde
        },
        props: {
            value: {
                type: String,
                default: '',
            },
        },
        data() {
            return {
                text: '',
                config: {
                    autoDownloadFontAwesome: false,
                    spellChecker: false,
                    placeholder: 'Click here to enter a description...',
                    toolbar: [
                        'heading-1',
                        'heading-2',
                        'heading-3',
                        'heading-smaller',
                        'heading-bigger',
                        '|',
                        'bold',
                        'italic',
                        'strikethrough',
                        'code',
                        'quote',
                        'unordered-list',
                        'ordered-list',
                        '|',
                        'clean-block',
                        'link',
                        'image',
                        'table',
                        'horizontal-rule',
                        '|',
                        'preview',
                        'side-by-side',
                        'fullscreen',
                        'guide',
//                        {
//                            name: 'bold',
//                            title: 'Bold',
//                            iconElement: '<span>test</span>' // This relies on an extra thing added in node_modules/easymde/src/js/easymde.js:145
//                        },
                    ]
                },
            }
        },
        watch: {
            value(newVal) {
                this.text = newVal
            },
            text() {
                this.bubble()
            }
        },
        methods: {
            bubble() {
                this.$emit('input', this.text)
                this.$emit('change')
            }
        },
    }
</script>

<style lang="scss">
    @import '../../../node_modules/easymde/dist/easymde.min.css';

    .CodeMirror {
        padding: 0;
    }

    .CodeMirror-scroll {
        padding: .5em;
    }

    .editor-toolbar {
        background: #ffffff;
    }

    pre.CodeMirror-line{
        margin-bottom: 0 !important;
    }
</style>
