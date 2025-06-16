import type { Meta, StoryObj } from '@storybook/vue3'
import ProjectCardGrid from './ProjectCardGrid.vue'
import ProjectModel from '@/models/project'

const meta: Meta<typeof ProjectCardGrid> = {
    title: 'Project/ProjectCardGrid',
    component: ProjectCardGrid,
}
export default meta

type Story = StoryObj<typeof ProjectCardGrid>

export const Default: Story = {
    render: () => ({
        components: { ProjectCardGrid },
        setup() {
            const projects = [
                new ProjectModel({ id: 1, title: 'One' }),
                new ProjectModel({ id: 2, title: 'Two' }),
            ]
            return { projects }
        },
        template: '<ProjectCardGrid :projects="projects" />',
    }),
}
