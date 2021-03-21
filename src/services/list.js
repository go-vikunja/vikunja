import AbstractService from './abstractService'
import ListModel from '../models/list'
import TaskService from './task'
import {formatISO} from 'date-fns'
import {colorFromHex} from '@/helpers/color/colorFromHex'

export default class ListService extends AbstractService {
	constructor() {
		super({
			create: '/namespaces/{namespaceId}/lists',
			get: '/lists/{id}',
			getAll: '/lists',
			update: '/lists/{id}',
			delete: '/lists/{id}',
		})
	}

	processModel(model) {
		model.created = formatISO(new Date(model.created))
		model.updated = formatISO(new Date(model.updated))
		return model
	}

	modelFactory(data) {
		return new ListModel(data)
	}

	beforeUpdate(model) {
		let taskService = new TaskService()
		model.tasks = model.tasks.map(task => {
			return taskService.beforeUpdate(task)
		})
		model.hexColor = colorFromHex(model.hexColor)
		return model
	}

	beforeCreate(list) {
		list.hexColor = colorFromHex(list.hexColor)
		return list
	}

	background(list) {
		if (list.background === null) {
			return Promise.resolve('')
		}

		return this.http({
			url: `/lists/${list.id}/background`,
			method: 'GET',
			responseType: 'blob',
		})
			.then(response => {
				return window.URL.createObjectURL(new Blob([response.data]))
			})
			.catch(e => {
				return e
			})
	}

	removeBackground(list) {
		const cancel = this.setLoading()

		return this.http.delete(`/lists/${list.id}/background`, list)
			.then(response => {
				return Promise.resolve(response.data)
			})
			.catch(error => {
				return this.errorHandler(error)
			})
			.finally(() => {
				cancel()
			})
	}
}