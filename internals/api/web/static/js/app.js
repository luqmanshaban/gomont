// app.js — shared across authenticated pages (dashboard, settings)
const API_BASE = "";

function getToken() {
  return localStorage.getItem("gomont_token");
}

function requireAuth() {
  const token = getToken();
  if (!token) {
    window.location.href = "/signup";
    return null;
  }
  return token;
}

function logout() {
  localStorage.removeItem("gomont_token");
  window.location.href = "/signup";
}

async function api(path, options = {}) {
  const token = getToken();
  const headers = {
    "Content-Type": "application/json",
    ...(options.headers || {}),
  };
  if (token) headers["Authorization"] = `Bearer ${token}`;

  const res = await fetch(`${API_BASE}${path}`, { ...options, headers });

  if (res.status === 401) {
    logout();
    throw new Error("Session expired. Please log in again.");
  }

  let data = null;
  const text = await res.text();
  if (text) {
    try {
      data = JSON.parse(text);
    } catch (_) {
      data = text;
    }
  }

  if (!res.ok) {
    const message = (data && data.message) || `Request failed (${res.status})`;
    throw new Error(message);
  }
  return data;
}

function showToast(message, isError = false) {
  let toast = document.querySelector("#toast");
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

function escapeHtml(str) {
  const div = document.createElement("div");
  div.textContent = str;
  return div.innerHTML;
}

function timeAgo(isoString) {
  if (!isoString) return "never";
  const diff = (Date.now() - new Date(isoString).getTime()) / 1000;
  if (diff < 60) return "just now";
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
  if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`;
  return `${Math.floor(diff / 86400)}d ago`;
}

window.Gomont = { getToken, requireAuth, logout, api, showToast, escapeHtml, timeAgo };