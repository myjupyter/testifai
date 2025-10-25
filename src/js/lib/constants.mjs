import 'dotenv/config';

const isDev = process.env.NODE_ENV === 'development';

export const TESTIFAI_URL = isDev ? (process.env.TESTIFAI_URL ?? '') : 'https://testifai.com/api/v1';
export const IS_DEV = isDev;
