import {AuthenticatedHTTPFactory, apiV2Url} from '@/helpers/fetcher'
import type {IMcpInfo} from '@/modelTypes/IApiTokenSettings'

export async function getMcpInfo(): Promise<IMcpInfo> {
	const {data} = await AuthenticatedHTTPFactory().get<IMcpInfo>(apiV2Url('mcp/info'))
	return data
}
