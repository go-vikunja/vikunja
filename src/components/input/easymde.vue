<template>
    <!-- TODO: Fix the icons -->
    <vue-easymde v-model="text" :configs="config" @change="bubble"/>
</template>

<script>
    import VueEasymde from 'vue-easymde'
    import EasyMDE from 'easymde'

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
                iconFillColor: '#000',
                config: {
                    autoDownloadFontAwesome: false,
                    spellChecker: false,
                    placeholder: 'Click here to enter a description...',
                    toolbar: [
                        {
                            name: 'heading-1',
                            action: EasyMDE.toggleHeading1,
                            title: 'Heading 1',
                            icon: '<svg width="24" height="24"><rect width="24" height="24" fill="none" rx="0" ry="0"/><path fill-rule="evenodd" clip-rule="evenodd" d="M18.2773 19.25L12.5773 4.34995C12.5773 4.34995 12.5773 4.24995 12.4773 4.24995C12.4773 4.24995 12.4773 4.14995 12.3773 4.14995C12.3773 4.14995 12.2773 4.14995 12.2773 4.04995L12.1773 3.94995H12.0773H11.9773C11.8773 3.94995 11.8773 3.94995 11.8773 3.94995H11.7773C11.6773 4.04995 11.6773 4.14995 11.5773 4.14995C11.5773 4.14995 11.5773 4.14995 11.4773 4.14995C11.4773 4.14995 11.4773 4.24995 11.3773 4.24995L11.2773 4.34995L5.67733 19.25C5.57733 19.55 5.67733 19.95 5.97733 20.05C6.07733 20.05 6.07733 20.05 6.17733 20.05C6.37733 20.05 6.67733 19.95 6.77733 19.65L7.87733 16.85H16.1773L17.2773 19.65C17.3773 19.85 17.5773 20.05 17.8773 20.05C17.9773 20.05 17.9773 20.05 18.0773 20.05C18.2773 19.85 18.4773 19.55 18.2773 19.25ZM8.27733 15.65L11.9773 6.24995L15.6773 15.65H8.27733Z" fill="'+ this.iconFillColor +'"/></svg>',
                        },
                        {
                            name: 'heading-2',
                            action: EasyMDE.toggleHeading2,
                            title: 'Heading 2',
                            icon: '<svg width="24" height="24"><rect width="24" height="24" fill="none" rx="0" ry="0"/><path fill-rule="evenodd" clip-rule="evenodd" d="M18.2773 19.25L12.5773 4.34995C12.5773 4.34995 12.5773 4.24995 12.4773 4.24995C12.4773 4.24995 12.4773 4.14995 12.3773 4.14995C12.3773 4.14995 12.2773 4.14995 12.2773 4.04995L12.1773 3.94995H12.0773H11.9773C11.8773 3.94995 11.8773 3.94995 11.8773 3.94995H11.7773C11.6773 4.04995 11.6773 4.14995 11.5773 4.14995C11.5773 4.14995 11.5773 4.14995 11.4773 4.14995C11.4773 4.14995 11.4773 4.24995 11.3773 4.24995L11.2773 4.34995L5.67733 19.25C5.57733 19.55 5.67733 19.95 5.97733 20.05C6.07733 20.05 6.07733 20.05 6.17733 20.05C6.37733 20.05 6.67733 19.95 6.77733 19.65L7.87733 16.85H16.1773L17.2773 19.65C17.3773 19.85 17.5773 20.05 17.8773 20.05C17.9773 20.05 17.9773 20.05 18.0773 20.05C18.2773 19.85 18.4773 19.55 18.2773 19.25ZM8.27733 15.65L11.9773 6.24995L15.6773 15.65H8.27733Z" fill="'+ this.iconFillColor +'"/></svg>',
                        },
                        {
                            name: 'heading-3',
                            action: EasyMDE.toggleHeading3,
                            title: 'Heading 3',
                            icon: '<svg width="24" height="24"><rect width="24" height="24" fill="none" rx="0" ry="0"/><path fill-rule="evenodd" clip-rule="evenodd" d="M18.2773 19.25L12.5773 4.34995C12.5773 4.34995 12.5773 4.24995 12.4773 4.24995C12.4773 4.24995 12.4773 4.14995 12.3773 4.14995C12.3773 4.14995 12.2773 4.14995 12.2773 4.04995L12.1773 3.94995H12.0773H11.9773C11.8773 3.94995 11.8773 3.94995 11.8773 3.94995H11.7773C11.6773 4.04995 11.6773 4.14995 11.5773 4.14995C11.5773 4.14995 11.5773 4.14995 11.4773 4.14995C11.4773 4.14995 11.4773 4.24995 11.3773 4.24995L11.2773 4.34995L5.67733 19.25C5.57733 19.55 5.67733 19.95 5.97733 20.05C6.07733 20.05 6.07733 20.05 6.17733 20.05C6.37733 20.05 6.67733 19.95 6.77733 19.65L7.87733 16.85H16.1773L17.2773 19.65C17.3773 19.85 17.5773 20.05 17.8773 20.05C17.9773 20.05 17.9773 20.05 18.0773 20.05C18.2773 19.85 18.4773 19.55 18.2773 19.25ZM8.27733 15.65L11.9773 6.24995L15.6773 15.65H8.27733Z" fill="'+ this.iconFillColor +'"/></svg>',
                        },
                        {
                            name: 'heading-smaller',
                            action: EasyMDE.toggleHeadingSmaller,
                            title: 'Heading Smaller',
                            icon: '<svg width="24" height="24"><rect width="24" height="24" fill="none" rx="0" ry="0"/><path fill-rule="evenodd" clip-rule="evenodd" d="M18.2773 19.25L12.5773 4.34995C12.5773 4.34995 12.5773 4.24995 12.4773 4.24995C12.4773 4.24995 12.4773 4.14995 12.3773 4.14995C12.3773 4.14995 12.2773 4.14995 12.2773 4.04995L12.1773 3.94995H12.0773H11.9773C11.8773 3.94995 11.8773 3.94995 11.8773 3.94995H11.7773C11.6773 4.04995 11.6773 4.14995 11.5773 4.14995C11.5773 4.14995 11.5773 4.14995 11.4773 4.14995C11.4773 4.14995 11.4773 4.24995 11.3773 4.24995L11.2773 4.34995L5.67733 19.25C5.57733 19.55 5.67733 19.95 5.97733 20.05C6.07733 20.05 6.07733 20.05 6.17733 20.05C6.37733 20.05 6.67733 19.95 6.77733 19.65L7.87733 16.85H16.1773L17.2773 19.65C17.3773 19.85 17.5773 20.05 17.8773 20.05C17.9773 20.05 17.9773 20.05 18.0773 20.05C18.2773 19.85 18.4773 19.55 18.2773 19.25ZM8.27733 15.65L11.9773 6.24995L15.6773 15.65H8.27733Z" fill="'+ this.iconFillColor +'"/></svg>',
                        },
                        {
                            name: 'heading-bigger',
                            action: EasyMDE.toggleHeadingBigger,
                            title: 'Heading Bigger',
                            icon: '<svg width="24" height="24"><rect width="24" height="24" fill="none" rx="0" ry="0"/><path fill-rule="evenodd" clip-rule="evenodd" d="M18.2773 19.25L12.5773 4.34995C12.5773 4.34995 12.5773 4.24995 12.4773 4.24995C12.4773 4.24995 12.4773 4.14995 12.3773 4.14995C12.3773 4.14995 12.2773 4.14995 12.2773 4.04995L12.1773 3.94995H12.0773H11.9773C11.8773 3.94995 11.8773 3.94995 11.8773 3.94995H11.7773C11.6773 4.04995 11.6773 4.14995 11.5773 4.14995C11.5773 4.14995 11.5773 4.14995 11.4773 4.14995C11.4773 4.14995 11.4773 4.24995 11.3773 4.24995L11.2773 4.34995L5.67733 19.25C5.57733 19.55 5.67733 19.95 5.97733 20.05C6.07733 20.05 6.07733 20.05 6.17733 20.05C6.37733 20.05 6.67733 19.95 6.77733 19.65L7.87733 16.85H16.1773L17.2773 19.65C17.3773 19.85 17.5773 20.05 17.8773 20.05C17.9773 20.05 17.9773 20.05 18.0773 20.05C18.2773 19.85 18.4773 19.55 18.2773 19.25ZM8.27733 15.65L11.9773 6.24995L15.6773 15.65H8.27733Z" fill="'+ this.iconFillColor +'"/></svg>',
                        },
                        '|',
                        {
                            name: 'bold',
                            action: EasyMDE.toggleBold,
                            title: 'Bold',
                            icon: '<svg width="24" height="24"><rect width="24" height="24" fill="none" rx="0" ry="0"/><path fill-rule="evenodd" clip-rule="evenodd" d="M3.5 3H6.5H15.25C18.15 3 20.5 5.36 20.5 8.25C20.5 9.8 19.81 11.19 18.73 12.15C20.37 13.04 21.5 14.76 21.5 16.75C21.5 19.64 19.15 22 16.25 22H6.5H3.5C2.95 22 2.5 21.55 2.5 21C2.5 20.45 2.95 20 3.5 20H5.5V5H3.5C2.95 5 2.5 4.55 2.5 4C2.5 3.45 2.95 3 3.5 3ZM7.5 20H16.25C18.04 20 19.5 18.54 19.5 16.75C19.5 14.96 18.04 13.5 16.25 13.5H7.5V20ZM7.5 11.5H15.25C17.04 11.5 18.5 10.04 18.5 8.25C18.5 6.46 17.04 5 15.25 5H7.5V11.5Z" fill="'+ this.iconFillColor +'"/></svg>',
                        },
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
