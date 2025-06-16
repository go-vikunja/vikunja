import type { Meta, StoryObj } from '@storybook/vue3'
import LinkSharing from './LinkSharing.vue'
import LinkShareService from '@/services/linkShare'
import LinkShareModel from '@/models/linkShare'

const meta: Meta<typeof LinkSharing> = {
    title: 'Sharing/LinkSharing',
    component: LinkSharing,
}
export default meta

type Story = StoryObj<typeof LinkSharing>

export const Default: Story = {
    render: () => ({
        components: { LinkSharing },
        setup() {
            // stub service
            LinkShareService.prototype.getAll = async () => [
                new LinkShareModel({ id: 1, name: 'Public Link' }),
            ]
            LinkShareService.prototype.create = async () => new LinkShareModel({ id: 2 })
            LinkShareService.prototype.delete = async () => {}
            return {}
        },
        template: '<LinkSharing project-id="1" />',
    }),
}
