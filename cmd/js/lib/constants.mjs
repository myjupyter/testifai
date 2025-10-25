import 'dotenv/config';

const isDev = process.env.NODE_ENV === 'development';

export const { TESTIFAI_URL = '' } = process.env;
export const IS_DEV = isDev;
