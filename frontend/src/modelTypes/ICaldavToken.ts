import type {IAbstract} from './IAbstract'

export interface ICaldavToken extends IAbstract {
        id: number;

        created: Date;

       /**
        * The actual token value is only returned when creating a new token.
        */
       token?: string;
}
