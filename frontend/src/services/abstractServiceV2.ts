import AbstractService from '@/services/abstractService'
import type { IAbstract } from '@/modelTypes/IAbstract'

export default abstract class AbstractServiceV2<Model extends IAbstract = IAbstract> extends AbstractService<Model> {
  basePath = '/api/v2'
}
