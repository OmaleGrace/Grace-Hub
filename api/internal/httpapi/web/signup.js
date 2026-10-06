if (isLoggedIn()) {
  window.location.replace("/");
}
renderNav();

const form = document.getElementById("signup-form");
const message = document.getElementById("message");

form.addEventListener(
  "invalid",
  (ev) => {
    message.textContent = ev.target.validationMessage;
  },
  true
);

form.addEventListener("submit", async (ev) => {
  ev.preventDefault();
  const button = form.querySelector("button");
  button.disabled = true;
  message.textContent = "";
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
      message.textContent = (r.data && r.data.error) || "Sign up failed. Please try again.";
      return;
    }
    const result = await login(username, password);
    if (!result.ok) {
      window.location.assign("/login");
      return;
    }
    window.location.assign("/");
  } catch (err) {
    message.textContent = "Could not reach the server. Please try again.";
  } finally {
    button.disabled = false;
  }
});
