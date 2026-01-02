const tokenInput = document.getElementById("token");
const saveTokenBtn = document.getElementById("saveToken");
const programList = document.getElementById("programList");

const state = {
  token: "",
};

function loadToken() {
  const stored = localStorage.getItem("goapp_token");
  if (stored) {
    state.token = stored;
    tokenInput.value = stored;
  }
}

function saveToken() {
  state.token = tokenInput.value.trim();
  localStorage.setItem("goapp_token", state.token);
  refreshPrograms();
}

async function api(path, options = {}) {
  const headers = options.headers || {};
  if (state.token) {
    headers["Authorization"] = `Bearer ${state.token}`;
  }
  options.headers = headers;
  const res = await fetch(path, options);
  if (!res.ok) {
    throw new Error(await res.text());
  }
  return res.json();
}

async function refreshPrograms() {
  try {
    const programs = await api("/api/programs");
    renderPrograms(programs);
  } catch (err) {
    programList.innerHTML = `<div class="program">Error: ${err.message}</div>`;
  }
}

function renderPrograms(programs) {
  programList.innerHTML = "";
  programs.forEach((p) => {
    const el = document.createElement("div");
    el.className = "program";

    const info = document.createElement("div");
    info.className = "info";
    info.innerHTML = `<span class="name">${p.name}</span>
      <span class="status ${p.running ? "running" : "stopped"}">
        ${p.running ? `Running (PID ${p.pid})` : "Stopped"}
      </span>`;

    const actions = document.createElement("div");
    actions.className = "actions";

    const startBtn = document.createElement("button");
    startBtn.textContent = "Start";
    startBtn.onclick = () => doAction(p.name, "start");

    const stopBtn = document.createElement("button");
    stopBtn.textContent = "Stop";
    stopBtn.onclick = () => doAction(p.name, "stop");

    const restartBtn = document.createElement("button");
    restartBtn.textContent = "Restart";
    restartBtn.onclick = () => doAction(p.name, "restart");

    actions.append(startBtn, stopBtn, restartBtn);
    el.append(info, actions);
    programList.appendChild(el);
  });
}

async function doAction(name, action) {
  try {
    await api(`/api/programs/${name}/${action}`, { method: "POST" });
    await refreshPrograms();
  } catch (err) {
    alert(`${action} failed: ${err.message}`);
  }
}

saveTokenBtn.addEventListener("click", saveToken);
loadToken();
refreshPrograms();
setInterval(refreshPrograms, 5000);
