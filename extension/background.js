// background.js - service worker

function getSettings() {
    return new Promise((resolve) => {
        chrome.storage.local.get(
            { backendUrl: "http://localhost:8080/api", minScore: 0, watchInterval: 2880 },
            resolve
        );
    });
}

// Open side panel when toolbar icon is clicked
chrome.action.onClicked.addListener((tab) => {
    chrome.sidePanel.open({ tabId: tab.id });
});

async function setupAlarm(intervalMinutes) {
    await chrome.alarms.clear("poll");
    if (intervalMinutes > 0) {
        chrome.alarms.create("poll", { periodInMinutes: intervalMinutes });
    }
}

// Set up alarm on install using stored interval
chrome.runtime.onInstalled.addListener(async () => {
    const s = await getSettings();
    await setupAlarm(s.watchInterval);
    updateBadge();
});

chrome.runtime.onStartup.addListener(() => {
    updateBadge();
});

async function updateBadge() {
    try {
        const s = await getSettings();
        const res = await fetch(`${s.backendUrl}/jobs?new=true`);
        const jobs = await res.json();
        const count = Array.isArray(jobs) ? jobs.length : 0;
        chrome.action.setBadgeText({ text: count > 0 ? String(count) : "" });
        chrome.action.setBadgeBackgroundColor({ color: "#2563eb" });
    } catch {
        chrome.action.setBadgeText({ text: "" });
    }
}

chrome.alarms.onAlarm.addListener(async (alarm) => {
    if (alarm.name !== "poll") return;

    try {
        const s = await getSettings();
        const res = await fetch(`${s.backendUrl}/jobcart/scan`, { method: "POST" });
        const data = await res.json();

        if (data.new_jobs > 0) {
            chrome.notifications.create({
                type: "basic",
                iconUrl: "icons/icon16.png",
                title: "JobGo - New Jobs Found",
                message: `${data.new_jobs} new job(s) matched your profile`,
            });
        }
        updateBadge();
    } catch (err) {
        console.error("JobGo poll failed:", err)
    }
});

// Add a company from a career page and put it in the cart
async function addToJobGo({ name, platform, slug }) {
    const s = await getSettings();
    const base = s.backendUrl;

    let company;
    const createRes = await fetch(`${base}/companies`, {
        method: "POST",
        headers: {"Content-Type": "application/json" },
        body: JSON.stringify({ name, platform, slug }),
    });

    if (createRes.ok) {
        company = await createRes.json();
    } else {
        const all = await fetch(`${base}/companies`).then((r) => r.json());
        company = all.find((c) => c.slug === slug && c.platform === platform);
        if (!company) throw new Error("Could not create or find company");
    }

    await fetch(`${base}/jobcart/${company.id}`, { method: "POST" });
    return { ok: true };
}

// Handle messages from the side panel
chrome.runtime.onMessage.addListener((msg, _sender, sendResponse) => {
    if (msg.type === "SETTINGS_CHANGED") {
        setupAlarm(msg.watchInterval).then(() => {
            updateBadge();
            sendResponse({ ok: true });
        });
        return true;
    }
    if (msg.type === "SCAN_NOW") {
        getSettings().then((s) =>
            fetch(`${s.backendUrl}/jobcart/scan`, { method: "POST" })
                .then((r) => r.json())
                .then((data) => sendResponse({ ok: true, data }))
                .catch((err) => sendResponse({ ok: false, error: err.message }))
        );
        return true;
    }
    if (msg.type === "CLEAR_BADGE") {
        chrome.action.setBadgeText({ text: "" });
    }
    if (msg.type === "ADD_TO_JOBGO") {
        addToJobGo(msg)
            .then(sendResponse)
            .catch((err) => sendResponse({ ok: false, error: err.message }));
        return true;
    }
});
