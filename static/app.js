const input = document.getElementById("search-input");
const resultsEl = document.getElementById("results");
const statusEl = document.getElementById("status");

let debounceTimer = null;

// Load status on page load
fetchStatus();

input.addEventListener("input", () => {
    clearTimeout(debounceTimer);
    const query = input.value.trim();

    if (!query) {
        resultsEl.innerHTML = "";
        statusEl.textContent = "";
        return;
    }

    statusEl.textContent = "Searching...";

    debounceTimer = setTimeout(() => {
        doSearch(query);
    }, 300);
});

// Ctrl/Cmd+K to focus search
document.addEventListener("keydown", (e) => {
    if ((e.metaKey || e.ctrlKey) && e.key === "k") {
        e.preventDefault();
        input.focus();
        input.select();
    }
});

async function doSearch(query) {
    try {
        const resp = await fetch(`/api/search?q=${encodeURIComponent(query)}&limit=20`);
        if (!resp.ok) throw new Error(`HTTP ${resp.status}`);
        const data = await resp.json();

        if (data.results && data.results.length > 0) {
            statusEl.textContent = `${data.total} results`;
            renderResults(data.results);
        } else {
            statusEl.textContent = "No results found";
            resultsEl.innerHTML = '<div class="empty-state">No matching bookmarks found</div>';
        }
    } catch (err) {
        statusEl.textContent = `Error: ${err.message}`;
        resultsEl.innerHTML = "";
    }
}

function renderResults(results) {
    resultsEl.innerHTML = results.map((r) => {
        const score = (r.score * 100).toFixed(1);
        const domain = extractDomain(r.url);
        const tags = (r.tags || [])
            .map((t) => `<span class="tag-pill">${escapeHtml(t)}</span>`)
            .join("");
        const highlights = (r.highlights || [])
            .slice(0, 2)
            .map((h) => escapeHtml(truncate(h, 150)))
            .join(" ... ");

        return `
            <div class="result-card">
                <div class="result-header">
                    <span class="result-title"><a href="${escapeAttr(r.url)}" target="_blank" rel="noopener">${escapeHtml(r.title || "Untitled")}</a></span>
                    <span class="score-badge">${score}%</span>
                </div>
                <div class="result-url">${escapeHtml(domain)}</div>
                ${r.snippet ? `<div class="result-snippet">${escapeHtml(r.snippet)}</div>` : ""}
                ${highlights ? `<div class="result-highlights">${highlights}</div>` : ""}
                <div class="result-meta">
                    ${tags}
                    ${r.createdAt ? `<span class="result-date">${r.createdAt}</span>` : ""}
                </div>
            </div>
        `;
    }).join("");
}

async function fetchStatus() {
    try {
        const resp = await fetch("/api/status");
        if (!resp.ok) return;
        const data = await resp.json();

        const parts = [];
        parts.push(`${data.indexCount} bookmarks indexed`);
        if (!data.ollamaOk) parts.push("Ollama offline");
        statusEl.textContent = parts.join(" · ");
    } catch {
        // Ignore
    }
}

function extractDomain(url) {
    try {
        return new URL(url).hostname;
    } catch {
        return url;
    }
}

function truncate(str, len) {
    if (str.length <= len) return str;
    return str.slice(0, len) + "...";
}

function escapeHtml(str) {
    const div = document.createElement("div");
    div.textContent = str;
    return div.innerHTML;
}

function escapeAttr(str) {
    return str.replace(/&/g, "&amp;").replace(/"/g, "&quot;").replace(/'/g, "&#39;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
}
