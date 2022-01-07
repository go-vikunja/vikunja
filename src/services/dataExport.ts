import AbstractService from './abstractService'
import {downloadBlob} from '../helpers/downloadBlob'

export default class DataExportService extends AbstractService {
	request(password) {
		return this.post('/user/export/request', {password: password})
	}
	
	async download(password) {
		const clear = this.setLoading()
		try {
			const url = await this.getBlobUrl('/user/export/download', 'POST', {password})
			downloadBlob(url, 'vikunja-export.zip')
		} finally {
			clear()
		}
	}
}