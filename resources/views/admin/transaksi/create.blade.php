<x-app-layout>
    <x-slot name="header">Tambah Transaksi</x-slot>

    <div class="max-w-3xl mx-auto p-4">
        @include('admin.transaksi._form', [
            'action' => route('admin.transaksi.store'),
            'isEdit' => false,
            'buttonText' => 'Simpan'
        ])
    </div>

    <script src="https://cdn.jsdelivr.net/npm/jquery@3.6.0/dist/jquery.min.js"></script>
    <link href="https://cdn.jsdelivr.net/npm/select2@4.1.0-rc.0/dist/css/select2.min.css" rel="stylesheet" />
    <script src="https://cdn.jsdelivr.net/npm/select2@4.1.0-rc.0/dist/js/select2.min.js"></script>

    <script>
    $(function () {
        $('.pelanggan-select').select2({ placeholder: "Cari nama pelanggan...", allowClear: true, width: '100%' });

        function formatRupiah(num) {
            return new Intl.NumberFormat('id-ID').format(Math.round(num));
        }

        function addRow() {
            var $template = $('#row-template').clone().removeAttr('id').show();
            $('#detail-table tbody').append($template);

            $template.find('.layanan-select').select2({ placeholder: "Pilih layanan", allowClear: true, width: '100%' });

            recalcAll();
        }

        function recalcRow($row) {
            var $layananOpt = $row.find('.layanan-select option:selected');
            var price = parseFloat($layananOpt.data('price') || 0);
            var berat = parseFloat($row.find('.berat-input').val() || 0);
            var subtotal = price * berat;
            $row.find('.subtotal-display').text(formatRupiah(subtotal));
            return subtotal;
        }

        function recalcAll() {
            var total = 0;
            $('#detail-table tbody tr').each(function () {
                total += recalcRow($(this));
            });
            $('#total-display').text(formatRupiah(total));
            $('#total-input').val(Math.round(total));
        }

        $('#add-row').on('click', function () {
            addRow();
        });

        $('#detail-table').on('change', '.layanan-select, .berat-input', function () {
            recalcAll();
        });

        $('#detail-table').on('click', '.remove-row', function () {
            $(this).closest('tr').remove();
            recalcAll();
        });

        addRow();
        
        $('#transaksi-form').on('submit', function (e) {
            if ($('#detail-table tbody tr').length === 0) {
                e.preventDefault();
                alert('Tambahkan minimal 1 layanan.');
            }
        });
    });
    </script>
</x-app-layout>
