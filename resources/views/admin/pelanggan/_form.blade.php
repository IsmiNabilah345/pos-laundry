<form action="{{ $action }}" method="POST">
    @csrf
    @if($isEdit)
        @method('PUT')
    @endif

    <div class="mb-4">
        <label for="nama">Nama</label>
        <input type="text" name="nama" value="{{ old('nama', $pelanggan->nama ?? '') }}" required>
    </div>

    <div class="mb-4">
        <label for="kontak">Kontak</label>
        <input type="text" name="kontak" value="{{ old('kontak', $pelanggan->kontak ?? '') }}">
    </div>

    <div class="mb-4">
        <label for="alamat">Alamat</label>
        <textarea name="alamat">{{ old('alamat', $pelanggan->alamat ?? '') }}</textarea>
    </div>

    <button type="submit">{{ $buttonText }}</button>
</form>