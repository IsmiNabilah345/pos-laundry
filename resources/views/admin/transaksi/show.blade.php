<x-app-layout>
    <x-slot name="header">Detail Transaksi</x-slot>

    <div class="p-4 bg-white rounded shadow">
        <p><strong>Tanggal:</strong> {{ $transaksi->tanggal }}</p>
        <p><strong>Pelanggan:</strong> {{ $transaksi->pelanggan->nama }}</p>
        <p><strong>Status:</strong> {{ $transaksi->status }}</p>
        <p><strong>Total:</strong> Rp {{ number_format($transaksi->total,0,',','.') }}</p>
        <p><strong>Metode Pembayaran:</strong> {{ ucfirst($transaksi->metode_pembayaran) }}</p>

        <hr class="my-4">

        <h3 class="font-semibold mb-2">Detail Layanan</h3>
        <table class="w-full border">
            <thead>
                <tr>
                    <th>Layanan</th>
                    <th>Berat (kg)</th>
                    <th>Subtotal</th>
                </tr>
            </thead>
            <tbody>
                @foreach($transaksi->detailTransaksi as $detail)
                    <tr>
                        <td>{{ $detail->layanan->nama }}</td>
                        <td>{{ $detail->berat }}</td>
                        <td>Rp {{ number_format($detail->subtotal,0,',','.') }}</td>
                    </tr>
                @endforeach
            </tbody>
        </table>

        <div class="mt-4">
            <a href="{{ route('admin.transaksi.index') }}" 
               class="px-4 py-2 bg-gray-500 text-white rounded">Kembali</a>
        </div>
    </div>
</x-app-layout>