<aside
    class="w-64 bg-gray-100 dark:bg-gray-900 h-screen fixed top-0 left-0 shadow-md z-50 transform transition-transform duration-300"
    :class="sidebarOpen ? 'translate-x-0' : '-translate-x-full'"
>
    <div class="p-4 text-lg font-bold text-gray-800 dark:text-gray-200">
        Admin Menu
    </div>
    <nav class="flex flex-col space-y-2 px-4">
        <a href="/admin/dashboard" class="text-gray-700 dark:text-gray-300 hover:underline">Dashboard</a>
        <a href="/admin/pelanggan" class="text-gray-700 dark:text-gray-300 hover:underline">Data Pelanggan</a>
        <a href="/admin/layanan" class="text-gray-700 dark:text-gray-300 hover:underline">Data Layanan</a>
        <form method="POST" action="{{ route('logout') }}">
            @csrf
            <button type="submit" class="text-red-500 hover:underline mt-4">Logout</button>
        </form>
    </nav>
</aside>