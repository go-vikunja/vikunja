import AbstractService from './abstractService'

export default class ProjectTemplateService extends AbstractService {
	constructor() {
		super({
			create: '/projects/{projectId}/template',
		})
	}
}
