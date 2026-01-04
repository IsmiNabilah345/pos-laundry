const API = "http://localhost:8080";
const token = localStorage.getItem("token");

if (!token) {
  window.location.href = "login.html";
}

let posState = {
  selectedService: null,
  qty: 1,
  pelangganId: null,
  tomSelectInstance: null
};

document.addEventListener("DOMContentLoaded", () => {
  const page = localStorage.getItem("currentPage") || "dashboard";
  if (page === "transaksi") loadTransaksi();
  else if (page === "pelanggan") loadPelanggan();
  else if (page === "layanan") loadLayanan();
  else if (page === "laporan") loadLaporan();
  else if (page === "pos") loadPOS();
  else loadDashboard();

  document.getElementById("menu-dashboard").addEventListener("click", () => nav("dashboard", loadDashboard));
  document.getElementById("menu-pelanggan").addEventListener("click", () => nav("pelanggan", loadPelanggan));
  document.getElementById("menu-layanan").addEventListener("click", () => nav("layanan", loadLayanan));
  document.getElementById("menu-transaksi").addEventListener("click", () => nav("transaksi", loadTransaksi));
  document.getElementById("menu-laporan").addEventListener("click", () => nav("laporan", loadLaporan));

  // POS Menu
  document.getElementById("menu-pos").addEventListener("click", () => {
    localStorage.setItem("currentPage", "pos");
    loadPOS();
  });

  document.getElementById("close-pos").addEventListener("click", () => {
    localStorage.setItem("currentPage", "dashboard");
    document.getElementById("pos-view").classList.add("hidden");
    loadDashboard();
  });

  const role = localStorage.getItem("role");
  if (role === "admin") {
    document.getElementById("menu-laporan").classList.remove("hidden");
  } else {
    document.getElementById("menu-laporan").classList.add("hidden");
  }

  document.getElementById("logout").addEventListener("click", () => {
    if (confirm("Keluar dari aplikasi?")) {
      localStorage.removeItem("token");
      localStorage.removeItem("role");
      window.location.href = "login.html";
    }
  });

  // POS Events
  document.getElementById("pos-qty").addEventListener("input", (e) => {
    posState.qty = parseFloat(e.target.value) || 1;
    updatePOSTotal();
  });

  document.getElementById("pos-remove-item").addEventListener("click", () => {
    posState.selectedService = null;
    renderPOSCart();
  });

  document.getElementById("pos-metode").addEventListener("change", (e) => {
    posState.metode_pembayaran = e.target.value;
  });

  document.getElementById("pos-checkout").addEventListener("click", checkoutPOS);
  document.getElementById("pos-search-layanan").addEventListener("input", filterPOSGrid);

  // Add Pelanggan Shortcut
  document.getElementById("pos-add-pelanggan").addEventListener("click", () => {
    openModal("Tambah Pelanggan", `
          <input type="text" id="nama" placeholder="Nama" class="border p-2 w-full mb-2 rounded" required>
          <input type="text" id="telepon" placeholder="Telepon" class="border p-2 w-full mb-2 rounded" required>
          <input type="text" id="alamat" placeholder="Alamat" class="border p-2 w-full mb-2 rounded">
        `, () => {
      const body = {
        nama: document.getElementById("nama").value,
        telepon: document.getElementById("telepon").value,
        alamat: document.getElementById("alamat").value
      };
      fetch(`${API}/pelanggan`, {
        method: "POST",
        headers: { "Content-Type": "application/json", Authorization: "Bearer " + token },
        body: JSON.stringify(body)
      }).then(res => res.json()).then(newP => {
        closeModal();
        // Reload POS to refresh pelanggan list
        loadPOS();
        alert("Pelanggan ditambahkan!");
      });
    });
  });
});

function nav(pageName, loadFunc) {
  localStorage.setItem("currentPage", pageName);
  document.getElementById("pos-view").classList.add("hidden");
  loadFunc();
}

const rupiah = (number) => {
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: "IDR",
    minimumFractionDigits: 0
  }).format(number);
};

// ... existing modal/print functions ...
function openModal(title, formHTML, onSave) {
  document.getElementById("modal-title").innerText = title;
  document.getElementById("modal-form").innerHTML = formHTML;
  document.getElementById("modal").classList.remove("hidden");

  document.getElementById("modal-save").onclick = () => {
    onSave();
  };
  document.getElementById("modal-cancel").onclick = closeModal;
}

function closeModal() {
  document.getElementById("modal").classList.add("hidden");
}

function printStruk(transaksi, namaPelanggan, namaLayanan) {
  const printArea = document.getElementById("print-area");
  const tanggal = transaksi.created_at ? transaksi.created_at.split("T")[0] : new Date().toISOString().split("T")[0];

  printArea.innerHTML = `
    <div style="font-family: monospace; width: 300px; margin: 0 auto; border: 1px solid #ccc; padding: 20px;">
      <h2 style="text-align: center; margin-bottom: 5px;">WELLCLEAN LAUNDRY</h2>
      <p style="text-align: center; font-size: 12px; margin-top: 0;">Jl. Laundry No. 1, Kota Bandung</p>
      <hr style="border-top: 1px dashed black;">
      
      <table style="width: 100%; font-size: 12px;">
        <tr><td>No. Order</td><td style="text-align: right;">${transaksi.kode}</td></tr>
        <tr><td>Tanggal</td><td style="text-align: right;">${tanggal}</td></tr>
        <tr><td>Pelanggan</td><td style="text-align: right;">${namaPelanggan}</td></tr>
        <tr><td>Metode Pembayaran</td><td style="text-align: right;">${transaksi.metode_pembayaran}</td></tr>
      </table>
      
      <hr style="border-top: 1px dashed black;">
      
      <div style="font-size: 12px;">
        <p style="font-weight: bold;">${namaLayanan}</p>
        <div style="display: flex; justify-content: space-between;">
          <span>${transaksi.berat} x Layanan</span>
          <span>${rupiah(transaksi.total)}</span>
        </div>
      </div>
      
      <hr style="border-top: 1px dashed black;">
      
      <div style="display: flex; justify-content: space-between; font-weight: bold; font-size: 14px;">
        <span>TOTAL</span>
        <span>${rupiah(transaksi.total)}</span>
      </div>
      
      <p style="text-align: center; margin-top: 20px; font-size: 10px;">
        Terima Kasih atas kepercayaan Anda.<br>
        Barang yang tidak diambil > 30 hari<br>bukan tanggung jawab kami.
      </p>
    </div>
  `;

  window.print();
}

// POS LOGIC
let allServices = [];

async function loadPOS() {
  document.getElementById("pos-view").classList.remove("hidden");

  try {
    const [pRes, lRes] = await Promise.all([
      fetch(`${API}/pelanggan`, { headers: { Authorization: "Bearer " + token } }),
      fetch(`${API}/layanan`, { headers: { Authorization: "Bearer " + token } })
    ]);

    const pel = await pRes.json();
    const lay = await lRes.json();

    const pelanggan = Array.isArray(pel) ? pel : pel.data || [];
    allServices = Array.isArray(lay) ? lay : lay.data || [];

    // Render Grid
    renderPOSGrid(allServices);

    // Setup TomSelect
    const sel = document.getElementById("pos-pelanggan");
    sel.innerHTML = '<option value="">Pilih Pelanggan...</option>' +
      pelanggan.map(p => `<option value="${p.id}">${p.nama} - ${p.telepon || ''}</option>`).join('');

    if (posState.tomSelectInstance) {
      posState.tomSelectInstance.destroy();
    }
    if (window.TomSelect) {
      posState.tomSelectInstance = new TomSelect("#pos-pelanggan", {
        create: false,
        sortField: { field: "text", direction: "asc" }
      });
    }

  } catch (err) {
    alert("Gagal memuat data POS: " + err.message);
  }
}

function renderPOSGrid(services) {
  const grid = document.getElementById("pos-grid");
  grid.innerHTML = services.map(s => `
        <div class="bg-white p-4 rounded shadow cursor-pointer border hover:border-blue-500 hover:shadow-lg transition flex flex-col justify-between h-32"
             onclick='selectPOSService(${JSON.stringify(s)})'>
            <h3 class="font-bold text-gray-800 line-clamp-2">${s.nama}</h3>
            <div>
                <p class="text-blue-600 font-bold">${rupiah(s.harga)}</p>
                <p class="text-xs text-gray-500">/${s.satuan}</p>
            </div>
        </div>
    `).join('');
}

function filterPOSGrid(e) {
  const key = e.target.value.toLowerCase();
  const filtered = allServices.filter(s => s.nama.toLowerCase().includes(key));
  renderPOSGrid(filtered);
}

function selectPOSService(service) {
  // Single Item Limitation Logic
  posState.selectedService = service;
  posState.qty = 1;
  document.getElementById("pos-qty").value = 1;
  renderPOSCart();
}

// Make selectPOSService global for HTML onclick
window.selectPOSService = selectPOSService;
window.updateQty = (delta) => {
  let newQty = posState.qty + delta;
  if (newQty < 0.1) newQty = 0.1;
  posState.qty = parseFloat(newQty.toFixed(1)); // avoid float errors
  document.getElementById("pos-qty").value = posState.qty;
  updatePOSTotal();
};

function renderPOSCart() {
  const emptyState = document.getElementById("pos-cart-empty");
  const itemState = document.getElementById("pos-cart-item");
  const checkoutBtn = document.getElementById("pos-checkout");

  if (!posState.selectedService) {
    emptyState.classList.remove("hidden");
    itemState.classList.add("hidden");
    checkoutBtn.disabled = true;
    document.getElementById("pos-total").innerText = rupiah(0);
    return;
  }

  emptyState.classList.add("hidden");
  itemState.classList.remove("hidden");
  checkoutBtn.disabled = false;

  document.getElementById("pos-item-name").innerText = posState.selectedService.nama;
  document.getElementById("pos-item-price").innerText = `@ ${rupiah(posState.selectedService.harga)} / ${posState.selectedService.satuan}`;

  updatePOSTotal();
}

function updatePOSTotal() {
  if (!posState.selectedService) return;
  const total = posState.selectedService.harga * posState.qty;
  document.getElementById("pos-total").innerText = rupiah(total);
}

async function checkoutPOS() {
  const pelangganId = document.getElementById("pos-pelanggan").value;
  if (!pelangganId) {
    alert("Mohon pilih pelanggan terlebih dahulu!");
    return;
  }

  if (!posState.selectedService) return;

  const confirmMsg = `Konfirmasi Pembayaran:\n\nLayanan: ${posState.selectedService.nama}\nBerat: ${posState.qty} ${posState.selectedService.satuan}\nTotal: ${document.getElementById("pos-total").innerText}\n\nLanjut proses?`;

  if (!confirm(confirmMsg)) return;

  const total = posState.selectedService.harga * posState.qty;

  const metode = document.getElementById("pos-metode").value;

  const body = {
    kode: "TRX-" + Math.floor(Math.random() * 100000),
    pelanggan_id: pelangganId,
    layanan_id: posState.selectedService.id,
    berat: posState.qty,
    total: total,
    status: "proses",
    metode_pembayaran: metode
  };

  try {
    const res = await fetch(`${API}/transaksi`, {
      method: "POST",
      headers: { "Content-Type": "application/json", Authorization: "Bearer " + token },
      body: JSON.stringify(body)
    });

    if (res.ok) {
      // Success
      // Fetch Transaksi Data just created (for Print) OR just construct mockup
      // Since Supabase doesn't return the full object on POST usually (depends on Prefer header)
      // Let's print immediately with known data

      // Get Pelanggan Name
      const sel = document.getElementById("pos-pelanggan");
      const pName = sel.options[sel.selectedIndex].text.split(" - ")[0];

      const dummyTrans = {
        kode: body.kode,
        created_at: new Date().toISOString(),
        berat: body.berat,
        total: body.total,
        metode_pembayaran: body.metode_pembayaran
      };

      printStruk(dummyTrans, pName, posState.selectedService.nama);

      // Reset Cart
      posState.selectedService = null;
      renderPOSCart();
      alert("Transaksi Berhasil!");
    } else {
      const err = await res.json();
      alert("Gagal: " + (err.error || "Unknown"));
    }
  } catch (e) {
    alert("Error: " + e.message);
  }
}


// --- EXISTING LOAD FUNCTIONS ---
async function loadDashboard() {
  const content = document.getElementById("content");
  content.innerHTML = '<p class="text-gray-500">Memuat statistik...</p>';

  try {
    const res = await fetch(`${API}/dashboard`, {
      headers: { Authorization: "Bearer " + token }
    });
    const data = await res.json();

    const transaksi = data.transaksi || [];
    const userEmail = data.user || "User";

    const totalPendapatan = transaksi.reduce((sum, t) => sum + (parseInt(t.total) || 0), 0);
    const jumlahTransaksi = transaksi.length;
    const transaksiProses = transaksi.filter(t => t.status && t.status.toLowerCase() === 'proses').length;

    content.innerHTML = `
      <h2 class="text-xl font-bold mb-4">Halo, ${userEmail}</h2>
      
      <div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-8">
        <div class="bg-blue-600 text-white p-6 rounded-lg shadow">
          <h3 class="text-lg opacity-90">Total Pendapatan</h3>
          <p class="text-3xl font-bold mt-2">${rupiah(totalPendapatan)}</p>
        </div>
        <div class="bg-green-500 text-white p-6 rounded-lg shadow">
          <h3 class="text-lg opacity-90">Total Transaksi</h3>
          <p class="text-3xl font-bold mt-2">${jumlahTransaksi} Order</p>
        </div>
        <div class="bg-yellow-500 text-white p-6 rounded-lg shadow">
          <h3 class="text-lg opacity-90">Sedang Proses</h3>
          <p class="text-3xl font-bold mt-2">${transaksiProses} Cucian</p>
        </div>
      </div>

      <div class="bg-white rounded shadow overflow-hidden">
        <div class="px-6 py-4 border-b">
          <h3 class="font-bold text-gray-700">5 Transaksi Terakhir</h3>
        </div>
        <table class="min-w-full divide-y divide-gray-200">
          <thead class="bg-gray-50">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Tanggal</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Kode</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Total</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
            </tr>
          </thead>
          <tbody class="bg-white divide-y divide-gray-200">
            ${transaksi.slice(0, 5).map(t => `
              <tr>
                <td class="px-6 py-4 text-sm text-gray-500">${t.created_at ? t.created_at.split('T')[0] : '-'}</td>
                <td class="px-6 py-4 text-sm font-medium text-gray-900">${t.kode || '-'}</td>
                <td class="px-6 py-4 text-sm text-gray-500">${rupiah(t.total)}</td>
                <td class="px-6 py-4 text-sm">
                  <span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${t.status === 'selesai' ? 'bg-green-100 text-green-800' : 'bg-yellow-100 text-yellow-800'
      }">
                    ${t.status}
                  </span>
                </td>
              </tr>
            `).join('')}
          </tbody>
        </table>
      </div>
    `;

  } catch (err) {
    content.innerHTML = `<p class="text-red-500">Error: ${err.message}</p>`;
  }
}

function loadPelanggan() {
  fetch(`${API}/pelanggan`, {
    headers: { Authorization: "Bearer " + token }
  })
    .then(res => res.json())
    .then(data => {
      const rows = Array.isArray(data) ? data : data.data || [];
      const content = document.getElementById("content");

      content.innerHTML = `
        <h2 class="text-xl font-bold mb-4">Data Pelanggan</h2>
        <button id="btn-tambah" class="bg-blue-600 text-white px-4 py-2 rounded mb-4 shadow hover:bg-blue-700">Tambah Pelanggan</button>
        <input type="text" id="search" placeholder="Cari pelanggan..." class="border p-2 w-full mb-4 rounded"/>
        <div id="tableContainer" class="bg-white shadow rounded overflow-hidden">
          ${renderPelangganTable(rows)}
        </div>
      `;

      document.getElementById("search").addEventListener("input", (e) => {
        const key = e.target.value.toLowerCase();
        const filtered = rows.filter(p =>
          p.nama.toLowerCase().includes(key) ||
          (p.telepon || "").toLowerCase().includes(key)
        );
        document.getElementById("tableContainer").innerHTML = renderPelangganTable(filtered);
        attachPelangganEvents(filtered);
      });

      document.getElementById("btn-tambah").addEventListener("click", () => {
        openModal("Tambah Pelanggan", `
          <input type="text" id="nama" placeholder="Nama" class="border p-2 w-full mb-2 rounded" required>
          <input type="text" id="telepon" placeholder="Telepon" class="border p-2 w-full mb-2 rounded" required>
          <input type="text" id="alamat" placeholder="Alamat" class="border p-2 w-full mb-2 rounded">
        `, () => {
          const body = {
            nama: document.getElementById("nama").value,
            telepon: document.getElementById("telepon").value,
            alamat: document.getElementById("alamat").value
          };

          const saveBtn = document.getElementById("modal-save");
          saveBtn.disabled = true;
          saveBtn.innerText = "Menyimpan...";

          fetch(`${API}/pelanggan`, {
            method: "POST",
            headers: { "Content-Type": "application/json", Authorization: "Bearer " + token },
            body: JSON.stringify(body)
          }).then(async (res) => {
            if (!res.ok) {
              const err = await res.json();
              alert("Gagal menambah pelanggan: " + (err.error || err.message || JSON.stringify(err)));
            } else {
              closeModal();
              loadPelanggan();
            }
          }).catch(err => alert("Error: " + err.message))
            .finally(() => {
              if (document.getElementById("modal-save")) {
                document.getElementById("modal-save").disabled = false;
                document.getElementById("modal-save").innerText = "Simpan";
              }
            });
        });
      });

      attachPelangganEvents(rows);
    });
}

function renderPelangganTable(data) {
  if (data.length === 0) return `<p class="p-4 text-gray-500">Data kosong.</p>`;
  return `
    <table class="min-w-full divide-y divide-gray-200">
      <thead class="bg-gray-50">
        <tr>
          <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Nama</th>
          <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Telepon</th>
          <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Alamat</th>
          <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Aksi</th>
        </tr>
      </thead>
      <tbody class="bg-white divide-y divide-gray-200">
        ${data.map(p => `
          <tr>
            <td class="px-6 py-4 whitespace-nowrap">${p.nama}</td>
            <td class="px-6 py-4 whitespace-nowrap">${p.telepon || '-'}</td>
            <td class="px-6 py-4 whitespace-nowrap">${p.alamat || '-'}</td>
            <td class="px-6 py-4 whitespace-nowrap">
              <button class="bg-yellow-500 text-white px-2 py-1 rounded text-xs hover:bg-yellow-600 btn-edit" 
                data-id="${p.id}" data-nama="${p.nama}" data-tel="${p.telepon}" data-alamat="${p.alamat}">Edit</button>
              <button class="bg-red-600 text-white px-2 py-1 rounded text-xs hover:bg-red-700 btn-delete" 
                data-id="${p.id}">Hapus</button>
            </td>
          </tr>
        `).join('')}
      </tbody>
    </table>
  `;
}

function attachPelangganEvents(rows) {
  document.querySelectorAll(".btn-delete").forEach(btn => {
    btn.onclick = () => {
      if (!confirm("Hapus data ini?")) return;
      fetch(`${API}/pelanggan?id=eq.${btn.dataset.id}`, {
        method: "DELETE",
        headers: { Authorization: "Bearer " + token }
      }).then(() => loadPelanggan());
    };
  });

  document.querySelectorAll(".btn-edit").forEach(btn => {
    btn.onclick = () => {
      openModal("Edit Pelanggan", `
        <input type="text" id="nama" value="${btn.dataset.nama}" class="border p-2 w-full mb-2 rounded">
        <input type="text" id="telepon" value="${btn.dataset.tel}" class="border p-2 w-full mb-2 rounded">
        <input type="text" id="alamat" value="${btn.dataset.alamat}" class="border p-2 w-full mb-2 rounded">
      `, () => {
        const body = {
          nama: document.getElementById("nama").value,
          telepon: document.getElementById("telepon").value,
          alamat: document.getElementById("alamat").value
        };
        fetch(`${API}/pelanggan?id=eq.${btn.dataset.id}`, {
          method: "PATCH",
          headers: { "Content-Type": "application/json", Authorization: "Bearer " + token },
          body: JSON.stringify(body)
        }).then(() => { closeModal(); loadPelanggan(); });
      });
    };
  });
}

function loadLayanan() {
  fetch(`${API}/layanan`, {
    headers: { Authorization: "Bearer " + token }
  })
    .then(res => res.json())
    .then(data => {
      const rows = Array.isArray(data) ? data : data.data || [];
      const content = document.getElementById("content");
      const isAdmin = localStorage.getItem("role") === "admin";

      content.innerHTML = `
        <h2 class="text-xl font-bold mb-4">Data Layanan</h2>
        ${isAdmin ? `<button id="btn-tambah" class="bg-blue-600 text-white px-4 py-2 rounded mb-4 shadow hover:bg-blue-700">Tambah Layanan</button>` : '<div class="mb-4 text-gray-500 italic">Mode Kasir: Anda hanya dapat melihat layanan.</div>'}
        <div class="bg-white shadow rounded overflow-hidden">
          ${renderLayananTable(rows, isAdmin)}
        </div>
      `;

      if (isAdmin) {
        document.getElementById("btn-tambah").addEventListener("click", () => {
          openModal("Tambah Layanan", `
            <input type="text" id="nama" placeholder="Nama Layanan" class="border p-2 w-full mb-2 rounded" required>
            <input type="number" id="harga" placeholder="Harga" class="border p-2 w-full mb-2 rounded" required>
            <input type="text" id="satuan" placeholder="Satuan (kg/pcs)" class="border p-2 w-full mb-2 rounded">
          `, () => {
            const body = {
              nama: document.getElementById("nama").value,
              harga: document.getElementById("harga").value,
              satuan: document.getElementById("satuan").value
            };
            fetch(`${API}/layanan`, {
              method: "POST",
              headers: { "Content-Type": "application/json", Authorization: "Bearer " + token },
              body: JSON.stringify(body)
            }).then(() => { closeModal(); loadLayanan(); });
          });
        });
      }

      attachLayananEvents();
    });
}

function renderLayananTable(data, isAdmin) {
  if (data.length === 0) return `<p class="p-4 text-gray-500">Data kosong.</p>`;
  return `
    <table class="min-w-full divide-y divide-gray-200">
      <thead class="bg-gray-50">
        <tr>
          <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Layanan</th>
          <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Harga</th>
          <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Satuan</th>
          ${isAdmin ? `<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Aksi</th>` : ''}
        </tr>
      </thead>
      <tbody class="bg-white divide-y divide-gray-200">
        ${data.map(l => `
          <tr>
            <td class="px-6 py-4 whitespace-nowrap">${l.nama}</td>
            <td class="px-6 py-4 whitespace-nowrap">${rupiah(l.harga)}</td>
            <td class="px-6 py-4 whitespace-nowrap">${l.satuan}</td>
            ${isAdmin ? `
            <td class="px-6 py-4 whitespace-nowrap">
              <button class="bg-red-600 text-white px-2 py-1 rounded text-xs hover:bg-red-700 btn-delete" 
                data-id="${l.id}">Hapus</button>
            </td>` : ''}
          </tr>
        `).join('')}
      </tbody>
    </table>
  `;
}

function attachLayananEvents() {
  document.querySelectorAll(".btn-delete").forEach(btn => {
    btn.onclick = () => {
      if (!confirm("Hapus layanan ini?")) return;
      fetch(`${API}/layanan?id=eq.${btn.dataset.id}`, {
        method: "DELETE",
        headers: { Authorization: "Bearer " + token }
      }).then(() => loadLayanan());
    };
  });
}

function loadTransaksi() {
  Promise.all([
    fetch(`${API}/pelanggan`, { headers: { Authorization: "Bearer " + token } }).then(r => r.json()),
    fetch(`${API}/layanan`, { headers: { Authorization: "Bearer " + token } }).then(r => r.json()),
    fetch(`${API}/transaksi`, { headers: { Authorization: "Bearer " + token } }).then(r => r.json())
  ]).then(([pelangganData, layananData, transaksiData]) => {
    const pelanggan = Array.isArray(pelangganData) ? pelangganData : pelangganData.data || [];
    const layanan = Array.isArray(layananData) ? layananData : layananData.data || [];
    const transaksi = Array.isArray(transaksiData) ? transaksiData : transaksiData.data || [];

    const pelangganMap = {};
    pelanggan.forEach(p => pelangganMap[p.id] = p.nama);
    const layananMap = {};
    layanan.forEach(l => layananMap[l.id] = l.nama);

    const content = document.getElementById("content");
    content.innerHTML = `
      <h2 class="text-xl font-bold mb-4">Data Transaksi</h2>
      <button id="btn-tambah" class="bg-blue-600 text-white px-4 py-2 rounded mb-4 shadow hover:bg-blue-700">Tambah Transaksi</button>
      <div class="bg-white shadow rounded overflow-hidden">
        <table class="min-w-full divide-y divide-gray-200">
          <thead class="bg-gray-50">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Kode</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Pelanggan</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Metode Pembayaran</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Total</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Aksi</th>
            </tr>
          </thead>
          <tbody class="bg-white divide-y divide-gray-200">
            ${transaksi.map(t => `
              <tr>
                <td class="px-6 py-4 whitespace-nowrap text-sm font-medium">${t.kode || '-'}</td>
                <td class="px-6 py-4 whitespace-nowrap text-sm">${pelangganMap[t.pelanggan_id] || '-'}</td>
                <td class="px-6 py-4 whitespace-nowrap text-sm">${t.metode_pembayaran || '-'}</td>
                <td class="px-6 py-4 whitespace-nowrap text-sm">${rupiah(t.total)}</td>
                <td class="px-6 py-4 whitespace-nowrap text-sm">
                  <span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${t.status === 'selesai' ? 'bg-green-100 text-green-800' : 'bg-yellow-100 text-yellow-800'
      }">${t.status}</span>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm space-x-1">
                   <button class="bg-gray-700 text-white px-2 py-1 rounded text-xs hover:bg-gray-800 btn-print" 
                    data-id="${t.id}">Cetak</button>

                   ${t.status !== 'selesai' ? `<button class="bg-green-600 text-white px-2 py-1 rounded text-xs hover:bg-green-700 btn-selesai" data-id="${t.id}">Selesai</button>` : ''}
                   
                   <button class="bg-red-600 text-white px-2 py-1 rounded text-xs hover:bg-red-700 btn-delete" 
                    data-id="${t.id}">Hapus</button>
                </td>
              </tr>
            `).join('')}
          </tbody>
        </table>
      </div>
    `;

    // Event Tambah Transaksi
    document.getElementById("btn-tambah").addEventListener("click", () => {
      openModal("Tambah Transaksi", `
        <label class="block text-sm mb-1">Pelanggan</label>
        <select id="pelanggan" class="border p-2 w-full mb-2 rounded">
          ${pelanggan.map(p => `<option value="${p.id}">${p.nama}</option>`).join('')}
        </select>
        <label class="block text-sm mb-1">Layanan</label>
        <select id="layanan" class="border p-2 w-full mb-2 rounded">
          ${layanan.map(l => `<option value="${l.id}">${l.nama} - ${rupiah(l.harga)}</option>`).join('')}
        </select>
        <label class="block text-sm mb-1">Berat / Jumlah</label>
        <input type="number" id="berat" class="border p-2 w-full mb-2 rounded" placeholder="Contoh: 3" required>
        <label class="block text-sm mb-1">Metode Pembayaran</label>
        <select id="metode_pembayaran" class="border p-2 w-full mb-2 rounded">
          <option value="Cash">Cash</option>
          <option value="QRIS">QRIS</option>
        </select>
        <label class="block text-sm mb-1">Kode Transaksi</label>
        <input type="text" id="kode" class="border p-2 w-full mb-2 rounded" value="TRX-${Math.floor(Math.random() * 10000)}" readonly>
      `, async () => {
        const layId = document.getElementById("layanan").value;
        const berat = parseFloat(document.getElementById("berat").value) || 1;

        const selectedLayanan = layanan.find(l => l.id == layId);
        const total = berat * (selectedLayanan ? selectedLayanan.harga : 0);

        const body = {
          kode: document.getElementById("kode").value,
          pelanggan_id: document.getElementById("pelanggan").value,
          layanan_id: layId,
          berat: berat,
          total: total,
          status: "proses",
          metode_pembayaran: document.getElementById("metode_pembayaran").value
        };

        fetch(`${API}/transaksi`, {
          method: "POST",
          headers: { "Content-Type": "application/json", Authorization: "Bearer " + token },
          body: JSON.stringify(body)
        }).then(() => { closeModal(); loadTransaksi(); });
      });

      // Fix TomSelect
      setTimeout(() => {
        if (window.TomSelect) {
          new TomSelect("#pelanggan");
          new TomSelect("#layanan");
        }
      }, 100);
    });

    // Event Delete
    document.querySelectorAll(".btn-delete").forEach(btn => {
      btn.onclick = () => {
        if (!confirm("Hapus transaksi ini?")) return;
        fetch(`${API}/transaksi?id=eq.${btn.dataset.id}`, {
          method: "DELETE",
          headers: { Authorization: "Bearer " + token }
        }).then(() => loadTransaksi());
      };
    });

    // Event Selesai
    document.querySelectorAll(".btn-selesai").forEach(btn => {
      btn.onclick = () => {
        if (!confirm("Tandai transaksi selesai?")) return;
        fetch(`${API}/transaksi?id=eq.${btn.dataset.id}`, {
          method: "PATCH",
          headers: { "Content-Type": "application/json", Authorization: "Bearer " + token },
          body: JSON.stringify({ status: "selesai" })
        }).then(() => loadTransaksi());
      };
    });

    // Event Cetak Struk (BARU)
    document.querySelectorAll(".btn-print").forEach(btn => {
      btn.onclick = () => {
        const t = transaksi.find(item => item.id == btn.dataset.id);
        const pName = pelangganMap[t.pelanggan_id] || "Umum";
        const lName = layananMap[t.layanan_id] || "Jasa Laundry";
        printStruk(t, pName, lName);
      };
    });

  });
}

function loadLaporan() {
  const content = document.getElementById("content");
  content.innerHTML = '<p class="text-gray-500">Memuat laporan...</p>';

  Promise.all([
    fetch(`${API}/laporan?type=riwayat`, { headers: { Authorization: "Bearer " + token } }).then(r => r.json()),
    fetch(`${API}/laporan?type=keuangan`, { headers: { Authorization: "Bearer " + token } }).then(r => r.json())
  ])
    .then(([riwayatRes, keuanganRes]) => {
      const items = riwayatRes.data || [];
      // const selesaiItems = items.filter(i =>
      //   (i.status || "").trim().toLowerCase() === "selesai"
      // );

      // const totalOmset = keuanganRes.total_omset ?? 0;

      // const today = new Date();
      // today.setHours(0, 0, 0, 0);

      // const daily = selesaiItems.filter(i => {
      //   if (!i.selesai_at) return false;

      //   const selesaiDate = new Date(i.selesai_at);
      //   selesaiDate.setHours(0, 0, 0, 0);

      //   return selesaiDate.getTime() === today.getTime();
      // });

      // const dailyTotal = daily.reduce(
      //   (acc, curr) => acc + (Number(curr.total) || 0),
      //   0
      // );

      const totalOmset = keuanganRes.total_omset || 0;
      const dailyTotal = keuanganRes.omset_harian || 0;

      const todayDate = new Date(); const todayLabel = todayDate.toLocaleDateString("id-ID");

      // console.log(
      //   "DAILY ITEMS:",
      //   daily.map(i => ({
      //     selesai_at: new Date(i.selesai_at).toLocaleString("id-ID"),
      //     total: i.total
      //   }))
      // );

      content.innerHTML = `
        <h2 class="text-xl font-bold mb-4">Laporan Keuangan</h2>
        <div class="flex items-center gap-2 mb-4">
          <select id="filter-keuangan" class="border p-2 rounded">
            <option value="harian">Harian</option>
            <option value="mingguan">Mingguan</option>
            <option value="bulanan">Bulanan</option>
            <option value="tahunan">Tahunan</option>
          </select>
          <button id="btn-export-keuangan" class="bg-purple-600 text-white px-4 py-2 rounded shadow hover:bg-purple-700">
            Export Excel
          </button>
        </div>
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4 mb-6">
           <div class="bg-indigo-600 text-white p-6 rounded shadow">
             <h3>Total Omset (Semua)</h3>
             <p class="text-2xl font-bold">${rupiah(totalOmset)}</p>
           </div>
           <div class="bg-green-600 text-white p-6 rounded shadow">
             <h3>Omset Tanggal Hari ini = (${todayLabel})</h3>
             <p class="text-2xl font-bold">${rupiah(dailyTotal)}</p>
           </div>
        </div>

        <h3 class="font-bold text-lg mb-2">Riwayat Transaksi</h3>
        <div class="flex items-center gap-2 mb-4">
          <select id="filter-periode" class="border p-2 rounded">
            <option value="harian">Harian</option>
            <option value="mingguan">Mingguan</option>
            <option value="bulanan">Bulanan</option>
            <option value="tahunan">Tahunan</option>
          </select>
          <button id="btn-export" class="bg-yellow-500 text-white px-4 py-2 rounded shadow hover:bg-yellow-600">
            Export Excel
          </button>
        </div>
        <div class="bg-white shadow rounded overflow-auto" style="max-height: 400px;">
           <table class="min-w-full divide-y divide-gray-200">
             <thead class="bg-gray-50 sticky top-0">
               <tr>
                 <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Tanggal</th>
                 <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Total</th>
               </tr>
             </thead>
             <tbody class="divide-y divide-gray-200">
              ${items.map(i => `
                <tr>
                  <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    ${i.selesai_at ? new Date(i.selesai_at).toLocaleString("id-ID") : (i.created_at ? new Date(i.created_at).toLocaleString("id-ID") : "-")}
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900 font-medium">
                    ${rupiah(i.total)}
                  </td>
                </tr>
              `).join('')}
            </tbody>
           </table>
        </div>
      `;
      document.getElementById("btn-export").addEventListener("click", () => {
        const periode = document.getElementById("filter-periode").value; // ambil periode
        fetch(`${API}/laporan/export?type=riwayat&periode=${periode}`, {
          headers: { Authorization: "Bearer " + token }
        })
          .then(res => res.blob())
          .then(blob => {
            const a = document.createElement("a");
            a.href = URL.createObjectURL(blob);
            a.download = `Riwayat_Transaksi_${periode}.xlsx`; // biar nama file sesuai periode
            a.click();
          });
      });

      document.getElementById("btn-export-keuangan").addEventListener("click", () => {
        const periode = document.getElementById("filter-keuangan").value; // ambil periode
        fetch(`${API}/laporan/export?type=keuangan&periode=${periode}`, {
          headers: { Authorization: "Bearer " + token }
        })
          .then(res => res.blob())
          .then(blob => {
            const a = document.createElement("a");
            a.href = URL.createObjectURL(blob);
            a.download = `Laporan_Keuangan_${periode}.xlsx`; // nama file sesuai periode
            a.click();
          });
      });
    })
    .catch(err => {
      content.innerHTML = `<p class="text-red-500">Gagal memuat laporan: ${err.message}</p>`;
    });
}

document.addEventListener("DOMContentLoaded", () => {
  const role = localStorage.getItem("role");
  if (role === "admin") {
    document.getElementById("menu-laporan").classList.remove("hidden");
    document.getElementById("menu-user").classList.remove("hidden");
  }

  document.getElementById("menu-user").addEventListener("click", loadUserManagement);
});

async function loadUserManagement() {
  const content = document.getElementById("content");
  content.innerHTML = `
        <div class="flex justify-between items-center mb-6">
            <h2 class="text-2xl font-bold text-gray-800">Manajemen Kasir</h2>
        </div>

        <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
            <div class="bg-gray-50 p-6 rounded-lg border border-gray-200 shadow-sm">
                <h3 class="font-bold text-indigo-700 mb-4">Tambah Kasir Baru</h3>
                <div class="space-y-4">
                    <div>
                        <label class="block text-xs font-bold text-gray-500 mb-1">NAMA LENGKAP</label>
                        <input type="text" id="reg-nama" class="w-full p-2 border rounded focus:ring-2 focus:ring-indigo-500 outline-none">
                    </div>
                    <div>
                        <label class="block text-xs font-bold text-gray-500 mb-1">USERNAME</label>
                        <input type="text" id="reg-user" class="w-full p-2 border rounded focus:ring-2 focus:ring-indigo-500 outline-none">
                    </div>
                    <div>
                        <label class="block text-xs font-bold text-gray-500 mb-1">PASSWORD</label>
                        <input type="password" id="reg-pass" class="w-full p-2 border rounded focus:ring-2 focus:ring-indigo-500 outline-none">
                    </div>
                    <button onclick="prosesSimpanKasir()" class="w-full bg-indigo-600 text-white py-2 rounded font-bold hover:bg-indigo-700 transition shadow">
                        Daftarkan Kasir
                    </button>
                </div>
            </div>

            <div class="lg:col-span-2 bg-white rounded-lg border shadow-sm overflow-hidden">
                <table class="w-full text-left">
                    <thead class="bg-gray-100 border-b">
                        <tr>
                            <th class="p-4 text-sm font-bold text-gray-600">Nama</th>
                            <th class="p-4 text-sm font-bold text-gray-600">Username</th>
                            <th class="p-4 text-sm font-bold text-gray-600">Role</th>
                            <th class="p-4 text-sm font-bold text-gray-600">Aksi</th>
                        </tr>
                    </thead>
                    <tbody id="user-list-body">
                        <tr><td colspan="4" class="p-4 text-center text-gray-400">Memuat data...</td></tr>
                    </tbody>
                </table>
            </div>
        </div>
    `;
  fetchUserList();
}

async function fetchUserList() {
  try {
    const res = await fetch(`${API}/users`, {
      headers: { "Authorization": "Bearer " + localStorage.getItem("token") }
    });

    const data = await res.json();
    console.log("Data dari backend:", data); 

    const tbody = document.getElementById("user-list-body");
    tbody.innerHTML = "";

    if (!Array.isArray(data)) {
      tbody.innerHTML = `<tr><td colspan="4" class="p-4 text-center text-red-500">Gagal memuat: ${data.error || 'Format data salah'}</td></tr>`;
      return;
    }

    data.forEach(u => {
      tbody.innerHTML += `
        <tr>
            <td class="p-4">${u.nama || '-'}</td>
            <td class="p-4">${u.username}</td>
            <td class="p-4">${u.role}</td>
            <td class="p-4">
                <button onclick="editUser('${u.id}', '${u.nama}', '${u.role}')" class="text-blue-500 mr-2">Edit</button>
                ${u.role !== 'admin' ? `<button onclick="hapusUser('${u.id}')" class="text-red-500">Hapus</button>` : ''}
            </td>
        </tr>
    `;
    });
  } catch (err) {
    console.error("Fetch error:", err);
  }
}

async function prosesSimpanKasir() {
  const nama = document.getElementById("reg-nama").value;
  const username = document.getElementById("reg-user").value;
  const password = document.getElementById("reg-pass").value;

  if (!nama || !username || !password) return alert("Harap isi semua kolom!");

  const res = await fetch(`${API}/register`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "Authorization": "Bearer " + localStorage.getItem("token")
    },
    body: JSON.stringify({ nama, username, password })
  });

  if (res.ok) {
    alert("Kasir berhasil didaftarkan!");
    loadUserManagement();
  } else {
    alert("Gagal mendaftarkan user. Pastikan username belum dipakai.");
  }
}

async function hapusUser(id) {
  if (!confirm("Yakin ingin menghapus user ini?")) return;
  const res = await fetch(`${API}/users/delete?id=${id}`, {
    method: "DELETE",
    headers: { "Authorization": "Bearer " + localStorage.getItem("token") }
  });
  if (res.ok) {
    alert("User dihapus!");
    loadUserManagement();
  }
}

async function editUser(id, namaLama, roleLama) {
  const nama = prompt("Nama Baru:", namaLama);
  const role = prompt("Role Baru (admin/kasir):", roleLama);

  if (!nama || !role) return;

  const res = await fetch(`${API}/users/update?id=${id}`, {
    method: "PATCH",
    headers: {
      "Content-Type": "application/json",
      "Authorization": "Bearer " + localStorage.getItem("token")
    },
    body: JSON.stringify({ nama, role })
  });

  if (res.ok) {
    alert("User diupdate!");
    loadUserManagement();
  }
}