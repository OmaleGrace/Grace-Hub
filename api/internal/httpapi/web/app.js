let currentEvents = [];

async function loadEvents() {
  const list = document.getElementById("events");
  list.textContent = "Loading...";

  try {
    const res = await fetch("/events?status=open");
    if (!res.ok) throw new Error("request failed");
    currentEvents = await res.json();
    renderEvents();
  } catch (err) {
    list.textContent = "Could not load events. Please try again later.";
  }
}

function renderEvents() {
  const list = document.getElementById("events");
  list.textContent = "";
  if (currentEvents.length === 0) {
    list.textContent = "No open events right now.";
    return;
  }
  for (const e of currentEvents) {
    list.appendChild(renderEvent(e));
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

    const pick = document.createElement("p");
  pick.className = "pick";
  pick.textContent = "System pick: loading...";
  card.append(title, meta, pick);
  loadSystemPick(e.id, pick);

  if (isLoggedIn()) {
    card.appendChild(buildPredictionForm(e));
  } else {
    const hint = document.createElement("p");
    hint.className = "hint";
    hint.append(makeLink("/login", "Log in"), document.createTextNode(" to make a prediction."));
    card.appendChild(hint);
  }
  return card;
}

async function loadSystemPick(eventId, el) {
  try {
    const res = await fetch("/events/" + eventId + "/system-prediction");
    if (res.status === 404) {
      el.textContent = "System pick: not ready yet";
      return;
    }
    if (!res.ok) throw new Error("request failed");
    const p = await res.json();
    el.textContent = "System pick: " + p.predicted + " (" + Math.round(p.confidence * 100) + "% confident)";
  } catch (err) {
    el.textContent = "System pick unavailable right now.";
  }
}

function buildPredictionForm(e) {
  const form = document.createElement("form");
  form.className = "predict";

  const choices = document.createElement("div");
  choices.className = "choices";
  for (const opt of e.options) {
    const label = document.createElement("label");
    label.className = "choice";
    const input = document.createElement("input");
    input.type = "radio";
    input.name = "predicted";
    input.value = opt;
    input.required = true;
    label.append(input, document.createTextNode(" " + opt));
    choices.appendChild(label);
  }

  const confLabel = document.createElement("label");
  const confText = document.createElement("span");
  confText.textContent = "Confidence: 60%";
  const slider = document.createElement("input");
  slider.type = "range";
  slider.name = "confidence";
  slider.min = "1";
  slider.max = "99";
  slider.value = "60";
  slider.addEventListener("input", () => {
    confText.textContent = "Confidence: " + slider.value + "%";
  });
  confLabel.append(confText, slider);

  const button = document.createElement("button");
  button.type = "submit";
  button.textContent = "Submit prediction";

  const status = document.createElement("p");
  status.className = "status";
  status.setAttribute("role", "status");

  form.append(choices, confLabel, button, status);

  form.addEventListener("submit", async (ev) => {
    ev.preventDefault();
    button.disabled = true;
    status.textContent = "";
    try {
      const r = await api("/events/" + e.id + "/predictions", {
        method: "POST",
        body: JSON.stringify({
          predicted: form.elements.predicted.value,
          confidence: Number(slider.value) / 100,
        }),
      });

      if (r.status === 401) {
        clearSession();
        window.location.assign("/login");
        return;
      }
      if (!r.ok) {
        status.className = "status error";
        status.textContent = (r.data && r.data.error) || "Could not save your prediction.";
        return;
      }
      status.className = "status ok";
      status.textContent =
        "Saved: " + r.data.predicted + " at " + Math.round(r.data.confidence * 100) + "%";
    } finally {
      button.disabled = false;
    }
  });

  return form;
}

renderNav();
loadEvents();