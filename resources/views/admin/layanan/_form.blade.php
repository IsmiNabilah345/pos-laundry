<form action="{{ $action }}" method="POST">
    @csrf
    @if($isEdit)
        @method('PUT')
    @endif

    <div class="mb-4">
        <label for="nama_layanan">Nama Layanan</label>
        <input type="text" name="nama_layanan" value="{{ old('nama_layanan', $layanan->nama_layanan ?? '') }}" required>
    </div>

    <div class="mb-4">
        <label for="harga_per_kg">Harga</label>
        <input type="text" step="0.01" name="harga_per_kg" value="{{ old('harga_per_kg', $layanan->harga_per_kg ?? '') }}">
    </div>

    <div class="mb-4">
        <label for="satuan">Satuan</label>
        <textarea name="satuan">{{ old('satuan', $layanan->satuan ?? '') }}</textarea>
    </div>

    <div class="mb-4">
        <label for="estimasi">Estimasi</label>
        <textarea name="estimasi">{{ old('estimasi', $layanan->estimasi ?? '') }}</textarea>
    </div>

    <button type="submit">{{ $buttonText }}</button>
</form>