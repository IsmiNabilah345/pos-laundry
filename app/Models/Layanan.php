<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Factories\HasFactory;

class Layanan extends Model
{
    use HasFactory;

    protected $table = 'layanan';
    protected $fillable = ['nama_layanan', 'harga_per_kg', 'satuan', 'estimasi'];

    public function detailTransaksi()
    {
        return $this->hasMany(DetailTransaksi::class);
    }
}
