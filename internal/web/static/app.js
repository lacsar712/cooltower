async function api(path, opts = {}) {
  const res = await fetch('/api' + path, opts);
  if (!res.ok) throw new Error(await res.text());
  return res.json();
}

function setText(id, text) {
  const el = document.getElementById(id);
  if (el) el.textContent = text;
}

async function refresh() {
  try {
    const [health, tel, alarms] = await Promise.all([
      api('/health'),
      api('/telemetry'),
      api('/alarms'),
    ]);
    setText('tower-id', health.tower || '—');
    setText('state', tel.state || '—');
    setText('drift', tel.drift_ppm != null ? tel.drift_ppm.toFixed(1) + ' ppm' : '—');
    setText('fan-speed', tel.fan_speed != null ? tel.fan_speed.toFixed(0) + '%' : '—');
    setText('spray-flow', tel.spray_gpm != null ? tel.spray_gpm.toFixed(1) + ' gpm' : '—');
    setText('basin-temp', tel.basin_temp != null ? tel.basin_temp.toFixed(1) + ' °C' : '—');
    const list = document.getElementById('alarms');
    list.innerHTML = '';
    if (!alarms || alarms.length === 0) {
      const li = document.createElement('li');
      li.className = 'empty';
      li.textContent = 'No active alarms';
      list.appendChild(li);
    } else {
      alarms.forEach(a => {
        const li = document.createElement('li');
        li.textContent = a.code + ': ' + a.message;
        list.appendChild(li);
      });
    }
  } catch (e) {
    console.error(e);
  }
}

document.getElementById('btn-start').addEventListener('click', async () => {
  try {
    await api('/tower/start', { method: 'POST' });
    await refresh();
  } catch (e) {
    alert('Start failed: ' + e.message);
  }
});

document.getElementById('btn-stop').addEventListener('click', async () => {
  try {
    await api('/tower/stop', { method: 'POST' });
    await refresh();
  } catch (e) {
    alert('Stop failed: ' + e.message);
  }
});

refresh();
setInterval(refresh, 5000);
