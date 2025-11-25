<x-app-layout>
    <x-slot name="header">Tambah Pelanggan</x-slot>
    @include('admin.pelanggan._form', [
        'action' => route('admin.pelanggan.store'),
        'isEdit' => false,
        'buttonText' => 'Simpan'
    ])
</x-app-layout>