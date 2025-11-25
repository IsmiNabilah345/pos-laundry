<x-app-layout>
    <x-slot name="header">Edit Pelanggan</x-slot>
    @include('admin.pelanggan._form', [
        'action' => route('admin.pelanggan.update', $pelanggan),
        'isEdit' => true,
        'buttonText' => 'Update'
    ])
</x-app-layout>