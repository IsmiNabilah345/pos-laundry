<x-app-layout>
    <x-slot name="header">
        <h2 class="font-semibold text-xl text-gray-800 dark:text-gray-200 leading-tight">
            {{ __('Dashboard Admin') }}
        </h2>
    </x-slot>

    <div class="py-12">
        <div class="max-w-7xl mx-auto sm:px-6 lg:px-8">
            <div class="bg-white dark:bg-gray-800 overflow-hidden shadow-sm sm:rounded-lg">
                <div class="p-6 text-gray-900 dark:text-gray-100">
                    {{ __("Selamat datang Admin!") }}
                </div>
            </div>
        </div>
    </div>

    <div class="grid grid-cols-2 gap-4 ">
        <a href="{{ route('admin.pelanggan.index') }}" class="bg-gray-600 text-white p-4 rounded shadow hover:bg-gray-600">
            Kelola Pelanggan
        </a>
        <a href="{{ route('admin.layanan.index') }}" class="bg-gray-600 text-white p-4 rounded shadow hover:bg-black-600">
            Kelola Layanan
        </a>
        <a href="{{ route('admin.transaksi.index') }}" class="bg-gray-600 text-white p-4 rounded shadow hover:bg-black-600">
            Transaksi
        </a>
    </div>
</x-app-layout>
