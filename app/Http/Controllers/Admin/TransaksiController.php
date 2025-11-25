<?php

namespace App\Http\Controllers\Admin;

use App\Http\Controllers\Controller;
use App\Models\Transaksi;
use App\Models\Pelanggan;
use App\Models\Layanan;
use App\Models\DetailTransaksi;
use Illuminate\Http\Request;

class TransaksiController extends Controller
{
    public function index()
    {
        $transaksis = Transaksi::with(['pelanggan', 'user', 'detailTransaksi.layanan'])
            ->whereNull('deleted_at')
            ->paginate(15);
        return view('admin.transaksi.index', compact('transaksis'));
    }

    public function create()
    {
        $pelanggan = Pelanggan::all();
        $layanan = Layanan::all();
        return view('admin.transaksi.create', compact('pelanggan', 'layanan'));
    }

    public function store(Request $request)
    {
        $request->validate([
            'pelanggan_id' => 'required|exists:pelanggan,id',
            'tanggal' => 'required|date',
            'layanan_id.*' => 'required|exists:layanan,id',
            'berat.*' => 'required|numeric|min:0.1',
        ]);

        $transaksis = Transaksi::create([
            'pelanggan_id' => $request->pelanggan_id,
            'user_id' => auth()->id(),
            'tanggal' => $request->tanggal,
            'total' => 0,
            'status' => 'proses',
            'metode_pembayaran' => $request->metode_pembayaran,
        ]);

        $total = 0;

        foreach ($request->layanan_id as $i => $layananId) {
            $layanan = Layanan::find($layananId);
            $berat = $request->berat[$i];
            $subtotal = $layanan->harga_per_kg * $berat;

            DetailTransaksi::create([
                'transaksi_id' => $transaksis->id,
                'layanan_id' => $layananId,
                'berat' => $berat,
                'subtotal' => $subtotal,
            ]);

            $total += $subtotal;
        }

        $transaksis->update(['total' => $total]);

        return redirect()->route('admin.transaksi.index')->with('success', 'Transaksi berhasil disimpan!');
    }

    public function edit($id)
    {
        $transaksi = Transaksi::with('detailTransaksi.layanan')->findOrFail($id);
        $pelanggan = Pelanggan::all();
        $layanan = Layanan::all();

        return view('admin.transaksi.edit', compact('transaksi', 'pelanggan', 'layanan'));
    }

    public function update(Request $request, $id)
    {
        $request->validate([
            'pelanggan_id' => 'required|exists:pelanggan,id',
            'tanggal' => 'required|date',
            'layanan_id.*' => 'required|exists:layanan,id',
            'berat.*' => 'required|numeric|min:0.1',
        ]);

        $transaksi = Transaksi::findOrFail($id);

        $transaksi->update([
            'pelanggan_id' => $request->pelanggan_id,
            'tanggal' => $request->tanggal,
            'status' => $request->status,
            'metode_pembayaran' => $request->metode_pembayaran,
        ]);

        $transaksi->detailTransaksi()->delete();

        $total = 0;

        foreach ($request->layanan_id as $i => $layananId) {
            $layanan = Layanan::find($layananId);
            $berat = $request->berat[$i];
            $subtotal = $layanan->harga_per_kg * $berat;

            DetailTransaksi::create([
                'transaksi_id' => $transaksi->id,
                'layanan_id' => $layananId,
                'berat' => $berat,
                'subtotal' => $subtotal,
            ]);

            $total += $subtotal;
        }

        $transaksi->update(['total' => $total]);

        return redirect()->route('admin.transaksi.index')
            ->with('success', 'Transaksi berhasil diperbarui!');
    }

    public function destroy(Transaksi $transaksi)
    {
        if ($transaksi->status !== 'proses') {
            return redirect()->route('admin.transaksi.index')->with('error', 'Transaksi tidak bisa dihapus karena sudah selesai atau diambil.');
        }

        $transaksi->delete();

        return redirect()->route('admin.transaksi.index')->with('success', 'Transaksi berhasil dihapus!');
    }

    public function show($id)
    {
        $transaksi = Transaksi::with(['pelanggan','user','detailTransaksi.layanan'])
            ->findOrFail($id);

        return view('admin.transaksi.show', compact('transaksi'));
    }

}
