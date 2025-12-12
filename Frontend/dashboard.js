document.addEventListener("DOMContentLoaded", () => {
  const token = localStorage.getItem("token");
  if (!token) {
    window.location.href = "login.html";
    return;
  }

  // 🔧 cek halaman terakhir dari localStorage
  const page = localStorage.getItem("currentPage") || "dashboard";
  if (page === "transaksi") {
    loadTransaksi(token);
  } else if (page === "pelanggan") {
    loadPelanggan(token);
  } else if (page === "layanan") {
    loadLayanan(token);
  } else {
    loadDashboard(token);
  }

  // menu handler tetap sama
  document.getElementById("menu-transaksi").addEventListener("click", () => {
    localStorage.setItem("currentPage", "transaksi");
    loadTransaksi(token);
  });

  document.getElementById("menu-pelanggan").addEventListener("click", () => {
    localStorage.setItem("currentPage", "pelanggan");
    loadPelanggan(token);
  });

  document.getElementById("menu-layanan").addEventListener("click", () => {
    localStorage.setItem("currentPage", "layanan");
    loadLayanan(token);
  });

  document.getElementById("menu-dashboard").addEventListener("click", () => {
    localStorage.setItem("currentPage", "dashboard");
    loadDashboard(token);
  });

  document.getElementById("logout").addEventListener("click", () => {
    localStorage.removeItem("token");
    window.location.href = "login.html";
  });
});

function openModal(title, formHTML, onSave) {
  document.getElementById("modal-title").innerText = title;
  document.getElementById("modal-form").innerHTML = formHTML;
  document.getElementById("modal").classList.remove("hidden");

  document.getElementById("modal-save").onclick = () => {
    onSave();
    closeModal();
  };
  document.getElementById("modal-cancel").onclick = closeModal;
}

function closeModal() {
  document.getElementById("modal").classList.add("hidden");
}

function loadDashboard(token) {
  fetch("http://localhost:8080/dashboard", {
    headers: { Authorization: "Bearer " + token }
  })
    .then(res => res.json())
    .then(data => {
      document.getElementById("content").innerHTML = `
        <h2 class="text-xl font-bold mb-2">Dashboard</h2>
        <p><strong>User:</strong> ${data.user}</p>
        <h3 class="mt-4 font-semibold">Transaksi</h3>
        <ul class="list-disc pl-5">
          ${data.data.map(item => `<li>${item}</li>`).join("")}
        </ul>
      `;
    });
}

document.addEventListener("DOMContentLoaded", () => {
  const token = localStorage.getItem("token");
  if (!token) {
    window.location.href = "login.html";
    return;
  }

  loadDashboard(token);

  document.getElementById("menu-dashboard").addEventListener("click", () => loadDashboard(token));
  document.getElementById("menu-pelanggan").addEventListener("click", () => loadPelanggan(token));
  document.getElementById("menu-transaksi").addEventListener("click", () => loadTransaksi(token));

  document.getElementById("logout").addEventListener("click", () => {
    localStorage.removeItem("token");
    window.location.href = "login.html";
  });
});

function loadDashboard(token) {
  fetch("http://localhost:8080/dashboard", {
    headers: { Authorization: "Bearer " + token }
  })
    .then(res => res.json())
    .then(data => {
      document.getElementById("content").innerHTML = `
        <h2 class="text-xl font-bold mb-2">Dashboard</h2>
        <p><strong>User:</strong> ${data.user}</p>
        <h3 class="mt-4 font-semibold">Transaksi</h3>
        <ul class="list-disc pl-5">
          ${data.data.map(item => `<li>${item}</li>`).join("")}
        </ul>
      `;
    });
}

function loadPelanggan(token) {
  fetch("http://localhost:8080/pelanggan", {
    headers: { Authorization: "Bearer " + token }
  })
    .then(res => res.json())
    .then(data => {
      const rows = Array.isArray(data) ? data : data.data || [];

      // Render awal: tombol tambah + search + tabel
      document.getElementById("content").innerHTML = `
        <h2 class="text-xl font-bold mb-2">Data Pelanggan</h2>
        <button id="btn-tambah-pelanggan" class="bg-blue-600 text-white px-4 py-2 rounded mb-4">Tambah Pelanggan</button>
        <input type="text" id="searchInput" placeholder="Cari pelanggan..." class="border p-2 w-full mb-4"/>
        <div id="tableContainer">
          ${renderTable(rows)}
        </div>
      `;

      // Event search
      document.getElementById("searchInput").addEventListener("input", (e) => {
        const keyword = e.target.value.toLowerCase();

        const filtered = rows.filter(p => {
          return (
            p.nama.toLowerCase().includes(keyword) ||
            (p.telepon ?? "").toLowerCase().includes(keyword) ||
            (p.alamat ?? "").toLowerCase().includes(keyword)
          );
        });

        document.getElementById("tableContainer").innerHTML = renderTable(filtered);
      });

      // Fungsi render tabel pelanggan
      function renderTable(data) {
        if (data.length === 0) {
          return `<p class="text-gray-600">Belum ada pelanggan.</p>`;
        }
        return `
          <table class="min-w-full border border-gray-300">
            <thead class="bg-gray-200">
              <tr>
                <th class="px-4 py-2 border">Nama</th>
                <th class="px-4 py-2 border">Telepon</th>
                <th class="px-4 py-2 border">Alamat</th>
                <th class="px-4 py-2 border">Aksi</th>
              </tr>
            </thead>
            <tbody>
              ${data.map(p => `
                <tr>
                  <td class="px-4 py-2 border">${p.nama}</td>
                  <td class="px-4 py-2 border">${p.telepon}</td>
                  <td class="px-4 py-2 border">${p.alamat}</td>
                  <td class="px-4 py-2 border">
                    <button class="bg-yellow-500 text-white px-2 py-1 rounded btn-edit"
                      data-id="${p.id}" data-nama="${p.nama}" data-telepon="${p.telepon}" data-alamat="${p.alamat}">
                      Edit
                    </button>
                    <button class="bg-red-600 text-white px-2 py-1 rounded btn-delete" data-id="${p.id}">Delete</button>
                  </td>
                </tr>
              `).join("")}
            </tbody>
          </table>
        `;
      }

      // Event Tambah
      document.getElementById("btn-tambah-pelanggan").addEventListener("click", () => {
        openModal("Tambah Pelanggan", `
          <input type="text" id="nama" placeholder="Nama" class="border p-2 w-full" required>
          <input type="text" id="telepon" placeholder="Telepon" class="border p-2 w-full" required>
          <input type="text" id="alamat" placeholder="Alamat" class="border p-2 w-full">
        `, () => {
          const newData = {
            nama: document.getElementById("nama").value,
            telepon: document.getElementById("telepon").value,
            alamat: document.getElementById("alamat").value
          };
          fetch("http://localhost:8080/pelanggan", {
            method: "POST",
            headers: { "Content-Type": "application/json", Authorization: "Bearer " + token },
            body: JSON.stringify(newData)
          }).then(() => { closeModal(); loadPelanggan(token); });
        });
      });

      // Event Delete
      document.querySelectorAll(".btn-delete").forEach(btn => {
        btn.addEventListener("click", () => {
          const id = btn.dataset.id;
          if (!confirm("Yakin mau hapus pelanggan ini?")) return;
          fetch(`http://localhost:8080/pelanggan?id=eq.${id}`, {
            method: "DELETE",
            headers: { Authorization: "Bearer " + token }
          }).then(() => loadPelanggan(token));
        });
      });

      // Event Edit
      document.querySelectorAll(".btn-edit").forEach(btn => {
        btn.addEventListener("click", () => {
          const id = btn.dataset.id;
          openModal("Edit Pelanggan", `
            <input type="text" id="nama" value="${btn.dataset.nama}" class="border p-2 w-full" required>
            <input type="text" id="telepon" value="${btn.dataset.telepon}" class="border p-2 w-full" required>
            <input type="text" id="alamat" value="${btn.dataset.alamat}" class="border p-2 w-full">
          `, () => {
            const updated = {
              nama: document.getElementById("nama").value,
              telepon: document.getElementById("telepon").value,
              alamat: document.getElementById("alamat").value
            };
            fetch(`http://localhost:8080/pelanggan?id=eq.${id}`, {
              method: "PATCH",
              headers: { "Content-Type": "application/json", Authorization: "Bearer " + token },
              body: JSON.stringify(updated)
            }).then(() => { closeModal(); loadPelanggan(token); });
          });
        });
      });
    });
}

function loadLayanan(token) {
  fetch("http://localhost:8080/layanan", {
    headers: { Authorization: "Bearer " + token }
  })
    .then(res => res.json())
    .then(data => {
      const rows = Array.isArray(data) ? data : data.data || [];

      // Render awal: tombol tambah + search + tabel
      document.getElementById("content").innerHTML = `
        <h2 class="text-xl font-bold mb-2">Data Layanan</h2>
        <button id="btn-tambah-layanan" class="bg-blue-600 text-white px-4 py-2 rounded mb-4">Tambah Layanan</button>
        <input type="text" id="searchInput" placeholder="Cari layanan..." class="border p-2 w-full mb-4"/>
        <div id="tableContainer">
          ${renderTable(rows)}
        </div>
      `;

      // Event search
      document.getElementById("searchInput").addEventListener("input", (e) => {
        const keyword = e.target.value.toLowerCase();

        const filtered = rows.filter(l => {
          return (
            l.nama.toLowerCase().includes(keyword) ||
            String(l.harga).toLowerCase().includes(keyword) ||
            (l.satuan ?? "").toLowerCase().includes(keyword) ||
            (l.deskripsi ?? "").toLowerCase().includes(keyword)
          );
        });

        document.getElementById("tableContainer").innerHTML = renderTable(filtered);
      });

      // Fungsi render tabel layanan
      function renderTable(data) {
        if (data.length === 0) {
          return `<p class="text-gray-600">Belum ada layanan.</p>`;
        }
        return `
          <table class="min-w-full border border-gray-300">
            <thead class="bg-gray-200">
              <tr>
                <th class="px-4 py-2 border">Nama</th>
                <th class="px-4 py-2 border">Harga</th>
                <th class="px-4 py-2 border">Satuan</th>
                <th class="px-4 py-2 border">Deskripsi</th>
                <th class="px-4 py-2 border">Aksi</th>
              </tr>
            </thead>
            <tbody>
              ${data.map(l => `
                <tr>
                  <td class="px-4 py-2 border">${l.nama}</td>
                  <td class="px-4 py-2 border">Rp ${l.harga}</td>
                  <td class="px-4 py-2 border">${l.satuan}</td>
                  <td class="px-4 py-2 border">${l.deskripsi || "-"}</td>
                  <td class="px-4 py-2 border">
                    <button class="bg-yellow-500 text-white px-2 py-1 rounded btn-edit"
                      data-id="${l.id}" data-nama="${l.nama}" data-harga="${l.harga}"
                      data-satuan="${l.satuan}" data-deskripsi="${l.deskripsi || ""}">
                      Edit
                    </button>
                    <button class="bg-red-600 text-white px-2 py-1 rounded btn-delete" data-id="${l.id}">Delete</button>
                  </td>
                </tr>
              `).join("")}
            </tbody>
          </table>
        `;
      }

      // Event Tambah
      document.getElementById("btn-tambah-layanan").addEventListener("click", () => {
        openModal("Tambah Layanan", `
          <input type="text" id="nama" placeholder="Nama" class="border p-2 w-full" required>
          <input type="number" id="harga" placeholder="Harga" class="border p-2 w-full" required>
          <input type="text" id="satuan" placeholder="Satuan" class="border p-2 w-full" required>
          <input type="text" id="deskripsi" placeholder="Deskripsi" class="border p-2 w-full">
        `, () => {
          const newData = {
            nama: document.getElementById("nama").value,
            harga: document.getElementById("harga").value,
            satuan: document.getElementById("satuan").value,
            deskripsi: document.getElementById("deskripsi").value
          };
          fetch("http://localhost:8080/layanan", {
            method: "POST",
            headers: { "Content-Type": "application/json", Authorization: "Bearer " + token },
            body: JSON.stringify(newData)
          }).then(() => {
            closeModal();
            loadLayanan(token);
          });
        });
      });

      // Event Delete
      document.querySelectorAll(".btn-delete").forEach(btn => {
        btn.addEventListener("click", () => {
          const id = btn.dataset.id;
          if (!confirm("Yakin mau hapus layanan ini?")) return;
          fetch(`http://localhost:8080/layanan?id=eq.${id}`, {
            method: "DELETE",
            headers: { Authorization: "Bearer " + token }
          }).then(() => loadLayanan(token));
        });
      });

      // Event Edit
      document.querySelectorAll(".btn-edit").forEach(btn => {
        btn.addEventListener("click", () => {
          const id = btn.dataset.id;
          openModal("Edit Layanan", `
            <input type="text" id="nama" value="${btn.dataset.nama}" class="border p-2 w-full" required>
            <input type="number" id="harga" value="${btn.dataset.harga}" class="border p-2 w-full" required>
            <input type="text" id="satuan" value="${btn.dataset.satuan}" class="border p-2 w-full" required>
            <input type="text" id="deskripsi" value="${btn.dataset.deskripsi}" class="border p-2 w-full">
          `, () => {
            const updated = {
              nama: document.getElementById("nama").value,
              harga: document.getElementById("harga").value,
              satuan: document.getElementById("satuan").value,
              deskripsi: document.getElementById("deskripsi").value
            };
            fetch(`http://localhost:8080/layanan?id=eq.${id}`, {
              method: "PATCH",
              headers: { "Content-Type": "application/json", Authorization: "Bearer " + token },
              body: JSON.stringify(updated)
            }).then(() => {
              closeModal();
              loadLayanan(token);
            });
          });
        });
      });
    });
}

function loadTransaksi(token) {
  Promise.all([
    fetch("http://localhost:8080/pelanggan", { headers: { Authorization: "Bearer " + token } }).then(res => res.json()),
    fetch("http://localhost:8080/layanan", { headers: { Authorization: "Bearer " + token } }).then(res => res.json()),
    fetch("http://localhost:8080/transaksi", { headers: { Authorization: "Bearer " + token } }).then(res => res.json())
  ]).then(([pelangganRes, layananRes, transaksiRes]) => {
    const pelangganRows = Array.isArray(pelangganRes) ? pelangganRes : (pelangganRes.data || []);
    const layananRows = Array.isArray(layananRes) ? layananRes : (layananRes.data || []);
    const rows = Array.isArray(transaksiRes) ? transaksiRes : (transaksiRes.data || []);

    // Map ID -> Nama
    const pelangganMap = {};
    pelangganRows.forEach(p => { pelangganMap[p.id] = p.nama; });

    const layananMap = {};
    layananRows.forEach(l => { layananMap[l.id] = l.nama; });

    // Render awal: tombol tambah + search + tabel
    document.getElementById("content").innerHTML = `
      <h2 class="text-xl font-bold mb-2">Data Transaksi</h2>
      <button id="btn-tambah-transaksi" class="bg-blue-600 text-white px-4 py-2 rounded mb-4">Tambah Transaksi</button>
      <input type="text" id="searchInput" placeholder="Cari transaksi..." class="border p-2 w-full mb-4"/>
      <div id="tableContainer">
        ${renderTable(rows)}
      </div>
    `;

    // Event search
    document.getElementById("searchInput").addEventListener("input", (e) => {
      const keyword = e.target.value.toLowerCase();

      const filtered = rows.filter(t => {
        const pelangganNama = pelangganMap[t.pelanggan_id] || "";
        const layananNama = layananMap[t.layanan_id] || "";
        return (
          (t.kode ?? "").toLowerCase().includes(keyword) ||
          pelangganNama.toLowerCase().includes(keyword) ||
          layananNama.toLowerCase().includes(keyword) ||
          (t.metode_pembayaran ?? "").toLowerCase().includes(keyword) ||
          (t.status ?? "").toLowerCase().includes(keyword)
        );
      });

      document.getElementById("tableContainer").innerHTML = renderTable(filtered);
    });

    // Fungsi render tabel transaksi
    function renderTable(data) {
      if (data.length === 0) {
        return `<p class="text-gray-600">Belum ada transaksi.</p>`;
      }
      return `
        <table class="min-w-full border border-gray-300">
          <thead class="bg-gray-200">
            <tr>
              <th class="px-4 py-2 border">Kode</th>
              <th class="px-4 py-2 border">Pelanggan</th>
              <th class="px-4 py-2 border">Layanan</th>
              <th class="px-4 py-2 border">Tanggal</th>
              <th class="px-4 py-2 border">Berat</th>
              <th class="px-4 py-2 border">Total</th>
              <th class="px-4 py-2 border">Metode Pembayaran</th>
              <th class="px-4 py-2 border">Status</th>
              <th class="px-4 py-2 border">Aksi</th>
            </tr>
          </thead>
          <tbody>
            ${data.map(t => `
              <tr>
                <td class="px-4 py-2 border">${t.kode ?? ""}</td>
                <td class="px-4 py-2 border">${pelangganMap[t.pelanggan_id] || t.pelanggan_id}</td>
                <td class="px-4 py-2 border">${layananMap[t.layanan_id] || t.layanan_id}</td>
                <td class="px-4 py-2 border">${t.tanggal || ""}</td>
                <td class="px-4 py-2 border">${t.berat ?? ""}</td>
                <td class="px-4 py-2 border">Rp ${t.total ?? 0}</td>
                <td class="px-4 py-2 border">${t.metode_pembayaran || ""}</td>
                <td class="px-4 py-2 border">${t.status || ""}</td>
                <td class="px-4 py-2 border">
                <button class="bg-yellow-500 text-white px-2 py-1 rounded btn-edit"
                  data-id="${t.id}"
                  data-pelanggan="${t.pelanggan_id}"
                  data-layanan="${t.layanan_id}"
                  data-berat="${t.berat ?? ""}"
                  data-tanggal="${t.tanggal || ""}"
                  data-total="${t.total ?? ""}"
                  data-metode="${t.metode_pembayaran || ""}"
                  data-status="${t.status || ""}">
                  Edit
                </button>
                <button class="bg-red-600 text-white px-2 py-1 rounded btn-delete" data-id="${t.id}">Delete</button>
                <button class="bg-green-600 text-white px-2 py-1 rounded btn-detail" data-id="${t.id}">Detail</button>
                </td>
              </tr>
            `).join("")}
          </tbody>
        </table>
      `;
    }
    // Tambah transaksi
    document.getElementById("btn-tambah-transaksi").addEventListener("click", () => {
      const pelangganOptions = pelangganRows
        .map(p => `<option value="${p.id}">${p.nama}</option>`).join("");

      const layananOptions = layananRows
        .map(l => `<option value="${l.id}">${l.nama}</option>`).join("");

      openModal("Tambah Transaksi", `
        <label class="block mb-1">Pelanggan</label>
  <select id="pelanggan" class="border p-2 w-full mb-2">
    ${pelangganRows.map(p => `<option value="${p.id}">${p.nama}</option>`).join("")}
  </select>

  <label class="block mb-1">Layanan</label>
  <select id="layanan" class="border p-2 w-full mb-2">
    ${layananRows.map(l => `<option value="${l.id}">${l.nama}</option>`).join("")}
  </select>

        <label class="block mb-1">Berat / Jumlah</label>
        <input type="number" id="berat" class="border p-2 w-full mb-2" required>

        <label class="block mb-1">Tanggal</label>
        <input type="date" id="tanggal" class="border p-2 w-full mb-2" required>

        <label class="block mb-1">Metode Pembayaran</label>
        <input type="text" id="metode_pembayaran" placeholder="Cash / Qris" class="border p-2 w-full mb-2">

        <label class="block mb-1">Status</label>
        <input type="text" id="status" placeholder="(kosongkan untuk default: proses)" class="border p-2 w-full mb-2">
      `, async () => {
        const layananId = document.getElementById("layanan").value;
        const berat = parseFloat(document.getElementById("berat").value);

        // Detail layanan untuk satuan & harga
        const layananData = await fetch(`http://localhost:8080/layanan?id=eq.${layananId}`, {
          headers: { Authorization: "Bearer " + token }
        }).then(res => res.json());
        const layananDetail = Array.isArray(layananData) ? layananData[0] : (layananData.data?.[0]);

        const harga = layananDetail?.harga ?? 0;
        const total = berat * harga;

        const statusVal = document.getElementById("status").value;
        const newData = {
          pelanggan_id: document.getElementById("pelanggan").value,
          layanan_id: layananId,
          berat: berat,
          tanggal: document.getElementById("tanggal").value,
          total: total,
          metode_pembayaran: document.getElementById("metode_pembayaran").value,
          ...(statusVal ? { status: statusVal } : { status: "proses" }) // default kalau kosong
        };

        fetch("http://localhost:8080/transaksi", {
          method: "POST",
          headers: { "Content-Type": "application/json", Authorization: "Bearer " + token },
          body: JSON.stringify(newData)
        }).then(() => { closeModal(); loadTransaksi(token); });
      });
      setTimeout(() => {
        new TomSelect("#pelanggan");
        new TomSelect("#layanan");
      }, 0);

    });

    // Delete transaksi
    document.querySelectorAll(".btn-delete").forEach(btn => {
      btn.addEventListener("click", () => {
        const id = btn.dataset.id;
        if (!confirm("Yakin mau hapus transaksi ini?")) return;
        fetch(`http://localhost:8080/transaksi?id=eq.${id}`, {
          method: "DELETE",
          headers: { Authorization: "Bearer " + token }
        }).then(() => loadTransaksi(token));
      });
    });

    // Edit transaksi
    document.querySelectorAll(".btn-edit").forEach(btn => {
      btn.addEventListener("click", () => {
        const id = btn.dataset.id;

        // Build dropdown dengan auto-select
        const pelangganOptions = pelangganRows
          .map(p => `<option value="${p.id}" ${p.id === btn.dataset.pelanggan ? "selected" : ""}>${p.nama}</option>`).join("");

        const layananOptions = layananRows
          .map(l => `<option value="${l.id}" ${l.id === btn.dataset.layanan ? "selected" : ""}>${l.nama}</option>`).join("");

        openModal("Edit Transaksi", `
          <label class="block mb-1">Pelanggan</label>
          <select id="pelanggan" class="border p-2 w-full mb-2">${pelangganOptions}</select>

          <label class="block mb-1">Layanan</label>
          <select id="layanan" class="border p-2 w-full mb-2">${layananOptions}</select>

          <label class="block mb-1">Berat / Jumlah</label>
          <input type="number" id="berat" value="${btn.dataset.berat}" class="border p-2 w-full mb-2" required>

          <label class="block mb-1">Tanggal</label>
          <input type="date" id="tanggal" value="${btn.dataset.tanggal}" class="border p-2 w-full mb-2" required>

          <label class="block mb-1">Total</label>
          <input type="number" id="total" value="${btn.dataset.total}" class="border p-2 w-full mb-2" required>

          <label class="block mb-1">Metode Pembayaran</label>
          <input type="text" id="metode_pembayaran" value="${btn.dataset.metode}" class="border p-2 w-full mb-2">

          <label class="block mb-1">Status</label>
          <input type="text" id="status" value="${btn.dataset.status}" class="border p-2 w-full mb-2" placeholder="(kosongkan untuk default: proses)">
        `, () => {
          const statusVal = document.getElementById("status").value;

          const updated = {
            pelanggan_id: document.getElementById("pelanggan").value,
            layanan_id: document.getElementById("layanan").value,
            berat: parseFloat(document.getElementById("berat").value),
            tanggal: document.getElementById("tanggal").value,
            total: parseFloat(document.getElementById("total").value),
            metode_pembayaran: document.getElementById("metode_pembayaran").value,
            ...(statusVal ? { status: statusVal } : { status: "proses" })
          };

          fetch(`http://localhost:8080/transaksi?id=eq.${id}`, {
            method: "PATCH",
            headers: { "Content-Type": "application/json", Authorization: "Bearer " + token },
            body: JSON.stringify(updated)
          }).then(() => { closeModal(); loadTransaksi(token); });
        });
      });
    });

    document.querySelectorAll(".btn-detail").forEach(btn => {
      btn.addEventListener("click", async () => {
        const id = btn.dataset.id;

        // Ambil detail transaksi dari Supabase
        const transaksiData = await fetch(`http://localhost:8080/transaksi?id=eq.${id}`, {
          headers: { Authorization: "Bearer " + token }
        }).then(res => res.json());

        const t = Array.isArray(transaksiData) ? transaksiData[0] : transaksiData.data[0];

        // Buka modal detail
        openModal("Detail Transaksi", `
      <p><strong>Pelanggan:</strong> ${pelangganMap[t.pelanggan_id]}</p>
      <p><strong>Layanan:</strong> ${layananMap[t.layanan_id]}</p>
      <p><strong>Tanggal:</strong> ${t.tanggal}</p>
      <p><strong>Berat:</strong> ${t.berat}</p>
      <p><strong>Total:</strong> Rp ${t.total}</p>
      <p><strong>Metode Pembayaran:</strong> ${t.metode_pembayaran}</p>
      <p><strong>Status:</strong> ${t.status}</p>
      <button id="btn-selesai" class="bg-blue-600 text-white px-4 py-2 rounded mt-4">Selesai</button>
    `, () => { });

        // Tombol selesai → update status
        document.getElementById("btn-selesai").addEventListener("click", () => {
          fetch(`http://localhost:8080/transaksi?id=eq.${id}`, {
            method: "PATCH",
            headers: { "Content-Type": "application/json", Authorization: "Bearer " + token },
            body: JSON.stringify({ status: "selesai" })
          }).then(() => { closeModal(); loadTransaksi(token); });
        });
      });
    });
  });
}