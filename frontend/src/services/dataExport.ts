import AbstractService from './abstractService'
import {downloadBlob} from '../helpers/downloadBlob'
import type {IMessage} from '@/modelTypes/IMessage'

const DOWNLOAD_NAME = 'vikunja-export.zip'

export default class DataExportService extends AbstractService<IMessage> {
	request(password: string) {
		return this.post('/user/export/request', {password, maxRight: null} as unknown as IMessage)
	}
	
	async download(password: string) {
		const clear = this.setLoading()
		try {
			const url = await this.getBlobUrl('/user/export/download', 'POST', {password})
			downloadBlob(url as string, DOWNLOAD_NAME)
		} finally {
			clear()
		}
	}
}
