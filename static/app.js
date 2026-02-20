const input = document.getElementById("search-input");
const resultsEl = document.getElementById("results");
const statusEl = document.getElementById("status");
const historyDropdown = document.getElementById("history-dropdown");

let debounceTimer = null;
let historyIndex = -1;

const HISTORY_KEY = "curius-search-history";
const MAX_HISTORY = 20;

// Load status on page load
fetchStatus();

input.addEventListener("input", () => {
    clearTimeout(debounceTimer);
    hideHistory();
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

input.addEventListener("focus", () => {
    if (!input.value.trim()) {
        showHistory();
    }
});

input.addEventListener("blur", () => {
    // Delay to allow click on dropdown items
    setTimeout(hideHistory, 150);
});

input.addEventListener("keydown", (e) => {
    const items = historyDropdown.querySelectorAll(".history-item");
    if (!items.length || historyDropdown.classList.contains("hidden")) {
        return;
    }

    if (e.key === "ArrowDown") {
        e.preventDefault();
        historyIndex = Math.min(historyIndex + 1, items.length - 1);
        updateHistoryHighlight(items);
    } else if (e.key === "ArrowUp") {
        e.preventDefault();
        historyIndex = Math.max(historyIndex - 1, -1);
        updateHistoryHighlight(items);
    } else if (e.key === "Enter" && historyIndex >= 0) {
        e.preventDefault();
        const query = items[historyIndex].dataset.query;
        input.value = query;
        hideHistory();
        doSearch(query);
    }
});

// Ctrl/Cmd+K to focus search
document.addEventListener("keydown", (e) => {
    if ((e.metaKey || e.ctrlKey) && e.key === "k") {
        e.preventDefault();
        input.focus();
        input.select();
    }
    if (e.key === "Escape") {
        hideHistory();
    }
});

async function doSearch(query) {
    try {
        const resp = await fetch(`/api/search?q=${encodeURIComponent(query)}&limit=20`);
        if (!resp.ok) throw new Error(`HTTP ${resp.status}`);
        const data = await resp.json();

        addToHistory(query);

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

async function doFindSimilar(id, title) {
    statusEl.textContent = "Finding similar...";
    try {
        const resp = await fetch(`/api/similar?id=${id}&limit=10`);
        if (!resp.ok) throw new Error(`HTTP ${resp.status}`);
        const data = await resp.json();

        if (data.results && data.results.length > 0) {
            statusEl.textContent = `${data.total} similar to "${truncate(title, 40)}"`;
            const banner = `<div class="similar-banner">
                <span>Similar to: <strong>${escapeHtml(truncate(title, 50))}</strong></span>
                <button class="btn-back" onclick="backToSearch()">Back</button>
            </div>`;
            resultsEl.innerHTML = banner + renderResultCards(data.results);
        } else {
            statusEl.textContent = "No similar bookmarks found";
        }
    } catch (err) {
        statusEl.textContent = `Error: ${err.message}`;
    }
}

function backToSearch() {
    const query = input.value.trim();
    if (query) {
        doSearch(query);
    } else {
        resultsEl.innerHTML = "";
        statusEl.textContent = "";
    }
}

function renderResults(results) {
    resultsEl.innerHTML = renderResultCards(results);
}

function renderResultCards(results) {
    return results.map((r) => {
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
                    <button class="btn-similar" onclick="doFindSimilar(${r.id}, ${escapeAttr(JSON.stringify(r.title || 'Untitled'))})">Find similar</button>
                    ${r.createdAt ? `<span class="result-date">${r.createdAt}</span>` : ""}
                </div>
            </div>
        `;
    }).join("");
}

// --- Search History ---

function getHistory() {
    try {
        return JSON.parse(localStorage.getItem(HISTORY_KEY)) || [];
    } catch {
        return [];
    }
}

function addToHistory(query) {
    let history = getHistory();
    history = history.filter((q) => q.toLowerCase() !== query.toLowerCase());
    history.unshift(query);
    if (history.length > MAX_HISTORY) history = history.slice(0, MAX_HISTORY);
    localStorage.setItem(HISTORY_KEY, JSON.stringify(history));
}

function removeFromHistory(query) {
    let history = getHistory();
    history = history.filter((q) => q !== query);
    localStorage.setItem(HISTORY_KEY, JSON.stringify(history));
    showHistory();
}

function showHistory() {
    const history = getHistory();
    if (history.length === 0) {
        hideHistory();
        return;
    }

    historyIndex = -1;
    historyDropdown.innerHTML =
        `<div class="history-label">Recent searches</div>` +
        history.map((q) =>
            `<div class="history-item" data-query="${escapeAttr(q)}" onmousedown="selectHistory('${escapeAttr(q)}')">
                <span>${escapeHtml(q)}</span>
                <span class="history-remove" onmousedown="event.stopPropagation(); removeFromHistory('${escapeAttr(q)}')">&times;</span>
            </div>`
        ).join("");

    historyDropdown.classList.remove("hidden");
}

function hideHistory() {
    historyDropdown.classList.add("hidden");
    historyIndex = -1;
}

function selectHistory(query) {
    input.value = query;
    hideHistory();
    doSearch(query);
}

function updateHistoryHighlight(items) {
    items.forEach((el, i) => {
        el.classList.toggle("active", i === historyIndex);
    });
}

// --- Utilities ---

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
