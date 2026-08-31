import type {IProjectView, IProjectViewBucketConfiguration, ProjectViewBucketConfigurationMode, ProjectViewKind} from '@/modelTypes/IProjectView'
import AbstractModel from '@/models/abstractModel'
import {objectToSnakeCase} from '@/helpers/case'
import type {IFilters} from '@/modelTypes/ISavedFilter'

export default class ProjectViewModel extends AbstractModel<IProjectView> implements IProjectView {
	id = 0
	title = ''
	projectId = 0
	viewKind: ProjectViewKind =  'list'

	filter: IProjectView['filters'] = {
		sort_by: [],
		order_by: [],
		filter: '',
		filter_include_nulls: true,
		s: '',
	}
	position = 0
	bucketConfiguration: IProjectViewBucketConfiguration[] = []
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

		// assignData camelCases nested objects; filters stay snake_case, as in SavedFilterModel.
		this.filter = objectToSnakeCase(this.filter) as IFilters
		this.bucketConfiguration = this.bucketConfiguration.map(bc => ({
			...bc,
			filter: objectToSnakeCase(bc.filter) as IFilters,
		}))
	}

	static createWithDefaultFilter(data: Partial<IProjectView> = {}): ProjectViewModel {
		const defaultFilter: IProjectView['filters'] = {
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
