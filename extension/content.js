// content.js — career page detector

(function () {
    function detectCareerPage() {
        const { hostname, pathname } = window.location;
        const parts = pathname.split("/").filter(Boolean);

        if (hostname.includes("greenhouse.io")) {
            return { platform: "greenhouse", slug: parts[0] };
        }
        if (hostname.includes("lever.co")) {
            return { platform: "lever", slug: parts[0] };
        }
        if (hostname.includes("ashbyhq.com")) {
            return { platform: "ashby", slug: parts[0] };
        }
        return null;
    }

    // Get company name from page title or og:site_name
    function getCompanyName(slug) {
        const og = document.querySelector('meta[property="og:site_name"]');
        if (og && og.content) return og.content;
        const title = document.title.split(/[|–—-]/)[0].trim();
        if (title && title.length < 60) return title;
        return slug.replace(/-/g, " ").replace(/\b\w/g, (c) => c.toUpperCase());
    }

    // Inject a floating "Track with JobGo" button
    function inject(info) {
        if (document.getElementById("jobgo-btn")) return;

        const btn = document.createElement("button");
        btn.id = "jobgo-btn";
        btn.textContent = "＋ Track with JobGo";
        Object.assign(btn.style, {
            position: "fixed",
            bottom: "20px",
            right: "20px",
            zIndex: "999999",
            padding: "10px 16px",
            background: "#2563eb",
            color: "#fff",
            border: "none",
            borderRadius: "8px",
            fontSize: "13px",
            fontFamily: "system-ui, sans-serif",
            fontWeight: "600",
            cursor: "pointer",
            boxShadow: "0 4px 12px rgba(0,0,0,0.2)",
            transition: "background 0.2s",
        });

        btn.addEventListener("mouseenter", () => (btn.style.background = "#1d4ed8"));
        btn.addEventListener("mouseleave", () => (btn.style.background = "#2563eb"));

        btn.addEventListener("click", async () => {
            btn.textContent = "Adding…";
            btn.disabled = true;

            chrome.runtime.sendMessage(
                { type: "ADD_TO_JOBGO", ...info },
                (res) => {
                if (res && res.ok) {
                    btn.textContent = "✓ Tracked!";
                    btn.style.background = "#22c55e";
                } else {
                    btn.textContent = "✗ Failed";
                    btn.style.background = "#ef4444";
                    setTimeout(() => {
                    btn.textContent = "＋ Track with JobGo";
                    btn.style.background = "#2563eb";
                    btn.disabled = false;
                    }, 2000);
                }
                }
            );
        });

        document.body.appendChild(btn);
    }

    const info = detectCareerPage();
    if (info && info.slug) {
        info.name = getCompanyName(info.slug);
        inject(info);
    }
})();
