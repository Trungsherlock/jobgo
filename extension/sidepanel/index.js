import { api } from "../utils/api.js";

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
    j.location?.toLowerCase().includes("remote") ? '<span class="tag tag-remote">remote</span>' : "",
  ].join("");

  return `<li>
    <div class="job-title">${j.title}<span class="score-badge">${Math.round(j.match_score)}%</span>${tags}</div>
    <div class="job-meta">${j.company_name ?? ""} · ${j.location ?? "N/A"}</div>
  </li>`;
}

["filter-new","filter-remote","filter-visa","filter-newgrad"].forEach((id) => {
  document.getElementById(id).addEventListener("change", loadJobs);
});

// --- Cart tab ---
async function loadCart() {
  const list = document.getElementById("cart-list");
  list.innerHTML = "<li>Loading…</li>";
  try {
    const companies = await api.listCart();
    if (!companies.length) { list.innerHTML = "<li>Cart is empty.</li>"; return; }
    list.innerHTML = companies.map(
      (c) => `<li><strong>${c.name}</strong> <span class="job-meta">${c.platform}/${c.slug}</span></li>`
    ).join("");
  } catch (err) {
    list.innerHTML = `<li style="color:red">${err.message}</li>`;
  }
}

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
    list.innerHTML = companies.map(
      (c) => `<li>
        <strong>${c.name}</strong>
        <span class="job-meta">${c.platform}/${c.slug}${c.sponsors_h1b ? " · H1B✓" : ""}</span>
      </li>`
    ).join("");
  } catch (err) {
    list.innerHTML = `<li style="color:red">${err.message}</li>`;
  }
}

// --- Init ---
checkHealth();
loadJobs();

document.querySelector('[data-tab="cart"]').addEventListener("click", loadCart);
document.querySelector('[data-tab="companies"]').addEventListener("click", loadCompanies);
