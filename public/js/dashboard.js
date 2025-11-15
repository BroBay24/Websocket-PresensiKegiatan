const tabelBody = document.getElementById('tabel-body');
const totalHadirEl = document.getElementById('total-hadir');
const exportBtn = document.getElementById('export-btn');

let kehadiranData = [];
let socket;
let reconnectTimeout;

function websocketURL() {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  return `${protocol}//${window.location.host}/ws?role=dashboard`;
}

function formatDateTime(value) {
  if (!value) return '-';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleString('id-ID', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  });
}

function renderTable(data) {
  if (!data.length) {
    tabelBody.innerHTML = '<tr><td colspan="6" class="placeholder">Belum ada data presensi.</td></tr>';
    return;
  }

  const rows = data
    .map((row, index) => `
      <tr>
        <td>${index + 1}</td>
        <td>${row.nama}</td>
        <td>${row.nim}</td>
        <td>${row.jurusan}</td>
        <td>${row.angkatan}</td>
        <td>${formatDateTime(row.waktu)}</td>
      </tr>
    `)
    .join('');

  tabelBody.innerHTML = rows;
}

function updateTotal(total) {
  totalHadirEl.textContent = total;
}

async function loadInitialData() {
  try {
    const response = await fetch('/api/kehadiran');
    const result = await response.json();

    if (result.success) {
      kehadiranData = result.data || [];
      renderTable(kehadiranData);
      updateTotal(result.total ?? kehadiranData.length);
    } else {
      throw new Error('Gagal mengambil data awal');
    }
  } catch (error) {
    console.error('Gagal memuat data awal:', error);
    tabelBody.innerHTML = '<tr><td colspan="6" class="placeholder">Gagal memuat data awal. Coba muat ulang halaman.</td></tr>';
  }
}

function handleMessage(event) {
  try {
    const message = JSON.parse(event.data);
    if (message.type === 'attendance:init') {
      const data = message.payload?.data || [];
      kehadiranData = data;
      renderTable(kehadiranData);
      updateTotal(message.payload?.total ?? data.length);
    } else if (message.type === 'attendance:new') {
      const record = message.payload?.record;
      const total = message.payload?.total;
      if (record) {
        kehadiranData = [record, ...kehadiranData];
        renderTable(kehadiranData);
      }
      if (typeof total === 'number') {
        updateTotal(total);
      }
    }
  } catch (error) {
    console.error('Gagal mengolah pesan websocket:', error);
  }
}

function createSocket() {
  if (reconnectTimeout) {
    clearTimeout(reconnectTimeout);
  }

  socket = new WebSocket(websocketURL());
  socket.addEventListener('message', handleMessage);

  socket.addEventListener('close', () => {
    reconnectTimeout = setTimeout(createSocket, 2000);
  });

  socket.addEventListener('error', () => {
    socket.close();
  });
}

exportBtn.addEventListener('click', () => {
  window.open('/export', '_blank');
});

createSocket();
loadInitialData();
