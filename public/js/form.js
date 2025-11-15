const form = document.getElementById('presensi-form');
const statusMessage = document.getElementById('status-message');
const submitBtn = document.getElementById('submit-btn');

let socket;
let reconnectTimeout;

function websocketURL() {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  return `${protocol}//${window.location.host}/ws?role=form`;
}

function setStatus(message, type = 'info') {
  statusMessage.textContent = message;
  statusMessage.className = `status status--${type}`;
}

function disableForm(disabled) {
  submitBtn.disabled = disabled;
  submitBtn.textContent = disabled ? 'Mengirim...' : 'Kirim Presensi';
}

function ensureSocketReady(callback) {
  if (!socket || socket.readyState === WebSocket.CLOSED) {
    createSocket();
  }

  if (socket.readyState === WebSocket.OPEN) {
    callback();
    return;
  }

  const handleOpen = () => {
    socket.removeEventListener('open', handleOpen);
    callback();
  };

  socket.addEventListener('open', handleOpen, { once: true });
}

function createSocket() {
  if (reconnectTimeout) {
    clearTimeout(reconnectTimeout);
  }

  socket = new WebSocket(websocketURL());

  socket.addEventListener('open', () => {
    setStatus('Silakan isi data presensi Anda.', 'info');
  });

  socket.addEventListener('message', (event) => {
    try {
      const response = JSON.parse(event.data);
      if (response.type === 'attendance:ack') {
        const { success, message } = response.payload || {};
        if (success) {
          setStatus(message || 'Presensi berhasil dicatat.', 'success');
          form.reset();
        } else {
          setStatus(message || 'Gagal mencatat presensi.', 'warning');
        }
        disableForm(false);
      }
    } catch (error) {
      console.error('Gagal memproses pesan dari server:', error);
    }
  });

  socket.addEventListener('close', () => {
    setStatus('Koneksi ke server terputus. Mencoba menghubungkan ulang...', 'warning');
    disableForm(false);
    reconnectTimeout = setTimeout(createSocket, 2000);
  });

  socket.addEventListener('error', () => {
    setStatus('Terjadi kesalahan koneksi. Mencoba ulang...', 'error');
    disableForm(false);
    socket.close();
  });
}

form.addEventListener('submit', (event) => {
  event.preventDefault();

  const payload = {
    nama: form.nama.value.trim(),
    nim: form.nim.value.trim(),
    jurusan: form.jurusan.value.trim(),
    angkatan: form.angkatan.value.trim()
  };

  if (!payload.nama || !payload.nim || !payload.jurusan || !payload.angkatan) {
    setStatus('Semua field wajib diisi.', 'error');
    return;
  }

  disableForm(true);
  setStatus('Mengirim data ke server...', 'info');

  ensureSocketReady(() => {
    socket.send(JSON.stringify({
      type: 'attendance:submit',
      payload
    }));
  });
});

createSocket();
