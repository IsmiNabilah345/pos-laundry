<x-app-layout>
    <x-slot name="header">Edit Layanan</x-slot>
    @include('admin.layanan._form', [
        'action' => route('admin.layanan.update', $layanan),
        'isEdit' => true,
        'buttonText' => 'Update'
    ])
</x-app-layout>