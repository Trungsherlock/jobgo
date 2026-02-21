import { api, configure } from "../utils/api.js";

// Clear badge when side panel opens
chrome.runtime.sendMessage({ type: "CLEAR_BADGE" }).catch(() => {});

// --- Load settings from storage ---
async function initSettings() {
  return new Promise((resolve) => {
    chrome.storage.local.get(
      { backendUrl: "http://localhost:8080/api", minScore: 0 },
      (s) => {
        configure({ base: s.backendUrl, minScore: s.minScore });
        document.getElementById("setting-url").value = s.backendUrl;
        document.getElementById("setting-score").value = s.minScore;
        document.getElementById("setting-score-label").textContent =
          s.minScore + "%";
        resolve();
      }
    );
  });
}

// --- Tab switching ---
document.querySelectorAll(".tab").forEach((btn) => {
  btn.addEventListener("click", () => {
    document.querySelectorAll(".tab").forEach((t) => t.classList.remove("active"));
    document.querySelectorAll(".tab-panel").forEach((p) => p.classList.add("hidden"));
    btn.classList.add("active");
    document.getElementById(`tab-${btn.dataset.tab}`).classList.remove("hidden");
  });
});

// --- Backend health check ---
async function checkHealth() {
  const dot = document.getElementById("status-dot");
  try {
    await api.getStats();
    dot.className = "dot online";
    dot.title = "Backend online";
  } catch {
    dot.className = "dot offline";
    dot.title = "Backend offline — start with: jobgo serve";
  }
}

// --- Jobs tab ---
async function loadJobs() {
  const list = document.getElementById("jobs-list");
  list.innerHTML = "<li>Loading…</li>";

  const params = {};
  if (document.getElementById("filter-new").checked)     params.new = "true";
  if (document.getElementById("filter-remote").checked)   params.remote = "true";
  if (document.getElementById("filter-visa").checked)     params.visa_friendly = "true";
  if (document.getElementById("filter-newgrad").checked)  params.new_grad = "true";

  try {
    const jobs = await api.listJobs(params);
    if (!jobs.length) { list.innerHTML = "<li>No jobs found.</li>"; return; }
    list.innerHTML = jobs.map(jobCard).join("");
  } catch (err) {
    list.innerHTML = `<li style="color:red">${err.message}</li>`;
  }
}

function jobCard(j) {
  const tags = [
    j.is_new_grad ? '<span class="tag tag-newgrad">new grad</span>' : "",
    j.visa_sentiment === "positive" ? '<span class="tag tag-visa">visa+</span>' : "",
    j.remote ? '<span class="tag tag-remote">remote</span>' : "",
  ].join("");

  const score = j.match_score != null ? Math.round(j.match_score) : "—";
  const desc = j.description
    ? j.description.replace(/<[^>]+>/g, "").slice(0, 220) + "…"
    : "No description available.";
  const reason = j.match_reason
    ? `<div class="detail-reason">Match: ${j.match_reason}</div>`
    : "";

  return `<li class="job-item" data-id="${j.id}">
    <div class="job-header">
      <div>
        <span class="job-title">${j.title}</span>
        <span class="score-badge">${score}%</span>${tags}
      </div>
      <div class="job-meta">${j.company_name ?? ""} · ${j.location ?? "N/A"}</div>
    </div>
    <div class="job-detail hidden">
      <p class="detail-desc">${desc}</p>
      ${reason}
      <a class="btn-apply" href="${j.url}" target="_blank">Apply ↗</a>
    </div>
  </li>`;
}

// Job expand — set up once via delegation
document.getElementById("jobs-list").addEventListener("click", (e) => {
  if (e.target.classList.contains("btn-apply")) return;
  const li = e.target.closest("li.job-item");
  if (!li) return;
  li.querySelector(".job-detail").classList.toggle("hidden");
  li.classList.toggle("expanded");
});

["filter-new", "filter-remote", "filter-visa", "filter-newgrad"].forEach((id) => {
  document.getElementById(id).addEventListener("change", loadJobs);
});

// --- Cart tab ---
async function loadCart() {
  const list = document.getElementById("cart-list");
  list.innerHTML = "<li>Loading…</li>";
  try {
    const companies = await api.listCart();
    if (!companies.length) { list.innerHTML = "<li>Cart is empty.</li>"; return; }
    list.innerHTML = companies
      .map(
        (c) => `<li class="company-item">
        <div class="company-row">
          <div>
            <strong>${c.name}</strong>
            <div class="job-meta">${c.platform}/${c.slug}${c.sponsors_h1b ? " · H1B✓" : ""}</div>
          </div>
          <button class="btn-cart in-cart" data-id="${c.id}">Remove</button>
        </div>
      </li>`
      )
      .join("");
  } catch (err) {
    list.innerHTML = `<li style="color:red">${err.message}</li>`;
  }
}

// Cart remove — set up once via delegation
document.getElementById("cart-list").addEventListener("click", async (e) => {
  const btn = e.target.closest(".btn-cart");
  if (!btn) return;
  btn.disabled = true;
  btn.textContent = "…";
  try {
    await api.removeFromCart(btn.dataset.id);
    loadCart();
  } catch (err) {
    btn.disabled = false;
    btn.textContent = "Remove";
    alert(err.message);
  }
});

document.getElementById("btn-scan").addEventListener("click", async () => {
  const status = document.getElementById("scan-status");
  status.textContent = "Scanning…";
  try {
    const data = await api.scanCart();
    status.textContent = `Done — ${data.new_jobs} new job(s)`;
    loadJobs();
  } catch (err) {
    status.textContent = `Error: ${err.message}`;
  }
});

// --- Companies tab ---
async function loadCompanies() {
  const list = document.getElementById("companies-list");
  list.innerHTML = "<li>Loading…</li>";
  try {
    const companies = await api.listCompanies();
    list.innerHTML = companies
      .map(
        (c) => `<li class="company-item">
        <div class="company-row">
          <div>
            <strong>${c.name}</strong>
            <div class="job-meta">${c.platform}/${c.slug}${c.sponsors_h1b ? " · H1B✓" : ""}</div>
          </div>
          <button class="btn-cart ${c.in_cart ? 'in-cart' : ''}"
                  data-id="${c.id}" data-in-cart="${c.in_cart}">
            ${c.in_cart ? 'In Cart' : '+ Cart'}
          </button>
        </div>
      </li>`
      )
      .join("");
  } catch (err) {
    list.innerHTML = `<li style="color:red">${err.message}</li>`;
  }
}

// Cart toggle — set up once via delegation
document.getElementById("companies-list").addEventListener("click", async (e) => {
  const btn = e.target.closest(".btn-cart");
  if (!btn) return;
  const inCart = btn.dataset.inCart === "true";
  btn.disabled = true;
  btn.textContent = "…";
  try {
    if (inCart) {
      await api.removeFromCart(btn.dataset.id);
    } else {
      await api.addToCart(btn.dataset.id);
    }
    loadCompanies();
  } catch (err) {
    btn.disabled = false;
    btn.textContent = inCart ? "In Cart" : "+ Cart";
  }
});

// --- Settings tab ---
document.getElementById("setting-score").addEventListener("input", (e) => {
  document.getElementById("setting-score-label").textContent = e.target.value + "%";
});

document.getElementById("btn-save-settings").addEventListener("click", () => {
  const url =
    document.getElementById("setting-url").value.trim() ||
    "http://localhost:8080/api";
  const score = parseInt(document.getElementById("setting-score").value, 10) || 0;

  chrome.storage.local.set({ backendUrl: url, minScore: score }, () => {
    configure({ base: url, minScore: score });
    document.getElementById("settings-saved").textContent = "Saved!";
    setTimeout(
      () => (document.getElementById("settings-saved").textContent = ""),
      2000
    );
    checkHealth();
    loadJobs();
  });
});

// --- Tab click handlers for lazy load ---
document.querySelector('[data-tab="cart"]').addEventListener("click", loadCart);
document.querySelector('[data-tab="companies"]').addEventListener("click", loadCompanies);

// --- Init ---
initSettings().then(() => {
  checkHealth();
  loadJobs();
});
