// dashboard.js — monitor list, add form, retry/delete actions, live updates
(function () {
  const { requireAuth, api, showToast, escapeHtml, timeAgo, getToken } =
    window.Gomont;

  if (!requireAuth()) return; // redirects to /login if no token

  const listEl = document.getElementById("monitor-list");
  const countEl = document.getElementById("monitor-count");
  const addForm = document.getElementById("add-monitor-form");
  const endpointField = document.getElementById("endpoint-field");
  const endpointInput = document.getElementById("endpoint");
  const intervalInput = document.getElementById("interval");
  const logoutLink = document.getElementById("logout-link");

  let monitors = [];
  let editingId = null; // tracks which row (if any) is mid-edit, so live updates don't clobber it

  function statusIcon(isUp) {
    return `<span class="status-badge ${isUp ? "up" : "down"}">
      <span class="status-dot ${isUp ? "up" : "down"}"></span>
      ${isUp ? "Up" : "Down"}
    </span>`;
  }

  function monitorRowHtml(m) {
    const isUp = m.is_healthy ?? m.isHealthy ?? false;
    const updated = m.updated_at || m.created_at;
    return `
      <div class="monitor-row" data-id="${m.id}" data-interval="${m.interval}">
        ${statusIcon(isUp)}
        <span class="monitor-url" title="${escapeHtml(m.endpoint)}">${escapeHtml(m.endpoint)}</span>
        <span class="monitor-meta" data-role="meta">every ${m.interval}m &middot; checked ${timeAgo(updated)}</span>
        <div class="monitor-actions">
          <button class="icon-btn" data-action="edit" data-id="${m.id}" title="Edit interval" aria-label="Edit interval">&#9998;</button>
          <button class="icon-btn" data-action="retry" data-id="${m.id}" title="Retry now" aria-label="Retry now">&#x21bb;</button>
          <button class="icon-btn danger" data-action="delete" data-id="${m.id}" title="Delete monitor" aria-label="Delete monitor">&#x2715;</button>
        </div>
      </div>`;
  }

  function emptyStateHtml() {
    return `
      <div class="empty-state">
        <svg width="48" height="48" viewBox="0 0 100 100" fill="none">
          <path d="M 10 50 Q 30 20, 50 50 T 90 50" stroke="currentColor" stroke-width="4" stroke-linecap="round" stroke-dasharray="8,8" />
        </svg>
        <h3>No monitors yet</h3>
        <p>Add a URL above to start watching it.</p>
      </div>`;
  }

  function render() {
    countEl.textContent = monitors.length
      ? `${monitors.length} monitor${monitors.length === 1 ? "" : "s"}`
      : "";

    if (!monitors.length) {
      listEl.innerHTML = emptyStateHtml();
      return;
    }

    listEl.innerHTML = monitors.map(monitorRowHtml).join("");
  }

  async function loadMonitors() {
    try {
      const data = await api("/urls");
      monitors = Array.isArray(data) ? data : [];
      render();
    } catch (err) {
      listEl.innerHTML = `<div class="empty-state"><p>${escapeHtml(err.message)}</p></div>`;
    }
  }

  // ---- Live updates via Server-Sent Events ----
  // One connection covers every monitor on this dashboard (see backend:
  // internals/sse). EventSource handles reconnection automatically on
  // dropped connections, so no manual retry loop is needed here.
  function connectLiveUpdates() {
    const token = getToken();
    if (!token) return;

    const source = new EventSource(
      `/events?token=${encodeURIComponent(token)}`,
    );

    source.addEventListener("status_change", (e) => {
      let payload;
      try {
        payload = JSON.parse(e.data);
      } catch (_) {
        return;
      }

      const id = payload.id;
      const idx = monitors.findIndex((m) => String(m.id) === String(id));
      if (idx === -1) return; // monitor not in current view (e.g. deleted since)

      monitors[idx] = {
        ...monitors[idx],
        is_healthy: payload.is_healthy,
        updated_at: payload.updated_at,
      };

      // Don't blow away a row the user is actively editing — re-rendering
      // would replace their in-progress interval input with static text.
      if (String(editingId) === String(id)) return;

      render();
    });

    source.onerror = () => {
      // EventSource retries automatically with backoff; nothing to do
      // here besides letting it. If the token has expired, the server
      // will keep returning 401 on each retry attempt and the browser
      // will keep trying — acceptable for now since reloading the page
      // (which re-checks requireAuth) is the natural recovery path.
    };

    return source;
  }

  addForm.addEventListener("submit", async (e) => {
    e.preventDefault();
    endpointField.classList.remove("field-error");
    const existingErr = endpointField.querySelector(".error-text");
    if (existingErr) existingErr.remove();

    const endpoint = endpointInput.value.trim();
    const interval = Number(intervalInput.value);

    if (!endpoint) {
      endpointField.classList.add("field-error");
      const err = document.createElement("p");
      err.className = "error-text";
      err.textContent = "Enter a URL to monitor.";
      endpointField.appendChild(err);
      return;
    }

    const btn = addForm.querySelector("button[type=submit]");
    const originalLabel = btn.innerHTML;
    btn.innerHTML = `<span class="spinner"></span> Adding...`;
    btn.disabled = true;

    try {
      const created = await api("/urls", {
        method: "POST",
        body: JSON.stringify({ endpoint, interval }),
      });
      monitors.unshift(created);
      render();
      endpointInput.value = "";
      intervalInput.value = "5";
      showToast("Monitor added.");
      setTimeout(function () {
        window.location.reload();
      }, 3000); // 3000 milliseconds = 3 seconds
    } catch (err) {
      showToast(err.message, true);
    } finally {
      btn.innerHTML = originalLabel;
      btn.disabled = false;
    }
  });

  listEl.addEventListener("click", async (e) => {
    const btn = e.target.closest("button[data-action]");
    if (!btn) return;

    const id = btn.dataset.id;
    const action = btn.dataset.action;
    const row = btn.closest(".monitor-row");

    if (action === "edit") {
      editingId = id;
      const metaEl = row.querySelector('[data-role="meta"]');
      const currentInterval = row.dataset.interval;
      metaEl.outerHTML = `
        <span class="monitor-meta" data-role="meta">
          every <input type="number" min="1" class="email-row-input" style="width:60px;display:inline-block;padding:0.25rem 0.4rem" value="${currentInterval}" data-role="interval-input" />m
        </span>`;
      row.querySelector('[data-role="interval-input"]').focus();
      row.querySelector(".monitor-actions").innerHTML = `
        <button class="icon-btn" data-action="save-edit" data-id="${id}" title="Save" aria-label="Save interval">&#10003;</button>
        <button class="icon-btn" data-action="cancel-edit" data-id="${id}" title="Cancel" aria-label="Cancel edit">&#x2715;</button>`;
      return;
    }

    if (action === "cancel-edit") {
      editingId = null;
      render();
      return;
    }

    if (action === "save-edit") {
      const input = row.querySelector('[data-role="interval-input"]');
      const newInterval = Number(input.value);
      if (!newInterval || newInterval < 1) {
        showToast("Interval must be at least 1 minute.", true);
        return;
      }
      btn.disabled = true;
      try {
        const updated = await api(`/urls/${id}`, {
          method: "PATCH",
          body: JSON.stringify({ interval: newInterval }),
        });
        monitors = monitors.map((m) =>
          String(m.id) === String(id) ? updated : m,
        );
        editingId = null;
        render();
        showToast("Monitor updated.");
      } catch (err) {
        showToast(err.message, true);
        btn.disabled = false;
      }
      return;
    }

    if (action === "retry") {
      btn.disabled = true;
      try {
        await api(`/urls/${id}/retry`, { method: "POST" });
        showToast("Retry scheduled.");
      } catch (err) {
        showToast(err.message, true);
      } finally {
        btn.disabled = false;
      }
      return;
    }

    if (action === "delete") {
      if (!confirm("Delete this monitor? This can't be undone.")) return;
      btn.disabled = true;
      try {
        await api(`/urls/${id}`, { method: "DELETE" });
        monitors = monitors.filter((m) => String(m.id) !== String(id));
        render();
        showToast("Monitor deleted.");
      } catch (err) {
        showToast(err.message, true);
        btn.disabled = false;
      }
    }
  });

  let liveSource = null;

  if (logoutLink) {
    logoutLink.addEventListener("click", (e) => {
      e.preventDefault();
      if (liveSource) liveSource.close();
      window.Gomont.logout();
    });
  }

  loadMonitors().then(() => {
    liveSource = connectLiveUpdates();
  });
})();
