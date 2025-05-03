document.addEventListener('DOMContentLoaded', function() {
    const select = document.getElementById('items-per-page');
    const tbody = document.getElementById('tabla-recortes');
    const pageInfo = document.getElementById('page-info');

    function actualizarTabla() {
        const rows = tbody.querySelectorAll('tr');
        const itemsPerPage = parseInt(select.value);

        rows.forEach((row, index) => {
            if (index < itemsPerPage) {
                row.style.display = '';
            } else {
                row.style.display = 'none';
            }
        });

        pageInfo.textContent = `Mostrando ${Math.min(itemsPerPage, rows.length)} de ${rows.length} registros`;
    }

    select.addEventListener('change', actualizarTabla);

    actualizarTabla();
});
