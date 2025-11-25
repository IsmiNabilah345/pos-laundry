<x-app-layout>
    <x-slot name="header">
        <h2 class="font-semibold text-xl text-gray-800 dark:text-gray-200 leading-tight">
            Data Transaksi
        </h2>
    </x-slot>

    @if(session('success'))
        <div class="bg-green-100 text-green-800 p-2 rounded mb-4">
            {{ session('success') }}
        </div>
    @endif

    @if(session('error'))
        <div class="bg-red-100 text-red-800 p-2 rounded mb-4">
            {{ session('error') }}
        </div>
    @endif

    <div class="py-12">
        <div class="max-w-7xl mx-auto sm:px-6 lg:px-8">
            <div class="bg-white dark:bg-gray-800 overflow-hidden shadow-sm sm:rounded-lg">
                <div class="p-6 text-gray-900 dark:text-gray-100">
                    <div class="mb-4">
                        <a href="{{ route('admin.transaksi.create') }}"
                        class="inline-block bg-gradient-to-r from-indigo-500 to-purple-500 hover:from-indigo-600 hover:to-purple-600 text-white font-semibold py-2 px-4 rounded shadow-md transition duration-300">
                            + Tambah Transaksi
                        </a>
                    </div>

                    <table class="table-auto w-full">
                        <thead>
                            <tr>
                                <th class="border px-4 py-2">Tanggal</th>
                                <th class="border px-4 py-2">Pelanggan</th>
                                <th class="border px-4 py-2">Total</th>
                                <th class="border px-4 py-2">Status</th>
                                <th class="border px-4 py-2">Metode Pembayaran</th>
                                <th class="border px-4 py-2">Aksi</th>
                            </tr>
                        </thead>
                        <tbody>
                            @forelse($transaksis as $transaksi)
                                <tr>
                                    <td class="border px-4 py-2">{{ $transaksi->tanggal }}</td>
                                    <td class="border px-4 py-2">{{ $transaksi->pelanggan->nama }}</td>
                                    <td class="border px-4 py-2">Rp {{ number_format($transaksi->total, 0, ',', '.') }}</td>
                                    <td class="border px-4 py-2 capitalize">{{ $transaksi->status }}</td>
                                    <td class="border px-4 py-2 capitalize">{{ $transaksi->metode_pembayaran }}</td>
                                    <td class="border px-4 py-2">
                                        <div class="flex gap-2 items-center">
                                        @if($transaksi->status === 'proses')
                                            <a href="{{ route('admin.transaksi.edit', $transaksi->id) }}" class="text-yellow-500 mr-2">Edit</a>

                                            <form action="{{ route('admin.transaksi.destroy', $transaksi->id) }}" method="POST" class="inline">
                                                @csrf
                                                @method('DELETE')
                                                <button type="submit" onclick="return confirm('Yakin ingin menghapus transaksi ini?')" class="text-red-500">Hapus</button>
                                            </form>
                                        @else
                                            <span class="text-gray-400 italic">Tidak bisa diubah</span>
                                        @endif
                                        <a href="{{ route('admin.transaksi.show', $transaksi->id) }}" class="text-blue-500">Detail</a>
                                        </div>
                                    </td>
                                </tr>
                            @empty
                                <tr>
                                    <td colspan="5" class="text-center py-4">Belum ada transaksi.</td>
                                </tr>
                            @endforelse
                        </tbody>
                    </table>
                </div>
            </div>
        </div>
    </div>
</x-app-layout>