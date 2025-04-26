// Variables globales para el estado de la paginación
let currentPage = 1;
let itemsPerPage = 10;
let allData = [];
let filteredData = [];

// Elementos del DOM
const elements = {
  inputPalabra: document.getElementById('input-palabra'),
  searchButton: document.getElementById('search-button'),
  itemsPerPageSelect: document.getElementById('items-per-page'),
  tableBody: document.getElementById('table-body'),
  paginationContainer: document.getElementById('pagination'),
  pageInfo: document.getElementById('page-info')
};

// Event Listeners
function setupEventListeners() {
  elements.searchButton.addEventListener('click', buscarPalabra);
  elements.inputPalabra.addEventListener('keypress', (e) => {
    if (e.key === 'Enter') buscarPalabra();
  });
  elements.itemsPerPageSelect.addEventListener('change', changeItemsPerPage);
}

// Cargar datos iniciales
async function cargarTabla() {
  try {
    const response = await fetch("/data");
    allData = await response.json();
    filteredData = [...allData];
    
    if (allData.length === 0) {
      showNoDataMessage("No se encontraron resultados");
      return;
    }
    
    updatePagination();
    renderTable();
  } catch (error) {
    console.error("Error cargando los datos:", error);
    showNoDataMessage("Error al cargar los datos");
  }
}

// Función para buscar palabras
async function loadData(palabra) {
  try {
    const response = await fetch(
      `/similitudes/${encodeURIComponent(palabra)}`
    );
    if (!response.ok) {
      throw new Error("No se pudieron cargar los datos");
    }

    allData = await response.json();
    filteredData = [...allData];
    currentPage = 1; // Resetear a la primera página
    
    if (filteredData.length === 0) {
      showNoDataMessage(`No se encontraron resultados para "${palabra}"`);
      return;
    }
    
    updatePagination();
    renderTable();
  } catch (error) {
    console.error("Error cargando los datos:", error);
    showNoDataMessage(`Error al buscar "${palabra}"`);
  }
}

// Función de búsqueda
function buscarPalabra() {
  const palabra = elements.inputPalabra.value.trim();
  
  if (palabra !== "") {
    loadData(palabra);
  } else {
    if (allData && allData.length > 0) {
      filteredData = [...allData];
      currentPage = 1;
      updatePagination();
      renderTable();
    } else {
      cargarTabla();
    }
  }
}

// Función para renderizar la tabla
function renderTable() {
  elements.tableBody.innerHTML = '';
  
  const startIndex = (currentPage - 1) * itemsPerPage;
  const endIndex = startIndex + itemsPerPage;
  const pageData = filteredData.slice(startIndex, endIndex);
  
  if (pageData.length === 0) {
    showNoDataMessage("No hay datos para mostrar en esta página");
    return;
  }
  
  pageData.forEach((item) => {
    const row = document.createElement("tr");

    ['autor', 'nombre', 'contenido', 'fecha'].forEach(key => {
      const cell = document.createElement("td");
      cell.textContent = item[key] || "-";
      row.appendChild(cell);
    });

    elements.tableBody.appendChild(row);
  });
  
  updatePageInfo();
}

// Función para actualizar la paginación
function updatePagination() {
  const totalPages = Math.ceil(filteredData.length / itemsPerPage);
  elements.paginationContainer.innerHTML = '';
  
  // Botón Anterior
  const prevButton = document.createElement("button");
  prevButton.innerHTML = "&laquo;";
  prevButton.onclick = () => goToPage(currentPage - 1);
  prevButton.disabled = currentPage === 1;
  elements.paginationContainer.appendChild(prevButton);
  
  // Botones de páginas
  const maxVisiblePages = 5;
  let startPage = Math.max(1, currentPage - Math.floor(maxVisiblePages / 2));
  let endPage = Math.min(totalPages, startPage + maxVisiblePages - 1);
  
  if (endPage - startPage + 1 < maxVisiblePages) {
    startPage = Math.max(1, endPage - maxVisiblePages + 1);
  }
  
  if (startPage > 1) {
    const firstPageButton = document.createElement("button");
    firstPageButton.textContent = "1";
    firstPageButton.onclick = () => goToPage(1);
    elements.paginationContainer.appendChild(firstPageButton);
    
    if (startPage > 2) {
      const ellipsis = document.createElement("span");
      ellipsis.textContent = "...";
      elements.paginationContainer.appendChild(ellipsis);
    }
  }
  
  for (let i = startPage; i <= endPage; i++) {
    const pageButton = document.createElement("button");
    pageButton.textContent = i;
    pageButton.onclick = () => goToPage(i);
    if (i === currentPage) {
      pageButton.classList.add("active");
    }
    elements.paginationContainer.appendChild(pageButton);
  }
  
  if (endPage < totalPages) {
    if (endPage < totalPages - 1) {
      const ellipsis = document.createElement("span");
      ellipsis.textContent = "...";
      elements.paginationContainer.appendChild(ellipsis);
    }
    
    const lastPageButton = document.createElement("button");
    lastPageButton.textContent = totalPages;
    lastPageButton.onclick = () => goToPage(totalPages);
    elements.paginationContainer.appendChild(lastPageButton);
  }
  
  // Botón Siguiente
  const nextButton = document.createElement("button");
  nextButton.innerHTML = "&raquo;";
  nextButton.onclick = () => goToPage(currentPage + 1);
  nextButton.disabled = currentPage === totalPages;
  elements.paginationContainer.appendChild(nextButton);
}

// Función para cambiar de página
function goToPage(page) {
  const totalPages = Math.ceil(filteredData.length / itemsPerPage);
  if (page < 1 || page > totalPages) return;
  
  currentPage = page;
  renderTable();
  updatePagination();
}

// Función para cambiar items por página
function changeItemsPerPage() {
  itemsPerPage = parseInt(elements.itemsPerPageSelect.value);
  currentPage = 1;
  updatePagination();
  renderTable();
}

// Función para actualizar info de página
function updatePageInfo() {
  const startItem = (currentPage - 1) * itemsPerPage + 1;
  const endItem = Math.min(currentPage * itemsPerPage, filteredData.length);
  const totalItems = filteredData.length;
  
  elements.pageInfo.textContent = 
    `Mostrando ${startItem} a ${endItem} de ${totalItems} registros`;
}

// Función para mostrar mensaje sin datos
function showNoDataMessage(message) {
  elements.tableBody.innerHTML = `<tr class="no-data"><td colspan="4">${message}</td></tr>`;
  elements.paginationContainer.innerHTML = '';
  elements.pageInfo.textContent = '';
}

// Inicialización
document.addEventListener('DOMContentLoaded', () => {
  setupEventListeners();
  cargarTabla();
});