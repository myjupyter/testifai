import ky from 'ky';
import { TESTIFAI_URL } from '../constants.mjs';

export const request = ky.create({
    credentials: 'include',
    timeout: 30000,
    prefixUrl: TESTIFAI_URL,
    throwHttpErrors: true,
    headers: {
        'Content-Type': 'application/json',
    },
});
