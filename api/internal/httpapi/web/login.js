if (isLoggedIn()) {
  window.location.replace("/");
}
renderNav();

const form = document.getElementById("login-form");
const message = document.getElementById("message");

form.addEventListener("submit", async (ev) => {
  ev.preventDefault();
  const button = form.querySelector("button");
  button.disabled = true;
  message.textContent = "";
  try {
    const result = await login(form.elements.login.value, form.elements.password.value);
    if (!result.ok) {
      message.textContent = result.error;
      return;
    }
    window.location.assign("/");
  } finally {
    button.disabled = false;
  }
});