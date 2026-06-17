// settings.js — profile form, notification channel emails, delete account
(function () {
  const { requireAuth, api, showToast, escapeHtml, logout } = window.Gomont;

  if (!requireAuth()) return;

  // ---- Profile ----
  const profileForm = document.getElementById("profile-form");
  const namesInput = document.getElementById("names");
  const namesField = document.getElementById("names-field");
  const profileEmail = document.getElementById("profile-email");
  const profileJoined = document.getElementById("profile-joined");

  async function loadProfile() {
    try {
      const user = await api("/users");
      namesInput.value = user.names || "";
      profileEmail.textContent = user.email || "—";
      profileJoined.textContent = user.created_at
        ? new Date(user.created_at).toLocaleDateString()
        : "—";
    } catch (err) {
      showToast(err.message, true);
    }
  }

  profileForm.addEventListener("submit", async (e) => {
    e.preventDefault();
    namesField.classList.remove("field-error");
    const btn = profileForm.querySelector("button[type=submit]");
    const original = btn.textContent;
    btn.disabled = true;
    btn.textContent = "Saving...";
    try {
      await api("/users", {
        method: "PATCH",
        body: JSON.stringify({ names: namesInput.value.trim() }),
      });
      showToast("Profile updated.");
    } catch (err) {
      showToast(err.message, true);
    } finally {
      btn.disabled = false;
      btn.textContent = original;
    }
  });

  // ---- Notification channel ----
  const emailListEl = document.getElementById("email-list");
  const emailAddForm = document.getElementById("email-add-form");
  const newEmailInput = document.getElementById("new-email");
  const newEmailField = document.getElementById("new-email-field");

  let channelId = null;
  let emails = [];

  function emailRowHtml(email) {
    return `
      <div class="email-row" data-email="${escapeHtml(email)}">
        <span class="email-row-value">${escapeHtml(email)}</span>
        <div class="email-row-actions">
          <button class="icon-btn" data-action="edit" title="Edit" aria-label="Edit email">&#9998;</button>
          <button class="icon-btn danger" data-action="remove" title="Remove" aria-label="Remove email">&#x2715;</button>
        </div>
      </div>`;
  }

  function renderEmails() {
    if (!emails.length) {
      emailListEl.innerHTML = `<p class="page-sub">No notification emails yet. Add one above.</p>`;
      return;
    }
    emailListEl.innerHTML = emails.map(emailRowHtml).join("");
  }

  async function loadChannel() {
    try {
      const channel = await api("/notifications/channels");
      channelId = channel.id;
      emails = channel.emails || [];
      renderEmails();
    } catch (err) {
      emailListEl.innerHTML = `<p class="page-sub">${escapeHtml(err.message)}</p>`;
    }
  }

  emailAddForm.addEventListener("submit", async (e) => {
    e.preventDefault();
    newEmailField.classList.remove("field-error");
    const email = newEmailInput.value.trim();
    if (!email || channelId === null) return;

    const btn = emailAddForm.querySelector("button[type=submit]");
    btn.disabled = true;
    try {
      const updated = await api(`/notifications/channels/${channelId}`, {
        method: "POST",
        body: JSON.stringify({ emails: [email] }),
      });
      emails = updated.emails || [];
      renderEmails();
      newEmailInput.value = "";
      showToast("Email added.");
    } catch (err) {
      showToast(err.message, true);
    } finally {
      btn.disabled = false;
    }
  });

  emailListEl.addEventListener("click", async (e) => {
    const btn = e.target.closest("button[data-action]");
    if (!btn || channelId === null) return;

    const row = btn.closest(".email-row");
    const currentEmail = row.dataset.email;

    if (btn.dataset.action === "remove") {
      if (!confirm(`Remove ${currentEmail} from notifications?`)) return;
      btn.disabled = true;
      try {
        const updated = await api(`/notifications/channels/${channelId}`, {
          method: "DELETE",
          body: JSON.stringify({ email: currentEmail }),
        });
        emails = updated.emails || [];
        renderEmails();
        showToast("Email removed.");
      } catch (err) {
        showToast(err.message, true);
        btn.disabled = false;
      }
      return;
    }

    if (btn.dataset.action === "edit") {
      const valueEl = row.querySelector(".email-row-value");
      const input = document.createElement("input");
      input.type = "email";
      input.className = "email-row-input";
      input.value = currentEmail;
      valueEl.replaceWith(input);
      input.focus();
      btn.closest(".email-row-actions").innerHTML = `
        <button class="icon-btn" data-action="save-edit" title="Save" aria-label="Save email">&#10003;</button>
        <button class="icon-btn" data-action="cancel-edit" title="Cancel" aria-label="Cancel edit">&#x2715;</button>`;
      row.dataset.editing = "true";
      return;
    }

    if (btn.dataset.action === "cancel-edit") {
      renderEmails();
      return;
    }

    if (btn.dataset.action === "save-edit") {
      const input = row.querySelector(".email-row-input");
      const newEmail = input.value.trim();
      if (!newEmail || newEmail === currentEmail) {
        renderEmails();
        return;
      }
      btn.disabled = true;
      try {
        const updated = await api(`/notifications/channels/${channelId}`, {
          method: "PATCH",
          body: JSON.stringify({ old_email: currentEmail, new_email: newEmail }),
        });
        emails = updated.emails || [];
        renderEmails();
        showToast("Email updated.");
      } catch (err) {
        showToast(err.message, true);
        renderEmails();
      }
    }
  });

  // ---- Delete account ----
  const deleteBtn = document.getElementById("delete-account-btn");
  deleteBtn.addEventListener("click", async () => {
    if (!confirm("Delete your account permanently? This removes all your monitors too. This cannot be undone.")) {
      return;
    }
    deleteBtn.disabled = true;
    try {
      await api("/users", { method: "DELETE" });
      showToast("Account deleted.");
      setTimeout(logout, 800);
    } catch (err) {
      showToast(err.message, true);
      deleteBtn.disabled = false;
    }
  });

  loadProfile();
  loadChannel();
})();