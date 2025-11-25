<!DOCTYPE html>
<html lang="{{ str_replace('_', '-', app()->getLocale()) }}">
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <meta name="csrf-token" content="{{ csrf_token() }}">

    <title>{{ config('app.name', 'Laundry') }}</title>

    <!-- Fonts -->
    <link rel="preconnect" href="https://fonts.bunny.net">
    <link href="https://fonts.bunny.net/css?family=figtree:400,500,600&display=swap" rel="stylesheet" />

    <!-- Scripts -->
    @vite(['resources/css/app.css', 'resources/js/app.js'])
</head>
<body x-data="{ sidebarOpen: window.innerWidth >= 768 }" class="font-sans antialiased">
    @php
        $role = auth()->user()->role ?? 'guest';
    @endphp

    @if ($role === 'admin')
        <aside
            class="w-64 bg-gray-100 dark:bg-gray-900 h-screen fixed top-0 left-0 shadow-md z-50 transform transition-transform duration-300"
            :class="sidebarOpen ? 'translate-x-0' : '-translate-x-full'"
            x-cloak
        >
            <div class="p-4 text-lg font-bold text-gray-800 dark:text-gray-200">
                Admin Menu
            </div>
            <nav class="flex flex-col space-y-2 px-4">
                <a href="/admin/dashboard" class="text-gray-700 dark:text-gray-300 hover:underline">Dashboard</a>
                <a href="/admin/pelanggan" class="text-gray-700 dark:text-gray-300 hover:underline">Data Pelanggan</a>
                <a href="{{ route('admin.layanan.index') }}" class="text-gray-700 dark:text-gray-300 hover:underline">Data Layanan</a>
                <a href="{{ route('admin.transaksi.index') }}" class="text-gray-700 dark:text-gray-300 hover:underline">Data Transaksi</a>
                <form method="POST" action="{{ route('logout') }}">
                    @csrf
                    <button type="submit" class="text-red-500 hover:underline mt-4">Logout</button>
                </form>
            </nav>
        </aside>
    @endif

    <div class="min-h-screen bg-gray-100 dark:bg-gray-900 transition-all duration-300 md:pl-64"
         :class="{ 'pl-64': sidebarOpen }">
        @include('layouts.navigation')

        @isset($header)
            <header class="bg-white dark:bg-gray-800 shadow">
                <div class="max-w-7xl mx-auto py-6 px-4 sm:px-6 lg:px-8">
                    {{ $header }}
                </div>
            </header>
        @endisset

        <main>
            {{ $slot }}
        </main>
    </div>
    @stack('scripts')
</body>
</html>