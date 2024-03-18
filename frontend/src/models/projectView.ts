import type {IProjectView, ProjectViewBucketConfigurationMode, ProjectViewKind} from '@/modelTypes/IProjectView'
import AbstractModel from '@/models/abstractModel'

export default class ProjectViewModel extends AbstractModel<IProjectView> implements IProjectView {
	id = 0
	title = ''
	projectId = 0
	viewKind: ProjectViewKind =  'list'

	filter = ''
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
}