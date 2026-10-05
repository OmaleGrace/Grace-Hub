const TOKEN_KEY = "token";
const USER_KEY = "username";

function readStore(key) {
  try { return localStorage.getItem(key); } catch { return null; }
}

function writeStore(key, value) {
  try { localStorage.setItem(key, value); } catch {}
}

function removeStore(key) {
  try { localStorage.removeItem(key); } catch {}
}

function saveSession(token, username) {
  writeStore(TOKEN_KEY, token);
  writeStore(USER_KEY, username);
}

function clearSession() {
  removeStore(TOKEN_KEY);
  removeStore(USER_KEY);
}

async function api(path, options = {}) {
  const headers = { "Content-Type": "application/json" };
  const token = readStore(TOKEN_KEY);
  if (token) headers["Authorization"] = "Bearer " + token;

  const res = await fetch(path, { ...options, headers });
  let data = null;
  try { data = await res.json(); } catch {}
  return { ok: res.ok, status: res.status, data };
}

function showMessage(text) {
  document.getElementById("auth-message").textContent = text;
}

function refreshAuthUI() {
  const token = readStore(TOKEN_KEY);
  const username = readStore(USER_KEY);
  const loggedIn = Boolean(token && username);

  document.getElementById("auth-panel").hidden = loggedIn;
  document.getElementById("session").hidden = !loggedIn;
  document.getElementById("whoami").textContent = loggedIn ? "Signed in as " + username : "";
}

async function login(loginValue, password) {
  const r = await api("/login", {
    method: "POST",
    body: JSON.stringify({ login: loginValue, password: password }),
  });
  if (!r.ok) {
    showMessage((r.data && r.data.error) || "Login failed. Please try again.");
    return false;
  }
  saveSession(r.data.token, r.data.user.username);
  showMessage("");
  refreshAuthUI();
  return true;
}

async function handleLogin(ev) {
  ev.preventDefault();
  const form = ev.target;
  const button = form.querySelector("button");
  button.disabled = true;
  try {
    const ok = await login(form.elements.login.value, form.elements.password.value);
    if (ok) form.reset();
  } finally {
    button.disabled = false;
  }
}

async function handleRegister(ev) {
  ev.preventDefault();
  const form = ev.target;
  const button = form.querySelector("button");
  button.disabled = true;
  try {
    const username = form.elements.username.value;
    const password = form.elements.password.value;
    const r = await api("/register", {
      method: "POST",
      body: JSON.stringify({
        username: username,
        email: form.elements.email.value,
        password: password,
      }),
    });
    if (!r.ok) {
      showMessage((r.data && r.data.error) || "Sign up failed. Please try again.");
      return;
    }
    const ok = await login(username, password);
    if (ok) form.reset();
  } finally {
    button.disabled = false;
  }
}

async function loadEvents() {
  const list = document.getElementById("events");
  list.textContent = "Loading...";

  try {
    const res = await fetch("/events?status=open");
    if (!res.ok) throw new Error("request failed");
    const events = await res.json();

    list.textContent = "";
    if (events.length === 0) {
      list.textContent = "No open events right now.";
      return;
    }
    for (const e of events) {
      list.appendChild(renderEvent(e));
    }
  } catch (err) {
    list.textContent = "Could not load events. Please try again later.";
  }
}

function renderEvent(e) {
  const card = document.createElement("article");
  card.className = "card";

  const title = document.createElement("h3");
  title.textContent = e.title;

  const meta = document.createElement("p");
  meta.className = "meta";
  meta.textContent = e.category + " · locks " + new Date(e.locks_at).toLocaleString();

  const options = document.createElement("p");
  options.textContent = "Options: " + e.options.join(", ");

  card.append(title, meta, options);
  return card;
}

document.getElementById("login-form").addEventListener("submit", handleLogin);
document.getElementById("register-form").addEventListener("submit", handleRegister);
document.getElementById("logout").addEventListener("click", () => {
  clearSession();
  refreshAuthUI();
});

refreshAuthUI();
loadEvents();