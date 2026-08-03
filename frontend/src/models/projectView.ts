import type {IProjectView, ProjectViewBucketConfigurationMode, ProjectViewKind} from '@/modelTypes/IProjectView'
import type {IFilters} from '@/modelTypes/ISavedFilter'
import AbstractModel from '@/models/abstractModel'

export default class ProjectViewModel extends AbstractModel<IProjectView> implements IProjectView {
	id = 0
	title = ''
	projectId = 0
	viewKind: ProjectViewKind =  'list'

	filter: IFilters = {
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

	constructor(data: Partial<IProjectView>) {
		super()
		this.assignData(data)
		
		if (!this.bucketConfiguration) {
			this.bucketConfiguration = []
		}
	}

	static createWithDefaultFilter(data: Partial<IProjectView> = {}): ProjectViewModel {
		// Do not bake a default sort here — once ViewEditForm preserves sort_by,
		// a done/id sort would override manual (position) ordering for new views.
		const defaultFilter: IFilters = {
			sort_by: [],
			order_by: [],
			filter: 'done = false',
			filter_include_nulls: true,
			s: '',
		}

		const instance = new ProjectViewModel(data)
		instance.filter = defaultFilter
		return instance
	}
}
