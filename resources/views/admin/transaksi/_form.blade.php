<form action="{{ $action }}" method="POST" id="transaksi-form">
    @csrf
    @if($isEdit)
        @method('PUT')
    @endif

    <div class="mb-4">
        <label for="tanggal">Tanggal</label>
        <input type="date" name="tanggal" class="w-full p-2 border rounded"
               value="{{ old('tanggal', $isEdit ? $transaksi->tanggal : date('Y-m-d')) }}" required>
    </div>

    <div class="mb-4">
        <label for="pelanggan_id">Pelanggan</label>
        <select name="pelanggan_id" class="w-full p-2 border rounded pelanggan-select" required>
            <option value="">-- Pilih Pelanggan --</option>
            @foreach($pelanggan as $p)
                <option value="{{ $p->id }}"
                    {{ old('pelanggan_id', $isEdit ? $transaksi->pelanggan_id : '') == $p->id ? 'selected' : '' }}>
                    {{ $p->nama }}
                </option>
            @endforeach
        </select>
    </div>
    <div class="mb-4">
        <label for="status">Status</label>
        <select name="status" class="w-full p-2 border rounded" required>
            <option value="proses" {{ old('status', $isEdit ? $transaksi->status : '') == 'proses' ? 'selected' : '' }}>Proses</option>
            <option value="selesai" {{ old('status', $isEdit ? $transaksi->status : '') == 'selesai' ? 'selected' : '' }}>Selesai</option>
            <option value="diambil" {{ old('status', $isEdit ? $transaksi->status : '') == 'diambil' ? 'selected' : '' }}>Diambil</option>
         </select>
    </div>

    <div class="mb-4">
        <label for="metode_pembayaran">Metode Pembayaran</label>
        <select name="metode_pembayaran" class="w-full p-2 border rounded" required>
            <option value="">-- Pilih Metode --</option>
            <option value="cash">Cash</option>
            <option value="transfer">Transfer</option>
            <option value="qris">QRIS</option>
        </select>
    </div>

    <hr class="my-4">

    <div>
        <label class="font-semibold">Detail Layanan</label>
        <table class="w-full" id="detail-table">
            <thead>
                <tr>
                    <th class="text-left">Layanan</th>
                    <th class="text-left">Berat (kg)</th>
                    <th class="text-left">Subtotal</th>
                    <th></th>
                </tr>
            </thead>
            <tbody>
                @if($isEdit && isset($transaksi))
                    @foreach($transaksi->detailTransaksi as $detail)
                        <tr>
                            <td>
                                <select name="layanan_id[]" class="layanan-select p-2 border rounded">
                                    @foreach($layanan as $l)
                                        <option value="{{ $l->id }}"
                                            data-price="{{ $l->harga_per_kg ?? $l->harga }}"
                                            {{ $detail->layanan_id == $l->id ? 'selected' : '' }}>
                                            {{ $l->nama }} Rp {{ number_format($l->harga_per_kg ?? $l->harga,0,',','.') }}/kg
                                        </option>
                                    @endforeach
                                </select>
                            </td>
                            <td>
                                <input type="number" step="0.1" min="0.1" name="berat[]"
                                       class="berat-input p-2 border rounded"
                                       value="{{ $detail->berat }}">
                            </td>
                            <td>
                                <span class="subtotal-display">{{ $detail->subtotal }}</span>
                            </td>
                            <td>
                                <button type="button" class="remove-row text-red-600">Hapus</button>
                            </td>
                        </tr>
                    @endforeach
                @endif
            </tbody>
        </table>

        <div class="mt-2">
            <button type="button" id="add-row" class="bg-green-500 text-white px-3 py-1 rounded">+ Tambah Layanan</button>
        </div>

        <div class="mt-4 text-right">
            <span class="font-semibold">Total: Rp </span><span id="total-display">0</span>
        </div>
    </div>

    <input type="hidden" name="total" id="total-input" value="0" />

    <div class="mt-6">
        <button type="submit" class="bg-indigo-600 text-white px-4 py-2 rounded">{{ $buttonText }}</button>
        <a href="{{ route('admin.transaksi.index') }}"
           class="px-4 py-2 bg-gray-500 text-white rounded hover:bg-gray-600">
            Kembali
        </a>
    </div>
</form>

<table style="display:none">
    <tbody>
        <tr id="row-template">
            <td>
                <select name="layanan_id[]" class="layanan-select p-2 border rounded">
                    <option value="">-- Pilih Layanan --</option>
                    @foreach($layanan as $l)
                        <option value="{{ $l->id }}" data-price="{{ $l->harga_per_kg ?? $l->harga }}">
                            {{ $l->nama }} Rp {{ number_format($l->harga_per_kg ?? $l->harga,0,',','.') }}/kg
                        </option>
                    @endforeach
                </select>
            </td>
            <td>
                <input type="number" step="0.1" min="0.1" name="berat[]" class="berat-input p-2 border rounded" value="1">
            </td>
            <td>
                <span class="subtotal-display">0</span>
            </td>
            <td>
                <button type="button" class="remove-row text-red-600">Hapus</button>
            </td>
        </tr>
    </tbody>
</table>
