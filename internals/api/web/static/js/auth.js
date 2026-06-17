// auth.js — shared logic for /login and /signup
// Both pages: step 1 collects email, step 2 collects OTP code.
// The only difference is which endpoint kicks off step 1 and the copy.

const API_BASE = ""; // adjust if your Go server mounts routes elsewhere

function qs(sel) {
  return document.querySelector(sel);
}

function showToast(message, isError = false) {
  let toast = qs("#toast");
  if (!toast) {
    toast = document.createElement("div");
    toast.id = "toast";
    toast.className = "toast";
    document.body.appendChild(toast);
  }
  toast.textContent = message;
  toast.classList.toggle("error", isError);
  requestAnimationFrame(() => toast.classList.add("show"));
  clearTimeout(toast._t);
  toast._t = setTimeout(() => toast.classList.remove("show"), 3500);
}

function setLoading(btn, loading, label) {
  if (loading) {
    btn.dataset.label = btn.innerHTML;
    btn.innerHTML = `<span class="spinner"></span> ${label || "Working..."}`;
    btn.disabled = true;
  } else {
    btn.innerHTML = btn.dataset.label || btn.innerHTML;
    btn.disabled = false;
  }
}

function fieldError(fieldEl, message) {
  fieldEl.classList.add("field-error");
  let err = fieldEl.querySelector(".error-text");
  if (!err) {
    err = document.createElement("p");
    err.className = "error-text";
    fieldEl.appendChild(err);
  }
  err.textContent = message;
}

function clearFieldError(fieldEl) {
  fieldEl.classList.remove("field-error");
  const err = fieldEl.querySelector(".error-text");
  if (err) err.remove();
}

async function postJSON(path, body) {
  const res = await fetch(`${API_BASE}${path}`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  let data = {};
  try {
    data = await res.json();
  } catch (_) {
    /* no body */
  }
  if (!res.ok) {
    throw new Error(data.message || `Request failed (${res.status})`);
  }
  return data;
}

/**
 * Wires up a two-step auth form.
 * @param {Object} opts
 * @param {string} opts.requestEndpoint - e.g. "/auth" or "/auth/login" (step 1)
 * @param {string} opts.verifyEndpoint - always "/auth/login" (step 2, code verify)
 */
function initAuthFlow({ requestEndpoint, verifyEndpoint = "/auth/login" }) {
  const emailStep = qs("#email-step");
  const codeStep = qs("#code-step");
  const emailForm = qs("#email-form");
  const codeForm = qs("#code-form");
  const emailField = qs("#email-field");
  const emailInput = qs("#email");
  const codeField = qs("#code-field");
  const codeInput = qs("#code");
  const codeEmailLabel = qs("#code-email-label");
  const resendBtn = qs("#resend-code");

  let submittedEmail = "";

  emailForm.addEventListener("submit", async (e) => {
    e.preventDefault();
    clearFieldError(emailField);
    const email = emailInput.value.trim();
    if (!email) {
      fieldError(emailField, "Enter your email to continue.");
      return;
    }

    const btn = emailForm.querySelector("button[type=submit]");
    setLoading(btn, true, "Sending code...");
    try {
      await postJSON(requestEndpoint, { email });
      submittedEmail = email;
      codeEmailLabel.textContent = email;
      emailStep.hidden = true;
      codeStep.hidden = false;
      codeInput.focus();
    } catch (err) {
      fieldError(emailField, err.message);
    } finally {
      setLoading(btn, false);
    }
  });

  codeForm.addEventListener("submit", async (e) => {
    e.preventDefault();
    clearFieldError(codeField);
    const code = codeInput.value.trim();
    if (!code) {
      fieldError(codeField, "Enter the 6-digit code from your inbox.");
      return;
    }

    const btn = codeForm.querySelector("button[type=submit]");
    setLoading(btn, true, "Verifying...");
    try {
      const data = await postJSON(verifyEndpoint, {
        email: submittedEmail,
        code: Number(code),
      });
      if (data.token) {
        localStorage.setItem("gomont_token", data.token);
      }
      window.location.href = "/dashboard";
    } catch (err) {
      fieldError(codeField, err.message);
    } finally {
      setLoading(btn, false);
    }
  });

  if (resendBtn) {
    resendBtn.addEventListener("click", async () => {
      if (!submittedEmail) return;
      setLoading(resendBtn, true, "Sending...");
      try {
        await postJSON(requestEndpoint, { email: submittedEmail });
        showToast("New code sent — check your inbox.");
      } catch (err) {
        showToast(err.message, true);
      } finally {
        setLoading(resendBtn, false);
      }
    });
  }
}

window.GomontAuth = { initAuthFlow, showToast, postJSON, fieldError, clearFieldError, setLoading };