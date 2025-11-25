<x-app-layout>
    <x-slot name="header">
        <h2 class="font-semibold text-xl text-gray-800 dark:text-gray-200 leading-tight">
            Data Layanan
        </h2>
    </x-slot>

    <div class="py-12">
        <div class="max-w-7xl mx-auto sm:px-6 lg:px-8">
            <div class="bg-white dark:bg-gray-800 overflow-hidden shadow-sm sm:rounded-lg">
                <div class="p-6 text-gray-900 dark:text-gray-100">
                    <div class="mb-4">
                        <a href="{{ route('admin.layanan.create') }}"
                        class="inline-block bg-gradient-to-r from-indigo-500 to-purple-500 hover:from-indigo-600 hover:to-purple-600 text-white font-semibold py-2 px-4 rounded shadow-md transition duration-300">
                            + Tambah Layanan
                        </a>
                    </div>

                    <table class="table-auto w-full">
                        <thead>
                            <tr>
                                <th class="px-4 py-2">Nama Layanan</th>
                                <th class="px-4 py-2">Harga</th>
                                <th class="px-4 py-2">Satuan</th>
                                <th class="px-4 py-2">Estimasi</th>
                                <th class="px-4 py-2">Aksi</th>
                            </tr>
                        </thead>
                        <tbody>
                            @foreach ($layanans as $l)
                                <tr>
                                    <td class="border px-4 py-2">{{ $l->nama_layanan }}</td>
                                    <td class="border px-4 py-2">Rp {{ number_format($l->harga_per_kg, 0, ',', '.') }}</td>
                                    <td class="border px-4 py-2">{{ $l->satuan }}</td>
                                    <td class="border px-4 py-2">{{ $l->estimasi ?? '-' }}</td>
                                    <td class="border px-4 py-2 space-x-2">
                                        <a href="{{ route('admin.layanan.edit', $l->id) }}" class="text-blue-500 hover:underline">Edit</a>
                                        <form action="{{ route('admin.layanan.destroy', $l->id) }}" method="POST" onsubmit="return confirm('Yakin hapus layanan ini?')" style="display:inline">
                                            @csrf
                                            @method('DELETE')
                                            <button type="submit" class="text-red-500 hover:underline">Hapus</button>
                                        </form>
                                    </td>
                                </tr>
                            @endforeach
                        </tbody>
                    </table>

                    <div class="mt-4">
                        {{ $layanans->links() }}
                    </div>

                    @if ($layanans->isEmpty())
                        <p colspan="5" class="text-center py-4">Belum ada layanan.</p>
                    @endif
                </div>
            </div>
        </div>
    </div>
</x-app-layout>