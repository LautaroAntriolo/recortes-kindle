document.addEventListener('DOMContentLoaded', function() {
    // Seleccionar elementos con verificaciones
    const selectSuperior = document.getElementById('items-per-page');
    const selectInferior = document.getElementById('items-per-page-inferior');
    const tbody = document.getElementById('tabla-recortes');
    const pageInfo = document.getElementById('page-info');
    const prevButton = document.getElementById('prev-page');
    const nextButton = document.getElementById('next-page');
    const pageIndicator = document.getElementById('page-indicator');

    // Verificar que los elementos esenciales existen
    if (!selectSuperior || !tbody) {
        console.error('Elementos esenciales no encontrados en el DOM');
        return;
    }

    let currentPage = 1;
    let totalPages = 1;
    let allRows = [];

    function updateTable() {
        const itemsPerPage = parseInt(selectSuperior.value) || 0;
        allRows = Array.from(tbody.querySelectorAll('tr'));
        
        if (itemsPerPage === 0) {
            totalPages = 1;
            allRows.forEach(row => row.style.display = '');
        } else {
            totalPages = Math.max(1, Math.ceil(allRows.length / itemsPerPage));
            const startIndex = (currentPage - 1) * itemsPerPage;
            const endIndex = startIndex + itemsPerPage;
            
            allRows.forEach((row, index) => {
                row.style.display = (index >= startIndex && index < endIndex) ? '' : 'none';
            });
        }

        updatePaginationControls(itemsPerPage);
        updatePageInfo(itemsPerPage);
    }

    function updatePaginationControls(itemsPerPage) {
        if (prevButton) prevButton.disabled = currentPage === 1;
        if (nextButton) nextButton.disabled = currentPage === totalPages || totalPages === 0;
        
        const paginationControls = document.querySelectorAll('.pagination-controls');
        if (paginationControls) {
            paginationControls.forEach(control => {
                control.style.display = itemsPerPage === 0 ? 'none' : 'flex';
            });
        }
    }

    function updatePageInfo(itemsPerPage) {
        if (!pageIndicator && !pageInfo) return;
        
        if (itemsPerPage === 0) {
            if (pageIndicator) pageIndicator.textContent = `Mostrando todos los registros (${allRows.length})`;
            if (pageInfo) pageInfo.textContent = `Mostrando todos los ${allRows.length} registros`;
        } else {
            const startItem = ((currentPage - 1) * itemsPerPage) + 1;
            const endItem = Math.min(currentPage * itemsPerPage, allRows.length);
            if (pageIndicator) {
                pageIndicator.textContent = `Página ${currentPage} de ${totalPages} (${startItem}-${endItem} de ${allRows.length})`;
            }
            if (pageInfo) {
                pageInfo.textContent = `Mostrando ${endItem - startItem + 1} de ${allRows.length} registros`;
            }
        }
    }

    function handleItemsPerPageChange(e) {
        const valor = e.target.value;
        // Solo sincronizar si ambos selectores existen
        if (selectSuperior && selectInferior) {
            selectSuperior.value = valor;
            selectInferior.value = valor;
        }
        currentPage = 1;
        updateTable();
    }

    // Agregar event listeners solo si los elementos existen
    if (selectSuperior) selectSuperior.addEventListener('change', handleItemsPerPageChange);
    if (selectInferior) selectInferior.addEventListener('change', handleItemsPerPageChange);
    
    if (prevButton) prevButton.addEventListener('click', function() {
        if (currentPage > 1) {
            currentPage--;
            updateTable();
        }
    });
    
    if (nextButton) nextButton.addEventListener('click', function() {
        if (currentPage < totalPages) {
            currentPage++;
            updateTable();
        }
    });

    // Inicializar
    updateTable();
});