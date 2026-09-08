import AbstractService from '../abstractService'
import {apiV2Url} from '@/helpers/fetcher'

interface CancelResponse {
	message: string
}

// The migration slot is claimed per account, not per migrator, so cancelling is
// one route for all of them.
export default class MigrationCancelService extends AbstractService {
	constructor() {
		super({
			update: apiV2Url('migration/cancel'),
		})
	}

	useUpdateInterceptor() {
		return false
	}

	cancel() {
		// The route takes no payload; the model shape only exists to satisfy the
		// abstract service.
		return this.update({maxPermission: null}) as unknown as Promise<CancelResponse>
	}
}
