// background.js - service worker

const API_BASE = "http://localhost:8080/api";

// Open side panel when toolbar icon is clicked
chrome.action.onClicked.addListener((tab) => {
    chrome.sidePanel.open({ tabId: tab.id });
});

// Set up a 30-minute alarm for background polling
chrome.runtime.onInstalled.addListener(() => {
    chrome.alarms.create("poll", { periodInMinutes: 30 });
    updateBadge();
});

chrome.runtime.onStartup.addListener(() => {
    updateBadge();
});

async function updateBadge() {
    try {
        const res = await fetch(`${API_BASE}/jobs?new=true`);
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
        const res = await fetch(`${API_BASE}/jobcart/scan`, { method: "POST" });
        const data = await res.json();

        if (data.new_jobs > 0) {
            chrome.notifications.create({
                type: "basic",
                iconUrl: "icons/icon16.png",
                title: "JobGo - New Jobs Found",
                message: `${data.new_jobs} new job(s) matched your profile`,
            });
        }
    } catch (err) {
        console.error("JobGo poll falied:", err)
    }
});

// Handle messages from the side panel
chrome.runtime.onMessage.addListener((msg, _sender, sendResponse) => {
    if (msg.type === "SCAN_NOW") {
        fetch(`${API_BASE}/jobcart/scan`, { method: "POST" })
            .then((r) => r.json())
            .then((data) => sendResponse({ ok: true, data }))
            .catch((err) => sendResponse({ ok: false, error: err.message }));
        return true;
    }
    if (msg.type === "CLEAR_BADGE") {
        chrome.action.setBadgeText({ text: "" });
    }
});