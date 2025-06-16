import type { Meta, StoryObj } from '@storybook/vue3'
import AssigneeList from './AssigneeList.vue'
import UserModel from '@/models/user'

const meta: Meta<typeof AssigneeList> = {
    title: 'Tasks/AssigneeList',
    component: AssigneeList,
}
export default meta

type Story = StoryObj<typeof AssigneeList>

export const Default: Story = {
    render: () => ({
        components: { AssigneeList },
        setup() {
            const assignees = [
                new UserModel({ id: 1, username: 'alice' }),
                new UserModel({ id: 2, username: 'bob' }),
            ]
            return { assignees }
        },
        template: '<AssigneeList :assignees="assignees" />',
    }),
}
