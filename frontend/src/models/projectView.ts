import type {IProjectView, ProjectViewBucketConfigurationMode, ProjectViewKind} from '@/modelTypes/IProjectView'
import AbstractModel from '@/models/abstractModel'

export default class ProjectViewModel extends AbstractModel<IProjectView> implements IProjectView {
	id = 0
	title = ''
	projectId = 0
	viewKind: ProjectViewKind =  'list'

	filter: IProjectView['filter'] = {
		sort_by: [],
		order_by: [],
		filter: '',
		filter_include_nulls: true,
		s: '',
	}
	position = 0
	
	bucketConfiguration = []
	bucketConfigurationMode: ProjectViewBucketConfigurationMode = 'manual'
	defaultBucketId = 0
	doneBucketId = 0

	created: Date = new Date()
	updated: Date = new Date()

	constructor(data: Partial<IProjectView> | Record<string, unknown>) {
		super()
		this.assignData(data)

		// Convert numeric view_kind from API to string values for frontend
		const rawData = data as Record<string, unknown>
		if (typeof data.viewKind === 'number' || typeof rawData.view_kind === 'number') {
			const numericViewKind = data.viewKind || rawData.view_kind as number
			switch (numericViewKind) {
				case 0:
					this.viewKind = 'list'
					break
				case 1:
					this.viewKind = 'gantt'
					break
				case 2:
					this.viewKind = 'table'
					break
				case 3:
					this.viewKind = 'kanban'
					break
				default:
					this.viewKind = 'list'
			}
		}

		if (!this.bucketConfiguration) {
			this.bucketConfiguration = []
		}
	}

	static createWithDefaultFilter(data: Partial<IProjectView> = {}): ProjectViewModel {
		const defaultFilter: IProjectView['filter'] = {
			sort_by: ['done', 'id'],
			order_by: ['asc', 'desc'],
			filter: 'done = false',
			filter_include_nulls: true,
			s: '',
		}

		const instance = new ProjectViewModel(data)
		instance.filter = defaultFilter
		return instance
	}
}
