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

function isLoggedIn() {
  return Boolean(readStore(TOKEN_KEY) && readStore(USER_KEY));
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

async function login(loginValue, password) {
  const r = await api("/login", {
    method: "POST",
    body: JSON.stringify({ login: loginValue, password: password }),
  });
  if (!r.ok) {
    return { ok: false, error: (r.data && r.data.error) || "Login failed. Please try again." };
  }
  saveSession(r.data.token, r.data.user.username);
  return { ok: true };
}

function makeLink(href, text) {
  const a = document.createElement("a");
  a.href = href;
  a.textContent = text;
  return a;
}

function renderNav() {
  const nav = document.getElementById("nav");
  nav.textContent = "";
  if (isLoggedIn()) {
    const who = document.createElement("span");
    who.textContent = "Signed in as " + readStore(USER_KEY);
    const button = document.createElement("button");
    button.type = "button";
    button.id = "logout";
    button.textContent = "Log out";
    button.addEventListener("click", () => {
      clearSession();
      window.location.assign("/");
    });
    nav.append(who, button);
  } else {
    nav.append(makeLink("/login", "Log in"), makeLink("/signup", "Sign up"));
  }
}