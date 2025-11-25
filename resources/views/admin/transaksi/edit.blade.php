<x-app-layout>
    <x-slot name="header">Edit Transaksi</x-slot>

    @include('admin.transaksi._form', [
        'action' => route('admin.transaksi.update', $transaksi->id),
        'isEdit' => true,
        'buttonText' => 'Update'
    ])

    @push('scripts')
    <script>
        document.addEventListener('DOMContentLoaded', function () {

            const addRowBtn = document.getElementById('add-row');
            const detailTableBody = document.querySelector('#detail-table tbody');
            const rowTemplate = document.getElementById('row-template');
            const totalDisplay = document.getElementById('total-display');
            const totalInput = document.getElementById('total-input');

            function updateTotal() {
                let total = 0;

                detailTableBody.querySelectorAll('tr').forEach(row => {
                    const layananSelect = row.querySelector('.layanan-select');
                    const beratInput = row.querySelector('.berat-input');
                    const subtotalDisplay = row.querySelector('.subtotal-display');

                    const price = layananSelect.selectedOptions[0]?.dataset.price || 0;
                    const berat = parseFloat(beratInput.value) || 0;
                    const subtotal = price * berat;

                    subtotalDisplay.textContent = subtotal.toLocaleString('id-ID');
                    total += subtotal;
                });

                totalDisplay.textContent = total.toLocaleString('id-ID');
                totalInput.value = total;
            }

            detailTableBody.querySelectorAll('tr').forEach(row => {
                row.querySelector('.layanan-select')?.addEventListener('change', updateTotal);
                row.querySelector('.berat-input')?.addEventListener('input', updateTotal);
                row.querySelector('.remove-row')?.addEventListener('click', function () {
                    row.remove();
                    updateTotal();
                });
            });

            updateTotal();

            addRowBtn.addEventListener('click', function () {
                const newRow = rowTemplate.cloneNode(true);
                newRow.removeAttribute('id');
                newRow.style.display = '';
                detailTableBody.appendChild(newRow);

                newRow.querySelector('.layanan-select').addEventListener('change', updateTotal);
                newRow.querySelector('.berat-input').addEventListener('input', updateTotal);
                newRow.querySelector('.remove-row').addEventListener('click', function () {
                    newRow.remove();
                    updateTotal();
                });

                updateTotal();
            });
        });
    </script>
    @endpush
</x-app-layout>