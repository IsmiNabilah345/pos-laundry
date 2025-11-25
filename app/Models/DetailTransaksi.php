<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Factories\HasFactory;

class DetailTransaksi extends Model
{
    use HasFactory;
    protected $table = 'detail_transaksi';

    protected $fillable = ['transaksi_id', 'layanan_id', 'berat', 'subtotal', 'keterangan'];

    public function transaksi()
    {
        return $this->belongsTo(Transaksi::class);
    }

    public function layanan()
    {
        return $this->belongsTo(Layanan::class);
    }
}
