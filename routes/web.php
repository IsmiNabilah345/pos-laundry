<?php

use App\Http\Controllers\ProfileController;
use App\Http\Controllers\AdminController;
use App\Http\Controllers\KasirController;
use App\Http\Controllers\PelangganController;
use App\Http\Controllers\Admin\LayananController;
use App\Http\Controllers\Admin\TransaksiController;
use Illuminate\Support\Facades\Route;

Route::get('/', function () {
    return view('welcome');
});

Route::get('/dashboard', function () {
    $role = auth()->user()->role ?? 'guest';
    return redirect($role === 'admin' ? '/admin/dashboard' : '/kasir/transaksi');
})->middleware(['auth', 'verified'])->name('dashboard');

Route::middleware('auth')->group(function () {
    Route::get('/profile', [ProfileController::class, 'edit'])->name('profile.edit');
    Route::patch('/profile', [ProfileController::class, 'update'])->name('profile.update');
    Route::delete('/profile', [ProfileController::class, 'destroy'])->name('profile.destroy');
});

Route::middleware(['auth', \App\Http\Middleware\CheckRole::class . ':admin'])->group(function () {
    Route::get('/admin/dashboard', [AdminController::class, 'index']);
});

Route::middleware(['auth', \App\Http\Middleware\CheckRole::class . ':kasir'])->group(function () {
    Route::get('/kasir/transaksi', [KasirController::class, 'index']);
});

Route::middleware(['auth', \App\Http\Middleware\CheckRole::class . ':admin'])->group(function () {
    Route::get('/admin/dashboard', [AdminController::class, 'index'])->name('admin.dashboard');
    Route::resource('/admin/pelanggan', PelangganController::class)->names('admin.pelanggan');
    Route::resource('/admin/layanan', LayananController::class)->names('admin.layanan');
    Route::resource('/admin/transaksi', TransaksiController::class)->names('admin.transaksi');

});

require __DIR__ . '/auth.php';
