// utils/api.js - thin fetch wrapper around the Go backend

const config = {
    base: "http://localhost:8080/api",
    minScore: 0,
};

export function configure(opts) {
    if (opts.base) config.base = opts.base.replace(/\/$/, "");
    if (opts.minScore !== undefined) config.minScore = Number(opts.minScore);
}

async function apiFetch(path, options = {}) {
    const res = await fetch(config.base + path, {
        headers: {"Content-Type": "application/json" },
        ...options,
    });
    if (!res.ok) {
        const err = await res.json().catch(() => ({ error: res.statusText }));
        throw new Error(err.error || res.statusText);
    }
    return res.json();
}

export const api = {
    // Jobs
    listJobs: (params = {}) => {
        if (config.minScore > 0 && !params.min_score)
            params.min_score = config.minScore;
        const q = new URLSearchParams(params).toString();
        return apiFetch(`/jobs${q ? "?" + q : ""}`);
    },
    getJob: (id) => apiFetch(`/jobs/${id}`),

    // Companies
    listCompanies: () => apiFetch("/companies"),
    addCompany: (body) => 
        apiFetch("/companies", { method: "POST", body: JSON.stringify(body) }),
    deleteCompany: (id) => apiFetch(`/companies/${id}`, { method: "DELETE" }),

    // JobCart
    listCart: () => apiFetch("/jobcart"),
    addToCart: (id) => apiFetch(`/jobcart/${id}`, { method: "POST" }),
    removeFromCart: (id) => apiFetch(`/jobcart/${id}`, { method: "DELETE" }),
    scanCart: () => apiFetch("/jobcart/scan", { method: "POST" }),

    // Profile & stats
    getProfile: () => apiFetch("/profile"),
    getStats: () => apiFetch("/stats"),

    // H1B
    listSponsors: () => apiFetch("/h1b/sponsors"),
    h1bStatus: () => apiFetch("/h1b/status"),
};