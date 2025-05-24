export type PartialWithId<T extends { id: unknown }> = Pick<T, 'id'> & Omit<Partial<T>, 'id'>
