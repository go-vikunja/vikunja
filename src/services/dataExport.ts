import AbstractService from './abstractService'
import {downloadBlob} from '../helpers/downloadBlob'

export default class DataExportService extends AbstractService {
	request(password: string) {
		return this.post('/user/export/request', {password})
	}
	
	async download(password: string) {
		const clear = this.setLoading()
		try {
			const url = await this.getBlobUrl('/user/export/download', 'POST', {password})
			downloadBlob(url, 'vikunja-export.zip')
		} finally {
			clear()
		}
	}
}