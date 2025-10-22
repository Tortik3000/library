import http from 'k6/http';
import { randomItem } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

const BASE = __ENV.BASE_URL || `http://localhost:8080`;
const hdrs = { headers: { 'Content-Type': 'application/json' } };

function randomNum() {
    return Math.floor(Math.random() * 1_000_000_000);
}

export const options = {
    scenarios: {
        register_author: {
            executor: 'constant-arrival-rate',
            rate: 10,
            timeUnit: '1s',
            duration: '2m',
            preAllocatedVUs: 20,
            exec: 'registerAuthor',
        },
        add_book: {
            executor: 'constant-arrival-rate',
            rate: 8,
            timeUnit: '1s',
            duration: '2m',
            preAllocatedVUs: 20,
            exec: 'addBook',
        },
        update_book: {
            executor: 'ramping-arrival-rate',
            startRate: 2,
            timeUnit: '1s',
            stages: [
                { target: 15, duration: '1m' },
                { target: 0,  duration: '30s' },
            ],
            preAllocatedVUs: 25,
            exec: 'updateBook',
        },
        get_book_info: {
            executor: 'per-vu-iterations',
            vus: 30,
            iterations: 20,
            exec: 'getBookInfo',
        },
        get_author_info: {
            executor: 'per-vu-iterations',
            vus: 30,
            iterations: 20,
            exec: 'getAuthorInfo',
        },
        get_author_books: {
            executor: 'shared-iterations',
            vus: 10,
            iterations: 100,
            exec: 'getAuthorBooks',
        },
    },
};

export function setup () {
    const authorIds = [];
    for (let i = 0; i < 20; i++) {
        const name = `SeedAuthor${i}`;
        const res = http.post(`${BASE}/v1/library/author`, JSON.stringify({ name }), hdrs);
        authorIds.push(JSON.parse(res.body).id);
    }

    const bookIds = [];
    authorIds.forEach((a, i) => {
        const res = http.post(`${BASE}/v1/library/book`, JSON.stringify({
            name: `SeedBook${i}`,
            author_id: [a],
        }), hdrs);
        bookIds.push(JSON.parse(res.body).book.id);
    });

    return { authorIds, bookIds };
}

export function registerAuthor () {
    const name = `Author${randomNum()}`;
    http.post(`${BASE}/v1/library/author`, JSON.stringify({ name }), hdrs);
}

export function addBook (data) {
    const name = `Book${randomNum()}`;
    http.post(`${BASE}/v1/library/book`, JSON.stringify({
        name,
        author_id: [randomItem(data.authorIds)],
    }), hdrs);
}

export function updateBook (data) {
    const name = `Updated${randomNum()}`;
    http.put(`${BASE}/v1/library/book`, JSON.stringify({
        id: randomItem(data.bookIds),
        name,
        author_id: [randomItem(data.authorIds)],
    }), hdrs);
}

export function getBookInfo (data) {
    http.get(`${BASE}/v1/library/book/${randomItem(data.bookIds)}`);
}

export function getAuthorInfo (data) {
    http.get(`${BASE}/v1/library/author/${randomItem(data.authorIds)}`);
}

export function getAuthorBooks (data) {
    http.get(`${BASE}/v1/library/author_books/${randomItem(data.authorIds)}`);
}
