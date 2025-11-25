<x-app-layout>
    <x-slot name="header">
        <h2 class="font-semibold text-xl text-gray-800 dark:text-gray-200 leading-tight">
            Data Pelanggan
        </h2>
    </x-slot>

    <div class="py-12">
        <div class="max-w-7xl mx-auto sm:px-6 lg:px-8">
            <div class="bg-white dark:bg-gray-800 overflow-hidden shadow-sm sm:rounded-lg">
                <div class="p-6 text-gray-900 dark:text-gray-100">
                    <div class="mb-4">
                        <a href="{{ route('admin.pelanggan.create') }}"
                        class="inline-block bg-gradient-to-r from-indigo-500 to-purple-500 hover:from-indigo-600 hover:to-purple-600 text-white font-semibold py-2 px-4 rounded shadow-md transition duration-300">
                            + Tambah Pelanggan
                        </a>
                    </div>

                    <table class="table-auto w-full">
                        <thead>
                            <tr>
                                <th class="px-4 py-2">Nama</th>
                                <th class="px-4 py-2">Alamat</th>
                                <th class="px-4 py-2">No HP</th>
                                <th class="px-4 py-2">Aksi</th>
                            </tr>
                        </thead>
                        <tbody>
                            @foreach ($pelanggan as $p)
                                <tr>
                                    <td class="border px-4 py-2">{{ $p->nama }}</td>
                                    <td class="border px-4 py-2">{{ $p->alamat }}</td>
                                    <td class="border px-4 py-2">{{ $p->kontak }}</td>
                                    <td class="border px-4 py-2">
                                        <a href="{{ route('admin.pelanggan.edit', $p->id) }}" class="text-blue-500 hover:underline" >Edit</a>
                                        <form action="{{ route('admin.pelanggan.destroy', $p->id) }}" method="POST" onsubmit="return confirm('Yakin hapus?')" style="display:inline">
                                            @csrf
                                            @method('DELETE')
                                            <button type="submit" class="text-red-500 hover:underline" >Hapus</button>
                                        </form>
                                    </td>
                                </tr>
                            @endforeach
                        </tbody>
                    </table>

                    @if ($pelanggan->isEmpty())
                        <p colspan="5" class="text-center py-4">Belum ada pelangan.</p>
                    @endif

                </div>
            </div>
        </div>
    </div>
</x-app-layout>