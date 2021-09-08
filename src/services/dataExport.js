import AbstractService from './abstractService'
import {downloadBlob} from '../helpers/downloadBlob'

export default class DataExportService extends AbstractService {
	request(password) {
		return this.post('/user/export/request', {password: password})
	}
	
	download(password) {
		const clear = this.setLoading()
		return this.getBlobUrl('/user/export/download', 'POST', {password})
			.then(url => downloadBlob(url, 'vikunja-export.zip'))
			.finally(() => clear())
	}
}