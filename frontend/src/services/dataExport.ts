import AbstractService from './abstractService'
import {downloadBlob} from '../helpers/downloadBlob'

const DOWNLOAD_NAME = 'vikunja-export.zip'

export default class DataExportService extends AbstractService {
	request(password: string) {
		return this.post('/user/export/request', {password})
	}

	status() {
		return this.getM('/user/export')
	}
	
	async download(password: string) {
		const clear = this.setLoading()
		try {
			const url = await this.getBlobUrl('/user/export/download', 'POST', {password})
			downloadBlob(url, DOWNLOAD_NAME)
		} finally {
			clear()
		}
	}
}
