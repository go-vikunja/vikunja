import type { Meta, StoryObj } from '@storybook/vue3'
import { ref } from 'vue'
import FancyCheckbox from './FancyCheckbox.vue'

const meta: Meta<typeof FancyCheckbox> = {
    title: 'Input/FancyCheckbox',
    component: FancyCheckbox,
}
export default meta

type Story = StoryObj<typeof FancyCheckbox>

export const Default: Story = {
    render: () => ({
        components: { FancyCheckbox },
        setup() {
            const isChecked = ref(false)
            return { isChecked }
        },
        template: '<FancyCheckbox v-model="isChecked">This is probably not important</FancyCheckbox>\n<input v-model="isChecked" type="checkbox"> {{ isChecked }}',
    }),
}

export const EnabledInitially: Story = {
    render: () => ({
        components: { FancyCheckbox },
        setup() {
            const isCheckedInitiallyEnabled = ref(true)
            return { isCheckedInitiallyEnabled }
        },
        template: '<FancyCheckbox v-model="isCheckedInitiallyEnabled">We want you to use this option</FancyCheckbox>\n<input v-model="isCheckedInitiallyEnabled" type="checkbox"> {{ isCheckedInitiallyEnabled }}',
    }),
}

export const Disabled: Story = {
    render: () => ({
        components: { FancyCheckbox },
        setup() {
            const isCheckedDisabled = ref(false)
            return { isCheckedDisabled }
        },
        template: '<FancyCheckbox disabled :model-value="isCheckedDisabled" />' ,
    }),
}

export const UndefinedInitial: Story = {
    render: () => ({
        components: { FancyCheckbox },
        setup() {
            const withoutInitialState = ref(undefined)
            return { withoutInitialState }
        },
        template: '<FancyCheckbox v-model="withoutInitialState">Not sure what the value should be</FancyCheckbox>\n<input v-model="withoutInitialState" type="checkbox" disabled> {{ withoutInitialState }}',
    }),
}

export const Overview: Story = {
    render: () => ({
        components: { FancyCheckbox },
        setup() {
            const model = ref(false)
            return { model }
        },
        template: `
            <div>
                <FancyCheckbox v-model="model">Default</FancyCheckbox>
                <FancyCheckbox v-model="model" disabled class="ml-2">Disabled</FancyCheckbox>
            </div>
        `,
    }),
}
