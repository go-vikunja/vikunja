import type { Meta, StoryObj } from '@storybook/vue3'
import ProjectCard from './ProjectCard.vue'
import ProjectModel from '@/models/project'

const meta: Meta<typeof ProjectCard> = {
    title: 'Project/ProjectCard',
    component: ProjectCard,
}
export default meta

type Story = StoryObj<typeof ProjectCard>

export const Default: Story = {
    render: () => ({
        components: { ProjectCard },
        setup() {
            const project = new ProjectModel({ id: 1, title: 'Storybook Project' })
            return { project }
        },
        template: '<ProjectCard :project="project" />',
    }),
}
