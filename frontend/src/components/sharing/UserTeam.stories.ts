import type { Meta, StoryObj } from '@storybook/vue3'
import UserTeam from './UserTeam.vue'
import UserProjectService from '@/services/userProject'
import UserProjectModel from '@/models/userProject'
import UserModel from '@/models/user'

const meta: Meta<typeof UserTeam> = {
    title: 'Sharing/UserTeam',
    component: UserTeam,
}
export default meta

type Story = StoryObj<typeof UserTeam>

export const Default: Story = {
    render: () => ({
        components: { UserTeam },
        setup() {
            UserProjectService.prototype.getAll = async () => [
                new UserModel({ id: 1, username: 'tester' }),
            ]
            UserProjectService.prototype.create = async () => new UserProjectModel({})
            UserProjectService.prototype.delete = async () => {}
            return {}
        },
        template: '<UserTeam share-type="user" type="project" :id="1" />',
    }),
}
