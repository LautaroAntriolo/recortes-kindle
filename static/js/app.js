function configurarSelectorArchivos() {
    const selector = document.getElementById('selector-archivos');
    const paginacion = document.getElementById('paginacion');

    // Configurar visibilidad inicial
    if (selector.value === "") {
        paginacion.style.display = 'none'; // Ocultar paginación si no hay archivo seleccionado
    } else {
        paginacion.style.display = 'block'; // Mostrar paginación si hay archivo seleccionado
    }

    selector.addEventListener('change', async function () {
        const archivoSeleccionado = this.value;

        // Actualizar visibilidad de paginación cada vez que cambia
        if (archivoSeleccionado === "") {
            paginacion.style.display = 'none';
        } else {
            paginacion.style.display = 'block';
        }

        const tbody = document.getElementById('tabla-recortes');
        if (tbody) {
            tbody.innerHTML = ''; // Limpiar tabla existente si existe
        }

        if (!archivoSeleccionado) {
            console.log("No se seleccionó archivo.");
            return;
        }

        try {
            const response = await fetch(`/archivo/${encodeURIComponent(archivoSeleccionado)}`);

            if (!response.ok) {
                throw new Error(`Error ${response.status}: ${response.statusText}`);
            }

            const textoRespuesta = await response.text();
            console.log("Respuesta completa del servidor:", textoRespuesta);

            let jsonLimpio;
            let inicioJSON = Math.min(
                textoRespuesta.indexOf('{') >= 0 ? textoRespuesta.indexOf('{') : Number.MAX_SAFE_INTEGER,
                textoRespuesta.indexOf('[') >= 0 ? textoRespuesta.indexOf('[') : Number.MAX_SAFE_INTEGER
            );

            if (inicioJSON !== Number.MAX_SAFE_INTEGER) {
                jsonLimpio = textoRespuesta.substring(inicioJSON);
            } else {
                // Intentar regex si no encuentra { o [
                const regexJSON = /[\[\{].*[\]\}]/s;
                const coincidencias = textoRespuesta.match(regexJSON);
                if (coincidencias && coincidencias[0]) {
                    jsonLimpio = coincidencias[0];
                } else {
                    throw new Error("No se encontró JSON en la respuesta.");
                }
            }

            const datos = JSON.parse(jsonLimpio);
            mostrarDatosEnTabla(datos);

        } catch (error) {
            console.error('Error al procesar el archivo:', error);
            alert('Error al procesar el archivo: ' + error.message);
        }
    });
}


// Función para mostrar los datos en la tabla
function mostrarDatosEnTabla(datos) {
    const tbody = document.getElementById('tabla-recortes');
    tbody.innerHTML = ''; // Limpiar tabla existente
    
    // Verificar si los datos son válidos
    if (!Array.isArray(datos)) {
        console.error('Datos no válidos:', datos);
        // Si datos es un objeto pero no un array, intentar ver si contiene un array
        if (datos && typeof datos === 'object') {
            // console.log(datos)
            // Buscar la primera propiedad que sea un array
            for (const prop in datos) {
                if (Array.isArray(datos[prop])) {
                    console.log(`Encontrado array en la propiedad '${prop}'`);
                    datos = datos[prop];
                    break;
                }
            }
        }
        
        // Si aún no es un array, intentar envolverlo
        if (!Array.isArray(datos)) {
            console.log('Intentando convertir a array');
            datos = [datos];
        }
    }
    
    // Llenar la tabla con los datos
    datos.forEach(item => {
        const fila = document.createElement('tr');
        
        // Formatear fecha y hora
        const fechaHora = `${item.fecha || ''} ${item.hora || ''}`.trim();
        
        // Crear celdas para cada campo
        fila.innerHTML = `
            <td>${item.autor || 'Sin autor'}</td>
            <td>${item.nombre ? item.nombre.replace(/\uFEFF|\u200B/g, '') : 'Sin nombre'}</td>
            <td class="contenido-celda">
                <div class="contenido-preview">${item.contenido || 'Sin contenido'}</div>
                ${item.pagina ? `<div class="pagina-info">Página ${item.pagina}</div>` : ''}
            </td>
            <td>${fechaHora || 'Sin fecha'}</td>
            <td class="acciones">
                <button class="btn-accion ver" data-id="${item.id || ''}"><a href="#">👁️</a></button>
                <button class="btn-accion etiquetar" data-id="${item.id || ''}"><a href="#">🏷️</a></button>
                <button class="btn-accion editar" data-id="${item.id || ''}"><a href="#">😍</a></button>
                <button class="btn-accion eliminar" data-id="${item.id || ''}"><a href="#">😎</a></button>
                ${item.visibilidad !== undefined ? 
                    (item.visibilidad ? 
                    '<span class="badge visible">Visible</span>' : 
                    '<span class="badge oculto">Oculto</span>') :
                    '<span class="badge">-</span>'}
            </td>
        `;
        
        tbody.appendChild(fila);
    });
    
    // Agregar event listeners a los botones
    document.querySelectorAll('.btn-accion.ver').forEach(btn => {
        btn.addEventListener('click', () => verDetalleCompleto(btn.dataset.id));
    });
    
    document.querySelectorAll('.btn-accion.etiquetar').forEach(btn => {
        btn.addEventListener('click', () => gestionarEtiquetas(btn.dataset.id));
    });
    
    document.querySelectorAll('.btn-accion.editar').forEach(btn => {
        btn.addEventListener('click', () => editarElemento(btn.dataset.id));
    });

    document.querySelectorAll('.btn-accion.eliminar').forEach(btn => {
        btn.addEventListener('click', () => eliminarElemento(btn.dataset.id));
    });
}

// Función para ver el detalle completo
function verDetalleCompleto(id) {
    console.log('Mostrar detalle completo del item:', id);
    // Aquí puedes implementar un modal o expandir la fila
}

// Función para gestionar etiquetas
function gestionarEtiquetas(id) {
    console.log('Gestionar etiquetas del item:', id);
    // Aquí puedes implementar un selector de etiquetas
}

// Función para editar el recorte
function editarElemento(id) {
    console.log('Editar recorte del item:', id);
    // Aquí puedes implementar un selector de etiquetas
}

// Función para eliminar el recorte
function eliminarElemento(id) {
    console.log('Eliminar recorte del item:', id);
    // Aquí puedes implementar un selector de etiquetas
}

// Inicialización cuando el DOM esté listo
document.addEventListener('DOMContentLoaded', function() {
    configurarSelectorArchivos();
    
    // Cargar lista de archivos (tu código existente)
    fetch('/mostrar-archivos')
        .then(response => response.json())
        .then(archivos => {
            const selector = document.getElementById('selector-archivos');
            archivos.forEach(archivo => {
                const option = document.createElement('option');
                option.value = archivo;
                option.textContent = archivo;
                selector.appendChild(option);
            });
        })
        .catch(error => console.error("Error al cargar archivos:", error));
});