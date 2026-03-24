import AbstractService from './abstractService'
import type {ITask} from '@/modelTypes/ITask'
import TaskModel from '@/models/task'

export default class TrashService extends AbstractService<ITask> {
	constructor() {
		super({
			getAll: '/trash',
		})
	}

	modelFactory(data) {
		return new TaskModel(data)
	}

	async restore(taskId: number) {
		return this.http.post(`/trash/${taskId}/restore`)
	}

	async deletePermanently(taskId: number) {
		return this.http.delete(`/trash/${taskId}`)
	}

	async emptyTrash() {
		return this.http.delete('/trash')
	}
}
