<x-app-layout>
    <x-slot name="header">Tambah Layanan</x-slot>
    @include('admin.layanan._form', [
        'action' => route('admin.layanan.store'),
        'isEdit' => false,
        'buttonText' => 'Simpan'
    ])
</x-app-layout>